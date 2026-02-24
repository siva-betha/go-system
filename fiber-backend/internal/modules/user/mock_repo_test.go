package user

import (
	"context"
	"errors"
	"time"
)

type MockRepo struct {
	Users map[string]*User
}

func NewMockRepo() *MockRepo {
	return &MockRepo{Users: make(map[string]*User)}
}

func (m *MockRepo) Create(ctx context.Context, u *User) error {
	u.ID = "test-uuid"
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	m.Users[u.ID] = u
	return nil
}

func (m *MockRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
	for _, u := range m.Users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockRepo) GetByID(ctx context.Context, id string) (*User, error) {
	u, ok := m.Users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *MockRepo) List(ctx context.Context) ([]User, error) {
	var out []User
	for _, u := range m.Users {
		out = append(out, *u)
	}
	return out, nil
}

func (m *MockRepo) Update(ctx context.Context, id string, u *User) error {
	if _, ok := m.Users[id]; !ok {
		return errors.New("not found")
	}
	m.Users[id] = u
	return nil
}

func (m *MockRepo) Delete(ctx context.Context, id string) error {
	delete(m.Users, id)
	return nil
}

func (m *MockRepo) GetRolesAndPermissions(ctx context.Context, userID string) ([]string, map[string][]string, error) {
	return []string{"admin"}, map[string][]string{"*": {"*"}}, nil
}

type MockTokenRepo struct {
	Tokens map[string]string
}

func NewMockTokenRepo() *MockTokenRepo {
	return &MockTokenRepo{Tokens: make(map[string]string)}
}

func (m *MockTokenRepo) StoreRefreshToken(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	m.Tokens[tokenHash] = userID
	return nil
}

func (m *MockTokenRepo) FindRefreshToken(ctx context.Context, tokenHash string) (string, time.Time, error) {
	uid, ok := m.Tokens[tokenHash]
	if !ok {
		return "", time.Time{}, errors.New("not found")
	}
	return uid, time.Now().Add(time.Hour), nil
}

func (m *MockTokenRepo) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	delete(m.Tokens, tokenHash)
	return nil
}
