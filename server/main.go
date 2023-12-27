package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type User struct {
	Username   string `json:"username"`
	Authorized bool   `json:"authorized"`
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
		str_message := strings.TrimSpace(string(msg))
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

				token := strings.Split(str_message, " ")
				if len(token) != 3 {
					err := ws.WriteMessage(websocket.TextMessage, []byte("Invalid input"))
					if err != nil {
						ws.Close()
						delete(clients, ws)
						return
					}
				}

				username := token[1]
				password := token[2]

				for _, user := range db_users {
					if user.Username == username {
						if user.Password == password {
							clients[ws] = User{Username: username, Authorized: true}
							err := ws.WriteMessage(websocket.TextMessage, []byte("Welcome "+username))
							if err != nil {
								ws.Close()
								delete(clients, ws)
								return
							}
						} else {
							err := ws.WriteMessage(websocket.TextMessage, []byte("Wrong password"))
							if err != nil {
								ws.Close()
								delete(clients, ws)
								return
							}
						}
						break
					}
				}
			} else if str_message == "/help" {
				help_message := []string{
					"available commands:",
					"\t/list: list all connected users",
					"\t/login <username> <password>: login with username and password",
					"\t/register <username> <password>: register with username and password",
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
			} else if len(str_message) >= 10 && str_message[:9] == "/register" {
				token := strings.Split(str_message, " ")
				if len(token) != 3 {
					err := ws.WriteMessage(websocket.TextMessage, []byte("Invalid input"))
					if err != nil {
						ws.Close()
						delete(clients, ws)
						return
					}
				}

				username := token[1]
				password := token[2]

				db_users = append(db_users, DB_User{UserID: len(db_users) + 1, Username: username, Password: password, Verified: false})

				clients[ws] = User{Username: username, Authorized: true}
				err := ws.WriteMessage(websocket.TextMessage, []byte("Welcome "+username))
				if err != nil {
					ws.Close()
					delete(clients, ws)
					return
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
	PORT := "8080"
	if os.Getenv("PORT") != "" {
		PORT = os.Getenv("PORT")
	}
	log.Println("Starting server at port " + PORT + "...")
	log.Println("Users:", db_users)

	http.HandleFunc("/ws", handleConnections)
	go handleMessages()

	err := http.ListenAndServe(":"+PORT, nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
