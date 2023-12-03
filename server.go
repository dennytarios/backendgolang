package main

import (
	"fmt"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	// Handle root / default route
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	r.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("About"))
	})

	r.HandleFunc("/search", SearchHandler).Methods("GET")

	http.Handle("/", r)
	fmt.Println("Server ready")
	http.ListenAndServe(":8989", nil)

}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	query := vars.Get("q")
	sA := vars.Get("a")
	sB := vars.Get("b")
	
	a, errA  := strconv.Atoi(sA)
	b, errB := strconv.Atoi(sB)

	if errA != nil || errB != nil {
        http.Error(w, "Parameter a dan b harus berupa bilangan", http.StatusBadRequest)
        return
    }

	c := a + b
	responseMessage := fmt.Sprintf("Hasil pencarian untuk: %s. Penjumlahan: %d+%d=%d", query,a,b,c)
	w.Write([]byte(responseMessage))
}
