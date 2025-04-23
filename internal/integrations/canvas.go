package canvas

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"
	"online-course-platform/internal/db"
	"online-course-platform/internal/models"
)

type CanvasCourse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	CourseCode string `json:"course_code"`
	StartAt    string `json:"start_at"`
	EndAt      string `json:"end_at"`
}

func GetCanvasCourses() ([]CanvasCourse, error) {
	canvasURL := os.Getenv("CANVAS_API_URL")
	token := os.Getenv("CANVAS_API_TOKEN")
	if canvasURL == "" || token == "" {
		return nil, fmt.Errorf("CANVAS_API_URL or CANVAS_API_TOKEN not set")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/courses", canvasURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Canvas API error: %s", resp.Status)
	}

	var courses []CanvasCourse
	err = json.NewDecoder(resp.Body).Decode(&courses)
	if err != nil {
		return nil, err
	}

	return courses, nil
}

func ensureDefaultTeacher(db *gorm.DB) uint {
	var teacher models.User
	if err := db.Where("role = ?", "teacher").First(&teacher).Error; err == nil {
		return teacher.ID
	}

	// создаём если не найден
	newTeacher := models.User{
		Name:     "default_teacher",
		Email:    "teacher@narxoz.kz",
		Role:     "teacher",
		Password: "",
	}
	db.Create(&newTeacher)
	return newTeacher.ID
}

func ImportCanvasCoursesToDB() string {
	db.DB.AutoMigrate(&models.Course{}, &models.User{}, &models.Enrollment{})

	teacherID := ensureDefaultTeacher(db.DB)

	courses, err := GetCanvasCourses()
	if err != nil {
		log.Println("Ошибка получения курсов Canvas:", err)
		return "Ошибка подключения к Canvas."
	}

	count := 0
	for _, c := range courses {
		if c.Name == "" {
			continue
		}
		course := models.Course{
			Title:       c.Name,
			Description: fmt.Sprintf("Canvas ID: %d, Код курса: %s", c.ID, c.CourseCode),
			TeacherID:   teacherID,
		}
		err := db.DB.Create(&course).Error
		if err == nil {
			count++
			fmt.Printf("📥 Импортирован курс: %s\n", course.Title)
		} else {
			log.Println("❌ Ошибка сохранения курса:", err)
		}
	}
	return fmt.Sprintf("Импортировано %d курсов из Canvas", count)
}
