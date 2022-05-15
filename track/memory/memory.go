package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/simwatch/track"
)

type MemoryReadWriter struct {
	tracks      map[string]*track.Track
	purgePeriod time.Duration
	configured  bool
	lock        sync.Mutex
}

var (
	ReadWriter = &MemoryReadWriter{tracks: make(map[string]*track.Track)}
	log        = logrus.WithField("module", "track.memory")

	ErrNotFound      = fmt.Errorf("track not found")
	ErrNotConfigured = fmt.Errorf("MemoryReadWriter not configured")
	ErrConfigInvalid = fmt.Errorf("invalid configuration for MemoryReadWriter, must be *memory.Config")
)

func init() {
	log.Info("setup memory track writer")
	track.RegisterTrackReadWriter(ReadWriter)
}

func (m *MemoryReadWriter) LoadTrackByID(id string) (*track.Track, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if t, found := m.tracks[id]; found {
		return t, nil
	}
	return nil, ErrNotFound
}

func (m *MemoryReadWriter) LoadTrack(p *merged.Pilot) (*track.Track, error) {
	if !m.configured {
		return nil, ErrNotConfigured
	}

	trackID, _ := track.ExtractTrackData(p)
	return m.LoadTrackByID(trackID)
}

func (m *MemoryReadWriter) WriteTrack(p *merged.Pilot) error {
	l := log.WithFields(logrus.Fields{
		"func":     "WriteTrack",
		"callsign": p.Callsign,
	})

	if !m.configured {
		return ErrNotConfigured
	}

	trackID, point := track.ExtractTrackData(p)
	l.WithFields(logrus.Fields{"track_id": trackID, "point": point}).Trace("extracted")

	m.lock.Lock()
	defer m.lock.Unlock()

	t, found := m.tracks[trackID]
	if !found {
		t = &track.Track{
			CreatedAt: time.Now(),
			Points:    make([]track.TrackPoint, 0),
		}
		m.tracks[trackID] = t
	}

	if len(t.Points) < 2 || (t.Points[len(t.Points)-1].NE(point)) {
		// the new point is different or there's not enough
		// room for compression
		t.Points = append(t.Points, point)
	} else {
		// the new point is identical to previous one, updating
		// timestamp for compression purposes
		t.Points[len(t.Points)-1].TimeStamp = point.TimeStamp
	}

	return nil
}

func (m *MemoryReadWriter) ListIDs() []string {
	ids := make([]string, len(m.tracks))
	i := 0
	for key := range m.tracks {
		ids[i] = key
		i++
	}
	return ids
}

func (m *MemoryReadWriter) Configure(cfg interface{}) error {
	if config, ok := cfg.(*Config); ok {
		m.configured = true
		m.purgePeriod = config.PurgePeriod
		go m.gc()
		return nil
	}
	return ErrConfigInvalid
}

func (m *MemoryReadWriter) gc() {
	t := time.NewTicker(m.purgePeriod)
	defer t.Stop()

	for range t.C {
		m.lock.Lock()
		tm := time.Now()

		for key, track := range m.tracks {
			if tm.Sub(track.CreatedAt) > m.purgePeriod {
				delete(m.tracks, key)
				log.WithField("track_id", key).Info("expired track removed")
			}
		}
		m.lock.Unlock()
	}
}
