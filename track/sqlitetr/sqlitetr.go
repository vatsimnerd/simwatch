package sqlitetr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/simwatch/config"
	"github.com/vatsimnerd/simwatch/track"
)

type SQLiteReadWriter struct {
	cfg        *config.TrackConfigOptions
	db         *sql.DB
	stop       chan struct{}
	configured bool
}

var (
	ReadWriter = &SQLiteReadWriter{stop: make(chan struct{})}
	log        = logrus.WithField("module", "track.memory")
)

func (r *SQLiteReadWriter) Configure(cfg *config.TrackConfigOptions) error {
	r.cfg = cfg
	err := r.setupClient()
	if err != nil {
		return err
	}

	go r.gc()
	r.configured = true
	return nil
}

func (r *SQLiteReadWriter) gc() {
	t := time.NewTicker(r.cfg.PurgePeriod)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			log.Info("running tracks garbage collect")
			expiredAt := time.Now().Add(-r.cfg.PurgePeriod)
			res, err := r.db.Exec("DELETE FROM tracks WHERE created_at < ?", expiredAt)
			if err != nil {
				log.WithError(err).Error("error garbage-collecting tracks")
			} else {
				count, err := res.RowsAffected()
				if err != nil {
					log.WithError(err).Error("error counting deleted tracks")
				} else {
					log.WithField("count", count).Error("tracks garbage-collected")
				}
			}
			res.RowsAffected()
		case <-r.stop:
			return
		}
	}
}

func (r *SQLiteReadWriter) setupClient() error {
	db, err := sql.Open("sqlite3", r.cfg.Path)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	r.db = db
	return nil
}

func (r *SQLiteReadWriter) LoadTrackByID(ctx context.Context, trackCode string) (*track.Track, error) {
	var id int64
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx,
		"SELECT id, created_at FROM tracks WHERE track_code = ?",
		trackCode).Scan(&id, &createdAt)

	if err != nil {
		return nil, err
	}

	stmt, err := r.db.PrepareContext(ctx, "SELECT latitude, longitude, altitude, heading, groundspeed, ts FROM track_points WHERE track_id = ? ORDER BY ts")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	cur, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	var lat, lng float64
	var alt, hdg, gs int64
	var ts time.Time

	points := make([]track.TrackPoint, 0)
	for cur.Next() {
		err = cur.Scan(&lat, &lng, &alt, &hdg, &gs, &ts)
		if err != nil {
			return nil, err
		}

		pt := track.TrackPoint{
			Latitude:    lat,
			Longitude:   lng,
			Altitude:    int(alt),
			Heading:     int(hdg),
			Groundspeed: int(gs),
			TimeStamp:   ts.Unix(),
		}
		points = append(points, pt)
	}

	if err = cur.Err(); err != nil {
		return nil, err
	}

	return &track.Track{
		CreatedAt: createdAt,
		Points:    points,
	}, nil
}

func (r *SQLiteReadWriter) ListIDs(ctx context.Context) ([]string, error) {
	stmt, err := r.db.PrepareContext(ctx, "SELECT track_code FROM tracks ORDER BY track_code")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	cur, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer cur.Close()

	var trackCode string
	res := make([]string, 0)
	for cur.Next() {
		cur.Scan(&trackCode)
		res = append(res, trackCode)
	}

	if err = cur.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *SQLiteReadWriter) WriteTrack(ctx context.Context, p *merged.Pilot) error {
	trackCode, point := track.ExtractTrackData(p)
	trackID, err := r.getTrackID(ctx, trackCode)
	if err != nil {
		return fmt.Errorf("error checking track: %w", err)
	}
	return r.writePoint(ctx, trackID, point)
}

func (r *SQLiteReadWriter) Close() error {
	r.stop <- struct{}{}
	return r.db.Close()
}

func (r *SQLiteReadWriter) getTrackID(ctx context.Context, trackCode string) (int64, error) {
	var trackID int64
	err := r.db.QueryRowContext(ctx, "SELECT id FROM tracks WHERE track_code = ?", trackCode).Scan(&trackID)
	if err != nil {
		stmt, err := r.db.PrepareContext(ctx, "INSERT INTO tracks (track_code) VALUES (?)")
		if err != nil {
			return 0, err
		}
		defer stmt.Close()

		res, err := stmt.Exec(trackCode)
		if err != nil {
			return 0, err
		}

		trackID, err = res.LastInsertId()
		if err != nil {
			return 0, err
		}
	}
	return trackID, nil
}

func (r *SQLiteReadWriter) writePoint(ctx context.Context, trackID int64, point track.TrackPoint) error {
	stmt, err := r.db.PrepareContext(ctx, "INSERT INTO track_points (track_id, latitude, longitude, altitude, heading, groundspeed) VALUES (?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		trackID,
		point.Latitude,
		point.Longitude,
		point.Altitude,
		point.Heading,
		point.Groundspeed,
	)

	return err
}
