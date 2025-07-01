package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var store *sessions.CookieStore

type User struct {
	ID int
	Name string
	Surname string
	Age uint
	Email string
	Password string
	Balance float64
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Файл .env не найден, будут использованы системные переменные")
	}
	key := os.Getenv("SESSION_KEY")
	if key == "" {
		panic("Error")
	}
	store = sessions.NewCookieStore([]byte(key))
}

func dbConnect() (*pgx.Conn, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DB URL error")
	}
	return pgx.Connect(context.Background(), url)
}

func save(name, surname, pass, email string, age string) {
	conn, err := dbConnect()
	if err != nil {
		log.Fatal("DB connection error")
	}
	defer conn.Close(context.Background())

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Hashing password error")
	}

	_, err = conn.Exec(context.Background(), `INSERT INTO "user" (name, surname, age, email, password) VALUES ($1, $2, $3, $4, $5)`, name, surname, age, email, hashedPass)
	if err != nil {
		log.Fatal("DB writing error")
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	surname := r.FormValue("surname")
	age := r.FormValue("age")
	email := r.FormValue("email")
	pass := r.FormValue("password")

	save(name, surname, pass, email, age)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func load() []User {
	conn, err := dbConnect()
	if err != nil {
		log.Fatal("DB connection error")
	}
	defer conn.Close(context.Background())

	query := `SELECT id, name, surname, age, email, password, balance FROM "user"`

	rows, err := conn.Query(context.Background(), query)
	users := []User {}
	if err != nil {
		log.Fatal("DB listening error")
	}
	for rows.Next() {
		var user User
		_ = rows.Scan(
			&user.ID,
			&user.Name,
			&user.Surname,
			&user.Age,
			&user.Email,
			&user.Password,
			&user.Balance,
		)
		users = append(users, user)
	}
	return users
}

func loadHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	session, _ := store.Get(r, "account")

	if r.Method != http.MethodPost {
		// Load user data from session
		id, ok := session.Values["id"].(int)
		if !ok {
			http.Error(w, "Invalid session or not logged in", http.StatusUnauthorized)
			return
		}

		user = User{
			ID:       id,
			Name:     session.Values["name"].(string),
			Surname:  session.Values["surname"].(string),
			Age:      session.Values["age"].(uint),
			Email:    session.Values["email"].(string),
			Password: session.Values["password"].(string),
		}
		conn, err := dbConnect()
		if err != nil {
			http.Error(w, "DB connection error", 500)
		}
		defer conn.Close(context.Background())

		var balance float64
		query := `SELECT balance FROM "user" WHERE id = $1;`
		err = conn.QueryRow(context.Background(), query, user.ID).Scan(&balance)
		if err != nil {
			http.Error(w, "Balance getting error", 500)
		}
		user.Balance = balance

	} else {
		// Parse login form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Form parsing error", http.StatusBadRequest)
			return
		}

		email := r.FormValue("emaill")
		pass := r.FormValue("passwordl")

		users := load()
		found := false

		for _, u := range users {
			if u.Email == email {
				// Compare hashed password
				err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass))
				if err != nil {
					http.Error(w, "Invalid password", http.StatusUnauthorized)
					return
				}

				user = u
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Set session
		session.Values["id"] = user.ID
		session.Values["name"] = user.Name
		session.Values["surname"] = user.Surname
		session.Values["age"] = user.Age
		session.Values["email"] = user.Email
		session.Values["password"] = user.Password
		session.Save(r, w) // Important: save the session
	}

	// Render template
	t, err := template.ParseFiles("templates/user.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	t.Execute(w, user)
}

func delete(id int) {
	conn, err := dbConnect()
	if err != nil {
		log.Fatal("DB connection error")
	}
	defer conn.Close(context.Background())
	query := `DELETE FROM "user" WHERE id = $1;`
	_, err = conn.Exec(context.Background(), query, id)
	if err != nil {
		log.Fatal("DB deleting error")
	}
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "account")
	id, ok := session.Values["id"].(int)
	if !ok {
		http.Error(w, "Getting ID error", 500)
	}
	delete(id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func pay(sum float64, id int) {
	conn, err := dbConnect()
	if err != nil {
		log.Fatal("DB connection error")
	}
	defer conn.Close(context.Background())

	query := `UPDATE "user" SET balance = balance + $1 WHERE id = $2;`

	conn.Exec(context.Background(), query, sum, id)

}

func payHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Form parsing error", 400)
		return
	}
	id := r.FormValue("id")
	sum := r.FormValue("sum")

	s, err := strconv.ParseFloat(sum, 64)
	if err != nil {
		http.Error(w, "Parsing error", 400)
		return
	}

	i, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Parsing error", 404)
		return
	}
	pay(s, i)
	http.Redirect(w, r, "/load", http.StatusSeeOther)
}

func index(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "File parsing error", 404)
		return
	}
	t.Execute(w, nil)
}

func handeFunc() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/save", saveHandler).Methods("POST")
	r.HandleFunc("/load", loadHandler)
	r.HandleFunc("/pay", payHandler).Methods("POST")
	r.HandleFunc("/delete", deleteHandler)

	http.Handle("/", r)

	http.ListenAndServe(":8080", nil)
}

func main() {
	handeFunc()
}