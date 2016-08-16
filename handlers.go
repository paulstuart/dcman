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
	//"github.com/davecgh/go-spew/spew"
)

type Common string

func HomePage(w http.ResponseWriter, r *http.Request) {
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

func RackAdjust(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		adjust := r.Form.Get("adjust")
		rid := r.Form.Get("rid")
		q1 := "update servers set ru=ru+? where rid=?"
		if err := dbExec(q1, adjust, rid); err != nil {
			log.Println("rack adjust err:", err)
		}
	}
}

func RackMoveUnit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		adjust := r.Form.Get("adjust")
		ru := r.Form.Get("ru")
		rid := r.Form.Get("rid")
		q1 := "update servers set ru=ru+? where rid=? and ru=?"
		if err := dbExec(q1, adjust, rid, ru); err != nil {
			log.Println("rack adjust err:", err)
		}
	}
}

func MacTable(w http.ResponseWriter, r *http.Request) {
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

func pingPage(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	uptime := time.Since(startTime)
	stats := strings.Join(dbStats(), "\n")
	fmt.Fprintf(w, "status: %s\nversion: %s\nhostname: %s\nstarted:%s\nuptime: %s\ndb stats:\n%s\n", status, version, Hostname, startTime, uptime, stats)
}

func DebugPage(w http.ResponseWriter, r *http.Request) {
	what := r.URL.Path
	on, _ := strconv.ParseBool(what)
	log.Println("DEBUG?", what, "ON:", on)
	dbDebug(on)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "db debug: %t\n", on)
}

func loginFailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("FAIL!")
	//httpError(w, r, "Login failed!")
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	auditLog(user.USR, RemoteHost(r), "Logout", user.Email)
	Authorized(w, false)
	Remember(w, nil)
	redirect(w, r, "/", 302)
}

func Authorized(w http.ResponseWriter, yes bool) {
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

func Remember(w http.ResponseWriter, u *User) {
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

func APISearch(w http.ResponseWriter, r *http.Request) {
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
func APIUnknown(w http.ResponseWriter, r *http.Request) {
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

func fullURL(path ...string) string {
	return "http://" + serverIP + httpServer + pathPrefix + strings.Join(path, "")
}

func APILogin(w http.ResponseWriter, r *http.Request) {
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
		remoteAddr := RemoteHost(r)
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
		Remember(w, user)
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

func APILogout(w http.ResponseWriter, r *http.Request) {
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

type KV struct {
	Key, Value string
}

type BackTalk struct {
	Script, Callback string
	Envy             []KV
}

var webHandlers = []HFunc{
	//{"/favicon.ico", FaviconPage},
	{"/static/", StaticPage},
	{"/api/credentials/get", IPMICredentialsGet},
	{"/api/credentials/set", IPMICredentialsSet},
	{"/api/db/pragmas", apiPragmas},
	{"/api/device/ips/", MakeREST(DeviceIPs{})},
	{"/api/device/type/", MakeREST(DeviceType{})},
	{"/api/device/view/", MakeREST(DeviceView{})},
	{"/api/device/", MakeREST(Device{})},
	{"/api/interface/view/", MakeREST(IFaceView{})},
	{"/api/interface/", MakeREST(IFace{})},
	{"/api/login", APILogin},
	{"/api/logout", APILogout},
	{"/api/part/type/", MakeREST(PartType{})},
	{"/api/part/view/", MakeREST(PartView{})},
	{"/api/part/", MakeREST(Part{})},
	{"/api/inventory/", MakeREST(Inventory{})},
	{"/api/mfgr/", MakeREST(Manufacturer{})},
	{"/api/network/circuit/view/", MakeREST(CircuitView{})},
	{"/api/network/circuit/list/", MakeREST(CircuitList{})},
	{"/api/network/circuit/", MakeREST(Circuit{})},
	{"/api/network/ip/type/", MakeREST(IPType{})},
	{"/api/network/ip/used/", MakeREST(IPsUsed{})},
	{"/api/network/ip/", MakeREST(IPAddr{})},
	{"/api/rack/view/", MakeREST(RackView{})},
	{"/api/rack/", MakeREST(Rack{})},
	//{"/api/rackunit/", MakeREST(RackUnit{})},
	{"/api/site/", MakeREST(Site{})},
	{"/api/summary/", MakeREST(Summary{})},
	{"/api/pings", BulkPings},
	{"/api/rma/view/", MakeREST(RMAView{})},
	{"/api/rma/", MakeREST(RMA{})},
	{"/api/tag/", MakeREST(Tag{})},
	{"/api/user/", MakeREST(User{})},
	//{"/api/upload", APIUpload},
	//{"/api/update", APIUpdate},
	{"/api/vendor/", MakeREST(Vendor{})},
	{"/api/vlan/view/", MakeREST(VLANView{})},
	{"/api/vlan/", MakeREST(VLAN{})},
	{"/api/vm/view/", MakeREST(VMView{})},
	{"/api/vm/", MakeREST(VM{})},
	{"/api/search/", APISearch},
	{"/api/", APIUnknown},
	{"/data/server/discover/", ServerDiscover},
	{"/data/mactable", MacTable},
	/*
		{"/data/upload", DataUpload},
		{"/data/parts/load", PartsLoad},
		{"/db/debug/", DebugPage},
		{"/dc/all", ListingPage},
		{"/dc/connections", ConnectionsPage},
		{"/dc/edit/", DCEdit},
		{"/dc/list", DCList},
		{"/dc/racklist/", DCRackList},
		{"/dc/racks/", DatacenterPage},
		{"/ip/dc/", IPInternalDC},
		{"/ip/internal/all", IPInternalAllPage},
		{"/ip/internal/list", IPInternalList},
		{"/ip/public/all", IPPublicAllPage},
		{"/loginfail", loginFailHandler},
		{"/login", LoginHandler},
		{"/logout", logoutPage},
		{"/mfgr/edit/", MfgrEdit},
		{"/network/add/", NetworkAdd},
		{"/network/audit/", NetworkAudit},
		{"/network/devices", NetworkDevices},
		{"/network/edit/", NetworkEdit},
		{"/network/next/", NetworkNext},
		{"/network/vlans", VlansPage},
		{"/part/edit/", PartEdit},
		{"/part/list/", PartList},
		{"/part/replace/", PartReplace},
		{"/part/use/", PartUse},
		{"/part/totals", PartTotals},
		{"/part/type/", TypeEdit},
		{"/part/types", TypeList},
		{"/partly/template", PartlyUsed},
		{"/data/partly/page", PartlyPage},
		{"/data/partly/json", PartlyJSON},
		{"/pdu/edit", PDUEdit},
		{"/ping", pingPage},
		{"/profile/view", ProfileView},
		{"/rack/add", RackEdit},
		{"/rack/adjust", RackAdjust},
		{"/rack/audit/", RackAudit},
		{"/rack/edit/", RackEdit},
		{"/rack/move", RackMoveUnit},
		{"/rack/network", RackNetwork},
		{"/rack/updates", RackUpdates},
		{"/rack/view/", TheRackView},
		{"/rack/zone/", RackZone},
		{"/reload", reloadPage},
		//{"/search", SearchPage},
		{"/rma/list", RMAList},
		{"/rma/add/", RMAAdd},
		{"/rma/edit/", RMAEdit},
		{"/rma/received/", RMAReceived},
		{"/rma/return/add/", RMAReturnAdd},
		{"/rma/return/", RMAReturn},
		//{"/server/add/", ServerEdit},
		{"/server/audit/", ServerAudit},
		{"/server/checkfit", ServerCheckFit},
		{"/server/dupes", ServerDupes},
		//{"/server/edit/", ServerEdit},
		//{"/device/edit/", deviceEdit},
		//{"/server/find", ServerFind},
		{"/server/parts/", ServerParts},
		{"/server/replace/", ServerReplace},
		{"/server/vms", VMListing},
		{"/settings", SettingsHandler},
		{"/sku/edit/", SKUEdit},
		{"/sku/list", SKUList},
		{"/tag/edit/", TagEdit},
		{"/tag/list", TagList},
		{"/user/add", UserEdit},
		{"/user/edit/", UserEdit},
		{"/user/list", usersListPage},
		{"/user/run/", UserRun},
		{"/vendor/edit/", VendorEdit},
		{"/vendor/list", VendorList},
		{"/vlan/edit/", VlanEdit},
		//{"/vm/add/", VMAdd},
		{"/vm/all", VMAllPage},
		{"/vm/audit/", VMAudit},
		//{"/vm/edit/", VMEdit},
	*/
	{"/", HomePage},
}
