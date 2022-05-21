package simwatch

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type middlewareApplyCORS struct {
	wrapped http.Handler
}

func (m *middlewareApplyCORS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(logrus.Fields{
		"func": "middlewareApplyCORS.ServeHTTP",
		"req":  r,
	})
	origin := r.Header.Get("Origin")
	if origin != "" {
		l.WithField("origin", origin).Debug("setting Access-Control-Allow-Origin")
		w.Header().Add("Access-Control-Allow-Origin", origin)
	} else {
		l.Debug("no Origin header, continue")
	}
	m.wrapped.ServeHTTP(w, r)
}

func applyCors(handler http.Handler) http.Handler {
	return &middlewareApplyCORS{wrapped: handler}
}
