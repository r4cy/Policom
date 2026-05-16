package broker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
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

// Метод для публикации события в топике
func (n *Nats) Publish(ctx context.Context, message string) error {
	n.log.Info("publish update message")
	err := n.conn.Publish("xkcd.db.updated", []byte(message))
	if err != nil {
		n.log.Error("could not publish message", "error", err)
		return err
	}
	if err := n.conn.Flush(); err != nil {
		n.log.Error("flush error", "error", err)
	}
	return nil
}
