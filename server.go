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

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"github.com/lib/pq"
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
	CookieList         []string   `json:"cookie_list"`
	CoinsList		   []string   `json:"coins_list"`
	RememberMe         bool   `json:"rememberMe"`
}

var db *sql.DB
var store = sessions.NewCookieStore([]byte("secret-key")) // Замените на свой секретный ключ

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
	if err != nil {
		log.Println("Ошибка при хэшировании пароля:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	signUpToken := generateSignUpToken()
	user.SignUpToken = &signUpToken
	t := time.Now().Add(time.Hour)
	user.SignUpTokenDelTime = &t

	user.RecoveryToken = nil
	user.RecoveryTokenDelTime = nil
	user.TelegramID = nil
	user.IsBanned = false
	user.IsAdmin = false
	user.CookieList = []string{}
	user.CoinsList = []string{}

	var existingUser  User
	err = db.QueryRow("SELECT login, email, sign_up_token FROM users WHERE email = $1", user.Email).Scan(&existingUser .Login, &existingUser .Email, &existingUser .SignUpToken)
	if err == nil {
		if existingUser .SignUpToken == nil {
			http.Error(w, "UserAlreadyExistsWithEmailAndNoToken", http.StatusConflict)
		} else {
			http.Error(w, "UserAlreadyExistsWithEmailAndHasToken", http.StatusConflict)
		}
		return
	} else if err != sql.ErrNoRows {
		log.Println("Ошибка при проверке существующего пользователя по email:", err)
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT login, email, sign_up_token FROM users WHERE login = $1", user.Login).Scan(&existingUser .Login, &existingUser .Email, &existingUser .SignUpToken)
	if err == nil {
		http.Error(w, "UserAlreadyExistsWithLogin", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		log.Println("Ошибка при проверке существующего пользователя по логину:", err)
		http.Error(w, "InternalServerError", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (login, email, password, telegram_id, is_banned, is_admin, sign_up_token, sign_up_token_del_time, recovery_token, recovery_token_del_time, cookie_list, coins_list) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)",
	user.Login, user.Email, user.Password, user.TelegramID, user.IsBanned, user.IsAdmin, user.SignUpToken, user.SignUpTokenDelTime, user.RecoveryToken, user.RecoveryTokenDelTime, pq.Array(user.CookieList), pq.Array(user.CoinsList))
	if err != nil {
		log.Println("Ошибка при вставке пользователя в базу данных:", err)
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
	json.NewEncoder(w).Encode(map[string]string{"message": "User  registered successfully"})
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

	session, _ := store.Get(r, "session-name")
    session.Values["authenticated"] = true
    session.Values["userLogin"] = storedUser .Login
    session.Save(r, w)

	// var cookieToken string
    if user.RememberMe {
        cookieToken := generateSignUpToken()
        _, err = db.Exec("UPDATE users SET cookie_list = array_append(cookie_list, $1) WHERE email = $2 OR login = $3", cookieToken, storedUser .Email, storedUser .Login)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        // Установите куки в ответе
        http.SetCookie(w, &http.Cookie{
			Name:     "rememberMeToken",
			Value:    cookieToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   true, 
		})
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}


func handleCheckCookie(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("rememberMeToken")
    if err != nil {
        if err == http.ErrNoCookie {
            json.NewEncoder(w).Encode(map[string]interface{}{"success": false})
            return
        }
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    token := cookie.Value

    var user User
    err = db.QueryRow("SELECT login, email FROM users WHERE $1 = ANY(cookie_list)", token).Scan(&user.Login, &user.Email)

    if err != nil {
        if err == sql.ErrNoRows {
            json.NewEncoder(w).Encode(map[string]interface{}{"success": false})
            return
        }
        http.Error(w, "InternalServerError", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}


func handleAuthentication(w http.ResponseWriter, r *http.Request) {
    // Проверяем куки
    cookie, err := r.Cookie("rememberMeToken")
    if err == nil {
        token := cookie.Value
        var user User
        err = db.QueryRow("SELECT login, email, is_admin, telegram_id, coins_list FROM users WHERE $1 = ANY(cookie_list)", token).Scan(&user.Login, &user.Email, &user.IsAdmin, &user.TelegramID, pq.Array(&user.CoinsList))

        if err == nil {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "success": true,
                "login":   user.Login,
                "email":   user.Email,
                "is_admin": user.IsAdmin,
                "telegram_id": user.TelegramID,
                "coins":   user.CoinsList,
            })
            return
        }
    }

    // Проверяем сессию
	session, err := store.Get(r, "session-name")
    if err == nil && session.Values["authenticated"] != nil {
        login := session.Values["userLogin"].(string)

        var coinsList []string
        var isAdmin bool
        var telegramID sql.NullString 
		var email string
        err = db.QueryRow("SELECT coins_list, email, is_admin, telegram_id FROM users WHERE login = $1", login).Scan(pq.Array(&coinsList), &email, &isAdmin, &telegramID)
        if err == nil {
            response := map[string]interface{}{
                "success": true,
                "login":   login,
				"email": email,
                "is_admin": isAdmin,
                "telegram_id": "",
                "coins":   coinsList,
            }

            if telegramID.Valid {
                response["telegram_id"] = telegramID.String
            }

            json.NewEncoder(w).Encode(response)
            return
        }
    }

    // Если аутентификация не удалась
    json.NewEncoder(w).Encode(map[string]interface{}{"success": false})
}



func logoutHandler(w http.ResponseWriter, r *http.Request) {
    // Получаем токен куки
    cookie, err := r.Cookie("rememberMeToken")
    
    // Проверяем, есть ли куки. Если нет, просто продолжаем выполнение.
    if err != nil {
        // Куки нет, просто продолжаем
    } else {
        token := cookie.Value

        // Удаляем куки из ответа
        http.SetCookie(w, &http.Cookie{
            Name:     "rememberMeToken",
            Value:    "",
            Path:     "/",
            HttpOnly: true,
            Secure:   true,
            MaxAge:   -1, // Устанавливаем MaxAge в -1 для удаления куки
        })

        // Проверяем, существует ли токен в базе данных
        var exists bool
        err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE $1 = ANY(cookie_list))", token).Scan(&exists)
        if err != nil {
            http.Error(w, "Failed to check cookie existence in database", http.StatusInternalServerError)
            return
        }

        // Если токен существует, удаляем его из базы данных
        if exists {
            _, err = db.Exec("UPDATE users SET cookie_list = array_remove(cookie_list, $1) WHERE $1 = ANY(cookie_list)", token)
            if err != nil {
                http.Error(w, "Failed to remove cookie from database", http.StatusInternalServerError)
                return
            }
        }
    }

    // Завершаем сессию
    session, _ := store.Get(r, "session-name")
    session.Values["authenticated"] = false
    session.Values["userLogin"] = ""
    session.Save(r, w)

    // Возвращаем успешный ответ
    json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}

func userPageHandler(w http.ResponseWriter, r *http.Request) {
    // Проверяем аутентификацию
    session, err := store.Get(r, "session-name")
    if err != nil || session.Values["authenticated"] == nil || session.Values["authenticated"] == false {
        // Если пользователь не аутентифицирован, перенаправляем его на главную страницу
        http.Redirect(w, r, "/public/CoinTracker.html", http.StatusFound)
        return
    }

    // Если аутентифицирован, отображаем страницу пользователя
    http.ServeFile(w, r, "./public/User.html")
}

func adminPageHandler(w http.ResponseWriter, r *http.Request) {
    // Проверяем аутентификацию
    session, err := store.Get(r, "session-name")
    if err != nil || session.Values["authenticated"] == nil || session.Values["authenticated"] == false {
        // Если пользователь не аутентифицирован, перенаправляем его на главную страницу
        http.Redirect(w, r, "/public/CoinTracker.html", http.StatusFound)
        return
    }

    // Если аутентифицирован, отображаем страницу пользователя
    http.ServeFile(w, r, "./public/Admin.html")
}

func profilePageHandler(w http.ResponseWriter, r *http.Request) {
    // Проверяем аутентификацию
    session, err := store.Get(r, "session-name")
    if err != nil || session.Values["authenticated"] == nil || session.Values["authenticated"] == false {
        // Если пользователь не аутентифицирован, перенаправляем его на главную страницу
        http.Redirect(w, r, "/public/CoinTracker.html", http.StatusFound)
        return
    }

    // Если аутентифицирован, отображаем страницу пользователя
    http.ServeFile(w, r, "./public/Profile.html")
}

func main() {
	initDB()
	defer db.Close()

	// http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	http.HandleFunc("/public/User.html", userPageHandler)
	http.HandleFunc("/public/Admin.html", adminPageHandler)
	http.HandleFunc("/public/Profile.html", profilePageHandler)
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
	http.HandleFunc("/api/checkCookie", handleCheckCookie)
	http.HandleFunc("/api/authenticate", handleAuthentication)
	http.HandleFunc("/api/logout", logoutHandler)

	go cleanUpExpiredTokens()

	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
