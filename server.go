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
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/sessions"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var templates *template.Template
var store = sessions.NewCookieStore([]byte("your-secret-key"))
var devMode bool

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
	devMode = os.Getenv("DEV_MODE") == "true" // TODO: mungkin akan kepakai

	initFirebaseAdmin()
	if !devMode {
		precompileTemplate()
	}
	r := mux.NewRouter()

	r.Use(sessionMiddleware)
	// Handle root / default route
	//r.HandleFunc("/", DummyHandler)
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/about", AboutHandler)
	r.HandleFunc("/search", SearchHandler).Methods("GET")
	r.HandleFunc("/login", LoginPageHandler).Methods("GET")
	r.HandleFunc("/login", LoginHandler).Methods("POST")
	r.HandleFunc("/dashboard", DashboardHandler)
	r.HandleFunc("/privacy-policy", PrivacyHandler)
	r.HandleFunc("/tos", TosHandler)
	r.HandleFunc("/verify-token", VerifyTokenHandler).Methods("POST")
	r.HandleFunc("/daftar", DaftarHandler)

	if devMode {
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
		log.Printf("Key: %s, Value: %v\n", key, kv[i+1])
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
	templateNames := templates.Templates()
	fmt.Println("Nama-nama template yang terdaftar:")
	for _, t := range templateNames {
		fmt.Println(t.Name())
	}

}

var FirebaseAuthClient *auth.Client
var FirestoreClient *firestore.Client

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

	FirestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error creating Firestore client: %v\n", err)
		return
	}
	fmt.Println("Firebase Auth | Firestore ready")
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

func DummyHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DummyHandler")
	w.Write([]byte("DummyHandler"))
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

// NOTE: berguna untuk membuat ringkas kode yang mengambil nilai dan atau menset nilai dari/ke session
func SetSessionValue(session *sessions.Session, key string, defaultValue interface{}) interface{} {
	if val, ok := session.Values[key]; ok {
		return val
	} else {
		session.Values[key] = defaultValue
		return defaultValue
	}
}

func GetSessionValue(session *sessions.Session, key string) interface{} {
	if val, ok := session.Values[key]; ok {
		return val
	}
	return nil
}

// TODO: Buat middleware ini mempersiapkan semua variable session yang PASTI dipakai
// di semua route. Nanti di tiap route dia akan mengekstrak data session yang sesuai
// Ini berarti harus ada IF untuk pemanggilan jika user BARU PERTAMA KALI membuka web ini.
// TODO kenapa ssession middleware dan routenya sptnya dipanggil 2x?
func sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("SessionMW")
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, "Gagal mendapatkan session", http.StatusInternalServerError)
			return
		}
		_, ok := session.Values["newcomer"]

		// User pertama kali membuka situs ini, set newcomer = true
		if !ok {
			log.Println("1st time")
			session.Values["newcomer"] = true

			//Gunakan tempat ini untuk melakukan setting nilai awal utnuk pengunjung baru
			_ = SetSessionValue(session, "isLoggedIn", false)
			_ = SetSessionValue(session, "user_name", "Default Name").(string)
			_ = SetSessionValue(session, "user_email", "email@default.com").(string)
			_ = SetSessionValue(session, "user_photo", "static/img/v1.jpeg").(string)
			_ = SetSessionValue(session, "user_membership", "").(string)

			err = session.Save(r, w)
			if err != nil {
				// Handle error saat menyimpan session
				log.Printf("Error saving session: %v", err)
			}

		}
		// Simpan session ke context request agar bisa diakses di handler selanjutnya
		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Halaman utama aplikasi
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("HomeHandler")
	//NOTE: Dulu data disiapkan disini sebagai variable local di function ini
	//Sekarang data diambil dari session, dimana session disiapkan di sessionMiddleWare
	//NOTE: lanjutkan sessionMiddleWare agar data bisa disiapkan dengan variable dari
	//session
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Gagal mendapatkan session", http.StatusInternalServerError)
		return
	}

	data := struct {
		IsLoggedIn bool
		DevMode    bool
		Name       string
		Email      string
		Photo      string
		Membership string
		TimeStamp  time.Time
	}{
		IsLoggedIn: GetSessionValue(session, "isLoggedIn").(bool),
		DevMode:    devMode,
		Name:       GetSessionValue(session, "user_name").(string),
		Email:      GetSessionValue(session, "user_email").(string),
		Photo:      GetSessionValue(session, "user_photo").(string),
		Membership: GetSessionValue(session, "user_membership").(string),
		TimeStamp:  time.Now(),
	}

	if devMode {
		tmpl, err := template.ParseFiles("static/base.html", "static/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	} else {
		log.Println("Production code /")
		err = templates.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

/*
NOTE Di halaman ini peserta bisa melakukan pembayaran untuk join membership.
Peserta memilih jenis membership dari drop down.
Kemudian tampilan akan menyesuaikan berdasarkan dropdown itu tadi.
Untuk payment gateway, akan dilanjutkan di course part-2 dari course ini.
Untuk saat ini, sekedar upload bukti transaksi ke rek BCA/QRIS.
Jika memilih QRIS maka QRcode akan muncul dan bisa disave / discan
Field yang akan disubmit pada saat pendaftaran adalah:
- Nama diambil dari gmail, namun bisa diedit.
- Domisili mungkin sekedar Propinsi.
- Kesibukan saat ini (Pelajar/Mahasiswa/PNS)
- Skill di bidang IT? (Checkboxes)
- Jika ga skill, muncul link ke video materi ttg switching careers DAN juga warningnya
- TTD digital ke PDF hak dan kewajiban.

CHECKBOXES:
- tentang kesadaran join RWID dengan catatan-catatan bukan jalur cepat menjadi kaya
- tidak menjamin mendapat gaji ribuan dollar, namun disupport LIFETIME belajar sampai menembus
- keberhasilan akan kembali ke komitmen dan disiplin belajar perhari
- Total belajar 300-400jam menembus job
*/
func DaftarHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DaftarHandler")
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Gagal mendapatkan session", http.StatusInternalServerError)
		return
	}

	data := struct {
		IsLoggedIn bool
		DevMode    bool
		Name       string
		Email      string
		Photo      string
		Membership string
		TimeStamp  time.Time
	}{
		IsLoggedIn: GetSessionValue(session, "isLoggedIn").(bool),
		DevMode:    devMode,
		Name:       GetSessionValue(session, "user_name").(string),
		Email:      GetSessionValue(session, "user_email").(string),
		Photo:      GetSessionValue(session, "user_photo").(string),
		Membership: GetSessionValue(session, "user_membership").(string),
		TimeStamp:  time.Now(),
	}

	if devMode {
		tmpl, err := template.ParseFiles("static/base.html", "static/daftar.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	} else {
		log.Println("Production code daftar")
		err = templates.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			log.Println("Ada error di daftar")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func VerifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Struct untuk mem-parsing request body
	var requestBody struct {
		Token    string `json:"token"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		PhotoURL string `json:"photoURL"`
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
	// fmt.Println(requestBody.Name)
	// fmt.Println(requestBody.Email)
	// fmt.Println(requestBody.PhotoURL)

	// AMBIL DATA USER YANG LEBIH LENGKAP DARI FIRESOTRE
	ctx = context.Background()
	docRef := FirestoreClient.Collection("user").Doc(token.UID)
	docSnapshot, err := docRef.Get(ctx)

	var userData map[string]interface{}
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Dokumen tidak ditemukan, insert data baru
			userData = map[string]interface{}{
				"membership": "",
			}

			_, err = docRef.Set(ctx, userData)
			if err != nil {
				// Handle error saat insert data
				log.Printf("Gagal menyimpan data pengguna baru: %v", err)
				return
			}
			log.Println("Data pengguna baru berhasil disimpan")
		} else {
			// Handle error lainnya
			log.Printf("Error saat mendapatkan dokumen: %v", err)
			return
		}
	} else if docSnapshot.Exists() {
		log.Println("User sudah ada kan?")
		userData = docSnapshot.Data()

	}

	membershipValue, ok := userData["membership"].(string)
	if ok {
		log.Println("Membership:", membershipValue)
	} else {
		log.Println("Membership tidak ditemukan atau bukan tipe string")
	}

	createSession(w, r,
		"user_id", token.UID,
		"user_name", requestBody.Name,
		"user_email", requestBody.Email,
		"user_photo", requestBody.PhotoURL,
		"user_membership", membershipValue,
	)

	// Kirim response sukses ke client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// Fungsi ini sekedar menunjukkan bagaimana cara membaca parameter dari request
// Contoh: http://localhost:8080/search?a=2&b=3&q=eko akan menghasilkan output:
// Hasil pencarian untuk: eko. Penjumlahan: 2+3=5
// Perhatikan cara mengakses nilai q, a dan b
// Go bisa mendeklarasikan dan sekaligus menginisialisasi
//
//	sA := vars.Get("a")
//
// adalah deklarasi sekaligus inisialisasi
// Bisa seperti ini:
//
//	var sA string
//	sA = vars.Get("a")
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
