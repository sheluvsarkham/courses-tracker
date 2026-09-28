package service

import (
	"testing"

	"diploma/course-tracker/internal/models"
)

type testStore struct {
	courses []models.Course
}

func (s *testStore) ListCourses() ([]models.Course, error) {
	return s.courses, nil
}

func (s *testStore) AddCourse(course models.Course) (models.Course, error) {
	s.courses = append(s.courses, course)
	return course, nil
}

func (s *testStore) UpdateCourse(course models.Course) (models.Course, error) {
	for i, item := range s.courses {
		if item.ID == course.ID {
			s.courses[i] = course
			return course, nil
		}
	}
	return models.Course{}, nil
}

func (s *testStore) DeleteCourse(id int) error {
	for i, item := range s.courses {
		if item.ID == id {
			s.courses = append(s.courses[:i], s.courses[i+1:]...)
			return nil
		}
	}
	return nil
}

func TestCourseService_AddAndStats(t *testing.T) {
	service := NewCourseService(&testStore{})

	course, err := service.AddCourse("Go Fundamentals", "Основы Go", "Backend", "Beginner", 12)
	if err != nil {
		t.Fatalf("AddCourse returned error: %v", err)
	}
	if course.ID == 0 {
		t.Fatalf("expected generated ID")
	}

	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats returned error: %v", err)
	}
	if stats.Total != 1 || stats.Completed != 0 || stats.InProgress != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	updated, err := service.UpdateCompletion(course.ID, true)
	if err != nil {
		t.Fatalf("UpdateCompletion returned error: %v", err)
	}
	if !updated.Completed {
		t.Fatal("course should be marked as completed")
	}
}

func TestCourseService_AddCourseRejectsInvalidInput(t *testing.T) {
	service := NewCourseService(&testStore{})

	_, err := service.AddCourse("", "Описание", "Backend", "Beginner", 10)
	if err == nil {
		t.Fatal("expected validation error for empty title")
	}

	_, err = service.AddCourse("Go", "Описание", "Backend", "Beginner", 0)
	if err == nil {
		t.Fatal("expected validation error for zero duration")
	}
}
