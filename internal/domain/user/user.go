/*
---------------------------------------------------------------------------------------
File: user.go
Autor: Aldenor
Data: 29-04-2026
----------------------------------------------------------------------------------------
Finalidade:
Entidades do domínio "user"
---------------------------------------------------------------------------------------
*/
package user

import "time"

type UserID int64

type User struct {
	UserID    UserID    `json:"user_id"`
	Userrole  string    `json:"userrole"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserCreate struct {
	Userrole string `json:"userrole"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type UserUpdate struct {
	Userrole string `json:"userrole"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email"`
}

type UserSearch struct {
	NmSearch string `json:"nm_search,omitempty"`
	NrDocs   int    `json:"nr_docs,omitempty"`
}
