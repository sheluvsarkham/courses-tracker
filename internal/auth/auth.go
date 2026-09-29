package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsername    = errors.New("имя пользователя: 3–32 символа, латинские буквы, цифры, точка, дефис или подчёркивание")
	ErrInvalidPassword    = errors.New("пароль должен содержать от 8 до 72 байт")
	ErrUsernameTaken      = errors.New("это имя пользователя уже занято")
	ErrInvalidCredentials = errors.New("неверное имя пользователя или пароль")
)

type User struct {
	Username string `json:"username"`
}

type userRecord struct {
	PasswordHash string `json:"passwordHash"`
}

type Service struct {
	path string
	mu   sync.Mutex
}

func NewService(path string) *Service {
	return &Service{path: path}
}

func (s *Service) Register(username, password string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	username = normalizeUsername(username)
	if !validUsername(username) {
		return User{}, ErrInvalidUsername
	}
	if len(password) < 8 || len(password) > 72 {
		return User{}, ErrInvalidPassword
	}

	users, err := s.readUsers()
	if err != nil {
		return User{}, err
	}
	if _, exists := users[username]; exists {
		return User{}, ErrUsernameTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	users[username] = userRecord{PasswordHash: string(hash)}
	if err := s.writeUsers(users); err != nil {
		return User{}, err
	}

	return User{Username: username}, nil
}

func (s *Service) Login(username, password string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	username = normalizeUsername(username)
	if !validUsername(username) || len(password) > 72 {
		return User{}, ErrInvalidCredentials
	}

	users, err := s.readUsers()
	if err != nil {
		return User{}, err
	}
	user, exists := users[username]
	if !exists || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return User{}, ErrInvalidCredentials
	}

	return User{Username: username}, nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func validUsername(username string) bool {
	if len(username) < 3 || len(username) > 32 {
		return false
	}
	for _, char := range username {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '.' || char == '_' || char == '-' {
			continue
		}
		return false
	}
	return true
}

func (s *Service) readUsers() (map[string]userRecord, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]userRecord{}, nil
		}
		return nil, err
	}

	users := map[string]userRecord{}
	if len(data) == 0 {
		return nil, errors.New("accounts file is empty")
	}
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	if users == nil {
		return nil, errors.New("accounts file must contain a JSON object")
	}
	return users, nil
}

func (s *Service) writeUsers(users map[string]userRecord) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
