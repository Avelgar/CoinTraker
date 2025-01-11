package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
)

type User struct {
    Login    string `json:"login"`
    Email    string `json:"email"`
    Password string `json:"password"` 
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        var user User
        err := json.NewDecoder(r.Body).Decode(&user)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        users, err := loadUsers()
        if err != nil {
            http.Error(w, "Не удалось загрузить пользователей", http.StatusInternalServerError)
            return
        }

        for _, existingUser  := range users {
            if existingUser .Login == user.Login || existingUser .Email == user.Email {
                http.Error(w, "Пользователь с таким логином или электронной почтой уже существует", http.StatusConflict)
                return
            }
        }

        users = append(users, user)

        if err := saveUsers(users); err != nil {
            http.Error(w, "Не удалось сохранить пользователя", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"message": "Регистрация успешна!"})
    } else {
        http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
    }
}

func loadUsers() ([]User , error) {
    var users []User   

    file, err := os.Open("users.json")
    if err != nil {
        if os.IsNotExist(err) {

            return users, nil
        }
        return nil, err
    }
    defer file.Close()

    data, err := ioutil.ReadAll(file)
    if err != nil {
        return nil, err
    }
    err = json.Unmarshal(data, &users)
    return users, err
}

func saveUsers(users []User ) error {

    data, err := json.MarshalIndent(users, "", "  ")
    if err != nil {
        return err
    }

    return ioutil.WriteFile("users.json", data, 0644)
}

func main() {

    http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/public/CoinTracker.html", http.StatusFound)
    })

    http.HandleFunc("/register", registerHandler)

    fmt.Println("Сервер запущен на http://localhost:8080")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println("Ошибка запуска сервера:", err)
    }
}
