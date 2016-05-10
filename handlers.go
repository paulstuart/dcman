package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbu "github.com/paulstuart/dbutil"
	"github.com/paulstuart/dmijson"
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
	Heading               []template.HTML
	Datacenters           []Datacenter
	Current               Datacenter
	User                  User
	PXEBoot               bool
}

func (c *Common) AddHeadings(headings ...string) {
	for _, h := range headings {
		c.Heading = append(c.Heading, template.HTML(h))
	}
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
	ErrBlankHostname  = fmt.Errorf("hostname cannot be blank")
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
		Current:     thisDC,
		User:        currentUser(r),
		Banner:      b,
		PXEBoot:     cfg.Main.PXEBoot,
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
		log.Println("UPLOADING DATA")
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
		} else {
			fmt.Fprintln(w, "ok")
		}
	} else {
		data := struct {
			Common Common
		}{
			Common: NewCommon(r, "Upload server data"),
		}
		renderTemplate(w, r, "upload", data)
	}
}

func PartsLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		data := r.Form.Get("data")
		did := r.Form.Get("DID")
		test := r.Form.Get("test")
		fmt.Println("test:", test)
		fmt.Println("DID:", did)
		fmt.Println("Data:", data)
		id, err := strconv.ParseInt(did, 0, 64)
		if err != nil {
			http.Error(w, "invalid datacenter id: "+did, http.StatusNotAcceptable)
			return
		}
		err = LoadParts(id, strings.Split(data, "\n"))
		log.Println("UPLOADING PARTS")
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
		} else {
			fmt.Fprintln(w, "ok")
		}
	} else {
		data := struct {
			Common Common
		}{
			Common: NewCommon(r, "Upload server data"),
		}
		renderTemplate(w, r, "loadparts", data)
	}
}
func DocumentEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)

		action := r.Form.Get("action")
		u := currentUser(r)
		doc := Document{}
		objFromForm(&doc, r.Form)
		doc.Modified = time.Now()
		doc.UID = u.ID
		doc.RemoteAddr = RemoteHost(r)
		switch {
		case action == "Update":
			if file, header, err := r.FormFile("file"); err == nil {
				// remove original file in case filename changed and we aren't able to simply over-write
				os.Remove(doc.Fullpath())
				doc.Filename = header.Filename
				if err := saveMultipartFile(doc.Fullpath(), file); err != nil {
					fmt.Fprintln(w, err)
					return
				}
			}
			if err := dbSave(&doc); err != nil {
				fmt.Fprintln(w, err)
				return
			}
		case action == "Delete":
			if err := os.Remove(doc.Fullpath()); err != nil {
				fmt.Fprintln(w, err)
				return
			}
			if err := dbDelete(&doc); err != nil {
				fmt.Fprintln(w, err)
				return
			}
		case action == "Add":
			file, header, err := r.FormFile("file")
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			doc.Filename = header.Filename
			err = dbAdd(&doc)
			if err != nil {
				fmt.Fprintln(w, err)
				return
			}
			if err = saveMultipartFile(doc.Fullpath(), file); err != nil {
				fmt.Fprintln(w, err)
				return
			}
		}
		redirect(w, r, "/document/list", http.StatusSeeOther)
	} else {
		var err error
		doc := Document{}
		if len(r.URL.Path) > 0 {
			if doc.ID, err = strconv.ParseInt(r.URL.Path, 0, 64); err != nil {
				notFound(w, r)
				log.Println(err)
				return
			}
			if err = dbFindSelf(&doc); err != nil {
				notFound(w, r)
				log.Println(err)
				return
			}
		}
		data := struct {
			Common   Common
			Document Document
		}{
			Common:   NewCommon(r, "Edit document"),
			Document: doc,
		}
		renderTemplate(w, r, "document", data)
	}
}

func AutoPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Common Common
	}{
		Common: NewCommon(r, "Autocomplete Test"),
	}
	renderTemplate(w, r, "auto", data)
}

func DocumentList(w http.ResponseWriter, r *http.Request) {
	t, err := dbTable("select id, did, dc, filename, title, login, modified, remote_addr from docview")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	setLinks(t, 3, "/document/get/%s", 0)
	setLinks(t, 4, "/document/edit/%s", 0)

	t.Hide(0, 1)
	data := Tabular{
		Common: NewCommon(r, "Document List"),
		Table:  t,
	}
	renderTemplate(w, r, "table", data)
}

func DocumentGet(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) == 0 {
		notFound(w, r)
		return
	}
	docid, err := strconv.ParseInt(r.URL.Path, 0, 64)
	if err != nil {
		notFound(w, r)
		log.Println(err)
		return
	}
	doc := Document{ID: docid}
	if err = dbFindSelf(&doc); err != nil {
		notFound(w, r)
		log.Println(err)
		return
	}
	log.Println("DOC:", doc)
	file, err := os.Open(doc.Fullpath())
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, r, doc.Filename, fi.ModTime(), file)
}

func ServerFind(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		h := r.Form.Get("hostname")
		s := serversByQuery("where hostname like ?", "%"+h+"%")
		if len(s) == 0 {
			ErrorPage(w, r, "No servers found matching hostname: "+h)
		} else if len(s) == 1 {
			redirectToServer(w, r, s[0])
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
		redirectToServer(w, r, s)
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
			redirectToServer(w, r, s)
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

func redirectToServer(w http.ResponseWriter, r *http.Request, s Server) {
	redirect(w, r, fmt.Sprintf("/server/edit/%d", s.ID), http.StatusSeeOther)
}

func searchServers(w http.ResponseWriter, r *http.Request, hostname string) {
	s := serversByQuery("where hostname like ?", "%"+hostname+"%")
	if len(s) == 0 {
		ErrorPage(w, r, "No servers found matching hostname: "+hostname)
	} else if len(s) == 1 {
		redirectToServer(w, r, s[0])
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

// check to see if a server can fit in a rack location
func ServerCheckFit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		rid := r.Form.Get("RID")
		hostname := r.Form.Get("Hostname")
		bot, err := strconv.Atoi(r.Form.Get("Bottom"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		top, err := strconv.Atoi(r.Form.Get("Top"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		const q = "select hostname from rackspace where rid=? and hostname != ? and (ru >= ?) and (top < ?) order by ru desc"
		hosts, err := dbRows(q, rid, hostname, bot, top)
		if err != nil || len(hosts) == 0 {
			//log.Println("check err:", err)
			fmt.Fprint(w, "ok")
		} else {
			//log.Println("conflicting hosts:", hosts)
			msg := "occupied by: " + strings.Join(hosts, ",")
			http.Error(w, msg, http.StatusBadRequest)
		}
	}
}

func ServerEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var s Server
		validate := func(action string) error {
			bad := len(s.Hostname) == 0
			switch {
			case action == "Add" && bad:
				return ErrBlankHostname
			case action == "Update" && bad:
				return ErrBlankHostname
			}
			return nil
		}
		if err := objPost(r, &s, validate); err != nil {
			log.Println("server error:", err)
		}
		dc := r.FormValue("DC")
		redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	} else {
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) < 1 {
			notFound(w, r)
			return
		}
		var server Server
		if len(bits) > 2 {
			dc := dcLookup[strings.ToUpper(bits[0])]
			ru, _ := strconv.Atoi(bits[2])
			rid := RackID(dc.ID, bits[1])
			server = Server{
				RU:     ru,
				RID:    rid,
				Height: 1,
			}
		} else {
			var err error
			if server, err = getServer("where id=?", bits[0]); err != nil {
				log.Println("server error:", err)
				notFound(w, r)
				return
			}
		}
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
}

func ServerReplace(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var s Server
		validate := func(action string) error {
			bad := len(s.Hostname) == 0
			switch {
			case action == "Add" && bad:
				return ErrBlankHostname
			case action == "Update" && bad:
				return ErrBlankHostname
			}
			return nil
		}
		if err := objPost(r, &s, validate); err != nil {
			log.Println("server error:", err)
		}
		dc := r.FormValue("DC")
		redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	} else {
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) < 1 {
			notFound(w, r)
			return
		}
		var server Server
		if len(bits) > 2 {
			dc := dcLookup[strings.ToUpper(bits[0])]
			ru, _ := strconv.Atoi(bits[2])
			rid := RackID(dc.ID, bits[1])
			server = Server{
				RU:     ru,
				RID:    rid,
				Height: 1,
			}
		} else {
			var err error
			if server, err = getServer("where id=?", bits[0]); err != nil {
				log.Println("server error:", err)
				notFound(w, r)
				return
			}
		}
		data := struct {
			Common
			Server Server
		}{
			Common: NewCommon(r, server.Hostname),
			Server: server,
		}
		renderTemplate(w, r, "replace", data)
	}
}

type Choice struct {
	ID    int64  `sql:"pid" table:"part_choices"`
	Label string `sql:"label"`
}

func PartReplace(w http.ResponseWriter, r *http.Request) {
	part := &Part{}
	if r.Method == "POST" {
		r.ParseForm()
		if err := dbFindByID(part, r.URL.Path); err != nil {
			http.Error(w, "part:"+err.Error(), http.StatusBadRequest)
			return
		}
		newPart := &Part{}
		if err := dbFindByID(newPart, r.Form.Get("newpart")); err != nil {
			log.Println("NEWPART:", r.Form.Get("newpart"))
			http.Error(w, "new part:"+err.Error(), http.StatusBadRequest)
			return
		}
		server := &Server{}
		if err := dbFindByID(server, part.SID); err != nil {
			http.Error(w, "server error:"+err.Error(), http.StatusBadRequest)
			return
		}
		u := currentUser(r)
		part.Log("removed from server: "+server.Hostname, u)
		newPart.Log("installed in server: "+server.Hostname, u)
		newPart.SID = part.SID
		part.SID = 0
		dbSave(part)
		dbSave(newPart)
		redirect(w, r, "/part/totals", http.StatusSeeOther)
	} else {
		if err := dbFindByID(part, r.URL.Path); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		const q = "select id, label from part_choices where did=? and tid=?"
		sku := part.SKU()
		if sku == nil {
			http.Error(w, "no sku for part", http.StatusBadRequest)
			return
		}
		ch, err := dbObjectListQuery(Choice{}, "where did=? and tid=?", part.DID, sku.TID)
		if err != nil {
			log.Println("query error:", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		choices := ch.([]Choice)
		data := struct {
			Common
			Server  *Server
			Part    *Part
			Choices []Choice
		}{
			Common:  NewCommon(r, "Replace part:"+part.Serial),
			Server:  part.Server(),
			Part:    part,
			Choices: choices,
		}
		renderTemplate(w, r, "replace", data)
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
		if dc, ok := dcLookup[r.URL.Path]; ok {
			for _, rack := range dc.Racks() {
				u := fmt.Sprintf("/rack/audit/%d", rack.ID)
				redirect(w, r, u, http.StatusSeeOther)
				return
			}
			notFound(w, r)
			return
		}
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
		data := struct {
			Common
			Rack Rack
		}{
			Common: NewCommon(r, fmt.Sprintf("Audit rack: %d (%s)", rack.Label, rack.DC())),
			Rack:   rack,
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

func rackItemAdd(r *http.Request, ridString, ruString string) error {
	log.Println("ADD RID:", ridString, "RU:", ruString)
	height, err := strconv.Atoi(r.Form.Get("height"))
	if err != nil {
		return err
	}
	ru, err := strconv.Atoi(ruString)
	if err != nil {
		return err
	}
	rid, err := strconv.ParseInt(ridString, 0, 64)
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
	log.Println("ADD SERVER:", server)
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
	}
}

func RackNetwork(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var rn RackNet
		objFromForm(&rn, r.Form)
		action := r.Form.Get("action")
		OriginalVID := r.Form.Get("OriginalVID")
		rn.MinIP = ipFromString(rn.FirstIP)
		rn.MaxIP = ipFromString(rn.LastIP)
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

func RMAAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		pid := r.URL.Path
		log.Println("**** RMA ADD ****** PID:", pid)
		//log.Println("**** RMA FORM:", r.Form)
		p := Part{}
		if err := dbFindByID(&p, pid); err != nil {
			log.Println("part err:", err, "PID:", pid)
			notFound(w, r)
			return
		}
		log.Println("**** part sid:", p.SID)
		// see if we already have an RMA for the server
		rma := RMA{}
		if err := objPost(r, &rma); err != nil {
			log.Println("rma error:", err)
		}
		log.Println("rma id:", rma.ID)
		p.RMAID = rma.ID
		if err := dbSave(&p); err != nil {
			log.Println("part error:", err)
		}
		log.Println("part:", p)
		redirect(w, r, "/rma/list", http.StatusSeeOther)
	} else {
		log.Println("PATH:", r.URL.Path)
		if len(r.URL.Path) == 0 {
			notFound(w, r)
			return
		}
		p := &Part{}
		if err := dbFindByID(p, r.URL.Path); err != nil {
			log.Println("PART ERR:", err)
			notFound(w, r)
			return
		}
		var did int64
		if dc := p.Datacenter(); dc != nil {
			did = dc.ID
		}
		rma := RMA{
			DID:    did,
			Opened: time.Now(),
		}
		const q = "select distinct rma_id from rma_detail where date_closed < '1970-01-01' and sid=?"
		ids, err := dbRow(q, p.SID)
		if err == nil {
			log.Println("rma_ids:", ids)
			if len(ids) > 1 {
				log.Println("too many matches:", ids)
				notFound(w, r)
				return
			}
			if len(ids) == 1 {
				if p.RMAID, err = strconv.ParseInt(ids[0], 0, 64); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if err := dbSave(p); err != nil {
					log.Println("part save err:", err)
				}
				redirect(w, r, "/rma/edit/"+ids[0], http.StatusSeeOther)
				return
			}
		} else {
			log.Println("rma_id err:", err)
		}
		//dbDebug(false)
		log.Println("PART:", *p)
		log.Println("PART NO:", p.PartNumber())
		v, err := dbObjectList(Vendor{})
		if err != nil {
			log.Println("vendor object list error:", err)
		}
		data := struct {
			Common
			RMA     RMA
			Part    *Part
			Parts   *dbu.Table
			Returns *dbu.Table
			Vendors []Vendor
		}{
			Common:  NewCommon(r, fmt.Sprintf("RMA")),
			RMA:     rma,
			Part:    p,
			Parts:   nil,
			Returns: nil,
			Vendors: v.([]Vendor),
		}
		renderTemplate(w, r, "rma", data)
	}
}

func RMAEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		rma := RMA{}
		if err := objPost(r, &rma); err != nil {
			log.Println("rma error:", err)
		}
		redirect(w, r, "/rma/list", http.StatusSeeOther)
	} else {
		rma := RMA{
			Opened: time.Now(),
		}
		log.Println("PATH:", r.URL.Path)
		if len(r.URL.Path) > 0 {
			if err := dbFindByID(&rma, r.URL.Path); err != nil {
				notFound(w, r)
				return
			}
		}
		v, err := dbObjectList(Vendor{})
		if err != nil {
			log.Println("vendor object list error:", err)
		}
		t, err := rma.Table()
		if err != nil {
			log.Println("rma table error:", err)
		}
		t.Name = "parts"
		returns, err := rma.Returns()
		if err != nil {
			log.Println("returns table error:", err)
		}
		returns.Name = "returns"
		p := Part{}
		data := struct {
			Common
			RMA     RMA
			Part    *Part
			Parts   *dbu.Table
			Returns *dbu.Table
			Vendors []Vendor
		}{
			Common:  NewCommon(r, fmt.Sprintf("RMA")),
			RMA:     rma,
			Part:    &p,
			Parts:   t,
			Returns: returns,
			Vendors: v.([]Vendor),
		}
		renderTemplate(w, r, "rma", data)
	}
}

// rma_id|return_id|did|vid|pid|sid|user_id|date_opened|date_closed|vendor_name|rma_no|dc|part_no|serial_no|jira|dc_ticket|hostname|rack|ru|note|login|ts|action

func RMAList(w http.ResponseWriter, r *http.Request) {
	const q = "select * from rma_report"
	table, err := dbTable(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	table.Adjustment(isBlank, 10)
	table.Hide(0, 1, 2, 3, 4, 5, 6)
	//setLinks(table, 3, "/server/edit/%s", 1)
	setLinks(table, 10, "/rma/edit/%s", 0)
	//setLinks(table, 9, "/vendor/edit/%s", 2)
	table.Adjustment(trimDate, 7, 8)
	renderTabular(w, r, table, "RMA Report")
}

func RMAReturnAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		ship := Return{}
		objFromForm(&ship, r.Form)
		action := r.Form.Get("action")
		pidstr := r.Form.Get("PID")
		pid, err := strconv.ParseInt(pidstr, 0, 64)
		if err != nil {
			log.Println("bad pid:", pidstr, "err:", err)
		}
		user := currentUser(r)
		ship.ModifiedBy(user.ID, time.Now())

		auditLog(user.ID, RemoteHost(r), action, "shipment")
		//dbDebug(false)
		switch {
		case action == "Add" && pid > 0 && ship.ReturnID > 0:
			sent := &Sent{ship.ReturnID, pid}
			if sent.Unsent() {
				dbAdd(sent)
			}
		case action == "Add" && pid > 0:
			dbAdd(&ship)
			sent := &Sent{ship.ReturnID, pid}
			if sent.Unsent() {
				dbAdd(sent)
			}
		case action == "Update":
			dbFindByID(&ship, r.FormValue(ship.KeyField()))
			dbSave(&ship)
		case action == "Delete":
			dbDelete(&ship)
		}
		log.Println("SHIP:", ship)
		/*
			if err := dbAdd(&sent); err != nil {
				log.Println("sent add err:", err)
			}
		*/
		var url string
		if ship.RMAID > 0 {
			url = fmt.Sprintf("/rma/edit/%d", ship.RMAID)
		} else {
			url = "/rma/list"
		}
		redirect(w, r, url, http.StatusSeeOther)
	} else {
		p := Part{}
		log.Println("RETURN ADD PATH:", r.URL.Path)
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) < 1 {
			log.Println("BAD PATH:", r.URL.Path)
			notFound(w, r)
			return
		}
		if err := dbFindByID(&p, bits[0]); err != nil {
			log.Println("BAD PID:", bits[0], "ERR:", err)
			notFound(w, r)
			return
		}
		ship := Return{RMAID: p.RMAID, Sent: time.Now()}
		if len(bits) > 1 && len(bits[1]) > 0 {
			if err := dbFindByID(&ship, bits[1]); err != nil {
				log.Println("BAD PID:", bits[1], "ERR:", err)
				notFound(w, r)
				return
			}
			sent := Sent{ship.ReturnID, p.PID}
			if sent.Unsent() {
				if err := dbAdd(&sent); err != nil {
					log.Println("db add err:", err)
				}
			}
			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
			return
		}

		// find other shipments associated w/ rma
		const q = "select distinct return_id from rma_returns where rma_id=?"
		ids, err := dbRow(q, p.RMAID)
		if err != nil {
			log.Println("ids err:", err)
		}
		if len(ids) == 1 {
			if err := dbFindByID(&ship, ids[0]); err != nil {
				log.Println("BAD PID ??:", ids[0], "ERR:", err)
				return
			}
		}
		/*
			return_id|rma_id|cr_id|tracking_no|user_id|date_sent|pid|part_no|serial_no
			1|1|2|fxid|1|2015-09-01 00:00:00|12|M393B2G70BH0-YH9|13A466B4
		*/
		const q2 = "select part_no,serial_no from rma_returned where return_id=?"
		parts, _ := dbTable(q2, ship.ReturnID)
		data := struct {
			Common
			Return   Return
			Part     *Part
			Parts    *dbu.Table
			Carriers []Carrier
		}{
			Common:   NewCommon(r, "RMA Returns"),
			Return:   ship,
			Part:     &p,
			Parts:    parts,
			Carriers: carriers(),
		}
		renderTemplate(w, r, "shipment", data)
	}
}

func RMAReturn(w http.ResponseWriter, r *http.Request) {
	ship := Return{}
	if r.Method == "POST" {
		if err := objPost(r, &ship); err != nil {
			log.Println("shipment error:", err)
		}
		log.Println("SHIP:", ship)
		var url string
		if ship.RMAID > 0 {
			url = fmt.Sprintf("/rma/edit/%d", ship.RMAID)
		} else {
			url = "/rma/list"
		}
		redirect(w, r, url, http.StatusSeeOther)
	} else {
		log.Println("RETURN PATH:", r.URL.Path)
		if len(r.URL.Path) == 0 {
			log.Println("EMPTY RETURN PATH")
			notFound(w, r)
			return
		}
		if err := dbFindByID(&ship, r.URL.Path); err != nil {
			log.Println("BAD PID:", r.URL.Path, "ERR:", err)
			notFound(w, r)
			return
		}
		const q2 = "select part_no,serial_no from rma_returned where return_id=?"
		parts, _ := dbTable(q2, ship.ReturnID)
		data := struct {
			Common
			Return   Return
			Carriers []Carrier
			Parts    *dbu.Table
			Part     *Part
		}{
			Common:   NewCommon(r, "RMA Return"),
			Return:   ship,
			Carriers: carriers(),
			Parts:    parts,
			Part:     nil,
		}
		renderTemplate(w, r, "shipment", data)
	}
}

func RMAReceived(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		pn := r.Form.Get("PartNumber")
		sn := r.Form.Get("SerialNumber")
		var tid int64
		rma := RMA{}
		if err := dbFindByID(&rma, r.URL.Path); err != nil {
			log.Println("BAD RMA ID:", r.URL.Path, "ERR:", err)
			notFound(w, r)
			return
		}
		part, err := AddDevicePart(rma.DID, 0, tid, "", pn, "", sn, "", "")
		if err != nil {
			log.Println("add device part err:", err)
		}
		log.Println("RMA ID URL:", r.URL.Path)
		log.Println("RMA:", rma)
		u := currentUser(r)
		rec := Received{
			RMAID: rma.ID,
			PID:   part.PID,
			TS:    time.Now(),
			UID:   u.ID,
		}
		if err := dbAdd(&rec); err != nil {
			log.Println("dbadd rec err:", err)
		}
		url := fmt.Sprintf("/rma/edit/%s", r.URL.Path)
		redirect(w, r, url, http.StatusSeeOther)
	} else {
		if len(r.URL.Path) == 0 {
			log.Println("EMPTY RECEIVE PATH")
			notFound(w, r)
			return
		}
		log.Println("RECEIVE PATH:", r.URL.Path)
		rma := RMA{}
		if err := dbFindByID(&rma, r.URL.Path); err != nil {
			log.Println("BAD RMA ID:", r.URL.Path, "ERR:", err)
			notFound(w, r)
			return
		}
		rec := Received{}
		data := struct {
			Common
			Received Received
			RMA      RMA
		}{
			Common:   NewCommon(r, "RMA Received"),
			Received: rec,
			RMA:      rma,
		}
		renderTemplate(w, r, "received", data)
	}
}

func MfgrEdit(w http.ResponseWriter, r *http.Request) {
	m := &Manufacturer{}
	if r.Method == "POST" {
		if err := objPost(r, m); err != nil {
			log.Println("mfgr edit error:", err)
		}
		redirect(w, r, "/mfgr/list", http.StatusSeeOther)
	} else {
		data, err := m.PageData(r)
		if err != nil {
			notFound(w, r)
			return
		}
		renderTemplate(w, r, "mfgr", data)
	}
}

func PartEdit(w http.ResponseWriter, r *http.Request) {
	s := &Part{}
	if r.Method == "POST" {
		if err := objPost(r, s); err != nil {
			log.Println("part edit error:", err)
		}
		//auditLog(user.ID, remote_addr, action, v.Name)
		redirect(w, r, "/part/totals", http.StatusSeeOther)
	} else {
		data, err := s.PageData(r)
		if err != nil {
			notFound(w, r)
			return
		}
		renderTemplate(w, r, "part", data)
	}
}

func SKUEdit(w http.ResponseWriter, r *http.Request) {
	s := &SKU{}
	if r.Method == "POST" {
		if err := objPost(r, s); err != nil {
			log.Println("part edit error:", err)
		}
		redirect(w, r, "/sku/list", http.StatusSeeOther)
	} else {
		data, err := s.PageData(r)
		if err != nil {
			notFound(w, r)
			return
		}
		renderTemplate(w, r, "sku", data)
	}
}

func ServerParts(w http.ResponseWriter, r *http.Request) {
	const q = "select * from pview where sid=?"
	id := r.URL.Path
	s := Server{}
	if err := dbFindByID(&s, id); err != nil {
		notFound(w, r)
		return
	}
	table, err := dbTable(q, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	/*
		pid|vid|sid|did|kid|tid|mid|rma_id|user_id|dc|serial_no|part_no|description|mfgr|location|part_type|login|modified|rack|ru|hostname|used
		1|1|0|1|1|1|1|0|1|AMS|fakesn|iSSD123x|ssd drive|Intel Corporation|somewhere|disk|pstuart|2015-09-29 20:08:13||||free

		pid|vid|sid|did|kid|tid|mid|rma_id|user_id|dc|serial_no|part_no|description|mfgr|location|part_type|login|modified
		3|0|2087|3|2|1|2|0|0|NYC|WD-WCC1P0740453WDC|WD2000FYYZ-01UL1B1|1.819TB|Western Digital|0,0,0|disk||0001-01-01 00:00:00

	*/
	table.Hide(0, 1, 2, 3, 4, 5, 6, 7, 8)
	table.Adjustment(isBlank, 10)
	//table.Adjustment(trimTime, 7)
	setLinks(table, 10, "/part/edit/%s", 0)
	setLinks(table, 11, "/sku/edit/%s", 4)
	setLinks(table, 13, "/mfgr/edit/%s", 6)
	heading := []string{"Part List for " + s.Hostname}
	if len(table.Rows) == 0 {
		button := scriptButton(id, "Find Parts", "/api/diskinfo/,/api/dmidecode/")
		heading = append(heading, button)
	}
	renderTabular(w, r, table, "Parts for server:"+s.Hostname, heading...)
}

func PartUse(w http.ResponseWriter, r *http.Request) {
	// pid|vid|sid|did|kid|mid|rma_id|user_id|dc|serial_no|part_no|description|mfgr|location|login|modified
	const q = "select * from pview where rma_id=0 and sid=0"
	table, err := dbTable(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	table.Hide(0, 1, 2, 3, 4, 5, 6, 7, 13)
	setLinks(table, 9, "/part/edit/%s", 0)
	setLinks(table, 10, "/sku/edit/%s", 4)
	table.Adjustment(trimTime, 15)
	heading := "Part Report"
	renderTabular(w, r, table, heading)
}

func PartTotals(w http.ResponseWriter, r *http.Request) {
	//1	1	AMS	2	iSSD123x	ssd drive
	const q = "select * from part_totals"
	table, err := dbTable(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	table.Hide(0, 1)
	/*
		setLinks(table, 9, "/part/edit/%s", 0)
		setLinks(table, 10, "/partlist/edit/%s", 4)
		table.Adjustment(trimTime, 15)
	*/
	//heading := fmt.Sprintf(`Part List <a href="%s/part/edit/">Add</a>`, pathPrefix)
	title := "Part Totals"
	heading := fmt.Sprintf(`%s <a href="%s/part/edit/">Add</a>`, title, pathPrefix)
	renderTabular(w, r, table, title, heading)
}

func partTable(dc, used, kid string) (*dbu.Table, error) {
	switch {
	case dc == "all" && len(kid) > 0 && (used == "used" || used == "free"):
		return dbTable("select * from partuse where used=? and kid=?", used, kid)
	case dc == "all" && len(kid) > 0:
		return dbTable("select * from partuse where kid=?", used, kid)
	case dc == "all" && (used == "used" || used == "free"):
		return dbTable("select * from partuse where used=?", used)
	case dc == "all":
		return dbTable("select * from partuse")
	case len(kid) > 0 && (used == "used" || used == "free"):
		return dbTable("select * from partuse where used=? and kid=? and dc=?", used, kid, dc)
	case len(kid) > 0 && used == "all":
		return dbTable("select * from partuse where kid=? and dc=?", kid, dc)
	case len(kid) > 0:
		return dbTable("select * from partuse where kid=? and dc=?", kid, dc)
	case (used == "used" || used == "free"):
		return dbTable("select * from partuse where used=? and dc=?", used, dc)
	case used == "all":
		return dbTable("select * from partuse where dc=?", dc)
	}
	return nil, fmt.Errorf("not found - dc:%s used:%s kid:%s", dc, used, kid)
}

// part/list/{{DC}|all}/{used|free|all}/{PID}}
func PartList(w http.ResponseWriter, r *http.Request) {
	bits := strings.Split(r.URL.Path, "/")
	if len(bits) < 2 {
		notFound(w, r)
		return
	}
	if len(bits) == 2 {
		bits = append(bits, "")
	}
	table, err := partTable(bits[0], bits[1], bits[2])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	/*
		pid|vid|sid|did|kid|tid|mid|rma_id|user_id|dc|serial_no|part_no|description|mfgr|location|part_type|login|modified|rack|ru|hostname|used
		3|0|2447|2|2|2|2|0|0|SFO|13A466B4|M393B2G70BH0-YH9|16384 MB 1333 MHz|Samsung|P1-DIMMA1|memory||0001-01-01 00:00:00|1|33|sfo1hyp079|used
	*/
	table.Hide(1, 2, 3, 4, 5, 6, 7, 8, 16)
	setLinks(table, 10, "/part/edit/%s", 0)
	setLinks(table, 11, "/sku/edit/%s", 4)
	/*
		table.Adjustment(trimTime, 7)
		setLinks(table, 5, "/mfgr/edit/%s", 1)
	*/
	heading := fmt.Sprintf(`Part List <a href="%s/part/edit/">Add</a>`, pathPrefix)
	renderTabular(w, r, table, heading)
}

func TypeEdit(w http.ResponseWriter, r *http.Request) {
	s := &PartType{}
	if r.Method == "POST" {
		if err := objPost(r, s); err != nil {
			log.Println("part type error:", err)
		}
		redirect(w, r, "/part/types", http.StatusSeeOther)
	} else {
		data, err := s.PageData(r)
		if err != nil {
			notFound(w, r)
			return
		}
		renderTemplate(w, r, "part_type", data)
	}
}

func TypeList(w http.ResponseWriter, r *http.Request) {
	table, err := dbTable("select tid, name from part_types")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	table.Hide(0)
	setLinks(table, 1, "/part/type/%s", 0)
	heading := fmt.Sprintf(`Part Types <a href="%s/part/type/">Add</a>`, pathPrefix)
	renderTabular(w, r, table, "Part Types", heading)
}

func SKUList(w http.ResponseWriter, r *http.Request) {
	const cols = "*"
	const q = "select " + cols + " from skuview"
	table, err := dbTable(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	/*
		kid|mid|tid|user_id|part_no|part_type|description|mfgr|login|modified
		1|1|1|1|iSSD123x|disk|ssd drive|Intel Corporation|pstuart|2015-09-29 18:22:42
	*/
	table.Hide(0, 1, 2, 3)
	setLinks(table, 4, "/sku/edit/%s", 0)
	setLinks(table, 7, "/mfgr/edit/%s", 1)
	title := "SKU List"
	if u := currentUser(r); u.Editor() {
		heading := fmt.Sprintf(`%s <a href="%s/sku/edit/">Add</a>`, title, pathPrefix)
		renderTabular(w, r, table, title, heading)
	} else {
		renderTabular(w, r, table, title)
	}
}

func TagEdit(w http.ResponseWriter, r *http.Request) {
	t := &Tag{}
	if r.Method == "POST" {
		if err := objPost(r, t); err != nil {
			log.Println("tag edit error:", err)
		}
		redirect(w, r, "/tag/list", http.StatusSeeOther)
	} else {
		data, err := t.PageData(r)
		if err != nil {
			notFound(w, r)
			return
		}
		renderTemplate(w, r, "tag", data)
	}
}

func TagList(w http.ResponseWriter, r *http.Request) {
	const q = "select * from tags"
	table, err := dbTable(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	table.Hide(0)
	setLinks(table, 1, "/tag/edit/%s", 0)
	title := "Tag List"
	if u := currentUser(r); u.Editor() {
		heading := fmt.Sprintf(`%s <a href="%s/tag/edit/">Add</a>`, title, pathPrefix)
		renderTabular(w, r, table, title, heading)
	} else {
		renderTabular(w, r, table, title)
	}
}

func VendorList(w http.ResponseWriter, r *http.Request) {
	const cols = "*"
	const q = "select " + cols + " from vendors"
	table, err := dbTable(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	table.Hide(0)
	table.Adjustment(trimDate, 12)
	table.Adjustment(userLogin, 11)
	setLinks(table, 1, "/vendor/edit/%s", 0)
	heading := fmt.Sprintf(`Vendor List <a href="%s/vendor/edit/">Add</a>`, pathPrefix)
	renderTabular(w, r, table, "Vendor List", heading)
}

func VendorEdit(w http.ResponseWriter, r *http.Request) {
	var v Vendor
	if r.Method == "POST" {
		r.ParseForm()
		objFromForm(&v, r.Form)
		user := currentUser(r)
		remoteAddr := RemoteHost(r)
		v.Modified = time.Now()
		v.RemoteAddr = remoteAddr
		v.UID = user.ID
		action := r.Form.Get("action")
		var err error
		if action == "Add" {
			err = dbAdd(&v)
		} else if action == "Update" {
			err = dbSave(&v)
		} else if action == "Delete" {
			err = dbDelete(&v)
		}
		if err != nil {
			log.Println("VENDOR ERR:", err)
		}
		auditLog(user.ID, remoteAddr, action, v.Name)
		redirect(w, r, "/vendor/list", http.StatusSeeOther)
	} else {
		dbFindByID(&v, r.URL.Path)
		data := struct {
			Common
			Vendor Vendor
		}{
			Common: NewCommon(r, v.Name),
			Vendor: v,
		}
		renderTemplate(w, r, "vendor", data)
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
		data.Common.AddHeadings(heading)
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

func RackZone(w http.ResponseWriter, r *http.Request) {
	bits := strings.Split(r.URL.Path, "/")
	if len(bits) < 1 {
		notFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
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
		table.Hide(0, 1, 2)
		renderTemplate(w, r, "table", data)
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

func DCRackList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dc, ok := dcLookup[r.URL.Path]
	if !ok {
		fmt.Fprintf(w, `{"error": "invalid dc - %s"}`, r.URL.Path)
		return
	}
	const q = "select id, rack from racks where did=? order by rack"
	t, err := dbTable(q, dc.ID)
	if err != nil {
		fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
	} else {
		j, _ := json.MarshalIndent(t.Rows, " ", "  ")
		fmt.Fprint(w, string(j))
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

type VMType struct {
	DC       string
	Server   string
	Hostname string
	VIP      string
	Internal string
	Public   string
	Profile  string
	Note     string
}

func APIVM(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var v VMType
		objFromForm(&v, r.Form)

		const q1 = "select id from sview where dc=? and hostname=?"
		sid, err := dbGetInt(q1, v.DC, v.Server)
		if err != nil {
			msg := fmt.Sprintf("could not find server %s in datacenter:%s", v.Server, v.DC)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		const q2 = "select id from vview where dc=? and sid=? and hostname=?"
		if chk, _ := dbGetInt(q2, v.DC, sid, v.Hostname); chk > 0 {
			msg := fmt.Sprintf("vm %s already exists on server %s in datacenter %s", v.Hostname, v.Server, v.DC)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		vm := VM{
			SID:      int64(sid),
			Hostname: v.Hostname,
			Private:  v.Internal,
			Public:   v.Public,
			VIP:      v.VIP,
			Profile:  v.Profile,
			Note:     v.Note,
		}
		if _, err := dbObjectInsert(vm); err != nil {
			badRequest(w, err)
			return
		}
		fmt.Fprintln(w, "ok")
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
	common.AddHeadings(fmt.Sprintf(`Datacenters <a href="%s/dc/edit/">Add</a>`, pathPrefix))
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
		if action == "Add" {
			if dc, ok := dcLookup[r.Form.Get("DC")]; ok {
				v.DID = dc.ID
				if _, err := v.Insert(); err != nil {
					fmt.Println("ADD ERROR:", err)
				}
				LoadVLANs()
			} else {
				fmt.Println("NO DC FOUND FOR:", r.Form.Get("DC"))
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
	table.Hide(0)
	renderTemplate(w, r, "table", data)
}

func ListingPage(w http.ResponseWriter, r *http.Request) {
	const query = "select id,dc,rack,ru,hostname,alias,tag,profile,assigned,ip_ipmi,ip_internal,ip_public,note,asset_tag,vendor_sku,sn from sview"
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "Physical Servers"),
		Table:  table,
	}
	ShowListing(w, r, data)
}

func ServerDupes(w http.ResponseWriter, r *http.Request) {
	// TODO: make this a view
	const query = `select * from server_dupes`
	table, _ := dbTable(query)
	data := Tabular{
		Common: NewCommon(r, "Duplicate Servers"),
		Table:  table,
	}
	ShowListing(w, r, data)
}

func ShowListing(w http.ResponseWriter, r *http.Request, t Tabular) {
	t.Table.Hide(0)
	setLinks(t.Table, 1, "/rack/view/%s", 1)
	setLinks(t.Table, 2, "/rack/view/%s/%s", 1, 2)
	setLinks(t.Table, 4, "/server/edit/%s", 0)
	t.Table.Adjustment(isBlank, 4)
	t.Table.AddSort(1, false)
	t.Table.AddSort(2, false)
	t.Table.AddSort(3, true)
	t.Table.SetType("ip-address", 7, 8, 9)
	renderTemplate(w, r, "table", t)
}

func renderTabular(w http.ResponseWriter, r *http.Request, table *dbu.Table, title string, heading ...string) {
	data := Tabular{
		Common: NewCommon(r, title),
		Table:  table,
	}
	data.Common.AddHeadings(heading...)
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
	data.Common.AddHeadings(fmt.Sprintf(`Internal VLANs <a href=%s/vlan/edit/">(add)</a>`, pathPrefix))
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
	// TODO:  make this a view
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
		remoteAddr := RemoteHost(r)
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")
		user, err := UserByEmail(username)
		if err != nil {
			msg = username + " is not authorized for access"
			auditLog(0, remoteAddr, "Login", msg)
		} else if Authenticate(username, password) {
			auditLog(user.ID, remoteAddr, "Login", "Login succeeded for "+username)
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
			auditLog(0, remoteAddr, "Login", "Invalid credentials for "+username)
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
		objFromForm(&a, r.Form)
		a.Hostname = strings.ToLower(a.Hostname)
		a.FQDN = a.Hostname
		i := strings.Index(a.Hostname, ".")
		if i < 0 {
			i = len(a.Hostname)
		}
		a.Hostname = a.Hostname[:i]
		a.IP = RemoteHost(r)
		log.Println(a.IP, a.Hostname)
		err := dbReplace(&a)
		if err != nil {
			log.Println("AUDIT ERR:", err)
		}
	} else if r.Method == "GET" {
		data := struct{ URL string }{baseURL + r.URL.Path}
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

func DiskPage(w http.ResponseWriter, r *http.Request) {
	//path := r.URL.Path[len(pathPrefix+"/api/diskinfo/"):]
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var d DiskInfo
		if err := decoder.Decode(&d); err != nil {
			log.Println("decode error:", err)
			badRequest(w, err)
			return
		}
		log.Println("DISKINFO:", d)
		if err := ServerImportDisks(d); err != nil {
			badRequest(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "ok")
	} else if r.Method == "GET" {
		data := struct{ URL string }{baseURL + r.URL.Path}
		renderTextTemplate(w, r, "diskinfo.sh", data)
	}
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

func APIUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	servers := serversByQuery("where ip_ipmi > '' and mac_eth0=''")
	for i, server := range servers {
		fmt.Fprintln(w, i, "S:", server.Hostname, "I:", server.IPIpmi, "M:", server.MacPort0)
		go server.FixMac()
		time.Sleep(1 * time.Second)
	}
}

func APIParts(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	for k, v := range r.Form {
		log.Println("K:", k, "V:", v)
	}
	pn := strings.ToLower(r.Form.Get("q")) + "%"
	const query = "select part_no from skus where part_no like ?"
	dbDebug(true)
	rows, err := dbRows(query, pn)
	dbDebug(false)
	if err != nil {
		log.Println("db error:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	j, _ := json.MarshalIndent(rows, " ", " ")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(j))
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

func scriptButton(sid, label, path string) string {
	const tmpl = `<form method="POST" action="%s/api/remote">
	<input type="submit" value="%s">
	<input type="hidden" name="path" value="%s">
	<input type="hidden" name="sid"  value="%s">
	</form>`
	return fmt.Sprintf(tmpl, pathPrefix, label, path, sid)
}

func ServerDmiDecode(w http.ResponseWriter, r *http.Request) {
	const here = "/api/dmidecode/"
	path := r.URL.Path[len(pathPrefix+here):]
	id, err := strconv.ParseInt(path, 0, 64)
	if err != nil && len(path) > 0 {
		log.Println("BAD INT:", err)
		notFound(w, r)
		return
	}

	if r.Method == "POST" {
		fmt.Println("dmidecoding")
		if id == 0 {
			badRequest(w, fmt.Errorf("id is 0"))
			return
		}
		if err := ServerImportDMI(id, r.Body); err != nil {
			badRequest(w, err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "")

	} else {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Println("dmidecode!")
		fmt.Fprintln(w, "dmidecode")
	}
}

func ServerDecodeScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if s, err := dmijson.Script(); err != nil {
		log.Println(err)
	} else {
		fmt.Fprintln(w, s)
	}
}

func APIScript(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	skip := len(pathPrefix + "/api/script/")
	script := r.URL.Path[skip:]
	if r.Method == "POST" {
		data := struct {
			BaseURL string
			UserID  int64
		}{
			BaseURL: baseURL,
			UserID:  u.ID,
		}
		renderTextTemplate(w, r, script, data)
	}
}

type KV struct {
	Key, Value string
}

type BackTalk struct {
	Script, Callback string
	Envy             []KV
}

var webHandlers = []HFunc{
	{"/favicon.ico", FaviconPage},
	{"/static/", StaticPage},
	{"/api/audit", APIAudit},
	{"/api/credentials/get", IPMICredentialsGet},
	{"/api/credentials/set", IPMICredentialsSet},
	{"/api/dmidecode/script", ServerDecodeScript},
	{"/api/dmidecode/", ServerDmiDecode},
	{"/api/diskinfo/", DiskPage},
	{"/api/parts", APIParts},
	{"/api/pings", BulkPings},
	{"/api/script/", APIScript},
	{"/api/upload", APIUpload},
	{"/api/update", APIUpdate},
	{"/api/vm", APIVM},
	{"/auto", AutoPage},
	{"/audit/log", auditPage},
	{"/data/server/discover/", ServerDiscover},
	{"/data/mactable", MacTable},
	{"/data/servers.csv", ServersCSV},
	{"/data/servers.json", ServersJSON},
	{"/data/servers.tab", ServersTab},
	{"/data/upload", DataUpload},
	{"/data/parts/load", PartsLoad},
	{"/data/vms.csv", VMsCSV},
	{"/data/vms.tab", VMsTab},
	{"/db/debug/", DebugPage},
	{"/dc/all", ListingPage},
	{"/dc/connections", ConnectionsPage},
	{"/dc/edit/", DCEdit},
	{"/dc/list", DCList},
	{"/dc/racklist/", DCRackList},
	{"/dc/racks/", DatacenterPage},
	{"/document/edit/", DocumentEdit},
	{"/document/get/", DocumentGet},
	{"/document/list", DocumentList},
	{"/excel", ExcelPage},
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
	{"/rack/zone/", RackZone},
	{"/reload", reloadPage},
	{"/search", SearchPage},
	{"/rma/list", RMAList},
	{"/rma/add/", RMAAdd},
	{"/rma/edit/", RMAEdit},
	{"/rma/received/", RMAReceived},
	{"/rma/return/add/", RMAReturnAdd},
	{"/rma/return/", RMAReturn},
	{"/server/add/", ServerEdit},
	{"/server/audit/", ServerAudit},
	{"/server/checkfit", ServerCheckFit},
	{"/server/dupes", ServerDupes},
	{"/server/edit/", ServerEdit},
	{"/server/find", ServerFind},
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
