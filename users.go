package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/paulstuart/secrets"
)

var (
	errEmptyCookie = fmt.Errorf("user cookie is empty")
)

type userLevel struct {
	ID   int
	Name string
}

var userLevels = []userLevel{{0, "User"}, {1, "Editor"}, {2, "Admin"}}

func userByID(id interface{}) (user, error) {
	return getUser("where usr=?", id)
}

func userByLogin(login string) (user, error) {
	return getUser("where login=?", login)
}

func userByEmail(email string) (user, error) {
	return getUser("where email=?", email)
}

func (user *user) Cookie() string {
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

func (user *user) FromCookie(cookie string) error {
	if len(cookie) == 0 {
		return errEmptyCookie
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
	u, err := userByEmail(username)
	if err != nil {
		fmt.Println("User error:", err)
		return ""
	}
	return u.Cookie()
}

func userFromCookie(cookie string) user {
	u := &user{}
	// ignore errors -- just return blank user if no cookie set
	u.FromCookie(cookie)
	return *u
}

type credentials struct {
	Username, Password string
}

func userAuth(username, password string) (*user, error) {
	user := fullUser{}
	if err := dbFindBy(&user, "email", username); err != nil {
		log.Println("user error:", err)
		return nil, fmt.Errorf("%s is not authorized for access", username)
	}
	if user.Local {
		if user.Email == username && notNull(user.Password) == password {
			return user.User(), nil
		}
		return nil, fmt.Errorf("invalid login")
	}
	if authenticate(username, password) {
		return user.User(), nil
	}
	return nil, fmt.Errorf("invalid credentials for %s", username)
}

// TODO: should probably cache results in a safe map
func userFromAPIKey(key string) (user, error) {
	if cfg.Main.NoKey {
		login := "open acccess"
		return user{Email: login, Level: 2}, nil
	}
	return getUser("where apikey=?", key)
}
