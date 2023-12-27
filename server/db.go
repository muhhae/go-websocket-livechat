package main

type DB_User struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Verified bool   `json:"verified"`
}

var db_users = []DB_User{
	{
		UserID: 1,
		Username: "admin",
		Password: "admin",
	},
	{
		UserID: 2,
		Username: "user",
		Password: "user",
	},
}
