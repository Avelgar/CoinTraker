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
	Login              string     `json:"login"`
	Email              string     `json:"email"`
	Password           string     `json:"password"`
	TelegramID         *string    `json:"telegram_id"`
	IsBanned           bool       `json:"is_banned"`
	IsAdmin            bool       `json:"is_admin"`
	SignUpToken        *string    `json:"sign_up_token"`
	SignUpTokenDelTime *time.Time  `json:"sign_up_token_del_time"`
	RecoveryToken      *string    `json:"recovery_token"`
	RecoveryTokenDelTime *time.Time `json:"recovery_token_del_time"`
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

	signUpToken := generateSignUpToken()
	user.SignUpToken = &signUpToken
	t := time.Now().Add(time.Hour)
	user.SignUpTokenDelTime = &t

	// Устанавливаем RecoveryToken и RecoveryTokenDelTime в nil
	user.RecoveryToken = nil
	user.RecoveryTokenDelTime = nil

	user.TelegramID = nil // Устанавливаем TelegramID в nil
	user.IsBanned = false
	user.IsAdmin = false

	var existingUser  User
	err = db.QueryRow("SELECT login, email, sign_up_token FROM users WHERE email = $1", user.Email).Scan(&existingUser .Login, &existingUser .Email, &existingUser .SignUpToken)
	if err == nil {
		if existingUser .SignUpToken == nil {
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

	_, err = db.Exec("INSERT INTO users (login, email, password, telegram_id, is_banned, is_admin, sign_up_token, sign_up_token_del_time, recovery_token, recovery_token_del_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		user.Login, user.Email, user.Password, user.TelegramID, user.IsBanned, user.IsAdmin, user.SignUpToken, user.SignUpTokenDelTime, user.RecoveryToken, user.RecoveryTokenDelTime)
	if err != nil {
		http.Error(w, "UserAlreadySignUp", http.StatusInternalServerError)
		return
	}

	subject := "Подтверждение регистрации"
	body := fmt.Sprintf("Пожалуйста, подтвердите вашу регистрацию, перейдя по следующей ссылке: http://localhost:8080/public/CoinTracker.html?token=%s", *user.SignUpToken)
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

		// Удаляем пользователей с истекшими токенами регистрации
		_, err := db.Exec("DELETE FROM users WHERE sign_up_token_del_time < NOW() AND sign_up_token IS NOT NULL AND sign_up_token <> ''")
		if err != nil {
			log.Println("Ошибка при удалении пользователей с истекшими токенами:", err)
		} else {
			log.Println("Удалены пользователи с истекшими токенами.")
		}

		// Обнуляем токены восстановления для пользователей с истекшим временем
		_, err = db.Exec("UPDATE users SET recovery_token = NULL, recovery_token_del_time = NULL WHERE recovery_token_del_time < NOW() AND recovery_token IS NOT NULL AND recovery_token <> ''")
		if err != nil {
			log.Println("Ошибка при обнулении токенов восстановления:", err)
		} else {
			log.Println("Обнулены токены восстановления для пользователей с истекшим временем.")
		}
	}
}

func handleCheckToken(w http.ResponseWriter, r *http.Request) {

    // Получаем токен из тела запроса
    var requestData struct {
        Token string `json:"token"`
    }
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil || requestData.Token == "" {
        json.NewEncoder(w).Encode(map[string]string{"error": "NoToken"})
        return
    }

    var user User
    err := db.QueryRow("SELECT login, email FROM users WHERE sign_up_token = $1 AND sign_up_token_del_time > NOW()", requestData.Token).Scan(&user.Login, &user.Email)

    if err != nil {
        if err == sql.ErrNoRows {
            json.NewEncoder(w).Encode(map[string]bool{"success": false})
            return
        }
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    _, err = db.Exec("UPDATE users SET sign_up_token = NULL, sign_up_token_del_time = NULL WHERE sign_up_token = $1", requestData.Token)
    if err != nil {
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]bool{"success": true})
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
		query = "SELECT login, email, password, is_banned, is_admin, sign_up_token, sign_up_token_del_time, recovery_token, recovery_token_del_time FROM users WHERE email = $1"
		err = db.QueryRow(query, user.Login).Scan(&storedUser .Login, &storedUser .Email, &storedUser .Password, &storedUser .IsBanned, &storedUser .IsAdmin, &storedUser .SignUpToken, &storedUser .SignUpTokenDelTime, &storedUser .RecoveryToken, &storedUser .RecoveryTokenDelTime)
	} else {
		query = "SELECT login, email, password, is_banned, is_admin, sign_up_token, sign_up_token_del_time, recovery_token, recovery_token_del_time FROM users WHERE login = $1"
		err = db.QueryRow(query, user.Login).Scan(&storedUser .Login, &storedUser .Email, &storedUser .Password, &storedUser .IsBanned, &storedUser .IsAdmin, &storedUser .SignUpToken, &storedUser .SignUpTokenDelTime, &storedUser .RecoveryToken, &storedUser .RecoveryTokenDelTime)
	}
	
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "UserNotFound", http.StatusUnauthorized)
			return
		}
		fmt.Println(err)
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	// Проверка на наличие токена
	if storedUser .SignUpToken != nil && *storedUser .SignUpToken != "" && storedUser .SignUpTokenDelTime != nil {
		http.Error(w, "UserHasToken", http.StatusForbidden)
		return
	}

	if storedUser .RecoveryToken != nil && *storedUser .RecoveryToken != "" && storedUser .RecoveryTokenDelTime != nil {
		http.Error(w, "UserHasRecoveryToken", http.StatusForbidden)
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

func recoveryHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Email string `json:"email"`
	}
	
	// Декодируем JSON из тела запроса
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	
	// Проверяем, существует ли пользователь
	var user User
	err = db.QueryRow("SELECT login, email, is_banned, sign_up_token, sign_up_token_del_time, recovery_token, recovery_token_del_time FROM users WHERE email = $1", 
	requestBody.Email).Scan(&user.Login, &user.Email, &user.IsBanned, &user.SignUpToken, &user.SignUpTokenDelTime, &user.RecoveryToken, &user.RecoveryTokenDelTime)
	
	if err == sql.ErrNoRows {
		http.Error(w, "UserNotFound", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	
	// Проверка на наличие токена регистрации
	if user.SignUpToken != nil && *user.SignUpToken != "" && user.SignUpTokenDelTime != nil {
		http.Error(w, "UserHasToken", http.StatusForbidden)
		return
	}
	
	// Проверка на забаненность пользователя
	if user.IsBanned {
		http.Error(w, "UserIsBanned", http.StatusForbidden)
		return
	}
	
	// Проверка на наличие токена восстановления
	if user.RecoveryToken != nil && *user.RecoveryToken != "" && user.RecoveryTokenDelTime != nil {
		http.Error(w, "UserHasRecoveryToken", http.StatusForbidden)
		return
	}
	
	// Генерация токена восстановления
	recoveryToken := generateSignUpToken()
	user.RecoveryToken = &recoveryToken
	t := time.Now().Add(time.Hour) // Токен действителен 1 час
	user.RecoveryTokenDelTime = &t
	
	// Обновление пользователя с новым токеном
	_, err = db.Exec("UPDATE users SET recovery_token = $1, recovery_token_del_time = $2 WHERE email = $3", user.RecoveryToken, user.RecoveryTokenDelTime, user.Email)
	if err != nil {
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}
	
	// Отправка письма с ссылкой на восстановление пароля
	subject := "Восстановление пароля"
	body := fmt.Sprintf("Пожалуйста, восстановите ваш пароль, перейдя по следующей ссылке: http://localhost:8080/public/Recovery.html?recovery_token=%s", *user.RecoveryToken)
	err = sendMessageEmail(user.Email, subject, body)
	if err != nil {
		log.Println("Ошибка при отправке письма:", err)
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Письмо с ссылкой на восстановление пароля отправлено!"})
}

func confirmRecoveryTokenHandler(w http.ResponseWriter, r *http.Request) {
    var requestData struct {
        RecoveryToken string `json:"recovery_token"`
    }

    // Декодируем JSON-данные из запроса
    if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil || requestData.RecoveryToken == "" {
        http.Error(w, "NoRecoveryToken", http.StatusBadRequest)
        return
    }

    // Проверяем наличие токена в базе данных
    var user User
    err := db.QueryRow("SELECT login FROM users WHERE recovery_token = $1 AND recovery_token_del_time > NOW()", requestData.RecoveryToken).Scan(&user.Login)

    if err != nil {
        if err == sql.ErrNoRows {
            // Если токен не найден, возвращаем ответ с успехом false
            json.NewEncoder(w).Encode(map[string]bool{"success": false})
            return
        }
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    // Если токен действителен, возвращаем ответ с успехом true
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}



func SubmitRecoveryTokenHandler(w http.ResponseWriter, r *http.Request) {
    var requestBody struct {
        RecoveryPassword string `json:"RecoveryPassword"`
        RecoveryToken    string `json:"recovery_token"`
    }

    // Декодируем JSON-данные из запроса
    err := json.NewDecoder(r.Body).Decode(&requestBody)
    if err != nil || requestBody.RecoveryToken == "" || requestBody.RecoveryPassword == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Проверяем сложность пароля
    if !isPasswordStrong(requestBody.RecoveryPassword) {
        http.Error(w, "PasswordIsTooWeak", http.StatusBadRequest)
        return
    }

    // Проверяем, существует ли пользователь по токену
    var user User
    err = db.QueryRow("SELECT login FROM users WHERE recovery_token = $1", requestBody.RecoveryToken).Scan(&user.Login)
    if err == sql.ErrNoRows {
        http.Error(w, "User  NotFound", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    // Хэшируем новый пароль
    hashedPassword, err := hashPassword(requestBody.RecoveryPassword)
    if err != nil {
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    // Обновляем пароль в базе данных
    _, err = db.Exec("UPDATE users SET password = $1 WHERE recovery_token = $2", hashedPassword, requestBody.RecoveryToken)
    if err != nil {
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    // Обнуляем токен и время удаления токена
    _, err = db.Exec("UPDATE users SET recovery_token = NULL, recovery_token_del_time = NULL WHERE recovery_token = $1", requestBody.RecoveryToken)
    if err != nil {
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    // Возвращаем успешный ответ
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Пароль успешно изменен!"})
}


func main() {
	initDB()
	defer db.Close()

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/public/CoinTracker.html", http.StatusFound)
	})

	http.HandleFunc("/api/checkToken", handleCheckToken)
	http.HandleFunc("/SignUp", registerHandler)
	http.HandleFunc("/LogIn", logInHandler)
	http.HandleFunc("/Recovery", recoveryHandler)
	http.HandleFunc("/api/confirmRecoveryToken", confirmRecoveryTokenHandler)
	http.HandleFunc("/api/SubmitRecovery", SubmitRecoveryTokenHandler)

	go cleanUpExpiredTokens()

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
