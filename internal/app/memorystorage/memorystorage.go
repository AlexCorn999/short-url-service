package memorystorage

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/AlexCorn999/short-url-service/internal/app/store"
)

type keyForMemory struct {
	id     string
	userID string
}

// MemoryStorage реализует хранение в мапе.
type MemoryStorage struct {
	store map[keyForMemory]string
}

// NewMemoryStorage инициализирует хранилище.
func NewMemoryStorage() *MemoryStorage {

	return &MemoryStorage{
		store: make(map[keyForMemory]string),
	}
}

// WriteURL добавляет URL в хранилище.
func (m *MemoryStorage) WriteURL(url *store.URL, id int, ssh *string) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}

	// создание двойного ключа
	var key keyForMemory
	key.id = *ssh
	key.userID = strconv.Itoa(id)

	//m.store[*ssh] = string(data)\
	m.store[key] = string(data)
	return nil
}

// ReadURL вычитывает url по ключу.
func (m *MemoryStorage) ReadURL(url *store.URL, id int, ssh string) error {

	// создание двойного ключа
	var key keyForMemory
	key.id = ssh
	key.userID = strconv.Itoa(id)

	value, ok := m.store[key]
	if !ok {
		return fmt.Errorf("not found %s", ssh)
	}

	if err := json.Unmarshal([]byte(value), url); err != nil {
		return err
	}
	return nil
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

func (m *MemoryStorage) Create(id int) (int, error) {
	return 0, nil
}

func (m *MemoryStorage) GetUser(userID int) error {
	return nil
}
