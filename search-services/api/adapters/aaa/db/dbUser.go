package dbUser

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/api/core"
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

// Метод для добавление в БД нового пользователя
func (db *DB) AddUser(ctx context.Context, username string, email string, password_hash string, role core.UserRole) error {
	if _, err := db.conn.ExecContext(ctx,
		"INSERT INTO users (username, email, password_hash, role) VALUES ($1, $2, $3, $4)",
		username, email, password_hash, role,
	); err != nil {
		db.log.Error(
			"failed to insert in users table",
			"username", username,
			"err", err,
		)
		return err
	}
	return nil
}

// Метод для проверки пользователя в БД
func (db *DB) CheckUser(ctx context.Context, username string) (core.UserLog, error) {
	var user core.UserLog
	if err := db.conn.GetContext(ctx, &user,
		`SELECT id, email, password_hash, role
		FROM users
		WHERE username = $1`,
		username,
	); err != nil {
		db.log.Error(
			"failed to check user in table",
			"username", username,
			"err", err,
		)
		return core.UserLog{}, err
	}
	return user, nil
}

// Метод для добавления лайка(сохранения) на комикс от пользователя
func (db *DB) LikeComics(ctx context.Context, user_id string, comics_id int) error {
	if _, err := db.conn.ExecContext(ctx,
		`INSERT INTO users_comics_saved (user_id, comics_id) VALUES ($1, $2)
		ON CONFLICT (user_id, comics_id) DO NOTHING`,
		user_id, comics_id,
	); err != nil {
		db.log.Error(
			"failed to like comics",
			"comics_id", comics_id,
			"err", err,
		)
		return err
	}
	return nil
}

// Метод для удаление лайка(сохранения) на комикс от пользователя
func (db *DB) DiselikeComics(ctx context.Context, user_id string, comics_id int) error {
	if _, err := db.conn.ExecContext(ctx,
		`DELETE FROM users_comics_saved 
		WHERE user_id = $1 AND comics_id = $2`,
		user_id, comics_id,
	); err != nil {
		db.log.Error(
			"failed to dislike comics",
			"comics_id", comics_id,
			"err", err,
		)
		return err
	}
	return nil
}

// Метод для сбора понравившихся комиксов
func (db *DB) LikesComics(ctx context.Context, user_id string) ([]core.Comics, error) {
	var comics []core.Comics
	err := db.conn.SelectContext(ctx, &comics,
		`SELECT c.id, c.title, c.url
		FROM comics c
		JOIN users_comics_saved s ON c.id = s.comics_id
		WHERE s.user_id = $1
		ORDER BY s.saved_at DESC`,
		user_id,
	)
	if err != nil {
		db.log.Error(
			"failed to look likes comics",
			"user_id", user_id,
			"err", err,
		)
	}
	return comics, err
}

// Метод удаление данных в БД, без миграции, не использовать пока что!
func (db *DB) Drop(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx,
		"DROP TABLE IF EXISTS users",
	); err != nil {
		db.log.Error(
			"faild to drop the tables for migrations",
			"err", err)
		return err
	}
	return nil
}
