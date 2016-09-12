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
	status := "ok" // yes, this should be baked into 's' but shouldn't this be progamtic?
	uptime := time.Since(startTime)
	stats := strings.Join(dbStats(), "\n")
	w.Header().Set("Content-Type", "text/plain")
	const s = "status: %s\n\nversion: %s\nhostname: %s\nstarted:%s\nuptime: %s\n\ndb stats:\nconnections: %d\n%s\n"
	fmt.Fprintf(w, s, status, version, hostname, startTime, uptime, datastore.DB.Stats().OpenConnections, stats)
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
		sendJSON(w, searchDB(strings.TrimSpace(what)))
		return
	}
	jsonError(w, "no search term specified", http.StatusBadRequest)
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
	jsonError(w, "Bad path: "+r.URL.Path, http.StatusBadRequest)
}

func pingMany(w http.ResponseWriter, r *http.Request) {
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

func userLogin(w http.ResponseWriter, r *http.Request) (*user, error) {
	method := strings.ToUpper(r.Method)
	switch method {
	case "POST":
		obj := &credentials{}
		content := r.Header.Get("Content-Type")
		if strings.Contains(content, "application/json") {
			if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
				return nil, err
			}
		} else {
			if err := objFromForm(&obj, r.Form); err != nil {
				return nil, err
			}
		}
		remoteAddr := remoteHost(r)
		log.Println("user:", obj.Username)
		user, err := userAuth(obj.Username, obj.Password)
		if err != nil {
			return nil, err
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
		return user, nil
	}
	return nil, fmt.Errorf("invalid http method: %s", r.Method)
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

		now := time.Now()
		s := &session{
			USR:    &user.USR,
			Remote: remoteAddr,
			Event:  "login",
			TS:     &now,
		}
		if err := dbAdd(s); err != nil {
			log.Println("session log error:", err)
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

func apiLogout(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	now := time.Now()
	s := &session{
		USR:    &u.USR,
		Remote: remoteHost(r),
		Event:  "logout",
		TS:     &now,
	}
	dbAdd(s)

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

func apiPragmas(w http.ResponseWriter, r *http.Request) {
	//dbDebug(true)
	pragmas, err := dbPragmas()
	//dbDebug(false)
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

func assumeUser(w http.ResponseWriter, r *http.Request) {
	debug := true
	query := r.URL.Query()
	apiKey := r.Header.Get("X-API-KEY")
	if len(apiKey) == 0 {
		apiKey = query.Get("X-API-KEY")
	}
	delete(query, "X-API-KEY")
	if dbq, ok := query["debug"]; ok {
		debug, _ = strconv.ParseBool(dbq[0])
		delete(query, "debug")
	}
	r.URL.RawQuery = query.Encode()
	if debug {
		dbDebug(debug)
		defer dbDebug(false)
	}
	u, err := userFromAPIKey(apiKey)
	if err != nil {
		log.Println("AUTH ERROR:", err)
		jsonError(w, err, http.StatusUnauthorized)
		return
	}
	if u.Level < 2 {
		jsonError(w, "access denied", http.StatusForbidden)
		return
	}
	/*
		if debug {
			log.Println("USER LOGIN NAME:", u.Login)
		}
	*/
	var id string
	i := strings.LastIndex(r.URL.Path, "/")
	if i < len(r.URL.Path)-1 {
		id = r.URL.Path[i+1:]
	}
	body := bodyCopy(r)
	log.Printf("(%s) PATH:%s ID:%s Q:%s BODY:%s", r.Method, r.URL.Path, id, r.URL.RawQuery, body)
	method := strings.ToUpper(r.Method)
	switch method {
	case "POST":
		assumed := &user{}
		if err := dbFindByID(assumed, id); err != nil {
			fmt.Println("***** ERR:", err)
			jsonError(w, err, http.StatusInternalServerError)
			return
		}
		auditLog(u.USR, remoteHost(r), "Assumed", "Assumed identity for "+assumed.Login+" by "+u.Login)
		cors(w)
		c := &http.Cookie{
			Name:    "X-API-KEY",
			Path:    "/",
			Expires: time.Now().Add(4 * time.Hour),
			Value:   assumed.APIKey,
		}
		http.SetCookie(w, c)
		remember(w, assumed)
		sendJSON(w, assumed)

		now := time.Now()
		s := &session{
			USR:    &u.USR,
			Remote: remoteHost(r),
			Event:  "assumed identity of: " + assumed.Login,
			TS:     &now,
		}
		dbAdd(s)
		return
	}
	jsonError(w, "invalid method:"+method, http.StatusBadRequest)
}

func deviceAudit(w http.ResponseWriter, r *http.Request) {
	dbDebug(true)
	defer dbDebug(false)
	//const query = "select * from devices_history where did=?"
	//list, err := datastore.ListQuery(&deviceHistory{}, query, r.URL.Path)
	list, err := datastore.ListQuery(&deviceHistory{}, "where did=?", r.URL.Path)
	if err != nil {
		log.Println("audit error:", err)
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	//log.Println("LIST LEN:", len(list.([]dbutil.DBObject)))
	/*
		table, _ := dbTable(query, r.URL.Path)
			if len(table.Rows) > 1 {
				var hostname string
				for i, col := range table.Columns {
					if col == "hostname" {
						hostname = table.Rows[
				skip := []string{"did", "usr", "login", "version", "ts"}
				t2 := table.Diff(true, skip...)
				cors(w)
				sendJSON(w, t2)
			}
	*/
	sendJSON(w, list)
}

func vmAudit(w http.ResponseWriter, r *http.Request) {
	dbDebug(true)
	defer dbDebug(false)
	list, err := datastore.ListQuery(&vmHistory{}, "where vmi=?", r.URL.Path)
	if err != nil {
		log.Println("audit error:", err)
		jsonError(w, err, http.StatusInternalServerError)
		return
	}
	sendJSON(w, list)
}

var webHandlers = []hFunc{
	{"/static/", StaticPage},
	{"/ping", pingPage},
	{"/api/db/pragmas", apiPragmas},
	{"/api/device/adjust/", MakeREST(deviceAdjust{})},
	//{"/api/device/history/", MakeREST(deviceHistory{})},
	{"/api/device/audit/", deviceAudit},
	{"/api/device/ips/", MakeREST(deviceIPs{})},
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
	{"/api/mactable", macTable},
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
	{"/api/session/", MakeREST(sessionView{})},
	{"/api/site/", MakeREST(site{})},
	{"/api/summary/", MakeREST(summary{})},
	{"/api/ping", pingMany},
	{"/api/rma/view/", MakeREST(rmaView{})},
	{"/api/rma/", MakeREST(rma{})},
	{"/api/tag/", MakeREST(tag{})},
	{"/api/user/assume/", assumeUser},
	{"/api/user/", MakeREST(user{})},
	{"/api/vendor/", MakeREST(vendor{})},
	{"/api/vlan/view/", MakeREST(vlanView{})},
	{"/api/vlan/", MakeREST(vlan{})},
	{"/api/vm/audit/", MakeREST(vmHistory{})},
	{"/api/vm/ips/", MakeREST(vmIPs{})},
	{"/api/vm/view/", MakeREST(vmView{})},
	{"/api/vm/", MakeREST(vm{})},
	{"/api/search/", apiSearch},
	{"/api/", apiUnknown},
	{"/", homePage},
}
