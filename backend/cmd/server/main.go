package main

import (
	"log"
	"net/http"

	"chatCall/internal/signal"
)

func main() {
	hub := signal.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		signal.ServeWS(hub, w, r)
	})

	log.Println("🚀 Voice server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
