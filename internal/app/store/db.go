package store

import (
	"encoding/json"
	"fmt"

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
	Store  *bolt.DB
	Bucket *bolt.Bucket
}

func NewDB(d *bolt.DB) *DB {
	return &DB{
		Store: d,
	}
}

func NextID(id *int) {
	*id++
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
