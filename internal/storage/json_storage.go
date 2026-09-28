package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"course-tracker/internal/models"
)

// Store describes the storage contract used by the service layer.
type Store interface {
	ListCourses() ([]models.Course, error)
	AddCourse(course models.Course) (models.Course, error)
	UpdateCourse(course models.Course) (models.Course, error)
	DeleteCourse(id int) error
}

type JSONStore struct {
	path string
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

func (s *JSONStore) ListCourses() ([]models.Course, error) {
	courses, err := s.readAll()
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func (s *JSONStore) AddCourse(course models.Course) (models.Course, error) {
	courses, err := s.readAll()
	if err != nil {
		return models.Course{}, err
	}

	courses = append(courses, course)
	if err = s.saveAll(courses); err != nil {
		return models.Course{}, err
	}

	return course, nil
}

func (s *JSONStore) UpdateCourse(course models.Course) (models.Course, error) {
	courses, err := s.readAll()
	if err != nil {
		return models.Course{}, err
	}

	updated := false
	for i, item := range courses {
		if item.ID == course.ID {
			courses[i] = course
			updated = true
			break
		}
	}

	if !updated {
		return models.Course{}, errors.New("course not found")
	}

	if err = s.saveAll(courses); err != nil {
		return models.Course{}, err
	}

	return course, nil
}

func (s *JSONStore) DeleteCourse(id int) error {
	courses, err := s.readAll()
	if err != nil {
		return err
	}

	index := -1
	for i, item := range courses {
		if item.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return errors.New("course not found")
	}

	courses = append(courses[:index], courses[index+1:]...)
	return s.saveAll(courses)
}

func (s *JSONStore) readAll() ([]models.Course, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Course{}, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return []models.Course{}, nil
	}

	var courses []models.Course
	if err = json.Unmarshal(data, &courses); err != nil {
		return nil, err
	}

	return courses, nil
}

func (s *JSONStore) saveAll(courses []models.Course) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(courses, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0o644)
}
