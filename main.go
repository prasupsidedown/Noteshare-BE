package main

import (
	"fmt"
	"net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello dari Go 🚀")
    })

    fmt.Println("Server jalan di http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}