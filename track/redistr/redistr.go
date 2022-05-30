package redistr

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/simwatch/config"
	"github.com/vatsimnerd/simwatch/track"
)

type RedisReadWriter struct {
	cfg        *config.TrackConfigOptions
	cli        *redis.Client
	configured bool
}

var (
	ReadWriter = &RedisReadWriter{}
	log        = logrus.WithField("module", "track.redistr")

	errPointNotFound = errors.New("point not found")
)

func (r *RedisReadWriter) setupClient() {
	r.cli = redis.NewClient(&redis.Options{
		Addr:     r.cfg.Addr,
		Password: r.cfg.Password,
		DB:       r.cfg.DB,
	})
}

func (r *RedisReadWriter) LoadTrackByID(ctx context.Context, trackID string) (*track.Track, error) {
	if !r.trackExists(ctx, trackID) {
		return nil, track.ErrNotFound
	}

	ck := trackCreatedKey(trackID)
	createdUx, err := r.getInt64(ctx, ck)
	if err != nil {
		return nil, err
	}

	tridcs, err := r.cli.LRange(ctx, pointsIndexKey(trackID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	tr := &track.Track{
		CreatedAt: time.Unix(createdUx, 0),
		Points:    make([]track.TrackPoint, 0, len(tridcs)),
	}

	for _, pix := range tridcs {
		p, err := r.getPoint(ctx, trackID, pix)
		if err != nil {
			continue
		}
		tr.Points = append(tr.Points, *p)
	}

	return tr, nil
}

func (r *RedisReadWriter) LoadTrack(ctx context.Context, p *merged.Pilot) (*track.Track, error) {
	trackID, _ := track.ExtractTrackData(p)
	return r.LoadTrackByID(ctx, trackID)
}

func (r *RedisReadWriter) WriteTrack(ctx context.Context, p *merged.Pilot) error {
	l := log.WithFields(logrus.Fields{
		"func":     "WriteTrack",
		"callsign": p.Callsign,
	})

	if !r.configured {
		return track.ErrNotConfigured
	}

	trackID, point := track.ExtractTrackData(p)
	l.WithFields(logrus.Fields{"track_id": trackID, "point": point}).Trace("extracted")

	if !r.trackExists(ctx, trackID) {
		err := r.createTrack(ctx, trackID)
		if err != nil {
			return fmt.Errorf("error creating track: %w", err)
		}
	}

	err := r.writeTrackPoint(ctx, trackID, point)
	if err != nil {
		return fmt.Errorf("error writing track point: %w", err)
	}

	return nil
}

func (r *RedisReadWriter) Configure(cfg *config.TrackConfigOptions) error {
	r.cfg = cfg
	r.configured = true
	r.setupClient()
	go r.gc()
	return nil
}

func (r *RedisReadWriter) ListIDs(ctx context.Context) []string {
	res, err := r.cli.SMembers(ctx, keyTrackIDs).Result()
	if err != nil {
		return []string{}
	}
	return res
}

func (r *RedisReadWriter) trackExists(ctx context.Context, trackID string) bool {
	val, err := r.cli.SIsMember(ctx, keyTrackIDs, trackID).Result()
	if err != nil {
		return false
	}
	return val
}

func (r *RedisReadWriter) createTrack(ctx context.Context, trackID string) error {
	err := r.cli.SAdd(ctx, keyTrackIDs, trackID).Err()
	if err != nil {
		return err
	}

	key := trackCreatedKey(trackID)
	return r.cli.Set(ctx, key, time.Now().Unix(), 0).Err()
}

func (r *RedisReadWriter) getPoint(ctx context.Context, trackID string, ts string) (*track.TrackPoint, error) {
	key := pointKey(trackID, ts)

	ex, err := r.cli.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if ex == 0 {
		return nil, errPointNotFound
	}

	alt, err := r.getHInt(ctx, key, "alt")
	if err != nil {
		return nil, err
	}
	gs, err := r.getHInt(ctx, key, "gs")
	if err != nil {
		return nil, err
	}
	hdg, err := r.getHInt(ctx, key, "hdg")
	if err != nil {
		return nil, err
	}
	lat, err := r.getHFloat(ctx, key, "lat")
	if err != nil {
		return nil, err
	}
	lng, err := r.getHFloat(ctx, key, "lng")
	if err != nil {
		return nil, err
	}
	its, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return nil, err
	}
	return &track.TrackPoint{
		Altitude:    alt,
		Groundspeed: gs,
		Heading:     hdg,
		Latitude:    lat,
		Longitude:   lng,
		TimeStamp:   its,
	}, nil

}

func (r *RedisReadWriter) writeTrackPoint(ctx context.Context, trackID string, point track.TrackPoint) error {
	l := log.WithFields(logrus.Fields{
		"func": "writeTrackPoint",
		"tid":  trackID,
		"pt":   point,
	})

	trackIdxKey := pointsIndexKey(trackID)
	trackSize, _ := r.cli.LLen(ctx, trackIdxKey).Result()

	if trackSize > 0 {
		// check if the new point is older than the last one
		last, err := r.cli.LIndex(ctx, trackIdxKey, trackSize-1).Result()
		if err != nil {
			return err
		}
		l = l.WithField("last", last)
		ilast, err := strconv.ParseInt(last, 10, 64)
		if err != nil {
			// this should never happen but just in case someone messed up data in redis
			return err
		}

		if ilast > point.TimeStamp {
			l.Trace("new point is older")
			// This is an older point, that happens sometimes because of inconsistent caching
			// on vatsim side. For now we just skip this point
			return nil
		}

		if trackSize > 1 {
			// We may want to compress this point if the aircraft is standing at the
			// gate

			lastPoint, err := r.getPoint(ctx, trackID, last)
			if err != nil {
				l.WithError(err).Error("error getting point")
				return err
			}

			if !point.NE(*lastPoint) {
				l.Trace("points match, removing old")
				// eligible for compression, remove last point
				err = r.cli.Del(ctx, pointKey(trackID, last)).Err()
				if err != nil {
					l.WithError(err).Error("error removing point")
				}
				err = r.cli.RPop(ctx, trackIdxKey).Err()
				if err != nil {
					l.WithError(err).Error("error removing point index")
				}
			} else {
				l.Trace("points don't match")
			}
		} else {
			l.Trace("track can't be compressed")
		}
	} else {
		l.Trace("track is empty")
	}

	ts := strconv.FormatInt(point.TimeStamp, 10)

	// push point itself
	pk := pointKey(trackID, ts)
	value := map[string]interface{}{
		"alt": point.Altitude,
		"gs":  point.Groundspeed,
		"hdg": point.Heading,
		"lat": point.Latitude,
		"lng": point.Longitude,
	}
	l.Trace("pushing point")
	err := r.cli.HSet(ctx, pk, value).Err()
	if err != nil {
		return err
	}

	// push point index
	l.Trace("pushing point index")
	return r.cli.RPush(ctx, trackIdxKey, ts).Err()
}

func (r *RedisReadWriter) deleteTrack(ctx context.Context, trackID string) error {
	l := log.WithFields(logrus.Fields{
		"func": "deleteTrack",
		"tid":  trackID,
	})

	if !r.trackExists(ctx, trackID) {
		return track.ErrNotFound
	}

	idxkey := pointsIndexKey(trackID)

	keys, err := r.cli.LRange(ctx, idxkey, 0, -1).Result()
	if err != nil {
		l.WithError(err).Error("error getting track point index")
		return err
	}

	err = r.cli.Del(ctx, keys...).Err()
	if err != nil {
		l.WithError(err).Error("error deleting points")
		return err
	}

	err = r.cli.Del(ctx, idxkey).Err()
	if err != nil {
		l.WithError(err).Error("error deleting point index")
		return err
	}

	ck := trackCreatedKey(trackID)
	err = r.cli.Del(ctx, ck).Err()
	if err != nil {
		l.WithError(err).Error("error deleting track created ts")
		return err
	}

	err = r.cli.SRem(ctx, keyTrackIDs, trackID).Err()
	if err != nil {
		l.WithError(err).Error("error deleting track id from index")
	}

	return err
}

func (r *RedisReadWriter) gc() {
	l := log.WithField("func", "gc")
	t := time.NewTicker(r.cfg.PurgePeriod)
	defer t.Stop()

	ctx := context.Background()

	for range t.C {
		now := time.Now()
		l.Debug("running tracks garbage collector")

		trackIDs, err := r.cli.SMembers(ctx, keyTrackIDs).Result()
		if err != nil {
			l.WithError(err).Error("error reading track ids")
			continue
		}

		for _, trackID := range trackIDs {
			ck := trackCreatedKey(trackID)
			createdUx, err := r.getInt64(ctx, ck)
			if err != nil {
				l.WithField("track_id", trackID).WithError(err).Error("error reading track created ts")
				continue
			}
			ctime := time.Unix(createdUx, 0)
			if now.Sub(ctime) > r.cfg.PurgePeriod {
				err = r.deleteTrack(ctx, trackID)
				if err != nil {
					l.WithField("track_id", trackID).WithError(err).Error("error deleting track")
					continue
				}
			}
		}
	}
}
