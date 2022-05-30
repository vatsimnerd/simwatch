package simwatch

import (
	"fmt"
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
	sendPaginated(w, r, pilots)
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
		sendError(w, 404, "pilot not found")
		return
	}

	apiPilot := ApiPilot{Pilot: pilot}
	tr, err := track.LoadTrack(r.Context(), pilot)
	if err != nil {
		l.WithError(err).Error("error loading track")
		sendError(w, 500, fmt.Sprintf("error loading track: %v", err))
		return
	}

	apiPilot.Track = tr.Points
	sendJSON(w, apiPilot)
}
