package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

type User struct {
	Login              string    `json:"login"`
	Email              string    `json:"email"`
	Password           string    `json:"password"`
	TelegramID         string    `json:"telegram_id"`
	IsBanned           bool      `json:"is_banned"`
	IsAdmin            bool      `json:"is_admin"`
	SignUpToken        string    `json:"sign_up_token"`
	SignUpTokenDelTime *time.Time `json:"sign_up_token_del_time"`
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

func sendMessageEmail(email, subject, body string) error {
	smtpHost := "smtp.yandex.ru"
	smtpPort := "587"
	username := "avelgart@yandex.ru"
	password := "tuzzwwjpckaqfdvh"

	msg := []byte("From: " + username + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	auth := smtp.PlainAuth("", username, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, username, []string{email}, msg)
	if err != nil {
		return fmt.Errorf("Ошибка при отправке: %w", err)
	}

	fmt.Println("Сообщение отправлено:", subject)
	return nil
}

func generateSignUpToken() string {
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

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

	if !isPasswordStrong(user.Password) {
		http.Error(w, "PasswordIsTooWeak", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(user.Password)
	user.Password = hashedPassword

	user.SignUpToken = generateSignUpToken()
	t := time.Now().Add(time.Hour)
	user.SignUpTokenDelTime = &t


	user.TelegramID = ""
	user.IsBanned = false
	user.IsAdmin = false

	var existingUser  User
	err = db.QueryRow("SELECT login, email, sign_up_token FROM users WHERE email = $1", user.Email).Scan(&existingUser .Login, &existingUser .Email, &existingUser .SignUpToken)
	if err == nil {
		if existingUser .SignUpToken == "" {
			http.Error(w, "UserAlreadyExistsWithEmailAndNoToken", http.StatusConflict)
		} else {
			http.Error(w, "UserAlreadyExistsWithEmailAndHasToken", http.StatusConflict)
		}
		return
	}

	err = db.QueryRow("SELECT login, email, sign_up_token FROM users WHERE login = $1", user.Login).Scan(&existingUser .Login, &existingUser .Email, &existingUser .SignUpToken)
	if err == nil {
		http.Error(w, "UserAlreadyExistsWithLogin", http.StatusConflict)
		return
	}

	_, err = db.Exec("INSERT INTO users (login, email, password, telegram_id, is_banned, is_admin, sign_up_token, sign_up_token_del_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		user.Login, user.Email, user.Password, user.TelegramID, user.IsBanned, user.IsAdmin, user.SignUpToken, user.SignUpTokenDelTime)
	if err != nil {
		http.Error(w, "UserAlreadySignUp", http.StatusInternalServerError)
		return
	}

	subject := "Подтверждение регистрации"
	body := fmt.Sprintf("Пожалуйста, подтвердите вашу регистрацию, перейдя по следующей ссылке: http://localhost:8080/public/CoinTracker.html?token=%s", user.SignUpToken)
	err = sendMessageEmail(user.Email, subject, body)
	if err != nil {
		log.Println("Ошибка при отправке письма:", err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func cleanUpExpiredTokens() {
	for {
		time.Sleep(time.Minute)

		_, err := db.Exec("DELETE FROM users WHERE sign_up_token_del_time < NOW() AND sign_up_token IS NOT NULL AND sign_up_token <> ''")
		if err != nil {
			log.Println("Ошибка при удалении пользователей с истекшими токенами:", err)
		} else {
			log.Println("Удалены пользователи с истекшими токенами.")
		}
	}
}

func checkTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	var user User
	err := db.QueryRow("SELECT login, email FROM users WHERE sign_up_token = $1 AND sign_up_token_del_time > NOW()", token).Scan(&user.Login, &user.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			json.NewEncoder(w).Encode(map[string]bool{"success": false})
			return
		}
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func confirmTokenHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Token string `json:"token"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE users SET sign_up_token = '', sign_up_token_del_time = NULL WHERE sign_up_token = $1", requestBody.Token)
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func logInHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var storedUser  User
	var query string
	if isEmail(user.Login) {
		query = "SELECT login, email, password, is_banned, is_admin, sign_up_token, sign_up_token_del_time FROM users WHERE email = $1"
		err = db.QueryRow(query, user.Login).Scan(&storedUser .Login, &storedUser .Email, &storedUser .Password, &storedUser .IsBanned, &storedUser .IsAdmin, &storedUser .SignUpToken, &storedUser .SignUpTokenDelTime)
	} else {
		query = "SELECT login, email, password, is_banned, is_admin, sign_up_token, sign_up_token_del_time FROM users WHERE login = $1"
		err = db.QueryRow(query, user.Login).Scan(&storedUser .Login, &storedUser .Email, &storedUser .Password, &storedUser .IsBanned, &storedUser .IsAdmin, &storedUser .SignUpToken, &storedUser .SignUpTokenDelTime)
	}
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "UserNotFound", http.StatusUnauthorized)
			return
		}
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	
	// Проверка на наличие токена
	if storedUser .SignUpToken != "" && storedUser .SignUpTokenDelTime != nil {
		http.Error(w, "UserHasToken", http.StatusForbidden)
		return
	}	

	if storedUser .IsBanned {
		http.Error(w, "UserIsBanned", http.StatusForbidden)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUser .Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "InvalidCredentials", http.StatusUnauthorized)
		return
	}

	// Если токен пустой, пропускаем пользователя
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}




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

	http.HandleFunc("/api/checkToken", checkTokenHandler)
	http.HandleFunc("/api/confirmToken", confirmTokenHandler)
	http.HandleFunc("/SignUp", registerHandler)
	http.HandleFunc("/LogIn", logInHandler)

	go cleanUpExpiredTokens()

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
