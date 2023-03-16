package storage

type MemStorage struct {
	urls map[string]string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		urls: make(map[string]string),
	}
}

func (m *MemStorage) Add(id, value string) {
	m.urls[id] = value
}

func (m *MemStorage) Get(id string) (string, bool) {
	value, ok := m.urls[id]
	return value, ok
}
