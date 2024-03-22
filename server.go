package main

// Package yang digunakan
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"

	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
)

var templates *template.Template
var store = sessions.NewCookieStore([]byte("your-secret-key"))

// Konfigurasi database postgre
const (
	host     = "localhost"
	port     = "5432"
	user     = "postgres"
	password = "admin"
	dbname   = "backendgolang"
)

func clearSessionHandler(w http.ResponseWriter, r *http.Request) {
    // Mendapatkan session
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Mengatur MaxAge menjadi -1 memaksa penghapusan cookie
    session.Options.MaxAge = -1

    // Menyimpan perubahan untuk menerapkan penghapusan
    err = session.Save(r, w)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Redirect atau memberikan response lain setelah "membersihkan" session
}



// Fungsi yang diakses pertama kali oleh go
func main() {
	devMode := os.Getenv("DEV_MODE") == "true" // TODO: mungkin akan kepakai

	initFirebaseAdmin()
	precompileTemplate()

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
	r.HandleFunc("/verify-token", VerifyTokenHandler).Methods("POST")

	if devMode{
		r.HandleFunc("/del", clearSessionHandler)
	}

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

// TODO: jadikan function ini bisa dipakai banyak variable session
func createSession(w http.ResponseWriter, r *http.Request, kv ...interface{}) {
    // Membuat atau mengambil session
    session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, "Gagal mendapatkan session", http.StatusInternalServerError)
        return
    }

    // Memastikan jumlah argumen kv adalah genap (key-value pairs)
    if len(kv)%2 != 0 {
        log.Println("createSession error: Argumen harus berpasangan (key-value)")
        return
    }

    // Menyimpan pasangan key-value ke dalam session
    for i := 0; i < len(kv); i += 2 {
        key, ok := kv[i].(string)
        if !ok {
            log.Println("createSession error: Key harus bertipe string")
            return
        }
        session.Values[key] = kv[i+1]
    }

    // Simpan perubahan session
    err = session.Save(r, w)
    if err != nil {
        // Handle error
        log.Printf("createSession error: %v\n", err)
    }
}


func precompileTemplate() {
	// Menggunakan template.ParseGlob untuk memuat semua file template
	templates = template.Must(template.ParseGlob("static/*.html"))
}

var FirebaseAuthClient *auth.Client

// Membuat aplikasi ini bisa mengakses layanan firebase
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
	FirebaseAuthClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("error creating Auth client: %v\n", err)
		return
	}
	fmt.Println("Firebase Admin ready")
}

// Dibutuhkan oleh Google Authentication
func TosHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/tos.html")
}

// Dibutuhkan oleh Google Authentication
func PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/privacy.html")
}

// Akan digantikan oleh static page
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("About"))
}

// Form login yang tidak akan dipakai lagi
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/login.html")
}

// Ini nanti hanya bisa diakses oleh member yang berhasil login
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/dashboard.html")
}

// ini sekedar menunjukkan bagaimana mengakses database PosgreSQL
// tidak untuk dicontoh karena password tentu tidak boleh disimpan as is
// wajib di hash (minimal) atau di encrypt
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

// Halaman utama aplikasi
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Ubah dari sekedar serving static file ke bentuk templating engine
	//http.ServeFile(w, r, "static/index.html")

	session, err := store.Get(r, "session-name")
    if err != nil {
        http.Error(w, "Gagal mendapatkan session", http.StatusInternalServerError)
        return
    }

	isLoggedIn := false // Default false sampai dibuktikan sebaliknya
    if _, ok := session.Values["user_id"]; ok {
        isLoggedIn = true
    }
	data := struct {
        IsLoggedIn bool
    }{
        IsLoggedIn: isLoggedIn,
    }

	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {
    // Struct untuk mem-parsing request body
    var requestBody struct {
        Token string `json:"token"`
    }
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
	ctx := r.Context()
    // Verifikasi ID token
    token, err := FirebaseAuthClient.VerifyIDToken(ctx, requestBody.Token)
    if err != nil {
        http.Error(w, "Invalid ID token", http.StatusUnauthorized) //TODO: jika ini terjadi, suruh login ulang.
        return
    }

    // Token valid, atur session untuk user
    createSession(w, r, "user_id", token.UID) // Fungsi createSession dari contoh sebelumnya

    // Kirim response sukses ke client
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Fungsi ini sekedar menunjukkan bagaimana cara membaca parameter dari request
// Contoh: http://localhost:8080/search?a=2&b=3&q=eko akan menghasilkan output:
// Hasil pencarian untuk: eko. Penjumlahan: 2+3=5
// Perhatikan cara mengakses nilai q, a dan b
// Go bisa mendeklarasikan dan sekaligus menginisialisasi 
//    sA := vars.Get("a")
// adalah deklarasi sekaligus inisialisasi
// Bisa seperti ini:
//    var sA string
//    sA = vars.Get("a")
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
