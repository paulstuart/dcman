package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	ttext "text/template"
	"time"

	//"github.com/davecgh/go-spew/spew"
	gorilla "github.com/gorilla/handlers"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/paulstuart/dbutil"
)

const (
	accessLog = "access.log"
	errorLog  = "error.log"
	cookieID  = "dcuser"
)

var (
	assetDir   = "assets"
	tdir       string
	serverIP   = myIP()
	htmlTmpl   map[string]*template.Template
	textTmpl   map[string]*ttext.Template
	httpServer string
	baseURL    string
	errorFile  *os.File
	authCookie string
)

type hFunc struct {
	Path string
	Func http.HandlerFunc
}

// for debugging, get a copy of the body text
func bodyCopy(r *http.Request) string {
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
	}
	// Restore the io.ReadCloser to its original state
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// Use the content
	return string(bodyBytes)
}

func remoteHost(r *http.Request) string {
	if remoteAddr := r.Header.Get("X-Forwarded-For"); len(remoteAddr) > 0 {
		return remoteAddr
	}
	remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "REMOTE ADDR ERR:", err)
	}
	// check if running on same host
	if len(remoteAddr) > 0 && remoteAddr[0] == ':' {
		remoteAddr = serverIP
	}
	return remoteAddr
}

// for loading an object from an http post
func objFromForm(obj interface{}, values map[string][]string) error {
	val := reflect.ValueOf(obj)
	base := reflect.Indirect(val)
	t := reflect.TypeOf(base.Interface())

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		b := base.Field(i)
		if val, ok := values[f.Name]; ok {
			switch b.Interface().(type) {
			case string:
				b.SetString(val[0])
			case int:
				i, _ := strconv.Atoi(val[0])
				b.SetInt(int64(i))
			case int64:
				i, _ := strconv.ParseInt(val[0], 0, 64)
				b.SetInt(i)
			case uint:
				i, _ := strconv.ParseUint(val[0], 0, 64)
				b.SetUint(i)
			case uint32:
				i, _ := strconv.ParseUint(val[0], 0, 32)
				b.SetUint(i)
			case time.Time:
				if len(val[0]) == 0 {
					continue
				}
				l := dateLayout
				if len(val[0]) > len(dateLayout) {
					l = timeLayout
				}
				if when, err := time.Parse(l, val[0]); err == nil {
					v := reflect.ValueOf(when)
					b.Set(v)
				} else {
					return err
				}
			default:
				return fmt.Errorf("unhandled field type for:%s type:%s", f.Name, b.Type())
			}
		}
	}
	return nil
}

func loadObj(r *http.Request, obj dbutil.DBObject) error {
	content := r.Header.Get("Content-Type")
	if strings.Contains(content, "application/json") {
		return json.NewDecoder(r.Body).Decode(obj)
	}
	return objFromForm(obj, r.Form)
}

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS, POST, PUT, DELETE")
	//w.Header().Set("Access-Control-Allow-Headers", "Origin,Accept,X-Requested-With,Content-Type,Access-Control-Request-Method,Access-Control-Request-Headers")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, Content-Type, X-API-KEY") // TODO: X-API-KEY should probably be X-Auth-Token
}

func sendJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Encoding", "gzip")
	cors(w)
	// Gzip data
	gz := gzip.NewWriter(w)
	if err := json.NewEncoder(gz).Encode(obj); err != nil {
		log.Println("marshal error:", err)
	}
	gz.Flush()
}

// Validator is a custom function for object validation
type Validator func(string) error

func objPost(r *http.Request, o dbutil.DBObject, validators ...Validator) error {
	r.ParseForm()
	objFromForm(o, r.Form)
	action := r.Form.Get("action")
	user := currentUser(r)
	o.ModifiedBy(user.USR, time.Now())
	name := fmt.Sprintf("%v", reflect.TypeOf(o))
	for _, v := range validators {
		if err := v(action); err != nil {
			log.Println("VALID ERR:", err)
			return err
		}
	}
	auditLog(user.USR, remoteHost(r), action, name)
	switch {
	case action == "Add":
		return dbAdd(o)
	case action == "Update":
		dbFindByID(o, r.FormValue(o.KeyField()))
		return dbSave(o)
	case action == "Delete":
		return dbDelete(o)
	}
	return fmt.Errorf("Unknown action: %s", action)
}

func loadTemplates() {
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
		if strings.HasSuffix(name, ".swp") {
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

func setLinks(t *dbutil.Table, id int, path string, args ...int) {
	t.SetLinks(id, pathPrefix+path, args...)
}

func setLinksWhen(t *dbutil.Table, fn dbutil.LinkFunc, id int, path string, args ...int) {
	t.SetLinksWhen(fn, id, pathPrefix+path, args...)
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
	return d.Format(dateLayout)
}

// render an html template that inherits the "base" template
func renderTemplate(w http.ResponseWriter, r *http.Request, tname string, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	name := string(tname + ".html")
	err := htmlTmpl[name].ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// render an html template with no inheritence
func renderPlainTemplate(w http.ResponseWriter, r *http.Request, tname string, data interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	name := string(tname + ".html")
	err := htmlTmpl[name].Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// render a plaintext template with no inheritence (e.g., for scripts)
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

func currentUser(r *http.Request) user {
	cookie, err := r.Cookie(cookieID)
	if err != nil {
		return user{}
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
		if !authenticate(login, password) {
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

func faviconPage(w http.ResponseWriter, r *http.Request) {
	fav := filepath.Join(assetDir, "static/images/favicon.ico")
	http.ServeFile(w, r, fav)
}

func invalidPage(w http.ResponseWriter, r *http.Request) {
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

// StaticPage returns a static file
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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	switch path.Ext(name) {
	case "css":
		w.Header().Set("Content-Type", "text/css")
	case "js":
		w.Header().Set("Content-Type", "application/javascript")
	}
	http.ServeContent(w, r, name, fi.ModTime(), file)
}

// ErrorLog logs errors
func ErrorLog(r *http.Request, msg string, args ...interface{}) {
	user := currentUser(r)
	remoteAddr := remoteHost(r)
	fmt.Fprintln(errorFile, time.Now().Format(logLayout), remoteAddr, user.USR, msg, args)
}

// allows logging to show user id
func userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.User == nil {
			user := currentUser(r)
			if user.USR > 0 {
				r.URL.User = url.User(user.Email)
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
				fmt.Println("AUTH FOR PAGE:", r.URL.Path)
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

/*
func StripPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			r.URL.Path = p
			h.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}
*/

func redirect(w http.ResponseWriter, r *http.Request, path string, status int) {
	http.Redirect(w, r, pathPrefix+path, status)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	redirect(w, r, "/error", http.StatusNotFound)
}

func goHome(w http.ResponseWriter, r *http.Request) {
	redirect(w, r, "/", http.StatusMovedPermanently)
}

func badRequest(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

func webServer(handlers []hFunc) {
	// mux all the handlers
	for _, h := range handlers {
		p := pathPrefix + h.Path
		switch {
		case strings.HasPrefix(h.Path, "/static/"):
			http.Handle(p, gorilla.CompressHandler(http.StripPrefix(pathPrefix, h.Func)))
		case strings.HasPrefix(h.Path, "/data/"):
			http.Handle(p, http.StripPrefix(p, h.Func))
		case strings.HasPrefix(h.Path, "/api/"):
			http.Handle(p, http.StripPrefix(p, h.Func))
		case strings.HasPrefix(h.Path, "/login"):
			http.Handle(p, http.StripPrefix(p, h.Func))
		default:
			//http.Handle(p, http.StripPrefix(p, authMiddleware(h.Func)))
			http.Handle(p, http.StripPrefix(p, h.Func))
		}
	}
	http.HandleFunc("/favicon.ico", faviconPage)

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

	httpServer = fmt.Sprintf(":%d", cfg.Main.Port)
	baseURL = fmt.Sprintf("http://%s:%d/%s", serverIP, cfg.Main.Port, pathPrefix)
	fmt.Printf("serve up web: http://%s%s/\n", serverIP, httpServer)
	err = http.ListenAndServe(httpServer, userMiddleware(gorilla.LoggingHandler(accessLog, http.DefaultServeMux)))
	if err != nil {
		panic(err)
	}
}

// convert url parameters to sql 'where' condition
func whereParams(m map[string][]string) string {
	cnt := 0
	var buf bytes.Buffer
	for k, v := range m {
		if len(v) > 0 {
			if cnt == 0 {
				buf.WriteString(" where ")
			} else {
				buf.WriteString(" and ")
			}
			buf.WriteString(k)
			buf.WriteString("=")
			_, err := strconv.ParseFloat(v[0], 64)
			if err == nil {
				buf.WriteString(v[0])
			} else {
				buf.WriteString("'")
				buf.WriteString(v[0])
				buf.WriteString("'")
			}
			cnt++
		}
	}
	return buf.String()
}

// build query string
func qstring(k []string) string {
	switch len(k) {
	case 0:
		return ""
	case 1:
		return "where " + k[0] + "=?"
	default:
		var b bytes.Buffer
		b.WriteString("where ")
		b.WriteString(k[0])
		b.WriteString("=? ")
		for _, v := range k[1:] {

			b.WriteString("and ")
			b.WriteString(v)
			b.WriteString("=? ")
		}
		return b.String()

	}
}

func fullQuery(r *http.Request) (string, []interface{}, error) {
	var val []interface{}
	if err := r.ParseForm(); err != nil {
		return "", val, err
	}
	keys := make([]string, 0, len(r.Form))
	for k := range r.Form {
		keys = append(keys, k)
	}
	q := qstring(keys)
	val = make([]interface{}, 0, len(keys))
	for _, k := range keys {
		val = append(val, r.Form[k][0])
	}
	return q, val, nil
}

func jsonError(w http.ResponseWriter, what interface{}, code int) {
	var msg string
	switch what := what.(type) {
	case string:
		msg = fmt.Sprintf(`{"Error": "%v"}`, what)
	case error:
		msg = fmt.Sprintf(`{"Error": "%v"}`, what.Error())
	default:
		j, err := json.MarshalIndent(what, " ", " ")
		if err != nil {
			log.Println("dang! error on error:", err)
		}
		msg = string(j)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	//http.Error(w, msg, code)
	fmt.Fprint(w, msg)
}

// MakeREST will generate a REST handler for a DBObject
func MakeREST(gen dbutil.DBGen) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newREST(gen.NewObj().(dbutil.DBObject), w, r)
	}
}

func apiKey(w http.ResponseWriter, r *http.Request) (*user, bool, error) {
	query := r.URL.Query()
	debug := false
	apiKey := r.Header.Get("X-API-KEY")
	if len(apiKey) == 0 {
		apiKey = query.Get("X-API-KEY")
	}
	//log.Println("check api key:", apiKey)
	delete(query, "X-API-KEY")
	if dbq, ok := query["debug"]; ok {
		debug, _ = strconv.ParseBool(dbq[0])
		//log.Println("*** debug set to:", debug)
		delete(query, "debug")
	} /*else {
		log.Println("*** debug:", debug)
	}
	*/
	r.URL.RawQuery = query.Encode()
	/*
		if debug {
			dbDebug(debug)
			//defer dbDebug(false)
		}
	*/
	user, err := userFromAPIKey(apiKey)
	return &user, debug, err
}

func newREST(obj dbutil.DBObject, w http.ResponseWriter, r *http.Request) {
	db := datastore
	user, debug, err := apiKey(w, r)
	if err != nil && !insecure {
		log.Println("AUTH ERROR:", err)
		jsonError(w, err, http.StatusUnauthorized)
		return
	}
	//log.Println("auth ok")
	//spew.Println("\nUSER:", user, "\n")

	var id string
	i := strings.LastIndex(r.URL.Path, "/")
	if i < len(r.URL.Path)-1 {
		id = r.URL.Path[i+1:]
	}
	body := bodyCopy(r)
	//log.Println("DEBUG:", debug, "BODY:", body)
	if debug {
		log.Printf("(%s) PATH:%s ID:%s T:%T Q:%s BODY:%s", r.Method, r.URL.Path, id, obj, r.URL.RawQuery, body)
		dbDebug(true)
	}
	method := strings.ToUpper(r.Method)
	//log.Println("M:", method, "ID:", id)
	switch method {
	case "PUT", "POST":
		bodyString := bodyCopy(r)
		content := r.Header.Get("Content-Type")
		//fmt.Println("CONTENT:", content)
		if strings.Contains(content, "application/json") {
			if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
				fmt.Println("***** BODY:", bodyString)
				fmt.Println("***** ERR:", err)
				jsonError(w, err, http.StatusInternalServerError)
				return
			}
		} else {
			objFromForm(obj, r.Form)
		}
		if !insecure {
			obj.ModifiedBy(user.USR, time.Now())
		}
	case "OPTIONS", "HEAD":
		cors(w)
		fmt.Fprintln(w, "options:", bodyCopy(r))
	}

	// Make the change
	switch method {
	case "GET":
		if len(id) == 0 {
			q, val, err := fullQuery(r)
			if err != nil {
				log.Println("query error:", err)
				jsonError(w, err, http.StatusInternalServerError)
				return
			}
			if debug && len(val) > 0 {
				log.Println("Q VAL:", val)
			}
			list, err := db.ListQuery(obj, q, val...)
			if err != nil {
				log.Println("list error:", err)
				jsonError(w, err, http.StatusInternalServerError)
				return
			}
			//log.Println("LIST LEN:", len(list.([]dbutil.DBObject)))
			//log.Println("GET LIST:", list)
			sendJSON(w, list)
			return
		}
		// column name meta data
		if strings.HasSuffix(r.URL.Path, "/columns") {
			log.Println("GET PATH------->", r.URL.Path)
			data := struct {
				Columns []string
			}{
				Columns: obj.Names(),
			}
			sendJSON(w, data)
			return
		}

		if err := db.FindByID(obj, id); err != nil {
			jsonError(w, err, http.StatusInternalServerError)
			return
		}
		sendJSON(w, obj)
		return
	case "DELETE":
		if len(id) == 0 {
			err := fmt.Errorf("delete requires object id")
			jsonError(w, err, http.StatusBadRequest)
			return
		}
		if err := db.DeleteByID(obj, id); err != nil {
			sqle := err.(sqlite3.Error)
			log.Printf("DELETE ERROR (%T) code:%d ext:%d %s\n", sqle, sqle.Code, sqle.ExtendedCode, sqle)
			//if sqle.Code == sqlite3.ErrConstraintForeignKey {
			if sqle.Code == sqlite3.ErrConstraint {
				jsonError(w, err, http.StatusConflict)
				return
			}
			jsonError(w, err, http.StatusInternalServerError)
			return
		}
		msg := struct{ Msg string }{"deleted object id: " + id}
		sendJSON(w, msg)
		return
	case "PATCH":
		if len(id) == 0 {
			jsonError(w, "no ID specified", http.StatusBadRequest)
			return
		}

		// if we're patching the object, we first fetch the original copy
		// then overwrite the new fields supplied by the patch object
		if err := db.FindByID(obj, id); err != nil {
			log.Println("PATCH ERR 1:", err)
			jsonError(w, err, http.StatusInternalServerError)
			return
		}

		if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
			log.Println("PATCH ERR 2:", err)
			jsonError(w, err, http.StatusInternalServerError)
			return
		}
		if !insecure {
			obj.ModifiedBy(user.USR, time.Now())
		}
		if err := db.Save(obj); err != nil {
			log.Println("PATCH ERR 3:", err)
			jsonError(w, err, http.StatusInternalServerError)
			return
		}
		db.FindSelf(obj) // load what DB has for verification
		sendJSON(w, obj)
	case "PUT":
		log.Println("PUT OBJ:", obj)
		if err := db.Save(obj); err != nil {
			jsonError(w, err, http.StatusInternalServerError)
			return
		}
		db.FindSelf(obj) // load what DB has for verification
		sendJSON(w, obj)
	case "POST":
		if err := db.Add(obj); err != nil {
			log.Println("add error:", err)
			jsonError(w, err, http.StatusInternalServerError)
			return
		}
		//w.WriteHeader(http.StatusCreated)
		db.FindSelf(obj) // load what DB has for verification
		sendJSON(w, obj)
	}
	// TODO: special debug flag for spew?
	//spew.Dump(obj)
}
