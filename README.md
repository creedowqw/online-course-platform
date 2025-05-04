# Online Course Platform Telegram Bot

Этот проект — мой Telegram-бот для управления университетскими курсами. Я написал его на Go с использованием Gin, GORM и PostgreSQL. Бот реализует три роли: **админ**, **преподаватель**, **студент**.

---

##  Возможности

###  Админ

- Авторизация через email `admin@narxoz.kz`
- Импорт курсов из Canvas
- Назначение учителя на курс
- Назначение ролей (student / teacher / admin)

###  Преподаватель

- Просмотр своих курсов
- Просмотр списка студентов, записанных на конкретный курс

###  Студент

- Авторизация через `@narxoz.kz`
- Просмотр доступных курсов
- Запись на курсы
- Просмотр профиля и списка записанных курсов

##  Технологии

- **Go (Golang)**
- **Gin** — HTTP роутер
- **GORM** — ORM для PostgreSQL
- **PostgreSQL**
- **Docker / docker-compose**
- **Telegram Bot API**

---

## Требования

- Go >= 1.20
- Docker / Docker Compose
- Аккаунт Telegram

---

##  Установка и запуск

```bash
git clone https://github.com/username/online-course-platform.git
cd online-course-platform
cp .env.example .env
```

### Запуск с Docker:

```bash
docker-compose up --build
```

Если нужно без кэша:

```bash
docker-compose build --no-cache
docker-compose up
```

---

## Структура проекта

```
/cmd               # main.go
/internal
  /bot            # Telegram bot logic
  /controllers    # REST API (optional)
  /db             # DB init & migrations
  /integrations/canvas # Canvas API importer
  /models         # GORM models
  /routes         # API routes
```

---

## .env переменные

```
TELEGRAM_BOT_TOKEN=...
SMTP_EMAIL=...
SMTP_PASS=...
CANVAS_API_URL=https://canvas.narxoz.kz
CANVAS_API_TOKEN=...
DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=online_courses
```

---

## Важное:

- После импорта курсов **только админ** может назначить учителя на курс
- При записи на курс количество мест уменьшается
- Один и тот же студент не может записаться дважды
- Для смены роли используется команда **"Назначить роль"**

---

##  Автор

- Я: @bekassyl.serik
