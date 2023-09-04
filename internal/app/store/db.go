package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	bolt "go.etcd.io/bbolt"
)

var (
	IDStorage = 1
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

// Структура БД для файла, БД или памяти.
type DB struct {
	dataBase *pgx.Conn
	Store    *bolt.DB
	Bucket   *bolt.Bucket
	MemoryDB map[string]string
}

func NewDB(d *bolt.DB) *DB {
	return &DB{
		Store: d,
	}
}

func NewMemoryDB() *DB {
	return &DB{
		MemoryDB: make(map[string]string),
	}
}

func NextID(id *int) {
	*id++
}

// OpenDB открывает подключение к базе данных
func (d *DB) OpenDB(addr string) (*DB, error) {
	db, err := pgx.Connect(context.Background(), addr)
	if err != nil {
		return nil, err
	}
	return &DB{
		dataBase: db,
	}, nil
}

// CloseDB закрывает подключение к базе данных
func (d *DB) CloseDB() error {
	return d.dataBase.Close(context.Background())
}

// InitTables первичная инициализация таблицы для хранения URL.
func (d *DB) InitTables() error {
	_, err := d.dataBase.Exec(context.Background(), "create table url(id varchar(255) not null, shorturl varchar(255) not null, originalurl varchar(255) not null)")
	return err
}

func (d *DB) CheckTables() error {
	rows, err := d.dataBase.Query(context.Background(), "SELECT id FROM url")
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
			return nil // The table already exists, no need to do anything
		}
	}
	return nil
}

// AddURL добавляет URL в базу данных.
func (d *DB) AddURL(url *URL, ssh string) error {
	_, err := d.dataBase.Exec(context.Background(), "insert into url (id, shorturl, originalurl) values ($1, $2, $3)", ssh, url.OriginalURL, url.ShortURL)
	return err
}

// AddrBack возвращает адрес по ключу из БД.
func (d *DB) AddrBack(ssh string) (string, error) {
	rows, err := d.dataBase.Query(context.Background(), "select shorturl from url where id = $1", ssh)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var link string
	for rows.Next() {
		if err = rows.Scan(&link); err != nil {
			return "", err
		}
	}

	if err = rows.Err(); err != nil {
		return "", err
	}

	if link == "" {
		return "", errors.New("there was no link to the address specified")
	}

	return link, nil
}

// CheckPing проверяет подключение к базе данных.
func (d *DB) CheckPing() error {
	return d.dataBase.Ping(context.Background())
}

// WriteURL записывает url по ключу
func (d *DB) WriteURL(url *URL, ssh string) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	d.Store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("URLBucket"))
		err := b.Put([]byte(ssh), data)
		return err
	})
	return nil
}

// ReadURL вычитывает url по ключу
func (d *DB) ReadURL(url *URL, ssh string) error {
	var v []byte
	d.Store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("URLBucket"))
		v = b.Get([]byte(ssh))
		return nil
	})

	if err := json.Unmarshal(v, url); err != nil {
		return err
	}
	return nil
}

// CreateBacketURL создает хранилище для url
func (d *DB) CreateBacketURL() error {
	return d.Store.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte("URLBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		d.Bucket = b
		return nil
	})
}
