package simwatch

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vatsimnerd/geoidx"
	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/simwatch/provider"
)

const (
	maxObjectsPerUpdate = 1500
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func (s *Server) handleApiUpdates(w http.ResponseWriter, r *http.Request) {
	l := log.WithField("func", "handleApiUpdates")

	sock, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.WithError(err).Error("error upgrading connection")
		w.WriteHeader(500)
		return
	}

	sub := s.provider.Subscribe(1024)
	defer s.provider.Unsubscribe(sub)

	mc := make(chan *Message, 1024)
	// every client request may result in a huge amount of
	// events sent via subscription channel so we must make sure
	// that those are processed independently in a separate thread
	//
	// also websocket doesn't allow concurrent writing so this
	// goroutine must be the only one writing to a ws connection
	go sendMessages(sock, sub, mc)
	defer close(mc)

	for {
		_, buf, err := sock.ReadMessage()
		l.WithField("buf", string(buf)).WithError(err).Trace("message from client")
		if err != nil {
			l.WithError(err).Error("error reading message")
			break
		}

		req := &Request{}
		err = json.Unmarshal(buf, req)

		if err != nil {
			l.WithError(err).Error("error parsing request")
			continue
		}

		switch req.Type {
		case RequestTypeBounds:
			err = json.Unmarshal(req.Payload, &req.Bounds)
		case RequestTypeAirportsFilter:
			err = json.Unmarshal(req.Payload, &req.AirportFilter)
		case RequestTypePilotsFilter:
			err = json.Unmarshal(req.Payload, &req.PilotFilter)
		case RequestTypeSubscribeID:
			fallthrough
		case RequestTypeUnsubscribeID:
			err = json.Unmarshal(req.Payload, &req.SubID)
		}

		if err != nil {
			l.WithError(err).WithField("req_type", req.Type).Error("error parsing request payload")
			sendErrorMessage(mc, req.ID, err)
			continue
		}

		log.WithField("req", req).Debug("request received")

		switch req.Type {
		case RequestTypeBounds:
			bounds := req.Bounds
			sub.SetBounds(bounds)
			sendStatusMessage(mc, req.ID, "bounds set")
		case RequestTypeAirportsFilter:
			sub.SetAirportFilter(req.AirportFilter.IncludeUncontrolled)
			sendStatusMessage(mc, req.ID, "airport filter set")
		case RequestTypePilotsFilter:
			sub.SetPilotFilter(req.PilotFilter.Query)
			sendStatusMessage(mc, req.ID, "pilot filter set")
		}
	}
}

func sendMessages(sock *websocket.Conn, sub *provider.Subscription, mc <-chan *Message) {
	var oType string
	var eType string
	var acc *ObjectUpdate

	flush := time.NewTicker(50 * time.Millisecond)
	defer flush.Stop()

	for {
		select {
		case event, ok := <-sub.Events():
			if !ok {
				return
			}

			switch event.Type {
			case geoidx.EventTypeSet:
				eType = "set"
			case geoidx.EventTypeDelete:
				eType = "del"
			}

			switch event.Obj.Value().(type) {
			case *merged.Airport:
				oType = "arpt"
			case *merged.Radar:
				oType = "rdr"
			case *merged.Pilot:
				oType = "plt"
			}

			// if acc is not created yet, create a new one
			if acc == nil {
				acc = makeObjectUpdate(eType, oType, maxObjectsPerUpdate)
			}

			// if acc contains updates of different type, flush it
			// and create a new one
			if acc.EType != eType || acc.OType != oType {
				if acc.hasData() {
					sock.WriteJSON(acc.message())
				}
				acc = makeObjectUpdate(eType, oType, maxObjectsPerUpdate)
			}

			if acc.add(event.Obj.Value()) {
				// if acc is full, send its contents and reset
				sock.WriteJSON(acc.message())
				acc.reset()
			}

		case <-flush.C:
			// periodically flush update buffers
			if acc != nil && acc.hasData() {
				// if acc has data, send its contents and reset
				sock.WriteJSON(acc.message())
				acc.reset()
			}

		case msg := <-mc:
			sock.WriteJSON(msg)
		}
	}
}

func sendErrorMessage(mc chan *Message, reqID string, err error) {
	msg := &Message{
		Type: MessageTypeError,
		Payload: struct {
			Error     string `json:"error"`
			RequestID string `json:"req_id"`
		}{
			Error:     err.Error(),
			RequestID: reqID,
		},
	}
	mc <- msg
}

func sendStatusMessage(mc chan *Message, reqID string, status string) {
	msg := &Message{
		Type: MessageTypeStatus,
		Payload: struct {
			Status    string `json:"status"`
			RequestID string `json:"req_id"`
		}{
			Status:    status,
			RequestID: reqID,
		},
	}
	mc <- msg
}
