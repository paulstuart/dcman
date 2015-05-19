package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"time"

	gorilla "github.com/gorilla/handlers"
)

const (
	logDir     = "logs"
	accessLog  = "access.log"
	errorLog   = "error.log"
	OKTACookie = "SAML"
	OKTAHash   = "VGhpc1Nob3VsZE5vdEJlRGVjb2RlZCoqU0FNUExFKipURVhU"
)

var (
	secure_url, insecure_url, ip string
	tmpl                         map[string]*template.Template
	tdir                         = "assets/templates"
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
	tmpl = make(map[string]*template.Template)
	files, err := filepath.Glob(tdir + "/*.html")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		name := filepath.Base(file)
		if name == "base.html" {
			//fmt.Println("skipping base.html")
			continue
		}
		//fmt.Println("COMPILE: ", name)
		t := template.New(name).Funcs(funcMap)
		//t := template.New(name)
		tmpl[name] = template.Must(t.ParseFiles(file, tdir+"/base.html"))
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

func renderTemplate(w http.ResponseWriter, r *http.Request, tname string, data interface{}) {
	name := string(tname + ".html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := tmpl[name].ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderPlainTemplate(w http.ResponseWriter, r *http.Request, tname string, data interface{}) {
	name := string(tname + ".html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := tmpl[name].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func currentUser(r *http.Request) User {
	cookie, err := r.Cookie("dcuser")
	if err != nil {
		fmt.Println("cookie error for user", err)
		return User{}
	}
	return userFromCookie(cookie.Value)
}

func reloadPage(w http.ResponseWriter, r *http.Request) {
	loadTemplates()
	fmt.Fprintln(w, "reloaded")
}

/*
func loginPage(w http.ResponseWriter, r *http.Request) {
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
		err := tmpl["login.html"].ExecuteTemplate(w, "base", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
*/

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

type statusLoggingResponseWriter struct {
	status int
	http.ResponseWriter
}

func (w *statusLoggingResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func StaticPage(w http.ResponseWriter, r *http.Request) {
	const skip = len(pathPrefix)
	name := filepath.Join(assets_dir, r.URL.Path[skip:])
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

// the site was originally designed to run at the web root
// it's been updated to run at a path below that
// -- this automatically does redirects for legacy links
func usePrefix(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("Redirect", r.URL.Path, "To", pathPrefix+r.URL.Path)
	if r.Method == "POST" {
		http.Redirect(w, r, pathPrefix+r.URL.Path, 307)
	} else {
		http.Redirect(w, r, pathPrefix+r.URL.Path, 302)
	}
}

func oktaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const u = pathPrefix + "/login"
		const s = pathPrefix + "/static"
		if r.Method == "GET" {
			// static pages should always be accessible (login page uses them!)
			if !(len(r.URL.Path) > len(s) && r.URL.Path[:len(s)] == s) {
				//fmt.Println("PATH", r.URL.Path, "CHECK", r.URL.Path[:len(s)])
				//cookie, err := r.Cookie(OKTACookie)
				cookie, err := r.Cookie(OKTACookie)
				//fmt.Println("PATH", r.URL.Path, "Cookie", cookie.Value)
				//TODO: we should check the cookie value to be valid, not just any string
				auth := (err == nil && len(cookie.Value) > 0)
				//if (err != nil && r.URL.Path != u) && len(cookie.Value) {
				//fmt.Println("PATH", r.URL.Path, "Cookie", cookie.Value, "AUTH", auth)
				if !(auth || r.URL.Path == u) {
					http.Redirect(w, r, u, 302)
					return
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func webServer(handlers []HFunc) {
	loadTemplates()
	ip = MyIp()
	http.HandleFunc("/", usePrefix)
	for _, h := range handlers {
		//fmt.Println("PATH", pathPrefix+h.Path, h.Func)
		//http.HandleFunc(h.Path, h.Func)
		// http.HandleFunc(pathPrefix+h.Path, h.Func)
		http.Handle(pathPrefix+h.Path, oktaMiddleware(h.Func))
	}

	http_server := fmt.Sprintf(":%d", http_port)
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

	fmt.Println("serve up web:", http_server)
	//http.ListenAndServe(http_server, nil)
	err = http.ListenAndServe(http_server, gorilla.CompressHandler(gorilla.LoggingHandler(accessLog, http.DefaultServeMux)))
	if err != nil {
		panic(err)
	}
}