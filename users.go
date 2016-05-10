package main

import (
	"encoding/json"
	"fmt"

	"github.com/paulstuart/secrets"
)

var (
	ErrEmptyCookie = fmt.Errorf("user cookie is empty")
)

func UserByID(id interface{}) (User, error) {
	return getUser("where id=?", id)
}

func userLogin(id string) string {
	if len(id) == 0 {
		return ""
	}
	if id == "0" {
		return ""
	}
	u, err := UserByID(id)
	if err != nil {
		return err.Error()
	}
	return u.Login
}

func UserByLogin(login string) (User, error) {
	return getUser("where login=?", login)
}

func UserByEmail(email string) (User, error) {
	return getUser("where email=?", email)
}

func (user *User) Cookie() string {
	text, e1 := json.Marshal(user)
	if e1 != nil {
		fmt.Println("Marshal user", user, "Error", e1)
		return ""
	}
	secret, e2 := secrets.EncryptString(string(text))
	if e2 != nil {
		fmt.Println("Encrypt text", text, "Error", e2)
		return ""
	}
	return secret
}

func (user *User) FromCookie(cookie string) error {
	if len(cookie) == 0 {
		return ErrEmptyCookie
	}
	plain, err := secrets.DecryptString(cookie)
	if err != nil {
		return fmt.Errorf("Decrypt text: %s error: %s", cookie, err)
	}
	if err = json.Unmarshal([]byte(plain), &user); err != nil {
		return fmt.Errorf("unmarshal text: %s error: %s", plain, err)
	}
	return nil
}

func userCookie(username string) string {
	u, err := UserByEmail(username)
	if err != nil {
		fmt.Println("User error:", err)
		return ""
	}
	return u.Cookie()
}

func userFromCookie(cookie string) User {
	u := &User{}
	// ignore errors -- just return blank user if no cookie set
	u.FromCookie(cookie)
	return *u
}
