package simwatch

import (
	"encoding/json"

	"github.com/vatsimnerd/geoidx"
)

type (
	RequestType string
	MessageType string

	Request struct {
		ID            string               `json:"id"`
		Type          RequestType          `json:"type"`
		Payload       json.RawMessage      `json:"payload"`
		AirportFilter RequestAirportFilter `json:"airport_filter"`
		PilotFilter   RequestPilotFilter   `json:"pilot_filter"`
		Bounds        RequestBounds        `json:"bounds"`
		SubID         RequestSubID         `json:"sub_id"`
	}

	RequestAirportFilter struct {
		IncludeUncontrolled bool `json:"include_uncontrolled"`
	}

	RequestPilotFilter struct {
		Query string `json:"query"`
	}

	RequestSubID struct {
		ID string `json:"id"`
	}

	Message struct {
		Type    MessageType `json:"type"`
		Payload interface{} `json:"payload"`
	}

	ObjectUpdate struct {
		EType     string        `json:"e_type"`
		OType     string        `json:"o_type"`
		Objects   []interface{} `json:"objects"`
		maxBucket int
	}

	RequestBounds = geoidx.Rect
)

func (o *ObjectUpdate) reset() {
	o.Objects = make([]interface{}, 0, o.maxBucket)
}

func (o *ObjectUpdate) hasData() bool {
	return len(o.Objects) > 0
}

func (o *ObjectUpdate) message() *Message {
	return &Message{
		Type:    MessageTypeUpdate,
		Payload: o,
	}
}

func (o *ObjectUpdate) add(obj interface{}) bool {
	ol := o.Objects
	ol = append(ol, obj)
	o.Objects = ol
	return len(ol) == o.maxBucket
}

func makeObjectUpdate(etype, otype string, maxBucket int) *ObjectUpdate {
	o := &ObjectUpdate{EType: etype, OType: otype, maxBucket: maxBucket}
	o.reset()
	return o
}

const (
	RequestTypeBounds         RequestType = "bounds"
	RequestTypeAirportsFilter RequestType = "airport_filter"
	RequestTypePilotsFilter   RequestType = "pilot_filter"
	RequestTypeSubscribeID    RequestType = "sub_id"
	RequestTypeUnsubscribeID  RequestType = "unsub_id"

	MessageTypeUpdate MessageType = "update"
	MessageTypeStatus MessageType = "status"
	MessageTypeError  MessageType = "error"
)
