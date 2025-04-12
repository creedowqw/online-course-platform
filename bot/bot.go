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

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Ошибка получения обновлений:", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		text := strings.ToLower(update.Message.Text)

		switch text {
		case "/start":
			reply := "Привет! Я бот университета. Используй /courses чтобы посмотреть курсы."
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, reply))

		case "/courses":
			msg := fetchCourses()
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))

		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Попробуй /start"))
		}
	}
}

func fetchCourses() string {
	resp, err := http.Get("http://app:8080/courses")
	if err != nil {
		return "Ошибка при обращении к API курсов"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Сервер вернул статус: %d", resp.StatusCode)
	}

	var courses []struct {
		ID          uint   `json:"ID"`
		Title       string `json:"title"`
		Description string `json:"description"`
		TeacherID   uint   `json:"teacher_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return "Ошибка обработки данных"
	}

	if len(courses) == 0 {
		return "Нет доступных курсов."
	}

	var result strings.Builder
	result.WriteString("📚 Курсы:\n\n")
	for _, c := range courses {
		result.WriteString(fmt.Sprintf("• %s\n", c.Title))
	}
	return result.String()
}
