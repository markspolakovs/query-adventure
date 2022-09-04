package auth

import "encoding/gob"

type UserData struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func init() {
	gob.Register(UserData{})
}
