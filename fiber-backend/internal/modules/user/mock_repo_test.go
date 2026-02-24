package user

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

// MockRepo is an in-memory implementation of Repository for testing.
type MockRepo struct {
	mu     sync.RWMutex
	users  map[string]*User
	nextID int64
}

var _ Repository = (*MockRepo)(nil)

func NewMockRepo() *MockRepo {
	return &MockRepo{
		users:  make(map[string]*User),
		nextID: 1,
	}
}

func (m *MockRepo) Create(ctx context.Context, u *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u.ID = m.nextID
	m.nextID++
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	if u.Role == "" {
		u.Role = "user"
	}

	stored := *u
	m.users[u.Email] = &stored
	return nil
}

func (m *MockRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.users[email]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	copy := *u
	return &copy, nil
}

func (m *MockRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.ID == id {
			copy := *u
			return &copy, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *MockRepo) List(ctx context.Context) ([]User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, *u)
	}
	return out, nil
}

func (m *MockRepo) Update(ctx context.Context, id int64, u *User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.users {
		if existing.ID == id {
			existing.Name = u.Name
			existing.Email = u.Email
			existing.UpdatedAt = time.Now()
			return nil
		}
	}
	return pgx.ErrNoRows
}

func (m *MockRepo) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for email, u := range m.users {
		if u.ID == id {
			delete(m.users, email)
			return nil
		}
	}
	return pgx.ErrNoRows
}

// ---------- MockTokenRepo ----------

// MockTokenRepo is an in-memory implementation of TokenRepository for testing.
type MockTokenRepo struct {
	mu     sync.RWMutex
	tokens map[string]struct {
		UserID    int64
		ExpiresAt time.Time
	}
}

var _ TokenRepository = (*MockTokenRepo)(nil)

func NewMockTokenRepo() *MockTokenRepo {
	return &MockTokenRepo{
		tokens: make(map[string]struct {
			UserID    int64
			ExpiresAt time.Time
		}),
	}
}

func (m *MockTokenRepo) StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens[tokenHash] = struct {
		UserID    int64
		ExpiresAt time.Time
	}{UserID: userID, ExpiresAt: expiresAt}
	return nil
}

func (m *MockTokenRepo) FindRefreshToken(ctx context.Context, tokenHash string) (int64, time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, ok := m.tokens[tokenHash]
	if !ok {
		return 0, time.Time{}, pgx.ErrNoRows
	}
	return t.UserID, t.ExpiresAt, nil
}

func (m *MockTokenRepo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tokens, tokenHash)
	return nil
}

func (m *MockTokenRepo) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for hash, t := range m.tokens {
		if t.UserID == userID {
			delete(m.tokens, hash)
		}
	}
	return nil
}
