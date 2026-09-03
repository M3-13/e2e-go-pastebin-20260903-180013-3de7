package store

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Paste struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PasteMeta struct {
	ID        string    `json:"id"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Store struct {
	mu    sync.Mutex
	items map[string]Paste
}

func New() *Store {
	return &Store{items: make(map[string]Paste)}
}

func (s *Store) Add(content, language string, expiresIn time.Duration) (PasteMeta, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return PasteMeta{}, err
	}
	id := hex.EncodeToString(b[:])

	now := time.Now()
	p := Paste{
		ID:        id,
		Content:   content,
		Language:  language,
		CreatedAt: now,
		ExpiresAt: now.Add(expiresIn),
	}

	s.mu.Lock()
	s.items[id] = p
	s.mu.Unlock()

	return PasteMeta{
		ID:        p.ID,
		Language:  p.Language,
		CreatedAt: p.CreatedAt,
		ExpiresAt: p.ExpiresAt,
	}, nil
}

func (s *Store) Get(id string) (Paste, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.items[id]
	if !ok {
		return Paste{}, false
	}
	if time.Now().After(p.ExpiresAt) {
		delete(s.items, id)
		return Paste{}, false
	}
	return p, true
}

func (s *Store) List() []PasteMeta {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	metas := make([]PasteMeta, 0, len(s.items))
	for id, p := range s.items {
		if now.After(p.ExpiresAt) {
			delete(s.items, id)
			continue
		}
		metas = append(metas, PasteMeta{
			ID:        p.ID,
			Language:  p.Language,
			CreatedAt: p.CreatedAt,
			ExpiresAt: p.ExpiresAt,
		})
	}
	return metas
}

func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}
