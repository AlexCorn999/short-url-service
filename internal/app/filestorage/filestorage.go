package filestorage

import (
	"encoding/json"
	"fmt"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
	bolt "go.etcd.io/bbolt"
)

// BoltDB реализует хранение в файле.
type BoltDB struct {
	Store  *bolt.DB
	Bucket *bolt.Bucket
}

// NewBoltDB инициализирует базу данных.
func NewBoltDB(filePath string) (*BoltDB, error) {
	db, err := bolt.Open(filePath, 0666, nil)
	if err != nil {
		return nil, err
	}

	var b *bolt.Bucket
	err = db.Update(func(tx *bolt.Tx) error {
		b, err = tx.CreateBucket([]byte("URLBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &BoltDB{
		Store:  db,
		Bucket: b,
	}, nil
}

// WriteURL записывает url по ключу
func (d *BoltDB) WriteURL(url *store.URL, id int, ssh *string) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	d.Store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("URLBucket"))
		err := b.Put([]byte(*ssh), data)
		return err
	})
	return nil
}

// ReadURL вычитывает url по ключу.
func (d *BoltDB) ReadURL(url *store.URL, ssh string) error {
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

func (d *BoltDB) GetAllURL(id int) ([]store.URL, error) {
	var userURL []store.URL

	d.Store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("URLBucket"))
		c := b.Cursor()
		for k, value := c.First(); k != nil; k, value = c.Next() {
			var url store.URL
			if err := json.Unmarshal([]byte(value), &url); err != nil {
				return err
			}
			if url.Creator == id {
				userURL = append(userURL, url)
			}

		}
		return nil
	})

	return userURL, nil
}

func (d *BoltDB) Close() error {
	return d.Store.Close()
}

// CheckPing проверяет подключение к базе данных.
func (d *BoltDB) CheckPing() error {
	return nil
}

func (d *BoltDB) Conflict(url *store.URL) (string, error) {
	return "", nil
}

func (d *BoltDB) RewriteURL(url *store.URL) error {
	return nil
}

func (d *BoltDB) InitID() (int, error) {
	return -1, nil
}
