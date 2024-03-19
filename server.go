package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	firebase "firebase.google.com/go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

const (
	host     = "localhost"
	port     = "5432"
	user     = "postgres"
	password = "admin"
	dbname   = "backendgolang"
)

func main() {
	initFirebaseAdmin()

	r := mux.NewRouter()

	// Handle root / default route
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/about", AboutHandler)
	r.HandleFunc("/search", SearchHandler).Methods("GET")
	r.HandleFunc("/login", LoginPageHandler).Methods("GET")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/dashboard", DashboardHandler)
	r.HandleFunc("/privacy-policy", PrivacyHandler)
	r.HandleFunc("/tos", TosHandler)

	http.Handle("/", r)
	fmt.Println("Server ready")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.HandleFunc("/style/global.css", func(w http.ResponseWriter, r *http.Request) {
		// Set tipe MIME menjadi 'text/css'
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "static/style/global.css")
	})
	http.ListenAndServe(":8080", nil)
}

func initFirebaseAdmin() {
	ctx := context.Background()

	// Konfigurasi Firebase Admin SDK dengan file konfigurasi yang diunduh dari Firebase Console
	opt := option.WithCredentialsFile("golangbackend-2cc64-firebase-adminsdk-wjnyf-cb03a5544d.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
		return
	}

	// Inisialisasi client Firebase Auth
	client, err := app.Auth(ctx)
	_ = client
	if err != nil {
		log.Fatalf("error creating Auth client: %v\n", err)
		return
	}
	fmt.Println("Firebase Admin ready")

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
