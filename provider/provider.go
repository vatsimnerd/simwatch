package provider

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/geoidx"
	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/simwatch/config"
	"github.com/vatsimnerd/simwatch/track"
	"github.com/vatsimnerd/simwatch/track/memory"
	"github.com/vatsimnerd/util/pubsub"
	"github.com/vatsimnerd/util/set"
)

var (
	log = logrus.WithField("module", "simwatch.provider")

	ErrNotFound = fmt.Errorf("object not found")
)

type Provider struct {
	vatsim *merged.Provider
	stop   chan bool
	idx    *geoidx.Index

	airports map[string]*merged.Airport
	pilots   map[string]*merged.Pilot
	radars   map[string]*merged.Radar

	airportTrace *set.SafeSet[string]

	dataLock sync.RWMutex
}

func New(cfg *config.Config) *Provider {
	return &Provider{
		vatsim: merged.New(&cfg.API, &cfg.Data, &cfg.Runways),
		stop:   make(chan bool),
		idx:    geoidx.NewIndex(),

		airports: make(map[string]*merged.Airport),
		pilots:   make(map[string]*merged.Pilot),
		radars:   make(map[string]*merged.Radar),

		airportTrace: set.NewSafe[string](),
	}
}

func (p *Provider) Start() error {
	track.Configure(&memory.Config{PurgePeriod: 24 * time.Hour})
	err := p.vatsim.Start()
	if err != nil {
		return err
	}
	go p.loop()
	return nil
}

func (p *Provider) Stop() {
	p.stop <- true
}

func (p *Provider) loop() {
	l := log.WithField("func", "loop")
	var err error

	defer p.vatsim.Stop()

	s := p.vatsim.Subscribe(32768)
	defer p.vatsim.Unsubscribe(s)

	count := 0

	for {
		select {
		case upd := <-s.Updates():
			count++
			if count%1000 == 0 {
				l.Debugf("accumulated %d updates from merged provider", count)
			}
			switch upd.UType {
			case pubsub.UpdateTypeSet:
				switch upd.OType {
				case merged.ObjectTypeAirport:
					err = p.setAirport(upd.Obj)
				case merged.ObjectTypePilot:
					err = p.setPilot(upd.Obj)
				case merged.ObjectTypeRadar:
					err = p.setRadar(upd.Obj)
				}
			case pubsub.UpdateTypeDelete:
				switch upd.OType {
				case merged.ObjectTypeAirport:
					err = p.deleteAirport(upd.Obj)
				case merged.ObjectTypePilot:
					err = p.deletePilot(upd.Obj)
				case merged.ObjectTypeRadar:
					err = p.deleteRadar(upd.Obj)
				}
			}

			if err != nil {
				l.WithField("upd", upd).Error("error deleting object")
				err = nil
			}

		case <-p.stop:
			return
		}
	}
}

func (p *Provider) setAirport(obj interface{}) error {
	l := log.WithFields(logrus.Fields{
		"func": "setAirport",
		"obj":  obj,
	})

	arpt, ok := obj.(merged.Airport)
	if !ok {
		return fmt.Errorf("unexpected type %T, expected to be Airport", obj)
	}

	trace := p.airportTrace.Has(arpt.Meta.ICAO)

	iobj := geoidx.NewObject(
		arpt.Meta.ICAO,
		squareCentered(arpt.Meta.Position.Lat, arpt.Meta.Position.Lng, airportSizeNM),
		&arpt,
	)

	if trace {
		l.Info("upserting airport geo object")
	} else {
		l.Trace("upserting airport geo object")
	}
	p.idx.Upsert(iobj)

	p.dataLock.Lock()
	if trace {
		l.Info("inserting airport to index")
	} else {
		l.Trace("inserting airport to index")
	}
	p.airports[arpt.Meta.ICAO] = &arpt
	p.dataLock.Unlock()

	return nil
}

func (p *Provider) deleteAirport(obj interface{}) error {
	l := log.WithFields(logrus.Fields{
		"func": "deleteAirport",
		"obj":  obj,
	})

	arpt, ok := obj.(merged.Airport)
	if !ok {
		return fmt.Errorf("unexpected type %T, expected to be Airport", obj)
	}

	l.WithField("ICAO", arpt.Meta.ICAO).Warn("DELETE AIRPORT")

	trace := p.airportTrace.Has(arpt.Meta.ICAO)

	iobj := geoidx.NewObject(
		arpt.Meta.ICAO,
		squareCentered(arpt.Meta.Position.Lat, arpt.Meta.Position.Lng, airportSizeNM),
		&arpt,
	)
	if trace {
		l.Info("deleting airport geo object")
	} else {
		l.Trace("deleting airport geo object")
	}
	p.idx.Delete(iobj)

	if trace {
		l.Info("deleting airport from index")
	} else {
		l.Trace("deleting airport from index")
	}
	p.dataLock.Lock()
	delete(p.airports, arpt.Meta.ICAO)
	p.dataLock.Unlock()

	return nil
}

func (p *Provider) setPilot(obj interface{}) error {
	l := log.WithFields(logrus.Fields{
		"func": "setPilot",
		"obj":  obj,
	})

	pilot, ok := obj.(merged.Pilot)
	if !ok {
		return fmt.Errorf("unexpected type %T, expected to be Pilot", obj)
	}

	iobj := geoidx.NewObject(
		pilot.Callsign,
		squareCentered(pilot.Latitude, pilot.Longitude, planeSizeNM),
		&pilot,
	)
	l.Trace("upserting pilot geo object")
	p.idx.Upsert(iobj)

	l.Trace("inserting pilot to index")
	p.dataLock.Lock()
	p.pilots[pilot.Callsign] = &pilot
	p.dataLock.Unlock()

	l.Trace("writing pilot's track")

	err := track.WriteTrack(&pilot)
	return err
}

func (p *Provider) deletePilot(obj interface{}) error {
	l := log.WithFields(logrus.Fields{
		"func": "deletePilot",
		"obj":  obj,
	})

	pilot, ok := obj.(merged.Pilot)
	if !ok {
		return fmt.Errorf("unexpected type %T, expected to be Pilot", obj)
	}

	l.WithField("CALLSIGN", pilot.Callsign).Warn("DELETE PILOT")

	iobj := geoidx.NewObject(
		pilot.Callsign,
		squareCentered(pilot.Latitude, pilot.Longitude, planeSizeNM),
		&pilot,
	)
	l.Trace("deleting pilot geo object")
	p.idx.Delete(iobj)

	l.Trace("deleting pilot from index")
	p.dataLock.Lock()
	delete(p.pilots, pilot.Callsign)
	p.dataLock.Unlock()

	return nil
}

func (p *Provider) setRadar(obj interface{}) error {
	l := log.WithFields(logrus.Fields{
		"func": "setRadar",
		"obj":  obj,
	})

	radar, ok := obj.(merged.Radar)
	if !ok {
		return fmt.Errorf("unexpected type %T, expected to be Radar", obj)
	}
	l = l.WithField("callsign", radar.Controller.Callsign)

	minLng := 1000.0
	minLat := 1000.0
	maxLng := -1000.0
	maxLat := -1000.0
	for _, fir := range radar.FIRs {
		if fir.Boundaries.Min.Lat < minLat {
			minLat = fir.Boundaries.Min.Lat
		}
		if fir.Boundaries.Min.Lng < minLng {
			minLng = fir.Boundaries.Min.Lng
		}
		if fir.Boundaries.Max.Lat > maxLat {
			maxLat = fir.Boundaries.Max.Lat
		}
		if fir.Boundaries.Max.Lng > maxLng {
			maxLng = fir.Boundaries.Max.Lng
		}
	}
	rect := geoidx.MakeRect(minLng, minLat, maxLng, maxLat)
	l = l.WithField("rect", rect)

	iobj := geoidx.NewObject(
		radar.Controller.Callsign,
		rect,
		&radar,
	)
	l.Trace("upserting radar geo object")
	p.idx.Upsert(iobj)

	l.Trace("inserting radar to index")
	p.dataLock.Lock()
	p.radars[radar.Controller.Callsign] = &radar
	p.dataLock.Unlock()

	return nil
}

func (p *Provider) deleteRadar(obj interface{}) error {
	l := log.WithFields(logrus.Fields{
		"func": "deleteRadar",
		"obj":  obj,
	})

	radar, ok := obj.(merged.Radar)
	if !ok {
		return fmt.Errorf("unexpected type %T, expected to be Radar", obj)
	}

	l.WithField("CALLSIGN", radar.Controller.Callsign).Warn("DELETE RADAR")

	rect := geoidx.MakeRect(0, 0, 0, 0)
	iobj := geoidx.NewObject(
		radar.Controller.Callsign,
		rect,
		&radar,
	)
	l.Trace("deleting radar geo object")
	p.idx.Delete(iobj)

	l.Trace("deleting radar from index")
	p.dataLock.Lock()
	delete(p.radars, radar.Controller.Callsign)
	p.dataLock.Unlock()

	return nil
}

func (p *Provider) Subscribe(chSize int) *Subscription {
	return &Subscription{
		Subscription:  p.idx.Subscribe(chSize),
		airportFilter: nil,
		pilotFilter:   nil,
	}
}

func (p *Provider) Unsubscribe(sub *Subscription) {
	p.idx.Unsubscribe(sub.Subscription)
}

func (p *Provider) GetPilots() []*merged.Pilot {
	p.dataLock.RLock()
	pilots := make([]*merged.Pilot, len(p.pilots))
	c := 0
	for _, pilot := range p.pilots {
		pilots[c] = pilot
		c++
	}
	p.dataLock.RUnlock()

	sort.Slice(pilots, func(i, j int) bool {
		return pilots[i].Callsign < pilots[j].Callsign
	})
	return pilots
}

func (p *Provider) GetPilotByCallsign(callsign string) (*merged.Pilot, error) {
	p.dataLock.RLock()
	defer p.dataLock.RUnlock()
	if pilot, found := p.pilots[callsign]; found {
		return pilot, nil
	}
	return nil, ErrNotFound
}

func (p *Provider) GetAirports() []*merged.Airport {
	p.dataLock.RLock()
	airports := make([]*merged.Airport, len(p.airports))
	c := 0
	for _, arpt := range p.airports {
		airports[c] = arpt
		c++
	}
	p.dataLock.RUnlock()

	sort.Slice(airports, func(i, j int) bool {
		return airports[i].Meta.ICAO < airports[j].Meta.ICAO
	})
	return airports
}

func (p *Provider) GetAirportByICAO(icao string) (*merged.Airport, error) {
	p.dataLock.RLock()
	defer p.dataLock.RUnlock()
	if arpt, found := p.airports[icao]; found {
		return arpt, nil
	}
	return nil, ErrNotFound
}

func (p *Provider) SetAirportTrace(icao string) {
	p.vatsim.SetAirportTrace(icao)
	p.airportTrace.Add(icao)
}

func (p *Provider) ResetAirportTrace(icao string) {
	p.vatsim.ResetAirportTrace(icao)
	p.airportTrace.Delete(icao)
}
