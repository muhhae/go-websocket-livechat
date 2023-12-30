package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	Username   string `json:"username"`
	Authorized bool   `json:"authorized"`
}

type Message struct {
	Sender User `json:"sender"`
	// Receiver User      `json:"receiver"`
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

var clients = make(map[*websocket.Conn]User)
var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4 * 1024,
	WriteBufferSize: 4 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second)
	cookie_token, err := r.Cookie("haechat-token")
	if err != nil {
		SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Bad Credentials",
		})
		return
	}
	token := cookie_token.Value
	if token == "" {
		SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Bad Credentials",
		})
		return
	}
	username := verifyToken(token)
	if username == "" {
		SendJson(w, http.StatusUnauthorized, map[string]interface{}{
			"status": "error",
			"error":  "invalid Credentials",
		})
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Error upgrading connection:", err)
		SendJson(w, http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"error":  "Internal Server Error",
		})
		return
	}
	defer ws.Close()

	clients[ws] = User{Username: username}
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message: ", err)
			delete(clients, ws)
			break
		}
		trimmedMessage := strings.TrimSpace(string(msg))
		if trimmedMessage != "" {
			message := Message{Sender: clients[ws], Message: trimmedMessage, Date: time.Now()}
			broadcast <- message
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			message, err := json.Marshal(msg)
			if err != nil {
				log.Println("Error marshalling message: ", err)
				client.Close()
				delete(clients, client)
			}
			err = client.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println("Error writing message: ", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
