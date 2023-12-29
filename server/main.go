package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	PORT := "8080"
	if os.Getenv("PORT") != "" {
		PORT = os.Getenv("PORT")
	}
	log.Println("Starting server at ", PORT, "...")

	MyHttpHandle("/register", register, cors)
	MyHttpHandle("/login", login, cors)
	MyHttpHandle("/ws", handleConnections)
	MyHttpHandle("/logout", logout, cors)

	go handleMessages()

	err := http.ListenAndServe(":"+PORT, nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}
