package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	name := filepath.Join(assetDir, "static/html/spa.html")
	file, err := os.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	w.Header().Set("Cache-control", "public, max-age=259200")
	cors(w)
	http.ServeContent(w, r, name, fi.ModTime(), file)
}

func macTable(w http.ResponseWriter, r *http.Request) {
	const h = "#%-19s %-15s %-10s %-20s %-15s  %s\n"
	const s = "%-20s %-15s %-10s %-20s %-15s  %s\n"
	fn := func(columns []string, count int, buffer []interface{}) {
		if count == 0 {
			cols := make([]interface{}, 0, len(columns))
			for _, c := range columns {
				cols = append(cols, c)
			}
			fmt.Fprintf(w, h, cols...)
		}
		fmt.Fprintf(w, s, buffer...)
	}
	w.Header().Set("Content-Type", "text/plain")
	dbStream(fn, "select * from mactable")
}

// TODO add as /api/status
func pingPage(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	uptime := time.Since(startTime)
	stats := strings.Join(dbStats(), "\n")
	fmt.Fprintf(w, "status: %s\nversion: %s\nhostname: %s\nstarted:%s\nuptime: %s\ndb stats:\n%s\n", status, version, hostname, startTime, uptime, stats)
}

func loginFailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("FAIL!")
	//httpError(w, r, "Login failed!")
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	auditLog(user.USR, remoteHost(r), "Logout", user.Email)
	isAuthorized(w, false)
	remember(w, nil)
	redirect(w, r, "/", 302)
}

func isAuthorized(w http.ResponseWriter, yes bool) {
	c := &http.Cookie{
		Name: authCookie,
		Path: "/",
	}
	if yes {
		c.Expires = time.Now().Add(sessionMinutes)
		c.Value = cfg.SAML.OKTAHash
	}
	http.SetCookie(w, c)
}

func remember(w http.ResponseWriter, u *user) {
	c := &http.Cookie{
		Name: cookieID,
		Path: "/",
	}
	if u != nil {
		c.Expires = time.Now().Add(sessionMinutes)
		c.Value = u.Cookie()
	}
	http.SetCookie(w, c)
	c = &http.Cookie{
		Name: "userinfo",
		Path: "/",
	}
	if u != nil {
		c.Expires = time.Now().Add(sessionMinutes)
		c.Value = b64(fmt.Sprintf(`{"username": "%s", "admin": %d}`, u.Login, u.Level))
	}
	http.SetCookie(w, c)
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	i := strings.LastIndex(r.URL.Path, "/")
	if i < len(r.URL.Path)-1 {
		what := r.URL.Path[i+1:]
		fmt.Println("WHAT:", what)
		sendJSON(w, searchDB(strings.TrimSpace(what)))
		return
	}
	jsonError(w, "no search term specified", http.StatusBadRequest)
}

func ServerDiscover(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ipmi := r.URL.Path
	if len(ipmi) == 0 {
		notFound(w, r)
		return
	}
	mac, _ := FindMAC(ipmi)
	d := struct {
		MacEth0 string
	}{
		MacEth0: mac,
	}
	j, _ := json.MarshalIndent(d, " ", " ")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(j))
}

// return the string after the last "/" of the url
func urlSuffix(r *http.Request) string {
	i := strings.LastIndex(r.URL.Path, "/")
	return r.URL.Path[i+1:]
}

// A catch-all if the api path is invalid
// otherwise, it the http router would default to "/"
// and return the home page
func apiUnknown(w http.ResponseWriter, r *http.Request) {
	/*
		msg := fmt.Sprintf(`{"Error": "Bad path: %s"}`, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, msg, http.StatusBadRequest)
	*/
	jsonError(w, "Bad path: "+r.URL.Path, http.StatusBadRequest)
}

func BulkPings(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	timeout := pingTimeout
	if text := r.Form.Get("timeout"); len(text) > 0 {
		if t, err := strconv.Atoi(text); err == nil {
			timeout = t
		}
	}
	if text := r.Form.Get("debug"); len(text) > 0 {
		if debug, err := strconv.ParseBool(text); err == nil && debug {
			for k, v := range r.Form {
				log.Println("K:", k, "(", len(k), ") V:", v)
			}
		}
	}
	if ips, ok := r.Form["ips[]"]; ok && len(ips) > 0 {
		pings := bulkPing(timeout, ips...)
		j, _ := json.MarshalIndent(pings, " ", " ")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(j))
	}
}

func IPMICredentialsGet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		ipmi := r.Form.Get("ipmi")
		username, password, err := GetCredentials(ipmi)
		if err != nil {
			log.Println("error getting creds for ipmp:", ipmi, "error:", err)
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, username, password)
	}
}

func IPMICredentialsSet(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		ipmi := r.Form.Get("ipmi")
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		err := SetCredentials(ipmi, username, password)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, err)
		}
	}
}

func ipRange(w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(r.Method)
	log.Println("IP RANGE METHOD:", method)
	switch method {
	case "POST":
		obj := struct{ From, To string }{}
		content := r.Header.Get("Content-Type")
		log.Println("IP RANGE CONTENT:", content)
		if strings.Contains(content, "application/json") {
			if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
				jsonError(w, err, http.StatusInternalServerError)
				return
			}
		} else {
			objFromForm(&obj, r.Form)
		}
		log.Println("RANGE OBJ:", obj)
		from := ipFromString(obj.From)
		to := ipFromString(obj.To)
		log.Println("RANGE FROM:", from, "TO:", to)
		dbDebug(true)
		defer dbDebug(false)
		list, err := dbObjectListQuery(ipAddr{}, "where ip32 >=? and ip32 <=?", from, to)
		if err != nil {
			log.Println("IP RANGE ERROR:", err)
			jsonError(w, err, http.StatusInternalServerError)
			return
		}
		sendJSON(w, list)
	}
}

func fullURL(path ...string) string {
	return "http://" + serverIP + httpServer + pathPrefix + strings.Join(path, "")
}

func apiLogin(w http.ResponseWriter, r *http.Request) {
	method := strings.ToUpper(r.Method)
	switch method {
	case "POST":
		obj := &credentials{}
		content := r.Header.Get("Content-Type")
		if strings.Contains(content, "application/json") {
			if err := json.NewDecoder(r.Body).Decode(obj); err != nil {
				fmt.Println("***** ERR:", err)
				jsonError(w, err, http.StatusInternalServerError)
				return
			}
		} else {
			objFromForm(obj, r.Form)
		}
		remoteAddr := remoteHost(r)
		log.Println("user:", obj.Username)
		user, err := userAuth(obj.Username, obj.Password)
		if err != nil {
			auditLog(0, remoteAddr, "Login", err.Error())
			jsonError(w, err, http.StatusUnauthorized)
			return
		}
		auditLog(user.USR, remoteAddr, "Login", "Login succeeded for "+obj.Username)
		cors(w)
		c := &http.Cookie{
			Name:    "X-API-KEY",
			Path:    "/",
			Expires: time.Now().Add(4 * time.Hour),
			Value:   user.APIKey,
		}
		http.SetCookie(w, c)
		remember(w, user)
		sendJSON(w, user)
	default:
		jsonError(w, "invalid http method:"+r.Method, http.StatusUnauthorized)
	}
}

func apiPragmas(w http.ResponseWriter, r *http.Request) {
	dbDebug(true)
	pragmas, err := dbPragmas()
	dbDebug(false)
	if err != nil {
		jsonError(w, err, http.StatusInternalServerError)
	} else {
		query := r.URL.Query()
		format := query.Get("format")
		if format == "plaintext" {
			for k, v := range pragmas {
				fmt.Fprintf(w, "%s: %s\n", k, v)
			}
		} else {
			sendJSON(w, pragmas)
		}
	}
}

func apiLogout(w http.ResponseWriter, r *http.Request) {
	cors(w)
	c := &http.Cookie{
		Name:    "SAML",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, c)
	c.Name = "dcuser"
	http.SetCookie(w, c)
	c.Name = "userinfo"
	http.SetCookie(w, c)
	c.Name = "redirect"
	http.SetCookie(w, c)
	c.Name = "X-API-KEY"
	http.SetCookie(w, c)
}

var webHandlers = []hFunc{
	//{"/favicon.ico", FaviconPage},
	{"/static/", StaticPage},
	{"/api/credentials/get", IPMICredentialsGet},
	{"/api/credentials/set", IPMICredentialsSet},
	{"/api/db/pragmas", apiPragmas},
	{"/api/device/adjust/", MakeREST(deviceAdjust{})},
	{"/api/device/ips/", MakeREST(deviceIPs{})},
	//{"/api/device/network/", MakeREST(DeviceNetwork{})},
	{"/api/device/pxe/", MakeREST(pxeDevice{})},
	{"/api/device/type/", MakeREST(deviceType{})},
	{"/api/device/view/", MakeREST(deviceView{})},
	{"/api/device/", MakeREST(device{})},
	{"/api/interface/view/", MakeREST(ifaceView{})},
	{"/api/interface/", MakeREST(iface{})},
	{"/api/login", apiLogin},
	{"/api/logout", apiLogout},
	{"/api/part/type/", MakeREST(partType{})},
	{"/api/part/view/", MakeREST(partView{})},
	{"/api/part/", MakeREST(part{})},
	{"/api/inventory/", MakeREST(inventory{})},
	{"/api/mfgr/", MakeREST(manufacturer{})},
	{"/api/network/circuit/view/", MakeREST(circuitView{})},
	{"/api/network/circuit/list/", MakeREST(circuitList{})},
	{"/api/network/circuit/", MakeREST(circuit{})},
	{"/api/network/ip/type/", MakeREST(ipType{})},
	{"/api/network/ip/used/", MakeREST(ipsUsed{})},
	{"/api/network/ip/range", ipRange},
	{"/api/network/ip/", MakeREST(ipAddr{})},
	{"/api/rack/view/", MakeREST(rackView{})},
	{"/api/rack/", MakeREST(rack{})},
	{"/api/site/", MakeREST(site{})},
	{"/api/summary/", MakeREST(summary{})},
	{"/api/pings", BulkPings},
	{"/api/rma/view/", MakeREST(rmaView{})},
	{"/api/rma/", MakeREST(rma{})},
	{"/api/tag/", MakeREST(tag{})},
	{"/api/user/", MakeREST(user{})},
	{"/api/vendor/", MakeREST(vendor{})},
	{"/api/vlan/view/", MakeREST(vlanView{})},
	{"/api/vlan/", MakeREST(vlan{})},
	{"/api/vm/view/", MakeREST(vmView{})},
	{"/api/vm/", MakeREST(vm{})},
	{"/api/search/", apiSearch},
	{"/api/", apiUnknown},
	{"/data/server/discover/", ServerDiscover},
	{"/data/mactable", macTable},
	{"/", homePage},
}
