package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	ttext "text/template"
	"time"

	gorilla "github.com/gorilla/handlers"
)

const (
	accessLog = "access.log"
	errorLog  = "error.log"
	cookieID  = "dcuser"
)

var (
	assetDir    = "assets"
	tdir        string
	ip          = MyIp()
	htmlTmpl    map[string]*template.Template
	textTmpl    map[string]*ttext.Template
	http_server string
	errorFile   *os.File
	authCookie  string
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
	// check if running on same host
	if len(remote_addr) > 0 && remote_addr[0] == ':' {
		remote_addr = ip
	}
	return remote_addr
}

func loadTemplates() {
	//fmt.Println("LOAD TEMPLATES DIR:", tdir)
	funcMap := template.FuncMap{
		"isBlank":   isBlank,
		"isTrue":    isTrue,
		"plusOne":   plusOne,
		"fixDate":   fixDate,
		"userLogin": userLogin,
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

// Creates a new file upload http request with optional extra params
func uploadFile(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return http.NewRequest("POST", uri, body)
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

func isBlank(s string) string {
	if len(s) > 0 {
		return s
	}
	return " * blank * "
}

func plusOne(in interface{}) string {
	val := in.(int)
	val++
	return strconv.Itoa(val)
}

func fixDate(d time.Time) string {
	if d.IsZero() {
		return ""
	}
	return d.Format(date_layout)
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
			http.Redirect(w, r, "/loginfail", http.StatusFound)
		} else {
			expires := time.Now()
			expires = expires.Add(time.Duration(time.Hour) * 24 * 365)
			c := http.Cookie{Name: "login", Value: login, Expires: expires, Path: "/"}
			http.SetCookie(w, &c)
			http.Redirect(w, r, "/", http.StatusFound)
		}
	} else {
		err := htmlTmpl["login.html"].ExecuteTemplate(w, "base", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func FaviconPage(w http.ResponseWriter, r *http.Request) {
	fav := filepath.Join(assetDir, "static/images/favicon.ico")
	http.ServeFile(w, r, fav)
}

func InvalidPage(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func saveMultipartFile(name string, file multipart.File) error {
	if len(name) == 0 {
		return fmt.Errorf("file name not specified")
	}
	defer file.Close()

	out, err := os.Create(name)
	if err != nil {
		return err
	}

	defer out.Close()
	_, err = io.Copy(out, file)
	return err
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
	name := filepath.Join(assetDir, r.URL.Path)
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

// allows logging to show user id
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

func authorized(s string) bool {
	return len(s) > 0
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			cookie, err := r.Cookie(authCookie)
			if err != nil || !authorized(cookie.Value) {
				// stash url attempted to redirect after successful login
				// not working right now, because we're stripping the registered path :-(
				expires := time.Now().Add(time.Hour * 24 * 365)
				path := r.URL.Path
				if len(r.URL.RawQuery) > 0 {
					path += "?" + r.URL.RawQuery
				}
				if len(path) > 0 {
					c := http.Cookie{Name: "redirect", Value: path, Expires: expires, Path: "/"}
					http.SetCookie(w, &c)
				}

				redirect(w, r, "/login", http.StatusFound)
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

func goHome(w http.ResponseWriter, r *http.Request) {
	redirect(w, r, "/", http.StatusMovedPermanently)
}

func webServer(handlers []HFunc) {
	loadTemplates()

	// mux all the handlers
	for _, h := range handlers {
		p := pathPrefix + h.Path
		switch {
		case strings.HasPrefix(h.Path, "/static/"):
			http.Handle(p, http.StripPrefix(pathPrefix, h.Func))
		case strings.HasPrefix(h.Path, "/data/"):
			http.Handle(p, http.StripPrefix(p, h.Func))
		case strings.HasPrefix(h.Path, "/api/"):
			http.Handle(p, h.Func)
		case strings.HasPrefix(h.Path, "/login"):
			http.Handle(p, http.StripPrefix(p, h.Func))
		default:
			http.Handle(p, http.StripPrefix(p, authMiddleware(h.Func)))
		}
	}
	if len(pathPrefix) > 0 {
		http.HandleFunc("/", goHome)
	}

	logDir := cfg.Main.LogDir
	if len(logDir) == 0 {
		logDir = "logs"
	}
	if !path.IsAbs(logDir) {
		logDir = path.Join(execDir, logDir)
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
