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
		ID      string      `json:"id"`
		Type    MessageType `json:"type"`
		Payload interface{} `json:"payload"`
	}

	RequestBounds = geoidx.Rect
)

const (
	RequestTypeBounds         RequestType = "bounds"
	RequestTypeAirportsFilter RequestType = "airports_filter"
	RequestTypePilotsFilter   RequestType = "pilots_filter"
	RequestTypeSubscribeID    RequestType = "sub_id"
	RequestTypeUnsubscribeID  RequestType = "unsub_id"

	MessageTypeUpdate MessageType = "update"
	MessageTypeStatus MessageType = "status"
	MessageTypeError  MessageType = "error"
)
