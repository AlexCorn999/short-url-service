package store

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	IDStorage    = 1
	ErrConfilict = errors.New("URL already exists in the database")
)

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewURL(short, original string) *URL {
	return &URL{
		ShortURL:    short,
		OriginalURL: original,
	}
}

// Database общая реализация базы данных.
type Database interface {
	WriteURL(url *URL, ssh string) error
	ReadURL(url *URL, ssh string) error
	Conflict(url *URL) (string, error)
	Close() error
	CheckPing() error
}

// Postgres реализует хранение в postgres.
type Postgres struct {
	store *sql.DB
}

func NextID(id *int) {
	*id++
}

func BackID(id *int) {
	*id--
}

// NewPostgres открывает подключение к базе данных.
func NewPostgres(addr string) (*Postgres, error) {
	db, err := goose.OpenDBWithDriver("pgx", addr)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	err = goose.Up(db, "./migrations")
	if err != nil {
		log.Fatalf(err.Error())
	}

	return &Postgres{
		store: db,
	}, nil
}

// CloseDB закрывает подключение к базе данных.
func (d *Postgres) Close() error {
	return d.store.Close()
}

// WriteURL добавляет URL в базу данных.
func (d *Postgres) WriteURL(url *URL, ssh string) error {
	result, err := d.store.Exec("insert into url (id, shorturl, originalurl) values ($1, $2, $3) on conflict (shorturl) do nothing", ssh, url.OriginalURL, url.ShortURL)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrConfilict
	}

	return nil
}

func (d *Postgres) Conflict(url *URL) (string, error) {
	rows, err := d.store.Query("select originalurl from url where shorturl = $1", url.OriginalURL)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var result string
	for rows.Next() {
		if err = rows.Scan(&result); err != nil {
			return "", err
		}
	}
	if err = rows.Err(); err != nil {
		return "", err
	}

	return result, nil
}

// ReadURL возвращает адрес по ключу из БД.
func (d *Postgres) ReadURL(url *URL, ssh string) error {
	row := d.store.QueryRow("select shorturl from url where id = $1", ssh)

	var link string

	if err := row.Scan(&link); err != nil {
		return err
	}

	if link == "" {
		return errors.New("there was no link to the address specified")
	}

	url.OriginalURL = link
	return nil
}

// CheckPing проверяет подключение к базе данных.
func (d *Postgres) CheckPing() error {
	return d.store.Ping()
}
