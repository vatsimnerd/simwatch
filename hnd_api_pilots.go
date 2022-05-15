package simwatch

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/simwatch/track"
)

type ApiPilot struct {
	*merged.Pilot
	Track []track.TrackPoint `json:"track"`
}

func (s *Server) handleApiPilots(w http.ResponseWriter, r *http.Request) {
	pilots := s.provider.GetPilots()
	data, err := json.Marshal(pilots)
	if err != nil {
		w.WriteHeader(500)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) handleApiPilotsGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	callsign := vars["id"]

	l := log.WithFields(logrus.Fields{
		"func":     "handleApiPilotsGet",
		"callsign": callsign,
	})

	pilot, err := s.provider.GetPilotByCallsign(callsign)

	if err != nil {
		l.WithError(err).Error("error searching for pilot")
		w.WriteHeader(404)
		return
	}

	apiPilot := ApiPilot{Pilot: pilot}
	tr, err := track.LoadTrack(pilot)
	if err != nil {
		l.WithError(err).Error("error loading track")
		w.WriteHeader(500)
		return
	}

	apiPilot.Track = tr.Points
	data, err := json.Marshal(apiPilot)
	if err != nil {
		l.WithError(err).Error("error marshaling data")
		w.WriteHeader(500)
		return
	}
	w.Write(data)

}
