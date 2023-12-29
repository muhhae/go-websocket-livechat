package main

import (
	"log"
	"net/http"
	"time"
)

func setSessionCookie(w http.ResponseWriter, value string) {
	var session_cookie = http.Cookie{
		Name:     "haechat-token",
		Value:    value,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, &session_cookie)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "invalid method",
		})
		return
	}

	body := ReqBodyToMap(r)

	username := GetBodyField(body, "username")
	password := GetBodyField(body, "password")

	if username == "" || password == "" {
		SendJson(w, http.StatusUnauthorized, map[string]interface{}{
			"status": "error",
			"error":  "empty username or password",
		})
		return
	}

	for _, user := range db_users {
		if user.Username == username {
			if user.Password == password {
				setSessionCookie(w, "DUMMY_TOKEN")
				SendJson(w, http.StatusOK, map[string]interface{}{
					"status": "success",
				})
				log.Println("User", username, "logged in")
			} else {
				SendJson(w, http.StatusUnauthorized, map[string]interface{}{
					"status": "error",
					"error":  "invalid password",
				})
			}
			return
		}
	}
	SendJson(w, http.StatusUnauthorized, map[string]interface{}{
		"status": "error",
		"error":  "invalid username or password",
	})
}

func register(w http.ResponseWriter, r *http.Request) {
	body := ReqBodyToMap(r)

	username := GetBodyField(body, "username")
	password := GetBodyField(body, "password")

	if username == "" || password == "" {
		SendJson(w, http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "invalid username or password",
		})
		return
	}

	for _, user := range db_users {
		if username == user.Username {
			SendJson(w, http.StatusConflict, map[string]interface{}{
				"status": "error",
				"error":  "username already exists",
			})
			return
		}
	}
	db_users = append(db_users, DB_User{
		UserID:   len(db_users) + 1,
		Username: username,
		Password: password,
		Verified: false,
	})
	setSessionCookie(w, "DUMMY_TOKEN")
	SendJson(w, http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}

func logout(w http.ResponseWriter, r *http.Request) {
	setSessionCookie(w, "")
	SendJson(w, http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}
