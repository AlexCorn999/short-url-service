package store

import (
	"context"
	"encoding/json"
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

// Bolt база данных
type DB struct {
	dataBase *pgx.Conn
	Store    *bolt.DB
	Bucket   *bolt.Bucket
}

func NewDB(d *bolt.DB) *DB {
	return &DB{
		Store: d,
	}
}

func NextID(id *int) {
	*id++
}

// OpenDB открывает подключение к базе данных
func (d *DB) OpenDB(addr string) error {
	db, err := pgx.Connect(context.Background(), addr)
	if err != nil {
		return err
	}

	d.dataBase = db
	return nil
}

// CloseDB закрывает подключение к базе данных
func (d *DB) CloseDB() error {
	return d.dataBase.Close(context.Background())
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
