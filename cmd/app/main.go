package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"diploma/course-tracker/internal/service"
	"diploma/course-tracker/internal/storage"
)

const defaultDataFile = "data/courses.json"

func main() {
	dataPath := filepath.Join(".", defaultDataFile)
	store := storage.NewJSONStore(dataPath)
	svc := service.NewCourseService(store)

	reader := bufio.NewReader(os.Stdin)
	for {
		printMenu()
		choice, err := readInt(reader, "Выберите пункт меню")
		if err != nil {
			fmt.Println("Ошибка: введите число.")
			continue
		}

		switch choice {
		case 1:
			addCourseInteractive(svc, reader)
		case 2:
			listCoursesInteractive(svc)
		case 3:
			completeCourseInteractive(svc, reader)
		case 4:
			deleteCourseInteractive(svc, reader)
		case 5:
			statsInteractive(svc)
		case 0:
			fmt.Println("До свидания!")
			return
		default:
			fmt.Println("Такого пункта нет. Попробуйте ещё раз.")
		}
		fmt.Println()
	}
}

func printMenu() {
	fmt.Println("==== Трекер курсов ====")
	fmt.Println("1. Добавить курс")
	fmt.Println("2. Показать список курсов")
	fmt.Println("3. Отметить курс завершённым")
	fmt.Println("4. Удалить курс")
	fmt.Println("5. Показать статистику")
	fmt.Println("0. Выход")
}

func addCourseInteractive(svc *service.CourseService, reader *bufio.Reader) {
	fmt.Println("\nДобавление нового курса")
	title := readString(reader, "Введите название курса: ")
	description := readString(reader, "Введите описание: ")
	category := readString(reader, "Введите категорию: ")
	level := readString(reader, "Введите уровень (Новичок/Средний/Продвинутый): ")
	duration, err := readInt(reader, "Введите длительность в часах: ")
	if err != nil {
		fmt.Println("Ошибка: длительность должна быть числом.")
		return
	}

	course, err := svc.AddCourse(title, description, category, level, duration)
	if err != nil {
		fmt.Println("Не удалось добавить курс:", err)
		return
	}

	fmt.Printf("Курс добавлен: ID=%d, Название=%s\n", course.ID, course.Title)
}

func listCoursesInteractive(svc *service.CourseService) {
	courses, err := svc.ListCourses()
	if err != nil {
		fmt.Println("Ошибка при загрузке курсов:", err)
		return
	}

	if len(courses) == 0 {
		fmt.Println("Список курсов пуст.")
		return
	}

	fmt.Println("\nСписок курсов:")
	for _, course := range courses {
		status := "В процессе"
		if course.Completed {
			status = "Завершён"
		}
		fmt.Printf("ID=%d | %s | %s | %s | %dh | %s\n", course.ID, course.Title, course.Category, course.Level, course.Duration, status)
	}
}

func completeCourseInteractive(svc *service.CourseService, reader *bufio.Reader) {
	courses, err := svc.ListCourses()
	if err != nil {
		fmt.Println("Ошибка при загрузке курсов:", err)
		return
	}
	if len(courses) == 0 {
		fmt.Println("Нет курсов для отметки.")
		return
	}

	listCoursesInteractive(svc)
	id, err := readInt(reader, "Введите ID курса для отметки как завершённого: ")
	if err != nil {
		fmt.Println("Ошибка: ID должен быть числом.")
		return
	}

	course, err := svc.UpdateCompletion(id, true)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Printf("Курс отмечен как завершённый: %s\n", course.Title)
}

func deleteCourseInteractive(svc *service.CourseService, reader *bufio.Reader) {
	courses, err := svc.ListCourses()
	if err != nil {
		fmt.Println("Ошибка при загрузке курсов:", err)
		return
	}
	if len(courses) == 0 {
		fmt.Println("Нет курсов для удаления.")
		return
	}

	listCoursesInteractive(svc)
	id, err := readInt(reader, "Введите ID курса для удаления: ")
	if err != nil {
		fmt.Println("Ошибка: ID должен быть числом.")
		return
	}

	if err = svc.DeleteCourse(id); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Printf("Курс с ID=%d удалён.\n", id)
}

func statsInteractive(svc *service.CourseService) {
	stats, err := svc.GetStats()
	if err != nil {
		fmt.Println("Ошибка при расчёте статистики:", err)
		return
	}

	fmt.Println("\nСтатистика:")
	fmt.Printf("Всего курсов: %d\n", stats.Total)
	fmt.Printf("Завершено: %d\n", stats.Completed)
	fmt.Printf("В процессе: %d\n", stats.InProgress)
}

func readString(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	value, _ := reader.ReadString('\n')
	return strings.TrimSpace(value)
}

func readInt(reader *bufio.Reader, prompt string) (int, error) {
	fmt.Print(prompt)
	value, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	value = strings.TrimSpace(value)
	return strconv.Atoi(value)
}
