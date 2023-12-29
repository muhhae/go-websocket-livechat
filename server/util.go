package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func SendJson(w http.ResponseWriter, status int, body interface{}) {
	jsonResponse, err := json.Marshal(body)
	if err != nil {
		log.Println("Error encoding json:", err)
		return
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Println("Error writing response:", err)
	}
}

func GetCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		log.Println("Error getting cookie:", err)
		return ""
	}
	return cookie.Value
}

func ReqBodyToMap(r *http.Request) map[string]interface{} {
	body := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Println("Error decoding json:", err)
		return nil
	}
	return body
}

func GetBodyField(body map[string]interface{}, field string) string {
	value, ok := body[field]
	if !ok {
		log.Println("Error getting field:", field)
		return ""
	}
	return value.(string)
}
