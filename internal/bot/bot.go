package bot

import (
	"fmt"
	"log"
	"math/rand"
	"net/smtp"
	"online-course-platform/internal/db"
	"online-course-platform/internal/integrations"
	"online-course-platform/internal/models"
	"os"
	"regexp"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type session struct {
	UserID       uint
	Role         string
	State        string
	TempCourseID uint
}

var (
	sessions       = make(map[int64]session)
	sessionsMutex  = sync.RWMutex{}
	authCodes      = make(map[string]string)
	authEmailState = make(map[int64]string)
)

func StartBot() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Ошибка инициализации бота:", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Ошибка получения обновлений:", err)
	}

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update.Message)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	text := strings.ToLower(strings.TrimSpace(message.Text))

	sessionsMutex.RLock()
	sess, loggedIn := sessions[chatID]
	sessionsMutex.RUnlock()

	if !loggedIn {
		handleAuth(bot, chatID, text)
		return
	}

	switch sess.State {
	case "choosing_course":
		handleCourseSelection(bot, chatID, text)
		return
	case "assigning_teacher":
		handleAssignTeacher(bot, chatID, text)
		return
	case "changing_role":
		handleChangeRole(bot, chatID, text)
		return
	}

	switch strings.ToLower(text) {
	case "курсы":
		showCourses(bot, chatID)
	case "записаться на курсы":
		startCourseEnrollment(bot, chatID)
	case "профиль":
		showProfile(bot, chatID)
	case "мои курсы":
		showUserCourses(bot, chatID)
	case "назначить роль":
		if sess.Role == "admin" {
			sess.State = "changing_role"
			sessions[chatID] = sess
			bot.Send(tgbotapi.NewMessage(chatID, "Введите: username новая_роль"))
		}

	case "импортировать курсы":
		if sess.Role == "admin" {
			importCoursesFromCanvas(bot, chatID)
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Недостаточно прав."))
		}
	case "выдать учителя на курс":
		if sess.Role == "admin" {
			startAssignTeacher(bot, chatID)
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Недостаточно прав."))
		}
	case "назад":
		resetState(chatID)
		showMainMenu(bot, chatID, sess.Role)
	case "выход":
		sessionsMutex.Lock()
		delete(sessions, chatID)
		sessionsMutex.Unlock()
		bot.Send(tgbotapi.NewMessage(chatID, "Вы вышли. Введите /start, чтобы войти снова."))
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда."))
	}
}

func handleAuth(bot *tgbotapi.BotAPI, chatID int64, text string) {
	switch {
	case strings.ToLower(text) == "/start":
		bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу корпоративную почту (@narxoz.kz):"))
	case strings.HasSuffix(text, "@narxoz.kz"):
		code := fmt.Sprintf("%06d", rand.Intn(1000000))
		authCodes[text] = code
		authEmailState[chatID] = text
		sendEmail(text, code)
		log.Printf("Код %s отправлен на почту: %s\n", code, text)
		bot.Send(tgbotapi.NewMessage(chatID, "Код отправлен на почту. Введите его:"))
	case len(text) == 6 && regexp.MustCompile(`^\d{6}$`).MatchString(text):
		email := authEmailState[chatID]
		if email != "" && authCodes[email] == text {
			user := registerOrGetUser(email)
			sessionsMutex.Lock()
			sessions[chatID] = session{UserID: user.ID, Role: user.Role, State: ""}
			sessionsMutex.Unlock()
			delete(authCodes, email)
			delete(authEmailState, chatID)
			showMainMenu(bot, chatID, user.Role)
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "Неверный код. Введите заново:"))
		}
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Введите /start чтобы начать регистрацию."))
	}
}

func startCourseEnrollment(bot *tgbotapi.BotAPI, chatID int64) {
	courses := fetchCourses()
	if len(courses) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Нет доступных курсов."))
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите курс для записи:")
	var rows [][]tgbotapi.KeyboardButton
	for _, course := range courses {
		row := tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(fmt.Sprintf("%s (%d мест)", course.Title, course.SeatsAvailable)))
		rows = append(rows, row)
	}
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Назад")))
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(rows...)
	bot.Send(msg)

	sessionsMutex.Lock()
	if sess, ok := sessions[chatID]; ok {
		sess.State = "choosing_course"
		sessions[chatID] = sess
	}
	sessionsMutex.Unlock()
}

func handleCourseSelection(bot *tgbotapi.BotAPI, chatID int64, text string) {
	courses := fetchCourses()

	if strings.ToLower(text) == "назад" {
		resetState(chatID)
		showMainMenu(bot, chatID, sessions[chatID].Role)
		return
	}

	for _, course := range courses {
		titleWithSeats := fmt.Sprintf("%s (%d мест)", course.Title, course.SeatsAvailable)
		if strings.ToLower(text) == strings.ToLower(titleWithSeats) {
			enrollUserToCourse(chatID, course.ID)
			bot.Send(tgbotapi.NewMessage(chatID, "Успешно зарегистрирован на курс!"))
			resetState(chatID)
			showMainMenu(bot, chatID, sessions[chatID].Role)
			return
		}
	}

	bot.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, выберите курс из списка кнопок ниже или нажмите «Назад»."))
}

func resetState(chatID int64) {
	sessionsMutex.Lock()
	if sess, ok := sessions[chatID]; ok {
		sess.State = ""
		sessions[chatID] = sess
	}
	sessionsMutex.Unlock()
}

func showMainMenu(bot *tgbotapi.BotAPI, chatID int64, role string) {
	var rows [][]tgbotapi.KeyboardButton

	switch role {
	case "admin":
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Импортировать курсы"),
			tgbotapi.NewKeyboardButton("Выдать учителя на курс"),
		))
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Назначить роль"),
			tgbotapi.NewKeyboardButton("Профиль"),
			tgbotapi.NewKeyboardButton("Выход"),
		))
	case "teacher":
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Курсы"),
			tgbotapi.NewKeyboardButton("Профиль"),
			tgbotapi.NewKeyboardButton("Выход"),
		))
	case "student":
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Курсы"),
			tgbotapi.NewKeyboardButton("Записаться на курсы"),
		))
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Мои курсы"),
			tgbotapi.NewKeyboardButton("Профиль"),
			tgbotapi.NewKeyboardButton("Выход"),
		))

	}

	msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(rows...)
	bot.Send(msg)
}

func importCoursesFromCanvas(bot *tgbotapi.BotAPI, chatID int64) {
	msg := canvas.ImportCanvasCoursesToDB()
	bot.Send(tgbotapi.NewMessage(chatID, msg))
}

func startAssignTeacher(bot *tgbotapi.BotAPI, chatID int64) {
	bot.Send(tgbotapi.NewMessage(chatID, "Введите через пробел: название_курса username_учителя"))
	sessionsMutex.Lock()
	if sess, ok := sessions[chatID]; ok {
		sess.State = "assigning_teacher"
		sessions[chatID] = sess
	}
	sessionsMutex.Unlock()
}
func showCourses(bot *tgbotapi.BotAPI, chatID int64) {
	courses := fetchCourses()
	if len(courses) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "📭 Курсы отсутствуют."))
		return
	}

	var sb strings.Builder
	sb.WriteString("📋 Список курсов:\n\n")
	for _, c := range courses {
		sb.WriteString(fmt.Sprintf("- %s (%d мест)\n", c.Title, c.SeatsAvailable))
	}
	bot.Send(tgbotapi.NewMessage(chatID, sb.String()))
}

func handleAssignTeacher(bot *tgbotapi.BotAPI, chatID int64, text string) {
	if strings.ToLower(text) == "назад" {
		resetState(chatID)
		showMainMenu(bot, chatID, sessions[chatID].Role)
		return
	}

	parts := strings.SplitN(text, " ", 2)
	if len(parts) != 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Формат: название_курса username_учителя"))
		return
	}

	courseTitle := parts[0]
	teacherName := parts[1]

	var course models.Course
	if err := db.DB.Where("LOWER(title) = ?", strings.ToLower(courseTitle)).First(&course).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Курс не найден."))
		return
	}

	var teacher models.User
	if err := db.DB.Where("name = ? AND role = ?", teacherName, "teacher").First(&teacher).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Учитель не найден."))
		return
	}

	course.TeacherID = teacher.ID
	db.DB.Save(&course)

	bot.Send(tgbotapi.NewMessage(chatID, "Учитель успешно назначен на курс!"))
	resetState(chatID)
	showMainMenu(bot, chatID, sessions[chatID].Role)
}

func showUserCourses(bot *tgbotapi.BotAPI, chatID int64) {
	sess := sessions[chatID]

	var enrollments []models.Enrollment
	db.DB.Where("user_id = ?", sess.UserID).Find(&enrollments)

	if len(enrollments) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Вы пока не записаны ни на один курс."))
		return
	}

	var courseTitles []string
	for _, e := range enrollments {
		var course models.Course
		db.DB.First(&course, e.CourseID)
		courseTitles = append(courseTitles, fmt.Sprintf("- %s", course.Title))
	}

	bot.Send(tgbotapi.NewMessage(chatID, " Ваши курсы:\n"+strings.Join(courseTitles, "\n")))
}

func showProfile(bot *tgbotapi.BotAPI, chatID int64) {
	sess := sessions[chatID]

	var user models.User
	db.DB.First(&user, sess.UserID)

	var enrollments []models.Enrollment
	db.DB.Where("user_id = ?", sess.UserID).Find(&enrollments)

	var courseTitles []string
	for _, enrollment := range enrollments {
		var course models.Course
		if err := db.DB.First(&course, enrollment.CourseID).Error; err == nil {
			courseTitles = append(courseTitles, course.Title)
		}
	}

	profile := fmt.Sprintf(
		"Имя: %s\nEmail: %s\nРоль: %s\nКурсы:\n- %s",
		user.Name,
		user.Email,
		user.Role,
		strings.Join(courseTitles, "\n- "),
	)

	if len(courseTitles) == 0 {
		profile += "-"
	}

	bot.Send(tgbotapi.NewMessage(chatID, profile))
}

func enrollUserToCourse(chatID int64, courseID uint) {
	sess := sessions[chatID]

	var existing models.Enrollment
	if err := db.DB.Where("user_id = ? AND course_id = ?", sess.UserID, courseID).First(&existing).Error; err == nil {
		return
	}

	enroll := models.Enrollment{
		UserID:   sess.UserID,
		CourseID: courseID,
	}
	if err := db.DB.Create(&enroll).Error; err != nil {
		log.Println("Ошибка сохранения enrollment:", err)
	} else {
		log.Println("Enrollment создан:", enroll.ID)
	}

	var course models.Course
	if err := db.DB.First(&course, courseID).Error; err == nil && course.SeatsAvailable > 0 {
		course.SeatsAvailable--
		db.DB.Save(&course)
	}
}

func sendEmail(to string, code string) {
	from := os.Getenv("SMTP_EMAIL")
	pass := os.Getenv("SMTP_PASS")
	host := "smtp.gmail.com"
	port := "587"
	auth := smtp.PlainAuth("", from, pass, host)
	msg := []byte("Subject: Код подтверждения\r\n\r\nВаш код: " + code)
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, msg)
	if err != nil {
		log.Println("Ошибка отправки письма:", err)
	}
}

func fetchCourses() []models.Course {
	var courses []models.Course
	if err := db.DB.Find(&courses).Error; err != nil {
		log.Println("Ошибка получения курсов:", err)
		return nil
	}
	return courses
}

func registerOrGetUser(email string) models.User {
	if db.DB == nil {
		db.InitDB()
	}
	name := strings.Split(email, "@")[0]
	var user models.User
	err := db.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		role := "student"
		if email == "admin@narxoz.kz" {
			role = "admin"
		}
		user = models.User{
			Name:     name,
			Email:    email,
			Password: "",
			Role:     role,
		}
		if err := db.DB.Create(&user).Error; err != nil {
			log.Println("Ошибка создания пользователя:", err)
		} else {
			log.Printf("Пользователь создан: %s с ролью %s", user.Email, role)
		}
	} else {
		log.Println("Пользователь уже существует:", user.Email)
	}
	return user
}

func handleChangeRole(bot *tgbotapi.BotAPI, chatID int64, text string) {
	if strings.ToLower(text) == "назад" {
		resetState(chatID)
		showMainMenu(bot, chatID, sessions[chatID].Role)
		return
	}

	parts := strings.SplitN(text, " ", 2)
	if len(parts) != 2 {
		bot.Send(tgbotapi.NewMessage(chatID, "Формат: username роль (student/teacher/admin)"))
		return
	}

	username := parts[0]
	newRole := parts[1]

	if newRole != "student" && newRole != "teacher" && newRole != "admin" {
		bot.Send(tgbotapi.NewMessage(chatID, "Неверная роль. Доступно: student, teacher, admin"))
		return
	}

	var user models.User
	if err := db.DB.Where("name = ?", username).First(&user).Error; err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Пользователь не найден."))
		return
	}

	user.Role = newRole
	db.DB.Save(&user)

	bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("%s теперь %s", username, newRole)))
	resetState(chatID)
	showMainMenu(bot, chatID, sessions[chatID].Role)
}
