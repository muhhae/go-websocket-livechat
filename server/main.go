package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type User struct {
	Username string `json:"username"`
}

type Message struct {
	Sender  User   `json:"sender"`
	Message string `json:"message"`
}

var clients = make(map[*websocket.Conn]User)
var broadcast = make(chan Message)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		panic(err)
	}

	defer ws.Close()

	clients[ws] = User{Username: "Anonymous"}

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message: ", err)
			delete(clients, ws)
			break
		}
		str_message := string(msg)
		log.Println(clients[ws].Username + ": " + str_message)

		if str_message[0] == '/' {
			if str_message == "/list" {
				err := ws.WriteMessage(websocket.TextMessage, []byte("Connected users:\n"))
				if err != nil {
					ws.Close()
					delete(clients, ws)
					return
				}
				for client := range clients {
					err := ws.WriteMessage(websocket.TextMessage, []byte("\t*"+clients[client].Username))
					if err != nil {
						ws.Close()
						delete(clients, ws)
						return
					}
				}
			} else if len(str_message) >= 6 && str_message[:6] == "/login" {
				client := clients[ws]
				client.Username = str_message[6:]
				clients[ws] = client

				err := ws.WriteMessage(websocket.TextMessage, []byte("Your username is now "+client.Username))
				if err != nil {
					ws.Close()
					delete(clients, ws)
					return
				}
			} else if str_message == "/help" {
				help_message := []string{
					"available commands:",
					"\t/list: list all connected users",
					"\t/login <username>: change your username",
					"\t/help: show this help message",
				}
				for _, str := range help_message {
					err := ws.WriteMessage(websocket.TextMessage, []byte(str))
					if err != nil {
						ws.Close()
						delete(clients, ws)
						return
					}
				}
			} else {
				err := ws.WriteMessage(websocket.TextMessage, []byte("Unknown command"))
				if err != nil {
					ws.Close()
					delete(clients, ws)
					return
				}
			}
		} else {
			message := Message{Sender: clients[ws], Message: string(msg)}
			broadcast <- message
		}
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

func main() {
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	http.ListenAndServe(":8080", nil)
}
