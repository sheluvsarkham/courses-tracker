package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestAccountsKeepCoursesSeparate(t *testing.T) {
	app, err := NewApp(t.TempDir())
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	if _, err := app.AddCourse("Курс", "Описание", "Разработка", "Новичок", 10, "first_user"); err == nil {
		t.Fatal("expected unauthenticated course operation to fail")
	}

	if _, err := app.Register("first_user", "password-one"); err != nil {
		t.Fatalf("Register first user returned error: %v", err)
	}
	if _, err := app.AddCourse("Go", "Основы языка", "Разработка", "Новичок", 12, "first_user"); err != nil {
		t.Fatalf("AddCourse returned error: %v", err)
	}
	app.Logout()

	if _, err := app.Register("second_user", "password-two"); err != nil {
		t.Fatalf("Register second user returned error: %v", err)
	}
	if _, err := app.GetCourses("first_user"); err == nil {
		t.Fatal("expected stale user request to be rejected")
	}
	if err := app.DeleteCourse(1, "first_user"); err == nil {
		t.Fatal("expected stale delete request to be rejected")
	}
	courses, err := app.GetCourses("second_user")
	if err != nil {
		t.Fatalf("GetCourses for second user returned error: %v", err)
	}
	if len(courses) != 0 {
		t.Fatalf("second user should not see first user's courses: %+v", courses)
	}
	app.Logout()

	if _, err := app.Login("FIRST_USER", "password-one"); err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	courses, err = app.GetCourses("first_user")
	if err != nil {
		t.Fatalf("GetCourses after login returned error: %v", err)
	}
	if len(courses) != 1 || courses[0].Title != "Go" {
		t.Fatalf("first user's course was not preserved: %+v", courses)
	}
	if courses[0].Completed {
		t.Fatalf("new course should be in progress: %+v", courses[0])
	}
}

func TestFirstAccountReceivesLegacyCourses(t *testing.T) {
	dataDir := t.TempDir()
	legacyPath := filepath.Join(dataDir, "courses.json")
	legacyData := `[{"id":7,"title":"Старый курс","description":"Описание","category":"Разработка","level":"Новичок","duration":8,"completed":false}]`
	if err := os.WriteFile(legacyPath, []byte(legacyData), 0o600); err != nil {
		t.Fatalf("failed to write legacy courses: %v", err)
	}

	app, err := NewApp(dataDir)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if _, err := app.Register("first_user", "password-one"); err != nil {
		t.Fatalf("Register first user returned error: %v", err)
	}
	courses, err := app.GetCourses("first_user")
	if err != nil {
		t.Fatalf("GetCourses returned error: %v", err)
	}
	if len(courses) != 1 || courses[0].Title != "Старый курс" {
		t.Fatalf("legacy course was not migrated: %+v", courses)
	}
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("expected legacy file to be moved, got error %v", err)
	}

	app.Logout()
	if _, err := app.Register("second_user", "password-two"); err != nil {
		t.Fatalf("Register second user returned error: %v", err)
	}
	courses, err = app.GetCourses("second_user")
	if err != nil {
		t.Fatalf("GetCourses for second user returned error: %v", err)
	}
	if len(courses) != 0 {
		t.Fatalf("second user should not receive migrated courses: %+v", courses)
	}
}

func TestExistingAccountDirectoryMigratesToHashedPath(t *testing.T) {
	dataDir := t.TempDir()
	legacyPath := filepath.Join(dataDir, "accounts", "legacy_user", "courses.json")
	legacyData := `[{"id":4,"title":"Сохранённый курс","description":"Описание","category":"Разработка","level":"Средний","duration":6,"completed":true}]`
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o700); err != nil {
		t.Fatalf("failed to create legacy account directory: %v", err)
	}
	if err := os.WriteFile(legacyPath, []byte(legacyData), 0o600); err != nil {
		t.Fatalf("failed to write legacy account courses: %v", err)
	}

	app, err := NewApp(dataDir)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if _, err := app.Register("legacy_user", "password-one"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	courses, err := app.GetCourses("legacy_user")
	if err != nil {
		t.Fatalf("GetCourses returned error: %v", err)
	}
	if len(courses) != 1 || courses[0].Title != "Сохранённый курс" || !courses[0].Completed {
		t.Fatalf("legacy account data was not migrated: %+v", courses)
	}
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("expected old account directory to be moved, got error %v", err)
	}
}

func TestConcurrentCourseWritesAreSerialized(t *testing.T) {
	app, err := NewApp(t.TempDir())
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if _, err := app.Register("con", "password-one"); err != nil {
		t.Fatalf("Register reserved Windows name returned error: %v", err)
	}

	const courseCount = 24
	var waitGroup sync.WaitGroup
	errorsFound := make(chan error, courseCount)
	for index := 0; index < courseCount; index++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			_, err := app.AddCourse(fmt.Sprintf("Course %d", index), "Description", "Category", "Beginner", 1, "con")
			if err != nil {
				errorsFound <- err
			}
		}(index)
	}
	waitGroup.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Errorf("concurrent AddCourse returned error: %v", err)
	}

	courses, err := app.GetCourses("con")
	if err != nil {
		t.Fatalf("GetCourses returned error: %v", err)
	}
	if len(courses) != courseCount {
		t.Fatalf("expected %d courses after concurrent writes, got %d", courseCount, len(courses))
	}
}
