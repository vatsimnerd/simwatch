package provider

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/geoidx"
	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/util/pubsub"
)

var (
	log = logrus.WithField("module", "simwatch.provider")
)

type Provider struct {
	vatsim *merged.Provider
	stop   chan bool
	idx    *geoidx.Index

	airports map[string]merged.Airport
	pilots   map[string]merged.Pilot
	radars   map[string]merged.Radar

	dataLock sync.RWMutex
}

func New() *Provider {
	return &Provider{
		vatsim: merged.New(),
		stop:   make(chan bool),
		idx:    geoidx.NewIndex(),

		airports: make(map[string]merged.Airport),
		pilots:   make(map[string]merged.Pilot),
		radars:   make(map[string]merged.Radar),
	}
}

func (p *Provider) Start() error {
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

	s := p.vatsim.Subscribe(1024)
	defer p.vatsim.Unsubscribe(s)

	for {
		select {
		case upd := <-s.Updates():
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
				l.WithField("upd", upd).Error("error updating object")
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

	iobj := geoidx.NewObject(
		arpt.Meta.ICAO,
		squareCentered(arpt.Meta.Position.Lat, arpt.Meta.Position.Lng, airportSizeNM),
		arpt,
	)
	l.Debug("upserting airport geo object")
	p.idx.Upsert(iobj)

	p.dataLock.Lock()
	l.Debug("inserting airport to index")
	p.airports[arpt.Meta.ICAO] = arpt
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

	iobj := geoidx.NewObject(
		arpt.Meta.ICAO,
		squareCentered(arpt.Meta.Position.Lat, arpt.Meta.Position.Lng, airportSizeNM),
		arpt,
	)
	l.Debug("deleting airport geo object")
	p.idx.Delete(iobj)

	l.Debug("deleting airport from index")
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
		pilot,
	)
	l.Debug("upserting pilot geo object")
	p.idx.Upsert(iobj)

	l.Debug("inserting pilot to index")
	p.dataLock.Lock()
	p.pilots[pilot.Callsign] = pilot
	p.dataLock.Unlock()

	return nil
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

	iobj := geoidx.NewObject(
		pilot.Callsign,
		squareCentered(pilot.Latitude, pilot.Longitude, planeSizeNM),
		pilot,
	)
	l.Debug("deleting pilot geo object")
	p.idx.Delete(iobj)

	l.Debug("deleting pilot from index")
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
		if fir.Boundaries.Max.Lng < maxLng {
			maxLng = fir.Boundaries.Max.Lng
		}
	}
	rect := geoidx.MakeRect(minLng, minLat, maxLng, maxLat)

	iobj := geoidx.NewObject(
		radar.Controller.Callsign,
		rect,
		radar,
	)
	l.Debug("upserting radar geo object")
	p.idx.Upsert(iobj)

	l.Debug("inserting radar to index")
	p.dataLock.Lock()
	p.radars[radar.Controller.Callsign] = radar
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
	rect := geoidx.MakeRect(0, 0, 0, 0)
	iobj := geoidx.NewObject(
		radar.Controller.Callsign,
		rect,
		radar,
	)
	l.Debug("deleting radar geo object")
	p.idx.Delete(iobj)

	l.Debug("deleting radar from index")
	p.dataLock.Lock()
	delete(p.radars, radar.Controller.Callsign)
	p.dataLock.Unlock()

	return nil
}

func (p *Provider) Subscribe(chSize int) *Subscription {
	geosub := p.idx.Subscribe(chSize)
	return &Subscription{
		id:     geosub.ID(),
		geosub: geosub,
	}
}
