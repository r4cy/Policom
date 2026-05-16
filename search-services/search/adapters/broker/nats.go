package broker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"yadro.com/course/search/core"
)

type Nats struct {
	conn *nats.Conn
	log  *slog.Logger
}

func NewClientNats(address string, log *slog.Logger) (*Nats, error) {
	log.Debug("running Nats")
	nc, err := nats.Connect("nats://" + address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, fmt.Errorf("problem connection to broker error: %w", err)
	}
	return &Nats{
		conn: nc,
		log:  log,
	}, nil
}

// Подписка на топик для отслеживания событий сервиса update
func (n *Nats) Subscribe(ctx context.Context, handler func(context.Context, core.EventDB)) error {
	n.log.Info("subsctibe to update server")
	_, err := n.conn.Subscribe("xkcd.db.updated", func(msg *nats.Msg) {
		handler(ctx, core.EventDB(msg.Data))
	})

	if err != nil {
		return err
	}

	<- ctx.Done()
	return nil
}