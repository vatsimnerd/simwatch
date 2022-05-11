package simwatch

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type (
	requestBundle struct {
		request *Request
		err     error
	}
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	log = logrus.WithField("module", "simwatch")
)

func (s *Server) handleApiUpdates(w http.ResponseWriter, r *http.Request) {
	l := log.WithField("func", "handleApiUpdates")

	sock, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.WithError(err).Error("error upgrading connection")
		w.WriteHeader(500)
	}
	reqChan := readRequests(sock)
	sub := s.provider.Subscribe(1024)

	for {
		select {
		case req := <-reqChan:
			if req.err != nil {
				continue
			}

			switch req.request.Type {
			case RequestTypeBounds:
				bounds := req.request.Bounds
				sub.SetBounds(bounds)
			case RequestTypeAirportsFilter:
				sub.SetAirportFilter(req.request.AirportFilter.IncludeUncontrolled)
			}
		case evt := <-sub.Events():
			fmt.Println(evt)
		}

	}
}

func readRequests(sock *websocket.Conn) <-chan *requestBundle {
	ch := make(chan *requestBundle, 1024)

	go func() {
		defer close(ch)

		for {
			_, buf, err := sock.ReadMessage()
			if err != nil {
				log.WithError(err).Error("error reading message")
				ch <- &requestBundle{
					request: nil,
					err:     err,
				}
				break
			}

			req := &Request{}
			err = json.Unmarshal(buf, req)

			if err != nil {
				log.WithError(err).Error("error parsing spy request")
				ch <- &requestBundle{
					request: req,
					err:     err,
				}
				continue
			}

			switch req.Type {
			case RequestTypeBounds:
				err = json.Unmarshal(req.Payload, &req.Bounds)
			case RequestTypeAirportsFilter:
				err = json.Unmarshal(req.Payload, &req.AirportFilter)
			case RequestTypePilotsFilter:
				err = json.Unmarshal(req.Payload, &req.AirportFilter)
			case RequestTypeSubscribeID:
				fallthrough
			case RequestTypeUnsubscribeID:
				err = json.Unmarshal(req.Payload, &req.SubID)
			}

			if err != nil {
				log.WithError(err).WithField("req_type", req.Type).Error("error parsing request payload")
			}

			ch <- &requestBundle{
				request: req,
				err:     err,
			}
		}
	}()

	return ch
}
