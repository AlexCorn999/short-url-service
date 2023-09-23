package memorystorage

import (
	"encoding/json"
	"fmt"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
)

type keyForMemory struct {
	id     string
	userID string
}

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
		return err
	}

	m.store[*ssh] = string(data)
	return nil
}

// ReadURL вычитывает url по ключу.
func (m *MemoryStorage) ReadURL(url *store.URL, ssh string) error {

	value, ok := m.store[ssh]
	if !ok {
		return fmt.Errorf("not found %s", ssh)
	}

	if err := json.Unmarshal([]byte(value), url); err != nil {
		return err
	}
	return nil
}

func (m *MemoryStorage) GetAllURL(id int) ([]store.URL, error) {
	var userURL []store.URL

	for _, value := range m.store {
		var url store.URL
		if err := json.Unmarshal([]byte(value), &url); err != nil {
			return nil, err
		}
		if url.Creator == id {
			userURL = append(userURL, url)
		}
	}
	return userURL, nil
}

// RewriteURL добавляет URL в базу данных.
func (m *MemoryStorage) RewriteURL(url *store.URL) error {
	return nil
}

func (m *MemoryStorage) Close() error {
	return nil
}

// CheckPing проверяет подключение к базе данных.
func (m *MemoryStorage) CheckPing() error {
	return nil
}

func (m *MemoryStorage) Conflict(url *store.URL) (string, error) {
	return "", nil
}
