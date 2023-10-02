package memorystorage

import (
	"encoding/json"
	"fmt"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
)

// MemoryStorage реализует хранение в мапе.
type MemoryStorage struct {
	store map[string]string
}

// NewMemoryStorage инициализирует хранилище.
func NewMemoryStorage() *MemoryStorage {

	return &MemoryStorage{
		store: make(map[string]string),
	}
}

// WriteURL добавляет URL в хранилище.
func (m *MemoryStorage) WriteURL(url *store.URL, id int, ssh *string) error {
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("error from local storage. can't convert url - %s ", err)
	}

	m.store[*ssh] = string(data)
	return nil
}

// ReadURL вычитывает url по ключу.
func (m *MemoryStorage) ReadURL(url *store.URL, ssh string) error {
	value, ok := m.store[ssh]
	if !ok {
		return fmt.Errorf("error from local storage. url not found %s", ssh)
	}

	if err := json.Unmarshal([]byte(value), url); err != nil {
		return fmt.Errorf("error from local storage. can't convert url - %s ", err)
	}

	if url.DeletedFlag {
		return store.ErrDeleted
	}

	return nil
}

// GetAllURL возвращает все сокращенные url пользователя.
func (m *MemoryStorage) GetAllURL(id int) ([]store.URL, error) {
	var userURL []store.URL

	for _, value := range m.store {
		var url store.URL
		if err := json.Unmarshal([]byte(value), &url); err != nil {
			return nil, fmt.Errorf("error from local storage. can't convert url - %s ", err)
		}
		if url.Creator == id {
			userURL = append(userURL, url)
		}
	}
	return userURL, nil
}

// DeleteURL удаляет url у текущего пользователя.
func (m *MemoryStorage) DeleteURL(tasks []store.Task) error {

	fmt.Println("из функции удаления ", tasks)

	for _, task := range tasks {
		for key, value := range m.store {
			var url store.URL
			if err := json.Unmarshal([]byte(value), &url); err != nil {
				return fmt.Errorf("error from local storage. can't convert url - %s ", err)
			}

			if url.ShortURL == task.Link && url.Creator == task.Creator {
				url.DeletedFlag = true

				data, err := json.Marshal(url)
				if err != nil {
					return fmt.Errorf("error from local storage. can't convert url - %s ", err)
				}

				fmt.Println(url)
				m.store[key] = string(data)
			}
		}
	}

	return nil

}

func (m *MemoryStorage) RewriteURL(url *store.URL) error {
	return nil
}

func (m *MemoryStorage) Close() error {
	return nil
}

func (m *MemoryStorage) CheckPing() error {
	return nil
}

func (m *MemoryStorage) Conflict(url *store.URL) (string, error) {
	return "", nil
}

func (m *MemoryStorage) InitID() (int, error) {
	return -1, nil
}
