package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	IDStorage    = 1
	ErrConfilict = errors.New("URL already exists in the database")
	ErrDeleted   = errors.New("has been deleted")
)

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	Creator     int
	DeletedFlag bool
}

func NewURL(short, original string, creator int) *URL {
	return &URL{
		ShortURL:    short,
		OriginalURL: original,
		Creator:     creator,
		DeletedFlag: false,
	}
}

// Database общая реализация базы данных.
type Database interface {
	WriteURL(url *URL, id int, ssh *string) error
	RewriteURL(url *URL) error
	ReadURL(url *URL, ssh string) error
	GetAllURL(id int) ([]URL, error)
	Conflict(url *URL) (string, error)
	DeleteURL(shortURL string, creator int) error
	Close() error
	InitID() (int, error)
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
func (d *Postgres) WriteURL(url *URL, id int, ssh *string) error {

	result, err := d.store.Exec("insert into url (shorturl, originalurl, user_id, deleted_flag) values ($1, $2, $3, $4) on conflict (shorturl) do nothing", url.OriginalURL, url.ShortURL, url.Creator, url.DeletedFlag)
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

	err = d.store.QueryRow("SELECT id FROM url WHERE shorturl = $1", url.OriginalURL).Scan(ssh)
	if err != nil {
		return err
	}

	return nil
}

// RewriteURL добавляет URL в базу данных.
func (d *Postgres) RewriteURL(url *URL) error {
	result, err := d.store.Exec("update url SET shorturl = $1, originalurl = $2 WHERE shorturl = $1", url.OriginalURL, url.ShortURL)
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
	deletedFlag := false
	row := d.store.QueryRow("select shorturl, deleted_flag from url where id = $1", ssh)

	var link string

	if err := row.Scan(&link, &deletedFlag); err != nil {
		return err
	}

	if deletedFlag {
		return ErrDeleted
	}

	if link == "" {
		return errors.New("there was no link to the address specified")
	}

	url.OriginalURL = link
	return nil
}

// GetAllURL возвращает все сокращенные url пользователя.
func (d *Postgres) GetAllURL(id int) ([]URL, error) {
	var urls []URL
	rows, err := d.store.Query("SELECT shorturl, originalurl FROM url WHERE user_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u URL
		err := rows.Scan(&u.ShortURL, &u.OriginalURL)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		urls = append(urls, u)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return urls, nil
}

// CheckPing проверяет подключение к базе данных.
func (d *Postgres) CheckPing() error {
	return d.store.Ping()
}

// InitID первичная инициализация.
func (d *Postgres) InitID() (int, error) {
	var maxID int
	err := d.store.QueryRow("select MAX(user_id) from url").Scan(&maxID)
	if err != nil {
		maxID = 1
	}
	return maxID, nil
}

// InitID первичная инициализация.
func (d *Postgres) DeleteURL(shortURL string, creator int) error {
	deletedFlag := true

	result, err := d.store.Exec("update url SET deleted_flag = $1 WHERE originalurl = $2 and user_id = $3", deletedFlag, shortURL, creator)
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
