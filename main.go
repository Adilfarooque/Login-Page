package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"text/template"
	
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))
var db *sql.DB

func init() {
	var err error

	connStr := "postgres://postgres:7356@localhost/logindb?sslmode=disable"
	db, err = sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	// this will be printed in the terminal, confirming the connection to the database
	fmt.Println("The database is connected")
}

func main() {

	r := mux.NewRouter()
	http.Handle("/", r)

	// Define routes
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/login", LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", LogoutHandler).Methods("GET")
	r.HandleFunc("/admin", AdminHandler).Methods("GET") //New admin route
	r.HandleFunc("/admin/update", UpdateHandler).Methods("GET", "POST")
	r.HandleFunc("/admin/delete", DeleteHandler).Methods("GET", "POST")
	r.HandleFunc("/signup", SignupHandler).Methods("GET", "POST")
	// Start the server
	http.ListenAndServe(":8080", nil)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		//To check the username already exists in the DB
		var existingUsername string
		err := db.QueryRow("SELECT username FROM login_data WHERE username = $1", username).Scan(&existingUsername)

		switch {
		case err == sql.ErrNoRows:
			//Proceed with registration

			//Insert new user into DB
			_, err := db.Exec("INSERT INTO login_data (username, password) VALUES($1, $2)", username, password)
			if err != nil {
				fmt.Fprintln(w, "Error creating user:", err)
				return
			}
			//Redirecting to the login page :)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return

		case err != nil:
			//Handler other error
			fmt.Fprintln(w, "Error checking username availability:", err)
			return

		default:
			//If the user already taken , display error msg
			fmt.Fprintln(w, "Username is already taken. Please choose a different username.")
			return

		}
	}

	tpl := template.Must(template.ParseFiles("templates/signup.html"))
	tpl.Execute(w, nil)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Render the home page
	tpl := template.Must(template.ParseFiles("templates/home.html"))
	tpl.Execute(w, username)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")

	if r.Method == "POST" {
		// Handle form submission
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Add your validation logic here, for example:
		var storedPassword string
		err := db.QueryRow("SELECT password FROM login_data WHERE username = $1", username).Scan(&storedPassword)
		if err == nil && password == storedPassword {
			session.Values["username"] = username
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else {
			fmt.Fprintf(w, "Incorrect username or password. Error: %v", err)
		}

	}
	// Render the login page
	tmpl := template.Must(template.ParseFiles("templates/login.html"))
	tmpl.Execute(w, nil)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"].(string)
	if !ok || username != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method == "POST" {
		r.ParseForm()
		userID := r.FormValue("user_id")
		newUsername := r.FormValue("username")
		newPassword := r.FormValue("password")

		_, err := db.Exec("UPDATE login_data SET username =$1 , password = $2 WHERE id = $3", newUsername, newPassword, userID)
		if err != nil {
			fmt.Fprintln(w, "Error updating user:", err)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	tpl := template.Must(template.ParseFiles("templates/update.html"))
	tpl.Execute(w, username)

}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"].(string)
	fmt.Println("Username in AdminHandler:", username) // Add this line for debugging
	if !ok || username != "admin" {
		fmt.Println("Redirecting to login page") // Add this line for debugging
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tpl := template.Must(template.ParseFiles("templates/admin.html"))
	tpl.Execute(w, username)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	username, ok := session.Values["username"].(string)

	if !ok || username != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		//Handle delete logic
		r.ParseForm()
		userID := r.FormValue("user_id")
		_, err := db.Exec("DELETE FROM login_data WHERE id = $1", userID)
		if err != nil {
			fmt.Fprintln(w, "Error deleting user:", userID)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	tpl := template.Must(template.ParseFiles("templates/delete.html"))
	tpl.Execute(w, username)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	delete(session.Values, "username")
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
