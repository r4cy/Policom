package postgres

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"yadro.com/course/search/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func NewDBClient(log *slog.Logger, address string) (*DB, error) {
	log.Debug("running db")
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Get(ctx context.Context, words []string, limit int) ([]core.Comics, error) {
	var comics []core.Comics
	err := db.conn.SelectContext(ctx, &comics,
		`SELECT c.id, c.title, c.url
		FROM comics c
		JOIN comics_keywords ck ON c.id = ck.comics_id
		WHERE ck.word = ANY($1)
		GROUP BY c.id, c.title, c.url
		ORDER BY COUNT(*) DESC
		LIMIT $2`,
		words, limit,
	)
	return comics, err
}

func (db *DB) GetByID(ctx context.Context, id int) (core.Comics, error) {
	var comics core.Comics
	err := db.conn.GetContext(ctx, &comics,
		`SELECT id, title, url, description, img_description
		FROM comics		
		WHERE id = $1`,
		id,
	)
	return comics, err
}

func (db *DB) GetPage(ctx context.Context, limit int) ([]core.Comics, error) {
	var comics []core.Comics
	err := db.conn.SelectContext(ctx, &comics,
		`SELECT id, title, url
		FROM comics
		ORDER BY RANDOM()
		LIMIT $1`,
		limit,
	)
	return comics, err
}

func (db *DB) Search(ctx context.Context, ids []int, limit int) ([]core.Comics, error) {
	var comics []core.Comics
	err := db.conn.SelectContext(ctx, &comics,
		`SELECT id, title, url
		FROM comics
		WHERE id = ANY($1)
		ORDER BY array_position($1, id)
		LIMIT $2`,
		pq.Array(ids), limit,
	)
	return comics, err
}

func (db *DB) GetAll(ctx context.Context) ([]core.ComicsFWords, error) {
	var comics []core.ComicsFWords
	err := db.conn.SelectContext(ctx, &comics,
		`SELECT comics_id as id, word
		FROM comics_keywords`,
	)
	return comics, err
}
