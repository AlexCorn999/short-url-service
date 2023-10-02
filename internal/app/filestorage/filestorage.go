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
		return nil, fmt.Errorf("error from file. can't open file - %s ", err)
	}

	var b *bolt.Bucket
	err = db.Update(func(tx *bolt.Tx) error {
		b, err = tx.CreateBucket([]byte("URLBucket"))
		if err != nil {
			return fmt.Errorf("error from file. create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error from file. can't create bucket for url - %s ", err)
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
		return fmt.Errorf("error from file. can't convert url for bucket - %s ", err)
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
		return fmt.Errorf("error from file. can't convert url from bucket - %s ", err)
	}

	if url.DeletedFlag {
		return store.ErrDeleted
	}

	return nil
}

// GetAllURL возвращает все сокращенные url пользователя.
func (d *BoltDB) GetAllURL(id int) ([]store.URL, error) {
	var userURL []store.URL

	d.Store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("URLBucket"))
		c := b.Cursor()
		for k, value := c.First(); k != nil; k, value = c.Next() {
			var url store.URL
			if err := json.Unmarshal([]byte(value), &url); err != nil {
				return fmt.Errorf("error from file. can't convert url from bucket - %s ", err)
			}
			if url.Creator == id {
				userURL = append(userURL, url)
			}

		}
		return nil
	})

	return userURL, nil
}

// DeleteURL удаляет url у текущего пользователя.
func (d *BoltDB) DeleteURL(tasks []store.Task) error {

	type valuesForDelete struct {
		key  string
		data []byte
	}

	var forDelete []valuesForDelete

	for _, task := range tasks {
		d.Store.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("URLBucket"))
			c := b.Cursor()
			for k, value := c.First(); k != nil; k, value = c.Next() {
				var url store.URL
				if err := json.Unmarshal([]byte(value), &url); err != nil {
					return fmt.Errorf("error from file. can't convert url from bucket - %s ", err)
				}

				if url.ShortURL == task.Link && url.Creator == task.Creator {
					url.DeletedFlag = true
					data, err := json.Marshal(url)
					if err != nil {
						return fmt.Errorf("error from file. can't convert url for bucket - %s ", err)
					}

					var val valuesForDelete
					val.key = string(k)
					val.data = data
					forDelete = append(forDelete, val)
				}

			}
			return nil
		})
	}

	// перезапись значения
	d.Store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("URLBucket"))
		for _, value := range forDelete {
			b.Put([]byte(value.key), value.data)
		}
		return nil
	})

	return nil
}

func (d *BoltDB) Close() error {
	return d.Store.Close()
}

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
