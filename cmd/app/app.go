package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"course-tracker/internal/auth"
	"course-tracker/internal/models"
	"course-tracker/internal/service"
	"course-tracker/internal/storage"
)

var errNotAuthenticated = errors.New("необходимо войти в аккаунт")

type App struct {
	dataDir  string
	accounts *auth.Service
	mu       sync.RWMutex
	user     *auth.User
	courses  *service.CourseService
}

func NewApp(dataDir string) (*App, error) {
	return &App{
		dataDir:  dataDir,
		accounts: auth.NewService(filepath.Join(dataDir, "users.json")),
	}, nil
}

func (a *App) GetSession() *auth.User {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.user == nil {
		return nil
	}
	user := *a.user
	return &user
}

func (a *App) Register(username, password string) (auth.User, error) {
	user, err := a.accounts.Register(username, password)
	if err != nil {
		return auth.User{}, err
	}
	if err := a.activateUser(user); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func (a *App) Login(username, password string) (auth.User, error) {
	user, err := a.accounts.Login(username, password)
	if err != nil {
		return auth.User{}, err
	}
	if err := a.activateUser(user); err != nil {
		return auth.User{}, err
	}
	return user, nil
}

func (a *App) Logout() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.user = nil
	a.courses = nil
}

func (a *App) activateUser(user auth.User) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	userHash := sha256.Sum256([]byte(user.Username))
	userDir := filepath.Join(a.dataDir, "accounts", hex.EncodeToString(userHash[:]))
	coursesPath := filepath.Join(userDir, "courses.json")
	if _, err := os.Stat(coursesPath); os.IsNotExist(err) {
		legacyPaths := []string{
			filepath.Join(a.dataDir, "accounts", user.Username, "courses.json"),
			filepath.Join(a.dataDir, "courses.json"),
		}
		for _, legacyPath := range legacyPaths {
			if _, err := os.Stat(legacyPath); err == nil {
				if err := os.MkdirAll(userDir, 0o700); err != nil {
					return err
				}
				if err := os.Rename(legacyPath, coursesPath); err != nil {
					return err
				}
				break
			} else if !os.IsNotExist(err) {
				return err
			}
		}
	} else if err != nil {
		return err
	}

	a.user = &user
	a.courses = service.NewCourseService(storage.NewJSONStore(coursesPath))
	return nil
}

func (a *App) GetCourses(username string) ([]models.Course, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	courses, err := a.authenticatedCourses(username)
	if err != nil {
		return nil, err
	}
	return courses.ListCourses()
}

func (a *App) GetStats(username string) (service.Stats, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	courses, err := a.authenticatedCourses(username)
	if err != nil {
		return service.Stats{}, err
	}
	return courses.GetStats()
}

func (a *App) AddCourse(title, description, category, level string, duration int, username string) (models.Course, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	courses, err := a.authenticatedCourses(username)
	if err != nil {
		return models.Course{}, err
	}
	return courses.AddCourse(title, description, category, level, duration)
}

func (a *App) UpdateCompletion(id int, completed bool, username string) (models.Course, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	courses, err := a.authenticatedCourses(username)
	if err != nil {
		return models.Course{}, err
	}
	return courses.UpdateCompletion(id, completed)
}

func (a *App) DeleteCourse(id int, username string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	courses, err := a.authenticatedCourses(username)
	if err != nil {
		return err
	}
	return courses.DeleteCourse(id)
}

func (a *App) authenticatedCourses(username string) (*service.CourseService, error) {
	if a.user == nil || a.courses == nil || a.user.Username != username {
		return nil, errNotAuthenticated
	}
	return a.courses, nil
}
