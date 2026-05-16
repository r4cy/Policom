package aaa

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
	dbUser "yadro.com/course/api/adapters/aaa/db"
	"yadro.com/course/api/core"
)

type AAA struct {
	db        *dbUser.DB
	secretKey []byte
	tokenTTL  time.Duration
	log       *slog.Logger
}

func NewAAAClient(dbAddress string, adminName string, adminPassword string, tokenTTL time.Duration, secret string, log *slog.Logger) (*AAA, error) {
	db, err := dbUser.NewDB(log, dbAddress)
	if err != nil {
		return nil, err
	}

	a := &AAA{
		db:        db,
		secretKey: []byte(secret),
		tokenTTL:  tokenTTL,
		log:       log,
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("bad password", "password", adminPassword)
		return &AAA{}, err
	}

	if err = a.db.AddUser(context.Background(), adminName, "admin@admin.ru", string(hash), core.RoleAdmin); err != nil {
		a.log.Error("can't to add user", "error", err)
		return &AAA{}, err
	}

	return a, nil
}

func (a *AAA) Ping(context.Context) error {
	return nil
}
