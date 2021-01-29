package main

import (
	"database/sql"
)

type login struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

type account struct {
	UserID string `json:"user_id"`
	Admin  bool   `json:"admin"`
	Email  string `json:"email"`
}

func (p *login) authenticate(db *sql.DB) (account, error) {
	var token account
	err := db.QueryRow("SELECT user_id, admin, email FROM booking.account WHERE user_id=$1 AND password = crypt($2, password)",
		p.UserID, p.Password).Scan(&token.UserID, &token.Admin, &token.Email)

	return token, err
}
