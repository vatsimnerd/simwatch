package memory

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/simwatch/config"
	"github.com/vatsimnerd/simwatch/track"
)

type MemoryReadWriter struct {
	tracks      map[string]*track.Track
	purgePeriod time.Duration
	configured  bool
	stop        chan struct{}
	lock        sync.Mutex
}

var (
	ReadWriter = &MemoryReadWriter{tracks: make(map[string]*track.Track), stop: make(chan struct{})}
	log        = logrus.WithField("module", "track.memory")
)

func (m *MemoryReadWriter) LoadTrackByID(ctx context.Context, id string) (*track.Track, error) {
	if !m.configured {
		return nil, track.ErrNotConfigured
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	if t, found := m.tracks[id]; found {
		return t, nil
	}
	return nil, track.ErrNotFound
}

func (m *MemoryReadWriter) WriteTrack(ctx context.Context, p *merged.Pilot) error {
	l := log.WithFields(logrus.Fields{
		"func":     "WriteTrack",
		"callsign": p.Callsign,
	})

	if !m.configured {
		return track.ErrNotConfigured
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

func (m *MemoryReadWriter) ListIDs(ctx context.Context) ([]string, error) {
	ids := make([]string, len(m.tracks))
	i := 0
	for key := range m.tracks {
		ids[i] = key
		i++
	}
	return ids, nil
}

func (m *MemoryReadWriter) Configure(cfg *config.TrackConfigOptions) error {
	m.configured = true
	m.purgePeriod = cfg.PurgePeriod
	go m.gc()
	return nil
}

func (m *MemoryReadWriter) gc() {
	t := time.NewTicker(m.purgePeriod)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			m.lock.Lock()
			tm := time.Now()

			for key, track := range m.tracks {
				if tm.Sub(track.CreatedAt) > m.purgePeriod {
					delete(m.tracks, key)
					log.WithField("track_id", key).Info("expired track removed")
				}
			}
			m.lock.Unlock()
		case <-m.stop:
			return
		}
	}
}

func (r *MemoryReadWriter) Close() error {
	r.stop <- struct{}{}
	close(r.stop)
	return nil
}
