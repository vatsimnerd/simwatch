package provider

import (
	"fmt"

	"github.com/vatsimnerd/simwatch/track"
	"github.com/vatsimnerd/simwatch/track/memory"
	"github.com/vatsimnerd/simwatch/track/redistr"
	"github.com/vatsimnerd/simwatch/track/sqlitetr"
)

func (p *Provider) setupTrackStore() error {
	switch p.tcfg.Engine {
	case "memory":
		log.Info("registering memory track engine")
		track.RegisterTrackReadWriter(memory.ReadWriter)
	case "redis":
		log.Info("registering redis track engine")
		track.RegisterTrackReadWriter(redistr.ReadWriter)
	case "sqlite":
		log.Info("registering sqlite track engine")
		track.RegisterTrackReadWriter(sqlitetr.ReadWriter)
	default:
		return fmt.Errorf("invalid track engine '%s'", p.tcfg.Engine)
	}
	return track.Configure(&p.tcfg.Options)
}
