package service

import (
	"fmt"
	"strings"

	"course-tracker/internal/models"
	"course-tracker/internal/storage"
)

// Stats contains summary information about courses.
type Stats struct {
	Total      int
	Completed  int
	InProgress int
}

type CourseService struct {
	store storage.Store
}

func NewCourseService(store storage.Store) *CourseService {
	return &CourseService{store: store}
}

func (s *CourseService) AddCourse(title, description, category, level string, duration int) (models.Course, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	category = strings.TrimSpace(category)
	level = strings.TrimSpace(level)

	if title == "" {
		return models.Course{}, fmt.Errorf("title is required")
	}
	if description == "" {
		return models.Course{}, fmt.Errorf("description is required")
	}
	if category == "" {
		return models.Course{}, fmt.Errorf("category is required")
	}
	if level == "" {
		return models.Course{}, fmt.Errorf("level is required")
	}
	if duration <= 0 {
		return models.Course{}, fmt.Errorf("duration must be greater than zero")
	}

	courses, err := s.store.ListCourses()
	if err != nil {
		return models.Course{}, err
	}

	nextID := 1
	for _, course := range courses {
		if course.ID >= nextID {
			nextID = course.ID + 1
		}
	}

	course := models.Course{
		ID:          nextID,
		Title:       title,
		Description: description,
		Category:    category,
		Level:       level,
		Duration:    duration,
		Completed:   false,
	}

	return s.store.AddCourse(course)
}

func (s *CourseService) ListCourses() ([]models.Course, error) {
	return s.store.ListCourses()
}

func (s *CourseService) UpdateCompletion(id int, completed bool) (models.Course, error) {
	courses, err := s.store.ListCourses()
	if err != nil {
		return models.Course{}, err
	}

	for i, course := range courses {
		if course.ID == id {
			courses[i].Completed = completed
			return s.store.UpdateCourse(courses[i])
		}
	}

	return models.Course{}, fmt.Errorf("course with id %d not found", id)
}

func (s *CourseService) DeleteCourse(id int) error {
	return s.store.DeleteCourse(id)
}

func (s *CourseService) GetStats() (Stats, error) {
	courses, err := s.store.ListCourses()
	if err != nil {
		return Stats{}, err
	}

	stats := Stats{Total: len(courses)}
	for _, course := range courses {
		if course.Completed {
			stats.Completed++
			continue
		}
		stats.InProgress++
	}

	return stats, nil
}
