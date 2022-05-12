package provider

import (
	"github.com/vatsimnerd/geoidx"
)

type Subscription struct {
	*geoidx.Subscription
	airportFilter geoidx.Filter
	pilotFilter   geoidx.Filter
}

func (s *Subscription) SetPilotFilter(query string) error {
	flt, err := pilotFilter(query)
	if err != nil {
		return err
	}
	s.pilotFilter = flt
	s.resetFilters()
	return nil
}

func (s *Subscription) SetAirportFilter(includeUncontrolled bool) {
	s.airportFilter = airportFilter(includeUncontrolled)
	s.resetFilters()
}

func (s *Subscription) resetFilters() {
	filters := make([]geoidx.Filter, 0)
	if s.airportFilter != nil {
		filters = append(filters, s.airportFilter)
	}
	if s.pilotFilter != nil {
		filters = append(filters, s.pilotFilter)
	}
	s.SetFilters(filters...)
}
