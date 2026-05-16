package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func NewDB(log *slog.Logger, address string) (*DB, error) {
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

// Метод для добавление в БД новый комикс
func (db *DB) Add(ctx context.Context, comics core.Comics) error {
	if _, err := db.conn.ExecContext(ctx,
		"INSERT INTO comics (id, url, title, description, img_description) VALUES ($1, $2, $3, $4, $5)",
		comics.ID, comics.URL, comics.Title, comics.Description, comics.ImgDescription,
	); err != nil {
		db.log.Error(
			"failed to insert in comics table",
			"comics_id", comics.ID,
			"err", err,
		)
		return err
	}

	for _, word := range comics.Words {
		if _, err := db.conn.ExecContext(ctx,
			"INSERT INTO comics_keywords (comics_id, word) VALUES ($1, $2)",
			comics.ID, word,
		); err != nil {
			db.log.Error(
				"failed to insert in comics_keywords table",
				"comics_id", comics.ID,
				"err", err,
			)
			return err
		}
	}

	return nil
}

// Метод для вывода статистики из БД
func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var wordsTotal, wordsUnique, comicsFetched int

	if err := db.conn.QueryRowContext(ctx, `
		SELECT 
			(SELECT COUNT(word) FROM comics_keywords),
			(SELECT COUNT(DISTINCT word) FROM comics_keywords),
			(SELECT COUNT(id) FROM comics)
	`).Scan(&wordsTotal, &wordsUnique, &comicsFetched); err != nil {
		db.log.Error(
			"failed to make stats from tables",
			"err", err,
		)
		return core.DBStats{}, err
	}

	return core.DBStats{
		WordsTotal:    wordsTotal,
		WordsUnique:   wordsUnique,
		ComicsFetched: comicsFetched,
	}, nil
}

// Метод, возвращает все загруженные id комиксов в БД
func (db *DB) IDs(ctx context.Context) ([]int, error) {
	ids := make([]int, 0)
	if err := db.conn.SelectContext(ctx,
		&ids,
		"SELECT id FROM comics",
	); err != nil {
		db.log.Error(
			"failed to download comics id's from comics table",
			"err", err,
		)
		return nil, err
	}

	return ids, nil
}

// Метод удаление данных в БД
func (db *DB) Drop(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx,
		"TRUNCATE TABLE users_comics_saved, comics_keywords, comics",
	); err != nil {
		db.log.Error(
			"faild to drop the tables for migrations",
			"err", err)
		return err
	}
	return nil
}
