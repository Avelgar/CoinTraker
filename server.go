package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
    "strings"

    "golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

type User struct {
	Login      string `json:"login"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	TelegramID string `json:"telegram_id"`
	IsBanned   bool   `json:"is_banned"`
	IsAdmin    bool   `json:"is_admin"`
}

var db *sql.DB

func initDB() {
	var err error
	connStr := "postgres://postgres:12345@localhost:5432/cointracker?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
}

// Функция для проверки сложности пароля
func isPasswordStrong(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}

// Функция для хеширования пароля
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Проверка сложности пароля
	if !isPasswordStrong(user.Password) {
		http.Error(w, "PasswordIsTooWeak", http.StatusBadRequest)
		return
	}

    // Хеширование пароля
	hashedPassword, err := hashPassword(user.Password)
	user.Password = hashedPassword

	// Заполнение новых полей по умолчанию
	user.TelegramID = ""
	user.IsBanned = false
	user.IsAdmin = false

	_, err = db.Exec("INSERT INTO users (login, email, password, telegram_id, is_banned, is_admin) VALUES ($1, $2, $3, $4, $5, $6)",
		user.Login, user.Email, user.Password, user.TelegramID, user.IsBanned, user.IsAdmin)
	if err != nil {
		http.Error(w, "UserAlreadySignUp", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func LogInHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Получаем пользователя из базы данных по логину или email
	var storedUser  User
	var query string
	if isEmail(user.Login) {
		query = "SELECT login, email, password, is_banned, is_admin FROM users WHERE email = $1"
		err = db.QueryRow(query, user.Login).Scan(&storedUser .Login, &storedUser .Email, &storedUser .Password, &storedUser .IsBanned, &storedUser .IsAdmin)
	} else {
		query = "SELECT login, email, password, is_banned, is_admin FROM users WHERE login = $1"
		err = db.QueryRow(query, user.Login).Scan(&storedUser .Login, &storedUser .Email, &storedUser .Password, &storedUser .IsBanned, &storedUser .IsAdmin)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "UserNotFound", http.StatusUnauthorized)
			return
		}
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	// Проверяем, забанен ли пользователь
	if storedUser .IsBanned {
		http.Error(w, "UserIsBanned", http.StatusForbidden)
		return
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(storedUser .Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "InvalidCredentials", http.StatusUnauthorized)
		return
	}

	// Успешный вход
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

// Функция для проверки, является ли строка электронной почтой
func isEmail(input string) bool {
	return strings.Contains(input, "@") && strings.Contains(input, ".")
}

func main() {
	initDB()
	defer db.Close()

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/public/CoinTracker.html", http.StatusFound)
	})

	http.HandleFunc("/SignUp", registerHandler)
    http.HandleFunc("/LogIn", LogInHandler)
	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
