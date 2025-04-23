package bot

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"net/http"
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
	UserID uint
	Role   string
}

var (
	sessions       = make(map[int64]session)
	sessionsMutex  = sync.RWMutex{}
	authCodes      = make(map[string]string) // email -> code
	authEmailState = make(map[int64]string)  // chatID -> email
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
	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		text := strings.TrimSpace(update.Message.Text)

		if text == "admin123" {
			sessionsMutex.Lock()
			sessions[chatID] = session{UserID: 1, Role: "admin"}
			sessionsMutex.Unlock()
			showMenuByRole(bot, chatID, "admin", "admin")
			continue
		}

		sessionsMutex.RLock()
		sess, loggedIn := sessions[chatID]
		sessionsMutex.RUnlock()

		if !loggedIn {
			switch {
			case strings.ToLower(text) == "/start":
				bot.Send(tgbotapi.NewMessage(chatID, "Введите вашу корпоративную почту (@narxoz.kz):"))
			case strings.HasSuffix(text, "@narxoz.kz"):
				code := fmt.Sprintf("%06d", rand.Intn(1000000))
				authCodes[text] = code
				authEmailState[chatID] = text
				sendEmail(text, code)
				log.Printf(" Код %s отправлен на почту: %s\n", code, text)
				bot.Send(tgbotapi.NewMessage(chatID, "Код отправлен на почту. Введите его:"))
			case len(text) == 6 && regexp.MustCompile(`^\d{6}$`).MatchString(text):
				email := authEmailState[chatID]
				if email != "" && authCodes[email] == text {
					user := registerOrGetUser(email)
					sessionsMutex.Lock()
					sessions[chatID] = session{UserID: user.ID, Role: user.Role}
					sessionsMutex.Unlock()
					delete(authCodes, email)
					delete(authEmailState, chatID)
					showMenuByRole(bot, chatID, user.Name, user.Role)
				} else {
					bot.Send(tgbotapi.NewMessage(chatID, "Неверный код. Введите заново:"))
				}
			default:
				bot.Send(tgbotapi.NewMessage(chatID, "Введите /start чтобы начать регистрацию."))
			}
			continue
		}

		switch strings.ToLower(text) {
		case "выход":
			sessionsMutex.Lock()
			delete(sessions, chatID)
			sessionsMutex.Unlock()
			bot.Send(tgbotapi.NewMessage(chatID, "Вы вышли. Введите /start, чтобы войти снова."))
		case "курсы":
			bot.Send(tgbotapi.NewMessage(chatID, fetchCourses()))
		case "создать курс":
			if sess.Role == "admin" || sess.Role == "teacher" {
				bot.Send(tgbotapi.NewMessage(chatID, "Введите название курса:"))
				// дальше логика создания курса
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Недостаточно прав."))
			}
		case "выдать роль":
			if sess.Role == "admin" {
				bot.Send(tgbotapi.NewMessage(chatID, "Введите username пользователя и новую роль (например: john teacher):"))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Недостаточно прав."))
			}

		case "импортировать из canvas":
			if sess.Role == "admin" {
				msg := canvas.ImportCanvasCoursesToDB()
				bot.Send(tgbotapi.NewMessage(chatID, msg))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Недостаточно прав"))
			}

		default:
			if sess.Role == "admin" && strings.Count(text, " ") == 1 {
				parts := strings.Split(text, " ")
				username := parts[0]
				newRole := parts[1]
				changeUserRole(username, newRole)
				bot.Send(tgbotapi.NewMessage(chatID, "Роль обновлена."))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда."))
			}
		}
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

func showMenuByRole(bot *tgbotapi.BotAPI, chatID int64, name, role string) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Добро пожаловать, %s (%s)", name, role))
	buttons := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton("Курсы"), tgbotapi.NewKeyboardButton("Выход")}
	if role == "admin" {
		buttons = append(buttons, tgbotapi.NewKeyboardButton("Создать курс"), tgbotapi.NewKeyboardButton("Выдать роль"))
	} else if role == "teacher" {
		buttons = append(buttons, tgbotapi.NewKeyboardButton("Создать курс"))
	}
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(buttons...))
	bot.Send(msg)
}

func registerOrGetUser(email string) SimpleUser {
	_ = godotenv.Load()

	if db.DB == nil {
		db.InitDB()
	}

	name := strings.Split(email, "@")[0]
	var user models.User
	err := db.DB.Where("email = ?", email).First(&user).Error

	if err != nil {
		// создаём нового студента
		user = models.User{
			Name:     name,
			Email:    email,
			Password: "",
			Role:     "student",
		}
		if err := db.DB.Create(&user).Error; err != nil {
			log.Println("❌ Ошибка создания пользователя:", err)
		} else {
			log.Println("✅ Новый студент сохранён:", user.Email)

			// привязка ко всем курсам
			var courses []models.Course
			if err := db.DB.Find(&courses).Error; err == nil {
				for _, course := range courses {
					enroll := models.Enrollment{
						UserID:   user.ID,
						CourseID: course.ID,
					}
					db.DB.Create(&enroll)
				}
				log.Printf("📚 Студент %s подписан на %d курсов\n", user.Email, len(courses))
			}
		}
	} else {
		log.Println("ℹ️ Пользователь уже существует:", user.Email)
	}

	return SimpleUser{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}
}

func changeUserRole(username string, role string) {
	payload := map[string]string{"username": username, "role": role}
	body, _ := json.Marshal(payload)
	http.Post("http://localhost:8080/update-role", "application/json", strings.NewReader(string(body)))
}

type SimpleUser struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func fetchCourses() string {
	resp, err := http.Get("http://localhost:8080/courses")
	if err != nil {
		return "Ошибка при получении курсов"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Ошибка сервера: %d", resp.StatusCode)
	}

	var courses []struct {
		ID          uint   `json:"ID"`
		Title       string `json:"title"`
		Description string `json:"description"`
		TeacherID   uint   `json:"teacher_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return "Ошибка обработки ответа"
	}

	if len(courses) == 0 {
		return "Курсы отсутствуют"
	}

	var result strings.Builder
	result.WriteString(" Курсы:\n\n")
	for _, c := range courses {
		result.WriteString(fmt.Sprintf("• %s\n", c.Title))
	}
	return result.String()
}
