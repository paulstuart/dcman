package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbu "github.com/paulstuart/dbutil"
)

type Fail struct {
	Title, Error string
}

type Tuple [2]string

type Tabular struct {
	Common
	Table *dbu.Table
}

type RackData struct {
	Common
	RackID string
	DC     string
	Server []Tuple
}

type Totals struct {
	Title string
	Table *dbu.Table
}

//
// for templates
//

type Common struct {
	Title, Prefix, Banner string
	Heading               template.HTML
	Datacenters           []Datacenter
	User                  User
}

type Summary struct {
	Common
	Physical Totals
	Profiles Totals
	VMs      []Totals
}

type DCRacks struct {
	Common
	DC    string
	Racks []Rack
}

type ServerInfo struct {
	Common
	Servers []Server
}

type VMInfo struct {
	Common
	VMs []VM
}

type VMTmpl struct {
	Common
	VM VM
}

const (
	vmExportColumns = "dc,server,vm,profile,private,public,vip"
	vmExportQuery   = "select " + vmExportColumns + " from vmlist"
)

var (
	serverExportQuery string
)

// skip the audit info
func getColumns() {
	t, _ := dbTable("select * from sview")
	columns := make([]string, 0, len(t.Columns))
	for _, c := range t.Columns {
		switch {
		case c == "modified":
		case c == "uid":
		case c == "remote_addr":
		default:
			columns = append(columns, c)
		}
	}
	serverExportQuery = "select " + strings.Join(columns, ",") + " from sview"
}

func NewCommon(r *http.Request, title string) Common {
	var b string
	if cfg.Main.ReadOnly {
		b = " ** READ ONLY ** "
	}
	if len(bannerText) > 0 {
		b = b + bannerText
	}
	return Common{
		Title:       title,
		Prefix:      pathPrefix,
		Datacenters: Datacenters,
		User:        currentUser(r),
		Banner:      b,
	}
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	t, _ := dbTable("select * from server_summary")
	physical := Totals{"Physical Servers", t}
	p, _ := dbTable("select profile,count(profile) as total from profiles group by profile;")
	profiles := Totals{"Profiles", p}
	vms := []Totals{}
	for _, dc := range Datacenters {
		e, err := dbTable("select * from vm_summary where dc=?", dc.Name)
		if err != nil {
			log.Println("DB ERR:", err)
		}
		if len(e.Rows) > 0 {
			vms = append(vms, Totals{dc.City, e})
		}
	}
	data := Summary{
		Common:   NewCommon(r, cfg.Main.Name+" DC Manager"),
		Physical: physical,
		Profiles: profiles,
		VMs:      vms,
	}
	renderTemplate(w, r, "index", data)
}

func ProfileView(w http.ResponseWriter, r *http.Request) {
	var t *dbu.Table
	r.ParseForm()
	profile := r.FormValue("profile")
	dc := r.FormValue("dc")
	switch {
	case len(profile) > 0 && len(dc) > 0:
		t, _ = dbTable("select * from profiles where profile=? and dc=?", profile, dc)
		setLinks(t, 2, "/profile/view?dc=%s", 2)
	case len(profile) > 0:
		t, _ = dbTable("select * from profiles where profile=?", profile)
		setLinks(t, 2, "/profile/view?dc=%s&profile=%s", 2, 4)
	case len(dc) > 0:
		t, _ = dbTable("select * from profiles where dc=?", dc)
		setLinks(t, 2, "/profile/view?dc=%s", 2)
	default:
		t, _ = dbTable("select * from profiles")
		setLinks(t, 2, "/profile/view?dc=%s", 2)
	}

	t.Hide(0, 1)
	setLinks(t, 3, "/%s/edit/%s", 1, 0)
	setLinks(t, 4, "/profile/view?profile=%s", 4)
	t.SetType("ip-address", 3, 4)
	data := Tabular{
		Common: NewCommon(r, "Profile List"),
		Table:  t,
	}
	renderTemplate(w, r, "table", data)
}

func DataUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		data := r.Form.Get("data")
		err := LoadServers(strings.Split(data, "\n"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
		} else {
			fmt.Fprintln(w, "ok")
		}
	} else {
		data := NewCommon(r, "Upload server data")
		renderTemplate(w, r, "upload", data)
	}
}

func ServerFind(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		h := r.Form.Get("hostname")
		s := serversByQuery("where hostname like ?", "%"+h+"%")
		if len(s) == 0 {
			ErrorPage(w, r, "No servers found matching hostname: "+h)
		} else if len(s) == 1 {
			ShowServer(w, r, s[0])
		} else {
			data := ServerInfo{
				Common:  NewCommon(r, "servers matching: "+h),
				Servers: s,
			}
			renderTemplate(w, r, "found", data)
		}
	}
}

func SearchPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		what := r.Form.Get("what")
		kind := r.Form.Get("kind")
		switch {
		case kind == "server":
			searchServers(w, r, what)
		case kind == "vm":
			searchVMs(w, r, what)
		case kind == "ip":
			searchIPs(w, r, what)
		case kind == "ipmi":
			searchIPMI(w, r, what)
		case kind == "rack":
			searchRack(w, r, what)
		}
	} else {
		redirect(w, r, "/", http.StatusSeeOther)
	}
}

func searchIPMI(w http.ResponseWriter, r *http.Request, ip string) {
	query := "select * from ipmstr where what='ipmi' and ip=?"
	table, _ := dbTable(query, ip)
	if table == nil || len(table.Rows) == 0 {
		ErrorPage(w, r, "No assets found matching IPMI address: "+ip)
	} else if len(table.Rows) == 1 {
		s, _ := getServer("where id = ?", table.Rows[0][0])
		ShowServer(w, r, s)
	}
}

func searchIPs(w http.ResponseWriter, r *http.Request, ip string) {
	query := "select * from ipmstr where ip=?"
	table, _ := dbTable(query, ip)
	if table == nil || len(table.Rows) == 0 {
		ErrorPage(w, r, "No assets found matching ip: "+ip)
	} else if len(table.Rows) == 1 {
		if table.Rows[0][1] == "server" {
			s, _ := getServer("where id = ?", table.Rows[0][0])
			ShowServer(w, r, s)
		} else if table.Rows[0][1] == "vm" {
			v, _ := getVM("where id = ?", table.Rows[0][0])
			ShowVM(w, r, v)
		}
	} else {
		data := Tabular{
			Common: NewCommon(r, "Internal IPs"),
			Table:  table,
		}
		ShowIPs(w, r, data)
	}
}

func searchServers(w http.ResponseWriter, r *http.Request, hostname string) {
	s := serversByQuery("where hostname like ?", "%"+hostname+"%")
	if len(s) == 0 {
		ErrorPage(w, r, "No servers found matching hostname: "+hostname)
	} else if len(s) == 1 {
		ShowServer(w, r, s[0])
	} else {
		data := ServerInfo{
			Common:  NewCommon(r, "servers matching: "+hostname),
			Servers: s,
		}
		renderTemplate(w, r, "found", data)
	}
}

func searchVMs(w http.ResponseWriter, r *http.Request, hostname string) {
	if r.Method == "POST" {
		r.ParseForm()
		v := vmsByQuery("where hostname like ?", "%"+hostname+"%")
		if len(v) == 0 {
			ErrorPage(w, r, "No vms found matching hostname: "+hostname)
		} else if len(v) == 1 {
			ShowVM(w, r, v[0])
		} else {
			data := VMInfo{
				Common: NewCommon(r, "VMs matching: "+hostname),
				VMs:    v,
			}
			renderTemplate(w, r, "vmfound", data)
		}
	}
}

func showIP(w http.ResponseWriter, r *http.Request, ip string) {
	query := "select * from ipmstr where ip=?"
	table, _ := dbTable(query, ip)
	if table == nil || len(table.Rows) == 0 {
		ErrorPage(w, r, "No assets found matching ip: "+ip)
	} else if len(table.Rows) == 1 {
		if table.Rows[0][1] == "server" {
			s, _ := getServer("where id = ?", table.Rows[0][0])
			ServerRack(w, r, s)
		} else if table.Rows[0][1] == "vm" {
			v, _ := getVM("where id = ?", table.Rows[0][0])
			ShowVM(w, r, v)
		}
	} else {
		data := Tabular{
			Common: NewCommon(r, "Internal IPs"),
			Table:  table,
		}
		ShowIPs(w, r, data)
	}
}

func searchRack(w http.ResponseWriter, r *http.Request, what string) {
	ip := net.ParseIP(what)
	if ip != nil {
		showIP(w, r, what)
	} else {
		searchServers(w, r, what)
	}
}

func VMFind(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		h := r.Form.Get("hostname")
		v := vmsByQuery("where hostname like ?", "%"+h+"%")
		if len(v) == 0 {
			ErrorPage(w, r, "No vms found matching hostname: "+h)
		} else if len(v) == 1 {
			ShowVM(w, r, v[0])
		} else {
			data := VMInfo{
				Common: NewCommon(r, "VMs matching: "+h),
				VMs:    v,
			}
			renderTemplate(w, r, "vmfound", data)
		}
	}
}

func ServerEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var s Server
		objFromForm(&s, r.Form)
		user := currentUser(r)
		remote_addr := RemoteHost(r)
		s.Modified = time.Now()
		s.RemoteAddr = remote_addr
		s.UID = user.ID
		action := r.Form.Get("action")
		var err error
		if action == "Add" {
			if len(s.Hostname) == 0 {
				ErrorPage(w, r, "Hostname cannot be blank")
				return
			}
			s.ID, err = dbObjectInsert(s)
			if err != nil {
				log.Println("SERVERADD ERR:", err)
			}
		} else if action == "Update" {
			if len(s.Hostname) == 0 {
				ErrorPage(w, r, "Hostname cannot be blank")
				return
			}
			s.Update()
		} else if action == "Delete" {
			s.Delete()
		}
		auditLog(user.ID, remote_addr, action, s.String())
		dc := r.FormValue("DC")
		redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	} else {
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) < 1 {
			notFound(w, r)
		} else {
			if len(bits) > 2 {
				dc := dcLookup[strings.ToUpper(bits[0])]
				ru, _ := strconv.Atoi(bits[2])
				rid := RackID(dc.ID, bits[1])
				server := Server{
					RU:     ru,
					RID:    rid,
					Height: 1,
				}
				ShowServer(w, r, server)
			} else {
				server, err := getServer("where id=?", bits[0])
				if err != nil {
					log.Println("server error:", err)
				}
				ShowServer(w, r, server)
			}
		}
	}
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

func RackAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		rid, err := strconv.ParseInt(r.URL.Path, 0, 64)
		if err != nil {
			log.Println("bad rack id:", err)
			notFound(w, r)
			return
		}
		rack := Rack{}
		if err := dbObjectLoad(&rack, "where id=?", rid); err != nil {
			log.Println("rack id:", rid, "not found:", err)
			notFound(w, r)
			return
		}
		/*
			ips := []string{}
			ipmis := []string{}
			units, err := rack.Units()
			if err != nil {
				notFound(w, r)
				return
			}
			for _, unit := range units {
				if len(unit.IPMI) > 0 {
					ipmis = append(ipmis, unit.IPMI)
				}
				if len(unit.Internal) > 0 {
					ips = append(ips, unit.Internal)
				}
			}
		*/
		data := struct {
			Common
			Rack Rack
			/*
				PingIPMI map[string]bool
				PingIP   map[string]bool
			*/
		}{
			Common: NewCommon(r, fmt.Sprintf("Audit rack: %d (%s)", rack.Label, rack.DC())),
			Rack:   rack,
			/*
				PingIPMI: bulkPing(pingTimeout, ipmis...),
				PingIP:   bulkPing(pingTimeout, ips...),
			*/
		}
		renderTemplate(w, r, "rackaudit", data)
	}
}

func rackItemUpdate(r *http.Request, rid, ru string) error {
	hostname := r.Form.Get("hostname")
	asset := r.Form.Get("asset")
	note := r.Form.Get("note")
	sn := r.Form.Get("sn")
	height, err := strconv.Atoi(r.Form.Get("height"))
	if err != nil {
		return err
	}
	server := Server{}
	if err = dbObjectLoad(&server, "where rid=? and ru=?", rid, ru); err != nil {
		// maybe it's a router?
		if router, err2 := getRouter("where rid=? and ru=?", rid, ru); err2 == nil {
			router.Hostname = hostname
			router.AssetTag = asset
			router.Height = height
			router.SerialNo = sn
			router.Note = note
			return dbSave(&router)
		}
		return err
	}
	server.Hostname = hostname
	server.AssetTag = asset
	server.Note = note
	server.SerialNo = sn
	server.Height = height
	err = dbSave(&server)
	return err
}

func rackItemAdd(r *http.Request, rid_string, ru_string string) error {
	height, err := strconv.Atoi(r.Form.Get("height"))
	if err != nil {
		return err
	}
	ru, err := strconv.Atoi(ru_string)
	if err != nil {
		return err
	}
	rid, err := strconv.ParseInt(rid_string, 0, 64)
	if err != nil {
		return err
	}
	server := Server{
		RU:       ru,
		RID:      rid,
		Hostname: r.Form.Get("hostname"),
		AssetTag: r.Form.Get("asset"),
		SerialNo: r.Form.Get("sn"),
		Note:     r.Form.Get("note"),
		Height:   height,
	}
	return dbAdd(&server)
}

func RackUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		action := r.Form.Get("action")
		rid := r.Form.Get("rid")
		ru := r.Form.Get("ru")
		switch {
		case action == "update":
			if err := rackItemUpdate(r, rid, ru); err != nil {
				log.Println("rack updates err:", err)
				notFound(w, r)
			}
		case action == "delete":
			if err := deleteServerFromRack(rid, ru); err != nil {
				log.Println("rack delete item err:", err)
				notFound(w, r)
			}
		case action == "add":
			if err := rackItemAdd(r, rid, ru); err != nil {
				log.Println("rack add err:", err)
				notFound(w, r)
			}
		default:
			log.Println("rack updates invalid action:", action)
			notFound(w, r)
		}
		//fmt.Fprint(w, "ok")
	}
}

func RackNetwork(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var rn RackNet
		objFromForm(&rn, r.Form)
		action := r.Form.Get("action")
		OriginalVID := r.Form.Get("OriginalVID")
		if action == "Add" {
			if _, err := dbObjectInsert(rn); err != nil {
				log.Println("Racknet add error:", err)
			}
		} else if action == "Update" {
			const q = "update racknet set vid=?,first_ip=?,last_ip=? where rid=? and vid=?"
			if err := dbExec(q, rn.VID, rn.FirstIP, rn.LastIP, rn.RID, OriginalVID); err != nil {
				log.Println("Racknet update error:", err)
			}
		} else if action == "Delete" {
			const q = "delete from racknet where rid=? and vid=?"
			dbExec(q, rn.RID, rn.VID)
		}
		user := currentUser(r)
		auditLog(user.ID, RemoteHost(r), action, rn.String())
		dc := r.FormValue("DC")
		redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	}
}

func RackEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		rack := Rack{}
		r.ParseForm()
		objFromForm(&rack, r.Form)
		action := r.Form.Get("action")
		dc := r.FormValue("DC")
		var err error
		switch {
		case action == "Add":
			_, err = dbObjectInsert(rack)
			dc = rack.DC()
		case action == "Update":
			err = dbObjectUpdate(rack)
		case action == "Delete":
			err = dbObjectDelete(rack)
		}
		if err != nil {
			log.Println("RACK", action, "Error:", err)
		} else {
			user := currentUser(r)
			auditLog(user.ID, RemoteHost(r), action, rack.String())
		}
		redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	} else {
		var rack Rack
		if len(r.URL.Path) == 0 {
			rack = Rack{RUs: 45}
		} else {
			var err error
			rack, err = getRack("where id=?", r.URL.Path)
			if err != nil {
				notFound(w, r)
				return
			}
		}
		dc := dcIDs[rack.DID]
		data := struct {
			Common
			DC   string
			Rack Rack
		}{
			Common: NewCommon(r, fmt.Sprintf("Edit Rack: %d (%s)", rack.Label, dc.Name)),
			DC:     dc.Name,
			Rack:   rack,
		}
		renderTemplate(w, r, "rack", data)
	}
}

func PDUEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		pdu := PDU{}
		r.ParseForm()
		objFromForm(&pdu, r.Form)
		action := r.Form.Get("action")
		var err error
		switch {
		case action == "Add":
			_, err = dbObjectInsert(pdu)
		case action == "Update":
			err = dbObjectUpdate(pdu)
		case action == "Delete":
			err = dbObjectDelete(pdu)
		}
		if err != nil {
			log.Println("PDU", action, "Error:", err)
			fmt.Fprintln(w, err)
		} else {
			user := currentUser(r)
			auditLog(user.ID, RemoteHost(r), action, pdu.IP)
			fmt.Fprintln(w, "ok")
		}
	}
}

func ShowRacks(w http.ResponseWriter, r *http.Request, bits ...string) {
	rack, table, err := RackTable(bits...)
	if err != nil {
		log.Println("RACK ERR", err)
		notFound(w, r)
		return
	}
	data := Tabular{
		Common: NewCommon(r, "Physical Servers"),
		Table:  table,
	}
	if rack.ID > 0 {
		heading := fmt.Sprintf(`Rack <a href="%s/rack/edit/%d">%d</a>`, cfg.Main.Prefix, rack.ID, rack.Label)
		data.Common.Heading = template.HTML(heading)
	}
	ShowListing(w, r, data)
}

func ServerRack(w http.ResponseWriter, r *http.Request, s Server) {
	ShowRacks(w, r, s.DC(), strconv.Itoa(s.Rack()))
}

func RackView(w http.ResponseWriter, r *http.Request) {
	bits := strings.Split(r.URL.Path, "/")
	ShowRacks(w, r, bits...)
}

func ServerAudit(w http.ResponseWriter, r *http.Request) {
	const query = "select * from servers_history where id=? order by rowid desc"
	id, err := strconv.Atoi(r.URL.Path)
	if err != nil {
		notFound(w, r)
	} else {
		table, _ := dbTable(query, id)
		skip := []string{"rowid", "id", "uid", "login", "modified", "remote_addr"}
		data := Tabular{
			Common: NewCommon(r, "Audit History"),
			Table:  table.Diff(true, skip...),
		}
		renderTemplate(w, r, "server_audit", data)
	}
}

func VMAudit(w http.ResponseWriter, r *http.Request) {
	const query = "select * from vms_history where id=? order by rowid desc"
	id, err := strconv.Atoi(r.URL.Path)
	if err != nil {
		notFound(w, r)
	} else {
		table, _ := dbTable(query, id)
		skip := []string{"rowid", "id", "uid", "login", "modified", "remote_addr"}
		data := Tabular{
			Common: NewCommon(r, "Audit History"),
			Table:  table.Diff(true, skip...),
		}
		renderTemplate(w, r, "server_audit", data)
	}
}

func NetworkAudit(w http.ResponseWriter, r *http.Request) {
	const query = "select * from routers_history where id=? order by rowid desc"
	id, err := strconv.Atoi(r.URL.Path)
	if err != nil {
		notFound(w, r)
	} else {
		table, _ := dbTable(query, id)
		skip := []string{"rowid", "id", "uid", "login", "modified", "remote_addr"}
		data := Tabular{
			Common: NewCommon(r, "Audit History"),
			Table:  table.Diff(true, skip...),
		}
		renderTemplate(w, r, "server_audit", data)
	}
}

func NetworkAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var n Router
		objFromForm(&n, r.Form)
		n.Modified = time.Now()
		_, err := n.Insert()
		if err != nil {
			log.Println("insert error:", err)
		}
		dc := r.FormValue("DC")
		redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	} else {
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) < 2 {
			notFound(w, r)
		} else {
			dc := dcLookup[strings.ToUpper(bits[0])]
			ru, _ := strconv.Atoi(bits[2])
			router := Router{
				RU:     ru,
				RID:    RackID(dc.ID, bits[1]),
				Height: 1,
			}
			ShowRouter(w, r, router)
		}
	}
}

func NetworkNext(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path
	i, _ := strconv.ParseInt(id, 0, 64)
	next, _ := NextIPs(i)
	j, _ := json.Marshal(next)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(j))
}

func ServersCSV(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment")
	dbStreamCSV(w, serverExportQuery)
}

func ServersTab(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment")
	dbStreamTab(w, serverExportQuery)
}

func VMsCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment")
	dbStreamCSV(w, vmExportQuery)
}

func VMsTab(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment")
	dbStreamTab(w, vmExportQuery)
}

func ServersJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := dbStreamJSON(w, serverExportQuery); err != nil {
		log.Println("JSON error:", err)
	}
}

func NetworkEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var n Router
		objFromForm(&n, r.Form)
		n.Modified = time.Now()
		n.RemoteAddr = RemoteHost(r)
		action := r.Form.Get("action")
		if action == "Update" {
			n.Update()
		} else if action == "Delete" {
			n.Delete()
		}
		dc := r.FormValue("DC")
		redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	} else {
		i := strings.LastIndex(r.URL.Path, "/")
		id, err := strconv.Atoi(r.URL.Path[i+1:])
		if err != nil {
			log.Println("NETWORK ERROR:", err)
			notFound(w, r)
		} else {
			router, err := getRouter("where id=?", id)
			if err != nil {
				log.Println("get router error:", err)
			}
			ShowRouter(w, r, router)
		}
	}
}

func ShowRouter(w http.ResponseWriter, r *http.Request, router Router) {
	dc := router.DC()
	data := struct {
		Common
		DC     string
		Router Router
	}{
		Common: NewCommon(r, router.Hostname+" in "+dc),
		Router: router,
		DC:     dc,
	}
	renderTemplate(w, r, "router", data)
}

func ShowServer(w http.ResponseWriter, r *http.Request, server Server) {
	IPs := make(map[string]string)
	if len(server.IPInternal) == 0 {
		IPs, _ = NextIPs(server.RID)
		for k, v := range IPs {
			log.Println("VLAN:", k, "IP:", v)
		}
	}
	data := struct {
		Common
		Server Server
		IPs    map[string]string
	}{
		Common: NewCommon(r, server.Hostname),
		Server: server,
		IPs:    IPs,
	}
	renderTemplate(w, r, "server", data)
}

func ShowVM(w http.ResponseWriter, r *http.Request, vm VM) {
	data := struct {
		Common
		VM VM
	}{
		Common: NewCommon(r, "VM: "+vm.Hostname),
		VM:     vm,
	}
	renderTemplate(w, r, "vm", data)
}

func VMAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var v VM
		objFromForm(&v, r.Form)
		var err error
		if v.ID, err = dbObjectInsert(v); err != nil {
			log.Println("VM ADD ERROR:", err)
		}
		url := fmt.Sprintf("/server/edit/%d", v.SID)
		redirect(w, r, url, http.StatusSeeOther)
	} else {
		if len(r.URL.Path) < 1 {
			notFound(w, r)
			return
		}
		id, _ := strconv.ParseInt(r.URL.Path, 0, 64)
		vm := VM{SID: id}
		data := VMTmpl{
			Common: NewCommon(r, "Add VM"),
			VM:     vm,
		}
		renderTemplate(w, r, "vm", data)
	}
}

func VMEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var v VM
		objFromForm(&v, r.Form)
		user := currentUser(r)
		v.Modified = time.Now()
		v.RemoteAddr = RemoteHost(r)
		v.UID = user.ID
		url := fmt.Sprintf("/server/edit/%d", v.SID)
		action := r.Form.Get("action")
		if action == "Add" {
			v.Insert()
		} else if action == "Update" {
			v.Update()
		} else if action == "Delete" {
			v.Delete()
		}
		redirect(w, r, url, http.StatusSeeOther)
	} else {
		id, err := strconv.ParseInt(r.URL.Path, 0, 64)
		if err != nil {
			log.Println("Bad VM ID:", err)
			notFound(w, r)
			return
		}
		vm, _ := getVM("where id=?", id)
		data := VMTmpl{
			Common: NewCommon(r, "VM: "+vm.Hostname),
			VM:     vm,
		}
		renderTemplate(w, r, "vm", data)
	}
}

func DCEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		dc := &Datacenter{}
		objFromForm(dc, r.Form)
		user := currentUser(r)
		dc.Modified = time.Now()
		dc.RemoteAddr = RemoteHost(r)
		dc.UID = user.ID
		action := r.Form.Get("action")
		if action == "Add" {
			dbAdd(dc)
		} else if action == "Update" {
			dbSave(dc)
		} else if action == "Delete" {
			dbDelete(dc)
		}
		redirect(w, r, "/dc/list", http.StatusSeeOther)
	} else {
		dc := Datacenter{}
		if len(r.URL.Path) > 0 {
			id, err := strconv.ParseInt(r.URL.Path, 0, 64)
			if err != nil {
				log.Println("Bad DC ID:", err)
			}
			dc.ID = id
			if err := dbFindSelf(&dc); err != nil {
				log.Println("DC not found:", err)
			}
		}
		data := struct {
			Common
			Datacenter Datacenter
		}{
			Common:     NewCommon(r, "DC: "+dc.City),
			Datacenter: dc,
		}
		renderTemplate(w, r, "dc", data)
	}
}

func MacTable(w http.ResponseWriter, r *http.Request) {
	sx, err := dbObjectList(Server{})
	if err != nil {
		log.Println("error loading objects:", err)
	}
	w.Header().Set("Content-Type", "text/plain")
	servers := sx.([]Server)
	for _, s := range servers {
		p := s.IPPublic
		if len(p) == 0 {
			p = "-"
		}
		if len(s.MacPort0) > 0 {
			if v, err := ipVLAN(s.IPInternal); err == nil {
				fmt.Fprintf(w, "%s  %-15s %s  %-25s %-15s  %s\n", s.MacPort0, s.Hostname, strings.ToLower(s.DC()), v.Profile, s.IPInternal, p)
			}
		}
	}
}

func DCList(w http.ResponseWriter, r *http.Request) {
	const query = "select id,name,city from datacenters"
	table, err := dbTable(query)
	if err != nil {
		log.Println("dc query error", err)
	}
	table.Hide(0)
	setLinks(table, 1, "/dc/edit/%s", 0)
	common := NewCommon(r, "Datacenters")
	common.Heading = template.HTML(fmt.Sprintf(`Datacenters <a href="%s/dc/edit/">Add</a>`, pathPrefix))
	data := Tabular{
		Common: common,
		Table:  table,
	}
	renderTemplate(w, r, "table", data)
}

func VlanEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var v VLAN
		objFromForm(&v, r.Form)
		action := r.Form.Get("action")
		fmt.Println("VLAN ACTION:", action, "VLAN:", v)
		if action == "Add" {
			if dc, ok := dcLookup[r.Form.Get("DC")]; ok {
				v.DID = dc.ID
				v.Insert()
				LoadVLANs()
			}
		} else if action == "Update" {
			v.Update()
		} else if action == "Delete" {
			v.Delete()
		}
		redirect(w, r, "/network/vlans", http.StatusSeeOther)
	} else {
		var vlan VLAN
		title := "Add VLAN"
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) == 2 {
			var err error
			vlan, err = dcVLAN(bits[0], bits[1])
			if err != nil {
				log.Println("VLAN ERR", err)
				notFound(w, r)
				return
			}
			title = fmt.Sprintf("VLAN: %d (%s) ", vlan.Name, bits[0])
		}
		data := struct {
			Common
			VLAN VLAN
		}{
			Common: NewCommon(r, title),
			VLAN:   vlan,
		}
		renderTemplate(w, r, "vlan", data)
	}
}

func auditPage(w http.ResponseWriter, r *http.Request) {
	const query = "select id,ts,ip,login,action,msg from audit_view order by id desc"
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "Audit Log"),
		Table:  table,
	}
	renderTemplate(w, r, "audit", data)
}

func ListingPage(w http.ResponseWriter, r *http.Request) {
	const query = "select id,dc,rack,ru,hostname,alias,profile,assigned,ip_ipmi,ip_internal,ip_public,note,asset_tag,vendor_sku,sn from sview"
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "Physical Servers"),
		Table:  table,
	}
	ShowListing(w, r, data)
}

func ServerDupes(w http.ResponseWriter, r *http.Request) {
	const query = `select a.id,a.dc,a.rack,a.ru,a.hostname,a.alias,a.profile,a.assigned,a.ip_ipmi,a.ip_internal,a.ip_public,a.asset_tag,a.vendor_sku,a.sn 
	from sview a, sview b
	where a.rid = b.rid
	  and a.ru  = b.ru
	    and a.id != b.id`
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "Duplicate Servers"),
		Table:  table,
	}
	ShowListing(w, r, data)
}

func setLinks(t *dbu.Table, id int, path string, args ...int) {
	t.SetLinks(id, pathPrefix+path, args...)
}

func ShowListing(w http.ResponseWriter, r *http.Request, t Tabular) {
	t.Table.Hide(0)
	setLinks(t.Table, 1, "/rack/view/%s", 1)
	setLinks(t.Table, 2, "/rack/view/%s/%s", 1, 2)
	setLinks(t.Table, 4, "/server/edit/%s", 0)
	t.Table.AddSort(1, false)
	t.Table.AddSort(2, false)
	t.Table.AddSort(3, true)
	t.Table.SetType("ip-address", 7, 8, 9)
	renderTemplate(w, r, "table", t)
}

func renderTabular(w http.ResponseWriter, r *http.Request, table *dbu.Table, title string) {
	data := Tabular{
		Common: NewCommon(r, title),
		Table:  table,
	}
	renderTemplate(w, r, "table", data)
}

func VlansPage(w http.ResponseWriter, r *http.Request) {
	const query = "select dc,name,profile,gateway,route,netmask from dcvlans"
	table, err := dbTable(query)
	if err != nil {
		log.Println("vlans query error", err)
	}
	setLinks(table, 1, "/vlan/edit/%s/%s", 0, 1)
	table.SetType("ip-address", 2, 3)
	data := Tabular{
		Common: NewCommon(r, "VLANS"),
		Table:  table,
	}
	data.Common.Heading = template.HTML(fmt.Sprintf(`Internal VLANs <a href=%s/vlan/edit/">(add)</a>`, pathPrefix))
	renderTemplate(w, r, "table", data)
}

func NetworkDevices(w http.ResponseWriter, r *http.Request) {
	const query = "select id,dc,rack,ru,hostname,make,model,note from nview"
	table, _ := dbTable(query)
	table.Hide(0)
	setLinks(table, 4, "/network/edit/%s", 0)
	renderTabular(w, r, table, "Network Devices")
}

func ConnectionsPage(w http.ResponseWriter, r *http.Request) {
	const columns = "id,datacenter,rack,ru,hostname,profile,ip_ipmi,ip_internal,ip_public,port_eth0,port_eth1,port_ipmi,cable_eth0,cable_eth1,cable_ipmi"
	const query = "select " + columns + " from dcview"
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "Physical Server Connections"),
		Table:  table,
	}
	ShowListing(w, r, data)
}

func IPInternalList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	for _, ip := range InternalIPs() {
		fmt.Fprintln(w, ipToString(ip))
	}
}

func IPInternalAllPage(w http.ResponseWriter, r *http.Request) {
	const query = "select * from ipinside"
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "Internal IPs"),
		Table:  table,
	}
	ShowIPs(w, r, data)
}

func IPInternalDC(w http.ResponseWriter, r *http.Request) {
	const query = "select * from ipinside where dc=?"
	dc := strings.ToUpper(r.URL.Path)
	table, _ := dbTable(query, dc)
	data := Tabular{
		Common: NewCommon(r, "Internal IPs for "+dc),
		Table:  table,
	}
	ShowIPs(w, r, data)
}

func ShowIPs(w http.ResponseWriter, r *http.Request, t Tabular) {
	t.Table.Hide(0, 1, 2)
	setLinks(t.Table, 3, "/ip/dc/%s", 3)
	setLinks(t.Table, 4, "/%s/edit/%s", 1, 0)
	t.Table.SetType("ip-address", 2)
	t.Table.AddSort(2, false)
	renderTemplate(w, r, "table", t)
}

func IPPublicAllPage(w http.ResponseWriter, r *http.Request) {
	const query = "select * from ippublic"
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "Public IPs"),
		Table:  table,
	}
	ShowIPs(w, r, data)
}

func VMAllPage(w http.ResponseWriter, r *http.Request) {
	query := "select * from vmlist"
	r.ParseForm()
	args := []string{}
	opts := []string{"profile", "dc"}
	for _, opt := range opts {
		v := r.Form.Get(opt)
		if len(v) > 0 {
			args = append(args, fmt.Sprintf("%s='%s'", opt, v))
		}
	}
	if len(args) > 0 {
		query += " where " + strings.Join(args, " and ")
	}
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "VMs"),
		Table:  table,
	}
	VMListLinks(w, r, data)
}

func VMOrphans(w http.ResponseWriter, r *http.Request) {
	query := "select * from vmbad"
	table, _ := dbTable(query)
	table.Hide(0)
	setLinks(table, 2, "/vm/orphan/%s", 0)
	for _, row := range table.Rows {
		if len(row[2]) == 0 {
			row[2] = "*blank*"
		}
	}
	renderTabular(w, r, table, "Orphaned VMs")
}

func orphan(w http.ResponseWriter, r *http.Request, vm Orphan, errmsg string) {
	data := struct {
		Title       string
		ErrorMsg    string
		User        User
		VM          Orphan
		Datacenters []Datacenter
	}{
		Title:       "VM: " + vm.Hostname,
		ErrorMsg:    errmsg,
		User:        currentUser(r),
		VM:          vm,
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "orphan", data)
}

func VMOrphaned(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var o Orphan
		objFromForm(&o, r.Form)
		action := r.Form.Get("action")
		if action == "Update" {
			const q = "select id from sview where dc=? and hostname=?"
			id, err := dbGetInt(q, o.DC, o.Server)
			if err != nil {
				orphan(w, r, o, "Can't find server "+o.Server)
				return
			}
			user := currentUser(r)
			v := VM{
				SID:      int64(id),
				Hostname: o.Hostname,
				Private:  o.Private,
				Public:   o.Public,
				VIP:      o.VIP,
				Note:     o.Note,
			}
			v.Modified = time.Now()
			v.RemoteAddr = RemoteHost(r)
			v.UID = user.ID
			v.Insert()
			o.Delete()
		} else if action == "Delete" {
			o.Delete()
		}
		redirect(w, r, "/vm/orphans", http.StatusSeeOther)
	} else {
		id := r.URL.Path
		var vm Orphan
		msg := ""
		if err := dbObjectLoad(&vm, "where rowid=?", id); err != nil {
			log.Println("ORPHAN ERR", err)
			msg = err.Error()
		}
		orphan(w, r, vm, msg)
	}
}

func VMListLinks(w http.ResponseWriter, r *http.Request, data Tabular) {
	data.Table.Hide(0, 1)
	setLinks(data.Table, 3, "/server/edit/%s", 0)
	setLinks(data.Table, 4, "/vm/edit/%s", 1)
	renderTemplate(w, r, "table", data)
}

func VMListPage(w http.ResponseWriter, r *http.Request) {
	const columns = "*"
	const query = "select " + columns + " from vmlist where dc=?"
	dc := strings.ToUpper(r.URL.Path)
	table, _ := dbTable(query, dc)
	data := Tabular{
		Common: NewCommon(r, "VMs"),
		Table:  table,
	}
	VMListLinks(w, r, data)
}

func usersListPage(w http.ResponseWriter, r *http.Request) {
	Users, _ := dbObjectList(User{})
	data := struct {
		Common
		Users []User
	}{
		Common: NewCommon(r, "Datacenter Admins"),
		Users:  Users.([]User),
	}
	renderTemplate(w, r, "user_list", data)
}

type userLevel struct {
	ID   int
	Name string
}

var userLevels = []userLevel{{0, "User"}, {1, "Editor"}, {2, "Admin"}}

func UserEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var u User
		objFromForm(&u, r.Form)
		var action string
		if u.ID == 0 {
			if _, err := userAdd(u); err != nil {
				log.Println("Add error", err)
			}
			action = "added"
		} else {
			action = "modified"
			if err := dbObjectUpdate(u); err != nil {
				log.Println("update error:", err)
			}
		}
		user := currentUser(r)
		auditLog(user.ID, RemoteHost(r), action, u.Login)
		redirect(w, r, "/user/list", http.StatusSeeOther)
	} else {
		var edit User
		title := "Add User"
		if len(r.URL.Path) > 0 {
			id := r.URL.Path
			edit, _ = UserByID(id)
			title = "Edit User"
		}
		data := struct {
			Common
			EditUser User
			Levels   []userLevel
		}{
			Common:   NewCommon(r, title),
			EditUser: edit,
			Levels:   userLevels,
		}
		renderTemplate(w, r, "user_edit", data)
	}
}

func UserRun(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if u.Admin() && u.RealID == 0 && len(r.URL.Path) > 0 {
		if as, err := UserByID(r.URL.Path); err == nil {
			as.RealID = u.ID
			Remember(w, &as)
			auditLog(u.ID, RemoteHost(r), "Impersonate", as.Login)
		} else {
			log.Println("RUN ERR:", err)
		}
	}
	redirect(w, r, "/", http.StatusSeeOther)
}

func VMListing(w http.ResponseWriter, r *http.Request) {
	serverVMs := ServerVMs{}.List()
	data := struct {
		Common
		Servers []ServerVMs
	}{
		Common:  NewCommon(r, "Server VMs"),
		Servers: serverVMs,
	}
	renderTemplate(w, r, "servervms", data)
}

func DatacenterPage(w http.ResponseWriter, r *http.Request) {
	dc := strings.ToUpper(r.URL.Path)
	datacenter := dcLookup[dc]
	rx, err := dbObjectListQuery(Rack{}, "where did=? order by rack", datacenter.ID)
	if err != nil {
		log.Println("error loading objects:", err)
	}
	racks := rx.([]Rack)
	data := DCRacks{
		Common: NewCommon(r, "Servers in "+dcLookup[dc].City),
		DC:     dc,
		Racks:  racks,
	}
	renderTemplate(w, r, "datacenter", data)
}

func pingPage(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	uptime := time.Since(start_time)
	stats := strings.Join(dbStats(), "\n")
	fmt.Fprintf(w, "status: %s\nversion: %s\nhostname: %s\nstarted:%s\nuptime: %s\ndb stats:\n%s\n", status, version, Hostname, start_time, uptime, stats)
}

func DebugPage(w http.ResponseWriter, r *http.Request) {
	what := r.URL.Path
	on, _ := strconv.ParseBool(what)
	log.Println("DEBUG?", what, "ON:", on)
	dbDebug(on)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "db debug: %t\n", on)
}

func ErrorPage(w http.ResponseWriter, r *http.Request, errmsg string) {
	data := struct {
		Common
		Error string
	}{
		Common: NewCommon(r, errmsg),
		Error:  errmsg,
	}
	renderTemplate(w, r, "fail", data)
}

func loginFailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("FAIL!")
	ErrorPage(w, r, "Login failed!")
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	auditLog(user.ID, RemoteHost(r), "Logout", user.Email)
	Authorized(w, false)
	Remember(w, nil)
	redirect(w, r, "/", 302)
}

func ExcelPage(w http.ResponseWriter, r *http.Request) {
	redirect(w, r, "/", 302)
	//w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

func Authorized(w http.ResponseWriter, yes bool) {
	c := &http.Cookie{
		Name: authCookie,
		Path: "/",
	}
	if yes {
		c.Expires = time.Now().Add(time.Minute * sessionMinutes)
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
		c.Expires = time.Now().Add(time.Minute * sessionMinutes)
		c.Value = u.Cookie()
	}
	http.SetCookie(w, c)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	msg := ""
	if r.Method == "POST" {
		remote_addr := RemoteHost(r)
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		user, err := UserByEmail(username)
		if err != nil {
			msg = username + " is not authorized for access"
			auditLog(0, remote_addr, "Login", msg)
		} else if Authenticate(username, password) {
			auditLog(user.ID, remote_addr, "Login", "Login succeeded for "+username)
			Authorized(w, true)
			Remember(w, &user)
			// did we timeout and need to login before accessing a page?
			c, err := r.Cookie("redirect")
			if err == nil && len(c.Value) > 0 {
				// clear it
				c := http.Cookie{Name: "login", MaxAge: -1, Path: "/"}
				http.SetCookie(w, &c)
				//log.Println("SAVED PATH:", c.Value)
				redirect(w, r, c.Value, 302)
			} else {
				redirect(w, r, "/", 302)
			}
			return
		} else {
			auditLog(0, remote_addr, "Login", "Invalid credentials for "+username)
			msg = "Invalid login credentials"
		}
	}
	data := struct {
		Common
		ErrorMsg    string
		Placeholder string
	}{
		Common:      NewCommon(r, "Login"),
		ErrorMsg:    msg,
		Placeholder: cfg.SAML.PlaceHolder,
	}
	renderPlainTemplate(w, r, "login", data)
}

func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		if b, ok := r.Form["banner"]; ok && len(b) > 0 {
			bannerText = b[0]
		}
		if _, ok := r.Form["mode"]; ok {
			cfg.Main.ReadOnly = true
		} else {
			cfg.Main.ReadOnly = false
		}
		redirect(w, r, "/", http.StatusSeeOther)
	} else {
		common := NewCommon(r, "Edit System Settings")
		common.Banner = bannerText
		data := struct {
			Common
			ReadOnly bool
		}{
			Common:   common,
			ReadOnly: cfg.Main.ReadOnly,
		}
		renderTemplate(w, r, "banner", data)
	}
}

func APIAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var a Audit
		/*
			for k, v := range r.Form {
				log.Println("K:", k, "V:", v)
			}
		*/
		objFromForm(&a, r.Form)
		a.Hostname = strings.ToLower(a.Hostname)
		a.FQDN = a.Hostname
		i := strings.Index(a.Hostname, ".")
		if i < 0 {
			i = len(a.Hostname)
		}
		a.Hostname = a.Hostname[:i]
		//log.Println(a.Hostname, a.VMs)
		a.IP = RemoteHost(r)
		log.Println(a.IP, a.Hostname)
		err := dbReplace(&a)
		if err != nil {
			log.Println("AUDIT ERR:", err)
		}
	} else if r.Method == "GET" {
		data := struct{ URL string }{"http://" + ip + http_server + r.URL.Path}
		renderTextTemplate(w, r, "audit.sh", data)
	}
}

func APIUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		name := r.Form.Get("name")
		if len(name) == 0 {
			fmt.Fprintln(w, "'name' not specified")
			return
		}
		name = filepath.Join(uploadDir, name)
		if err = saveMultipartFile(name, file); err != nil {
			fmt.Fprintln(w, err)
			return
		}

		fmt.Fprintf(w, "File uploaded successfully : ")
		fmt.Fprintf(w, header.Filename)
	}
}

func ServerDiscover(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ipmi := r.URL.Path
	if len(ipmi) == 0 {
		log.Println("DISCOVER IPMI:", ipmi)
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

func APIUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	servers := serversByQuery("where ip_ipmi > '' and mac_eth0=''")
	for i, server := range servers {
		fmt.Fprintln(w, i, "S:", server.Hostname, "I:", server.IPIpmi, "M:", server.MacPort0)
		go server.FixMac()
		time.Sleep(1 * time.Second)
	}
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

var webHandlers = []HFunc{
	{"/favicon.ico", FaviconPage},
	{"/static/", StaticPage},
	{"/api/audit", APIAudit},
	{"/api/credentials/get", IPMICredentialsGet},
	{"/api/credentials/set", IPMICredentialsSet},
	{"/api/pings", BulkPings},
	{"/api/upload", APIUpload},
	{"/api/update", APIUpdate},
	{"/audit/log", auditPage},
	{"/data/server/discover/", ServerDiscover},
	{"/data/mactable", MacTable},
	{"/data/servers.csv", ServersCSV},
	{"/data/servers.json", ServersJSON},
	{"/data/servers.tab", ServersTab},
	{"/data/upload", DataUpload},
	{"/data/vms.csv", VMsCSV},
	{"/data/vms.tab", VMsTab},
	{"/db/debug/", DebugPage},
	{"/dc/all", ListingPage},
	{"/dc/connections", ConnectionsPage},
	{"/dc/edit/", DCEdit},
	{"/dc/list", DCList},
	{"/dc/racks/", DatacenterPage},
	{"/excel", ExcelPage},
	{"/ip/dc/", IPInternalDC},
	{"/ip/internal/all", IPInternalAllPage},
	{"/ip/internal/list", IPInternalList},
	{"/ip/public/all", IPPublicAllPage},
	{"/loginfail", loginFailHandler},
	{"/login", LoginHandler},
	{"/logout", logoutPage},
	{"/network/add/", NetworkAdd},
	{"/network/audit/", NetworkAudit},
	{"/network/devices", NetworkDevices},
	{"/network/edit/", NetworkEdit},
	{"/network/next/", NetworkNext},
	{"/network/vlans", VlansPage},
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
	{"/rack/view/", RackView},
	{"/reload", reloadPage},
	{"/search", SearchPage},
	{"/server/add/", ServerEdit},
	{"/server/audit/", ServerAudit},
	{"/server/dupes", ServerDupes},
	{"/server/edit/", ServerEdit},
	{"/server/find", ServerFind},
	{"/server/vms", VMListing},
	{"/settings", SettingsHandler},
	{"/user/add", UserEdit},
	{"/user/edit/", UserEdit},
	{"/user/list", usersListPage},
	{"/user/run/", UserRun},
	{"/vlan/edit/", VlanEdit},
	{"/vm/add/", VMAdd},
	{"/vm/all", VMAllPage},
	{"/vm/audit/", VMAudit},
	{"/vm/edit/", VMEdit},
	{"/vm/find", VMFind},
	{"/vm/list/", VMListPage},
	{"/vm/orphans", VMOrphans},
	{"/vm/orphan/", VMOrphaned},
	{"/", HomePage},
}
