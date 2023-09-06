package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
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
	store *pgx.Conn
}

func NextID(id *int) {
	*id++
}

func BackID(id *int) {
	*id--
}

// NewPostgres открывает подключение к базе данных.
func NewPostgres(addr string) (*Postgres, error) {
	db, err := pgx.Connect(context.Background(), addr)
	if err != nil {
		return nil, err
	}
	return &Postgres{
		store: db,
	}, nil
}

// CloseDB закрывает подключение к базе данных.
func (d *Postgres) Close() error {
	return d.store.Close(context.Background())
}

// InitTables первичная инициализация таблицы для хранения URL.
func (d *Postgres) InitTables() error {
	_, err := d.store.Exec(context.Background(), "create table url(id varchar(255) not null primary key, shorturl varchar(255) not null unique, originalurl varchar(255) not null)")
	return err
}

func (d *Postgres) CheckTables() error {
	rows, err := d.store.Query(context.Background(), "SELECT id FROM url")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}

		if tableName == "id" {
			return nil
		}
	}
	return nil
}

// WriteURL добавляет URL в базу данных.
func (d *Postgres) WriteURL(url *URL, ssh string) error {
	result, err := d.store.Exec(context.Background(), "insert into url (id, shorturl, originalurl) values ($1, $2, $3) on conflict (shorturl) do nothing", ssh, url.OriginalURL, url.ShortURL)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return ErrConfilict
	}

	return nil
}

func (d *Postgres) Conflict(url *URL) (string, error) {
	rows, err := d.store.Query(context.Background(), "select originalurl from url where shorturl = $1", url.OriginalURL)
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

// ReadURL(url *store.URL, ssh string) error
//
// ReadURL возвращает адрес по ключу из БД.
func (d *Postgres) ReadURL(url *URL, ssh string) error {
	row := d.store.QueryRow(context.Background(), "select shorturl from url where id = $1", ssh)

	var link string

	if err := row.Scan(&link); err != nil {
		return err
	}

	if link == "" {
		return errors.New("there was no link to the address specified")
	}

	if err := json.Unmarshal([]byte(link), url); err != nil {
		return err
	}

	return nil
}

// CheckPing проверяет подключение к базе данных.
func (d *Postgres) CheckPing() error {
	return d.store.Ping(context.Background())
}
