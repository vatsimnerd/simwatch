package simwatch

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/simwatch/provider"
)

func (s *Server) handleApiAirports(w http.ResponseWriter, r *http.Request) {
	airports := s.provider.GetAirports()
	sendPaginated(w, r, airports)
}

func (s *Server) handleApiAirportsGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	icao := vars["id"]

	l := log.WithFields(logrus.Fields{
		"func": "handleApiAirportsGet",
		"icao": icao,
	})

	arpt, err := s.provider.GetAirportByICAO(icao)
	if err != nil {
		if err == provider.ErrNotFound {
			l.Error("airport not found")
			sendError(w, 404, "airport not found")
		} else {
			l.WithError(err).Error("error searching for airport")
			sendError(w, 500, err.Error())
		}
		return
	}

	sendJSON(w, arpt)
}
