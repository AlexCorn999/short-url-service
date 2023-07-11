package store

var (
	IDStorage = 1
)

type Storage struct {
	Data map[string]string
}

// NewStorage ...
func NewStorage() *Storage {
	return &Storage{Data: make(map[string]string)}
}
