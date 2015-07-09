package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/html"
)

var (
	domain   string
	fromURI  string
	remember = "true"
	login    = "Sign+In"
)

func saveToFile(name string, text []byte) {
	ioutil.WriteFile(name, text, 0644)
}

func SAMLSession(body io.Reader) string {
	doc, err := html.Parse(body)
	if err != nil {
		log.Fatal(err)
	}
	var found string
	var find func(*html.Node)
	find = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			ok := false
			value := ""
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "SAMLResponse" {
					ok = true
				}
				if a.Key == "value" {
					value = a.Val
				}
			}
			if ok {
				found = value
				return
			}
		}
		for c := n.FirstChild; c != nil && len(found) == 0; c = c.NextSibling {
			find(c)
		}
	}
	find(doc)
	if len(found) > 0 {
		data, err := base64.StdEncoding.DecodeString(found)
		if err != nil {
			fmt.Println("error:", err)
		}
		return string(data)
	}
	return found
}

func OktaAuth(username, password string) string {
	data := url.Values{
		"username":   {username},
		"password":   {password},
		"_xsrfToken": {cfg.SAML.Token},
		"fromURI":    {""},
		"remember":   {"true"},
		"login":      {"Sign+In"},
	}

	var options cookiejar.Options
	jar, cerr := cookiejar.New(&options)
	if cerr != nil {
		log.Fatal(cerr)
	}
	client := http.Client{Jar: jar}
	resp, err := client.PostForm(cfg.SAML.Login, data)

	if err != nil {
		fmt.Println("AUTH ERR:", err)
		return ""
	}
	/*
		body, _ := ioutil.ReadAll(resp.Body)
		ioutil.WriteFile("login.html", body, 0644)
		resp.Body.Close()
	*/

	resp, err = client.PostForm(cfg.SAML.URL, data)
	if err != nil {
		fmt.Println("POST ERR:", err)
	}
	if resp != nil {
		defer resp.Body.Close()
		return SAMLSession(resp.Body)
	}
	return ""
}

func Authenticate(username, password string) bool {
	reply := OktaAuth(username, password)
	//ioutil.WriteFile("saml.xml", []byte(reply), 0644)
	return len(reply) > 0
}
