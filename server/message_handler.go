package main

import (
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
	Sender   User      `json:"sender"`
	// Receiver User      `json:"receiver"`
	Message  string    `json:"message"`
	Date     time.Time `json:"date"`
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
	cookie_token, err := r.Cookie("haechat-token")
	if err != nil {
		log.Println("Error getting cookie:", err)
		return
	}
	token := cookie_token.Value
	if token == "" {
		SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "empty token",
		})
		return
	}

	if token != "DUMMY_TOKEN" {
		SendJson(w, http.StatusUnauthorized, map[string]interface{}{
			"status": "error",
			"error":  "invalid token",
		})
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer ws.Close()

	clients[ws] = User{}
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message: ", err)
			delete(clients, ws)
			break
		}
		str_message := strings.TrimSpace(string(msg))
		log.Println("Anonymous :" + str_message)
		message := Message{Sender: clients[ws], Message: string(msg)}
		broadcast <- message
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg.Sender.Username+": "+msg.Message))
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}
