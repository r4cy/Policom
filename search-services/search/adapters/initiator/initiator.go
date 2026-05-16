package initiator

import (
	"context"
	"log/slog"
	"time"

	"yadro.com/course/search/core"
)

type InitiatorIndex struct {
	log       *slog.Logger
	ttl       time.Duration
	rebuilder core.Rebuilder
}

func NewInitiatorIndex(log *slog.Logger, ttl time.Duration, rebuilder core.Rebuilder) *InitiatorIndex {
	return &InitiatorIndex{
		log:       log,
		ttl:       ttl,
		rebuilder: rebuilder,
	}
}

func (i *InitiatorIndex) Start(ctx context.Context) {
	i.log.Debug("running initiator adapter for index")
	_ = i.rebuilder.BuildIndex(ctx)
	ticker := time.NewTicker(i.ttl)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = i.rebuilder.BuildIndex(ctx)
				i.log.Info("start to rebuild the index")
			case <-ctx.Done():
				return
			}
		}
	}()
}
