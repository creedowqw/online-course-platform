package bot

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type authState struct {
	Step     int
	Username string
	UserID   uint
}

var authStates = make(map[int64]*authState)

type SimpleUser struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type userCreation struct {
	Step     int
	Role     string
	Username string
	Password string
}

var userCreateStates = make(map[int64]*userCreation)

type courseCreation struct {
	Step      int
	Title     string
	Desc      string
	TeacherID string
}

var courseCreateStates = make(map[int64]*courseCreation)

func StartBot() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Ошибка инициализации бота:", err)
	}

	log.Printf("Бот запущен: @%s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := strings.TrimSpace(update.Message.Text)

		if auth, ok := authStates[chatID]; ok {
			switch auth.Step {
			case 1:
				auth.Username = text
				auth.Step = 2
				bot.Send(tgbotapi.NewMessage(chatID, "Введите пароль:"))
				continue
			case 2:
				password := text
				user, role := checkUser(auth.Username, password)
				if user == nil {
					delete(authStates, chatID)
					bot.Send(tgbotapi.NewMessage(chatID, "Неверный логин или пароль. Попробуйте /start"))
					continue
				}
				delete(authStates, chatID)
				showMenuByRole(bot, chatID, user.Name, role)
				continue
			}
		}

		if state, ok := userCreateStates[chatID]; ok {
			switch state.Step {
			case 1:
				r := strings.ToLower(text)
				if r != "student" && r != "teacher" {
					bot.Send(tgbotapi.NewMessage(chatID, "Введите: student или teacher"))
					continue
				}
				state.Role = r
				state.Step = 2
				bot.Send(tgbotapi.NewMessage(chatID, "Введите username:"))
			case 2:
				state.Username = text
				state.Step = 3
				bot.Send(tgbotapi.NewMessage(chatID, "Введите пароль:"))
			case 3:
				state.Password = text
				sendUserToAPI(state)
				delete(userCreateStates, chatID)
				bot.Send(tgbotapi.NewMessage(chatID, "Пользователь создан!"))
			}
			continue
		}

		if state, ok := courseCreateStates[chatID]; ok {
			switch state.Step {
			case 1:
				state.Title = text
				state.Step = 2
				bot.Send(tgbotapi.NewMessage(chatID, "Введите описание курса:"))
			case 2:
				state.Desc = text
				state.Step = 3
				bot.Send(tgbotapi.NewMessage(chatID, "Введите ID преподавателя (или 0):"))
			case 3:
				state.TeacherID = text
				sendCourseToAPI(state)
				delete(courseCreateStates, chatID)
				bot.Send(tgbotapi.NewMessage(chatID, "Курс успешно создан!"))
			}
			continue
		}

		switch strings.ToLower(text) {
		case "/start":
			msg := tgbotapi.NewMessage(chatID, "Добро пожаловать в университетскую систему.")
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("Войти"),
				),
			)
			bot.Send(msg)

		case "войти":
			authStates[chatID] = &authState{Step: 1}
			bot.Send(tgbotapi.NewMessage(chatID, "Введите username:"))

		case "создать пользователя":
			userCreateStates[chatID] = &userCreation{Step: 1}
			bot.Send(tgbotapi.NewMessage(chatID, "Введите роль (student или teacher):"))

		case "создать курс":
			courseCreateStates[chatID] = &courseCreation{Step: 1}
			bot.Send(tgbotapi.NewMessage(chatID, "Введите название курса:"))

		case "курсы":
			msg := tgbotapi.NewMessage(chatID, fetchCourses())
			bot.Send(msg)

		default:
			bot.Send(tgbotapi.NewMessage(chatID, "Неизвестная команда. Введите /start"))
		}
	}
}

func checkUser(username, password string) (*SimpleUser, string) {
	payload := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post("http://localhost:8080/login", "application/json", strings.NewReader(string(body)))
	if err != nil || resp.StatusCode != 200 {
		return nil, ""
	}
	defer resp.Body.Close()

	var user SimpleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, ""
	}

	return &user, user.Role
}

func sendUserToAPI(data *userCreation) {
	payload := map[string]string{
		"name":     data.Username,
		"password": data.Password,
		"role":     data.Role,
	}
	body, _ := json.Marshal(payload)
	http.Post("http://localhost:8080/users", "application/json", strings.NewReader(string(body)))
}

func sendCourseToAPI(data *courseCreation) {
	payload := map[string]string{
		"title":       data.Title,
		"description": data.Desc,
		"teacher_id":  data.TeacherID,
	}
	body, _ := json.Marshal(payload)
	http.Post("http://localhost:8080/courses", "application/json", strings.NewReader(string(body)))
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
	result.WriteString("📚 Курсы:\n\n")
	for _, c := range courses {
		result.WriteString(fmt.Sprintf("• %s\n", c.Title))
	}
	return result.String()
}

func showMenuByRole(bot *tgbotapi.BotAPI, chatID int64, name, role string) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Добро пожаловать, %s (%s)", name, role))

	switch role {
	case "admin":
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Создать пользователя"),
				tgbotapi.NewKeyboardButton("Создать курс"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("Курсы"),
			),
		)
	default:
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	}
	bot.Send(msg)
}
