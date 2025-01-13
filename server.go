package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
    "os"

    _"github.com/jackc/pgx/v4/stdlib"
)

type User struct {
    Login    string `json:"login"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

var db *sql.DB

func init() {
    var err error
    // Замените строку подключения на вашу
    connStr := "postgres://postgres:12345@localhost:5432/cointracker"
    db, err = sql.Open("pgx", connStr)
    if err != nil {
        fmt.Println("DatabaseConnectionError:", err)
        os.Exit(1)
    }

    if err = db.Ping(); err != nil {
        fmt.Println("DatabasePingError:", err)
        os.Exit(1)
    }
}

func SignUp(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        var user User
        err := json.NewDecoder(r.Body).Decode(&user)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        var existingUser  User
        err = db.QueryRow("SELECT login, email FROM users WHERE login = $1 OR email = $2", user.Login, user.Email).Scan(&existingUser .Login, &existingUser .Email)
        if err == nil {
            http.Error(w, "UserExistSignUp", http.StatusConflict)
            return
        }

        // Попытка вставки нового пользователя
        _, err = db.Exec("INSERT INTO users (login, email, password) VALUES ($1, $2, $3)", user.Login, user.Email, user.Password)
        if err != nil {
            http.Error(w, "UserCannotBeSignUp", http.StatusInternalServerError)
            return
        }

        // Если все прошло успешно, отправляем сообщение об успешной регистрации
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Регистрация успешна!"})
    } else {
        http.Error(w, "MethodNotAllowedSignUp", http.StatusMethodNotAllowed)
    }
}



func main() {
    http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/public/CoinTracker.html", http.StatusFound)
    })

    http.HandleFunc("/SignUp", SignUp)

    fmt.Println("Сервер запущен на http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Ошибка запуска сервера:", err)
    }
}
