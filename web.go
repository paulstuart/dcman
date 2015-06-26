package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	ttext "text/template"
	"time"

	gorilla "github.com/gorilla/handlers"
)

const (
	logDir    = "logs"
	accessLog = "access.log"
	errorLog  = "error.log"
	cookieID  = "dcuser"
)

var (
	ip          = MyIp()
	htmlTmpl    map[string]*template.Template
	textTmpl    map[string]*ttext.Template
	tdir        = "assets/templates"
	http_server string
	//cookie_store                 = sessions.NewCookieStore([]byte("I can has cookies!"))
	errorFile *os.File
)

type HFunc struct {
	Path string
	Func http.HandlerFunc
}

func RemoteHost(r *http.Request) string {
	if remote_addr := r.Header.Get("X-Forwarded-For"); len(remote_addr) > 0 {
		return remote_addr
	}
	remote_addr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "REMOTE ADDR ERR:", err)
	}
	if len(remote_addr) > 0 && remote_addr[0] == ':' {
		remote_addr = MyIp()
	}
	return remote_addr
}

func loadTemplates() {
	//fmt.Println("LOAD TEMPLATES DIR:", tdir)
	funcMap := template.FuncMap{
		"isTrue": isTrue,
	}
	htmlTmpl = make(map[string]*template.Template)
	textTmpl = make(map[string]*ttext.Template)
	files, err := filepath.Glob(tdir + "/*.*")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		name := filepath.Base(file)
		if name == "base.html" {
			//fmt.Println("skipping base.html")
			continue
		}
		if strings.HasSuffix(name, ".html") {
			t := template.New(name).Funcs(funcMap)
			htmlTmpl[name] = template.Must(t.ParseFiles(file, tdir+"/base.html"))
			continue
		}
		t := ttext.New(name)
		textTmpl[name] = ttext.Must(t.ParseFiles(file))
	}
}

func isTrue(in interface{}) string {
	yes, err := strconv.ParseBool(in.(string))
	if err != nil {
		return err.Error()
	}
	if yes {
		return "true"
	}
	return "false"
}

// render a template that inherits the "base" template
func renderTemplate(w http.ResponseWriter, r *http.Request, tname string, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	name := string(tname + ".html")
	err := htmlTmpl[name].ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// render a template with no inheritence
func renderPlainTemplate(w http.ResponseWriter, r *http.Request, tname string, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	name := string(tname + ".html")
	err := htmlTmpl[name].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderTextTemplate(w http.ResponseWriter, r *http.Request, tname string, data interface{}) {
	w.Header().Set("Content-Type", "text/plain")
	t := textTmpl[tname]
	if t == nil {
		http.Error(w, "no template found for: "+tname, http.StatusInternalServerError)
		log.Println("no template found for: ", tname)
		return
	}
	err := t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func currentUser(r *http.Request) User {
	cookie, err := r.Cookie(cookieID)
	if err != nil {
		return User{}
	}
	return userFromCookie(cookie.Value)
}

func reloadPage(w http.ResponseWriter, r *http.Request) {
	loadTemplates()
	fmt.Fprintln(w, "reloaded")
}

func internalLoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		login := r.Form["login"][0]
		password := r.Form["password"][0]
		if !Authenticate(login, password) {
			http.Redirect(w, r, "/loginfail", 302)
		} else {
			expires := time.Now()
			expires = expires.Add(time.Duration(time.Hour) * 24 * 365)
			c := http.Cookie{Name: "login", Value: login, Expires: expires}
			http.SetCookie(w, &c)
			http.Redirect(w, r, "/", 302)
		}
	} else {
		err := htmlTmpl["login.html"].ExecuteTemplate(w, "base", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func FaviconPage(w http.ResponseWriter, r *http.Request) {
	fav := filepath.Join(assets_dir, "static/images/favicon.ico")
	http.ServeFile(w, r, fav)
}

func InvalidPage(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func init() {
	assets_dir, _ = filepath.Abs(assets_dir) // CWD may change at runtime
	tdir, _ = filepath.Abs(tdir)             // CWD may change at runtime
}

/*
type statusLoggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusLoggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
*/

func StaticPage(w http.ResponseWriter, r *http.Request) {
	name := filepath.Join(assets_dir, r.URL.Path)
	file, err := os.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	w.Header().Set("Cache-control", "public, max-age=259200")
	http.ServeContent(w, r, name, fi.ModTime(), file)
}

func ErrorLog(r *http.Request, msg string, args ...interface{}) {
	user := currentUser(r)
	remote_addr := RemoteHost(r)
	fmt.Fprintln(errorFile, time.Now().Format(log_layout), remote_addr, user.ID, msg, args)
}

func userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.User == nil {
			user := currentUser(r)
			if user.ID > 0 {
				r.URL.User = url.User(user.Login)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func oktaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			cookie, err := r.Cookie(cfg.SAML.OKTACookie)
			//TODO: we should check the cookie value to be valid, not just any string
			auth := (err == nil && len(cookie.Value) > 0)
			if !auth {
				redirect(w, r, "/login", 302)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func redirect(w http.ResponseWriter, r *http.Request, path string, status int) {
	http.Redirect(w, r, pathPrefix+path, status)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	redirect(w, r, "/error", http.StatusNotFound)
}

func webServer(handlers []HFunc) {
	loadTemplates()
	for _, h := range handlers {
		p := pathPrefix + h.Path
		switch {
		case strings.HasPrefix(h.Path, "/login"):
			http.Handle(p, http.StripPrefix(p, h.Func))
		case strings.HasPrefix(h.Path, "/static/"):
			http.Handle(p, http.StripPrefix(pathPrefix, h.Func))
		case strings.HasPrefix(h.Path, "/data/"):
			http.Handle(p, http.StripPrefix(p, h.Func))
		case strings.HasPrefix(h.Path, "/api/"):
			http.Handle(p, h.Func)
		default:
			http.Handle(p, http.StripPrefix(p, oktaMiddleware(h.Func)))
		}
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Panic(err)
	}
	logFile := filepath.Join(logDir, accessLog)
	accessLog, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Panic("Error opening access log:", err)
	}
	errorPath := filepath.Join(logDir, "error.log")
	errorFile, err = os.OpenFile(errorPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Panic("Error opening error log:", err)
	}

	http_server = fmt.Sprintf(":%d", cfg.Main.Port)
	fmt.Printf("serve up web: http://%s%s/\n", ip, http_server)
	err = http.ListenAndServe(http_server, gorilla.CompressHandler(userMiddleware(gorilla.LoggingHandler(accessLog, http.DefaultServeMux))))
	if err != nil {
		panic(err)
	}
}
