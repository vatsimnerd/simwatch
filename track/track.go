package track

import (
	"fmt"
	"time"

	"github.com/vatsimnerd/simwatch-providers/merged"
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
	WriteTrack(*merged.Pilot) error
	LoadTrack(*merged.Pilot) (*Track, error)
	LoadTrackByID(string) (*Track, error)
	ListIDs() []string
	Configure(interface{}) error
}

var (
	readWriter TrackReadWriter = nil
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

func WriteTrack(p *merged.Pilot) error {
	return readWriter.WriteTrack(p)
}

func LoadTrack(p *merged.Pilot) (*Track, error) {
	return readWriter.LoadTrack(p)
}

func LoadTrackByID(id string) (*Track, error) {
	return readWriter.LoadTrackByID(id)
}

func ListIDs() []string {
	return readWriter.ListIDs()
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

func Configure(cfg interface{}) error {
	return readWriter.Configure(cfg)
}
