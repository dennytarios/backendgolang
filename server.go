package main

import (
	"fmt"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"database/sql"
	"log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"context"
)

var (
	googleOAuthConfig = oauth2.Config{
		ClientID:     "239099467796-cjai71c4glp852dhq6idnhk8cmipkoil.apps.googleusercontent.com",
		ClientSecret: "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",
		RedirectURL:  "http://localhost:8080/callback", // Ganti dengan URL callback yang sesuai
		Scopes:       []string{"profile", "email"},
		Endpoint:     google.Endpoint,
	}

	oauthStateString = "randomstring"
)

const (
    host     = "localhost"
    port     = "5432"
    user     = "postgres"
    password = "admin"
    dbname   = "backendgolang"
)



func main() {
	
	r := mux.NewRouter()

	// Handle root / default route
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/about", AboutHandler)
	r.HandleFunc("/search", SearchHandler).Methods("GET")
	// r.HandleFunc("/login", LoginPageHandler).Methods("GET")
	// r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/dashboard", DashboardHandler)
	r.HandleFunc("/privacy-policy", PrivacyHandler)
	r.HandleFunc("/tos", TosHandler)
	r.HandleFunc("/login", GoogleLoginHandler)
	r.HandleFunc("/callback", GoogleCallbackHandler)

	http.Handle("/", r)
	fmt.Println("Server ready")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := googleOAuthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Println("Invalid OAuth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("Code exchange failed: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Anda dapat menggunakan token.AccessToken untuk mengambil informasi pengguna dari Google.
	// Selanjutnya, tangani informasi pengguna sesuai kebutuhan Anda.

	fmt.Printf("Access Token: %s\n", token.AccessToken)

	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}


func TosHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/tos.html")
}

func PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/privacy.html")
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("About"))
}

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/login.html")
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/dashboard.html")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	// String koneksi
	const connStr = "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"
	db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

	// Mengelola data masukan pengguna dari form login
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	// Anda dapat melakukan autentikasi pengguna di sini
	// Misalnya, memeriksa username dan password dengan data yang valid
	// Maksudnya, diquery dari basis data pengguna.

	// Jika autentikasi berhasil, Anda dapat mengarahkan pengguna ke halaman lain
	// atau memberikan respons sukses
	var dbUsername, dbPassword string 
	err = db.QueryRow("select username, password from users where username = $1", username).Scan(&dbUsername, &dbPassword)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Autentikasi gagal", http.StatusUnauthorized)
		return
	}

	 // Verifikasi kata sandi
	 if password != dbPassword {
		http.Error(w, "Kata sandi salah", http.StatusUnauthorized)
		return
	}
	
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	query := vars.Get("q")
	sA := vars.Get("a")
	sB := vars.Get("b")

	a, errA := strconv.Atoi(sA)
	b, errB := strconv.Atoi(sB)

	if errA != nil || errB != nil {
		http.Error(w, "Parameter a dan b harus berupa bilangan", http.StatusBadRequest)
		return
	}

	c := a + b
	responseMessage := fmt.Sprintf("Hasil pencarian untuk: %s. Penjumlahan: %d+%d=%d", query, a, b, c)
	w.Write([]byte(responseMessage))
}
