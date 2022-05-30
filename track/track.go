package track

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vatsimnerd/simwatch-providers/merged"
	"github.com/vatsimnerd/simwatch/config"
)

type TrackPoint struct {
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lng"`
	Heading     int     `json:"hdg"`
	Altitude    int     `json:"alt"`
	Groundspeed int     `json:"gs"`
	TimeStamp   int64   `json:"ts"`
}

type Track struct {
	CreatedAt time.Time
	Points    []TrackPoint
}

type TrackReadWriter interface {
	WriteTrack(context.Context, *merged.Pilot) error
	LoadTrack(context.Context, *merged.Pilot) (*Track, error)
	LoadTrackByID(context.Context, string) (*Track, error)
	ListIDs(context.Context) []string
	Configure(cfg *config.TrackConfigOptions) error
}

var (
	readWriter       TrackReadWriter = nil
	ErrNotFound                      = errors.New("track not found")
	ErrNotConfigured                 = errors.New("track writer not configured")
	ErrConfigInvalid                 = errors.New("invalid configuration type")
)

func (tp TrackPoint) NE(op TrackPoint) bool {
	return tp.Latitude != op.Latitude ||
		tp.Longitude != op.Longitude ||
		tp.Heading != op.Heading ||
		tp.Altitude != op.Altitude ||
		tp.Groundspeed != op.Groundspeed
}

func RegisterTrackReadWriter(trw TrackReadWriter) {
	readWriter = trw
}

func WriteTrack(ctx context.Context, p *merged.Pilot) error {
	return readWriter.WriteTrack(ctx, p)
}

func LoadTrack(ctx context.Context, p *merged.Pilot) (*Track, error) {
	return readWriter.LoadTrack(ctx, p)
}

func LoadTrackByID(ctx context.Context, id string) (*Track, error) {
	return readWriter.LoadTrackByID(ctx, id)
}

func ListIDs(ctx context.Context) []string {
	return readWriter.ListIDs(ctx)
}

func ExtractTrackData(p *merged.Pilot) (trackID string, point TrackPoint) {
	trackID = fmt.Sprintf("%s-%d-%d", p.Callsign, p.Cid, p.LogonTime.Unix())
	point = TrackPoint{
		Latitude:    p.Latitude,
		Longitude:   p.Longitude,
		Heading:     p.Heading,
		Altitude:    p.Altitude,
		Groundspeed: p.Groundspeed,
		TimeStamp:   time.Now().Unix(),
	}
	return
}

func Configure(cfg *config.TrackConfigOptions) error {
	return readWriter.Configure(cfg)
}
