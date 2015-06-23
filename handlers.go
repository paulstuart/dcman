package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
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
	Title       string
	Table       *dbu.Table
	User        User
	Datacenters []Datacenter
}

type RackData struct {
	Title  string
	RackID string
	DC     string
	User   User
	Server []Tuple
}

type Totals struct {
	Title string
	Table *dbu.Table
}

//
// for templates
//

type Summary struct {
	Title       string
	Datacenters []Datacenter
	User        User
	Physical    Totals
	Profiles    Totals
	VMs         []Totals
}

type DCRacks struct {
	Title       string
	Datacenters []Datacenter
	User        User
	DC          string
	Racks       []Rack
}

type BasicInfo struct {
	Title       string
	Datacenters []Datacenter
	User        User
}

type ServerInfo struct {
	Title       string
	User        User
	Datacenters []Datacenter
	Servers     []Server
}

type VMInfo struct {
	Title       string
	User        User
	Datacenters []Datacenter
	VMs         []VM
}

const (
	serverCols          = "rack,ru,hostname,profile,ip_ipmi,ip_internal,ip_public,asset_tag,vendor_sku,sn,port_eth0,port_eth1,port_ipmi,cable_eth0,cable_eth1,cable_ipmi,cpu,memory,power"
	serverQuery         = "select id," + serverCols + " from dcview where datacenter=? and hostname=?"
	serverExportColumns = "dc,rack,ru,height,asset_tag,vendor_sku,sn,profile,hostname,ip_internal,ip_ipmi,port_eth0,port_eth1,port_ipmi,cable_eth0,cable_eth1,cable_ipmi,cpu,memory,mac_port0,mac_port1,mac_ipmi,note"
	serverExportQuery   = "select " + serverExportColumns + " from sview"
	vmExportColumns     = "dc,server,vm,profile,private,public,vip"
	vmExportQuery       = "select " + vmExportColumns + " from vmlist"
)

var (
	sCols = strings.Split(serverCols, ",")
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	t, _ := dbServer.Table("select * from server_summary")
	physical := Totals{"Physical Servers", t}
	p, _ := dbServer.Table("select profile,count(profile) as total from profiles group by profile;")
	profiles := Totals{"Profiles", p}
	vms := []Totals{}
	for _, dc := range Datacenters {
		e, _ := dbServer.Table("select * from vm_summary where dc=?", dc.Name)
		vms = append(vms, Totals{dc.Location, e})
	}
	data := Summary{
		Title:       cfg.Main.Name + " DC Manager",
		Datacenters: Datacenters,
		User:        currentUser(r),
		Physical:    physical,
		Profiles:    profiles,
		VMs:         vms,
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
		t, _ = dbServer.Table("select * from profiles where profile=? and dc=?", profile, dc)
		setLinks(t, 2, "/profile/view?dc=%s", 2)
	case len(profile) > 0:
		t, _ = dbServer.Table("select * from profiles where profile=?", profile)
		setLinks(t, 2, "/profile/view?dc=%s&profile=%s", 2, 4)
	case len(dc) > 0:
		t, _ = dbServer.Table("select * from profiles where dc=?", dc)
		setLinks(t, 2, "/profile/view?dc=%s", 2)
	default:
		t, _ = dbServer.Table("select * from profiles")
		setLinks(t, 2, "/profile/view?dc=%s", 2)
	}

	t.Hide(0, 1)
	setLinks(t, 3, "/%s/edit/%s", 1, 0)
	setLinks(t, 4, "/profile/view?profile=%s", 4)
	t.SetType("ip-address", 3, 4)
	data := Tabular{
		Title:       "Profile List",
		Datacenters: Datacenters,
		User:        currentUser(r),
		Table:       t,
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
		data := BasicInfo{
			Title:       "Upload server data",
			Datacenters: Datacenters,
			User:        currentUser(r),
		}
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
				Title:       "servers matching: " + h,
				User:        currentUser(r),
				Datacenters: Datacenters,
				Servers:     s,
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
	}
}

func searchIPMI(w http.ResponseWriter, r *http.Request, ip string) {
	query := "select * from ipmstr where what='ipmi' and ip=?"
	table, _ := dbServer.Table(query, ip)
	if table == nil || len(table.Rows) == 0 {
		ErrorPage(w, r, "No assets found matching IPMI address: "+ip)
	} else if len(table.Rows) == 1 {
		s, _ := getServer("where id = ?", table.Rows[0][0])
		ShowServer(w, r, s)
	}
}

func searchIPs(w http.ResponseWriter, r *http.Request, ip string) {
	query := "select * from ipmstr where ip=?"
	table, _ := dbServer.Table(query, ip)
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
			Title:       "Internal IPs",
			Table:       table,
			User:        currentUser(r),
			Datacenters: Datacenters,
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
			Title:       "servers matching: " + hostname,
			User:        currentUser(r),
			Datacenters: Datacenters,
			Servers:     s,
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
				Title:       "VMs matching: " + hostname,
				User:        currentUser(r),
				Datacenters: Datacenters,
				VMs:         v,
			}
			renderTemplate(w, r, "vmfound", data)
		}
	}
}

func showIP(w http.ResponseWriter, r *http.Request, ip string) {
	query := "select * from ipmstr where ip=?"
	table, _ := dbServer.Table(query, ip)
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
			Title:       "Internal IPs",
			Table:       table,
			User:        currentUser(r),
			Datacenters: Datacenters,
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
				Title:       "VMs matching: " + h,
				User:        currentUser(r),
				Datacenters: Datacenters,
				VMs:         v,
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
			fmt.Println("ADD Server:", s)
			s.ID, err = dbServer.ObjectInsert(s)
			if err != nil {
				fmt.Println("SERVERADD ERR:", err)
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
					fmt.Println("server error:", err)
				}
				ShowServer(w, r, server)
			}
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
		if action == "Add" {
			if _, err := dbServer.ObjectInsert(rn); err != nil {
				fmt.Println("Racknet add error:", err)
			}
		} else if action == "Update" {
			const q = "update racknet set vid=?,first_ip=?,last_ip=? where rid=? and vid=?"
			if _, err := dbServer.Exec(q, rn.VID, rn.FirstIP, rn.LastIP, rn.RID, OriginalVID); err != nil {
				fmt.Println("Racknet update error:", err)
			}
		} else if action == "Delete" {
			const q = "delete from racknet where rid=? and vid=?"
			dbServer.Exec(q, rn.RID, rn.VID)
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
			_, err = dbServer.ObjectInsert(rack)
			dc = rack.DC()
		case action == "Update":
			err = dbServer.ObjectUpdate(rack)
		case action == "Delete":
			err = dbServer.ObjectDelete(rack)
		}
		if err != nil {
			fmt.Println("RACK", action, "Error:", err)
		} else {
			ip := strings.Split(r.RemoteAddr, ":")[0]
			user := currentUser(r)
			auditLog(user.ID, ip, action, rack.String())
		}
		redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	} else {
		var rack Rack
		bits := strings.Split(r.URL.Path, "/")
		if bits[1] == "add" {
			rack = Rack{RUs: 45}
		} else if len(bits) < 3 {
			notFound(w, r)
			return
		} else {
			rack, _ = getRack("where id=?", bits[2])
		}
		dc := dcIDs[rack.DID]
		data := struct {
			Title       string
			DC          string
			User        User
			Datacenters []Datacenter
			Rack        Rack
		}{
			Title:       fmt.Sprintf("Edit Rack: %d (%s)", rack.Label, dc.Name),
			DC:          dc.Name,
			Rack:        rack,
			User:        currentUser(r),
			Datacenters: Datacenters,
		}
		renderTemplate(w, r, "rack", data)
	}
}

func ShowRacks(w http.ResponseWriter, r *http.Request, bits ...string) {
	table, err := RackTable(bits...)
	if err != nil {
		fmt.Println("RACK ERR", err)
		notFound(w, r)
		return
	}
	data := Tabular{
		Title:       "Physical Servers",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
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
		table, _ := dbServer.Table(query, id)
		skip := []string{"rowid", "id", "uid", "login", "modified", "remote_addr"}
		data := Tabular{
			Title:       "Audit History",
			Table:       table.Diff(true, skip...),
			User:        currentUser(r),
			Datacenters: Datacenters,
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
		table, _ := dbServer.Table(query, id)
		skip := []string{"rowid", "id", "uid", "login", "modified", "remote_addr"}
		data := Tabular{
			Title:       "Audit History",
			Table:       table.Diff(true, skip...),
			User:        currentUser(r),
			Datacenters: Datacenters,
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
		table, _ := dbServer.Table(query, id)
		skip := []string{"rowid", "id", "uid", "login", "modified", "remote_addr"}
		data := Tabular{
			Title:       "Audit History",
			Table:       table.Diff(true, skip...),
			User:        currentUser(r),
			Datacenters: Datacenters,
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
			fmt.Println("insert error:", err)
		}
		dc := r.FormValue("DC")
		http.Redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
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
	fmt.Println("NEXT ID:", id)
	i, _ := strconv.ParseInt(id, 0, 64)
	next, _ := NextIPs(i)
	j, _ := json.Marshal(next)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(j))
}

func ServersCSV(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment")
	dbServer.StreamCSV(w, serverExportQuery)
}

func ServersTab(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment")
	dbServer.StreamTab(w, serverExportQuery)
}

func VMsCSV(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment")
	dbServer.StreamCSV(w, vmExportQuery)
}

func VMsTab(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment")
	dbServer.StreamTab(w, vmExportQuery)
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
		http.Redirect(w, r, "/dc/racks/"+dc, http.StatusSeeOther)
	} else {
		i := strings.LastIndex(r.URL.Path, "/")
		id, err := strconv.Atoi(r.URL.Path[i+1:])
		if err != nil {
			fmt.Println("NETWORK ERROR:", err)
			notFound(w, r)
		} else {
			router, err := getRouter("where id=?", id)
			if err != nil {
				fmt.Println("get router error:", err)
			}
			ShowRouter(w, r, router)
		}
	}
}

func ShowRouter(w http.ResponseWriter, r *http.Request, router Router) {
	dc := router.DC()
	data := struct {
		Title       string
		DC          string
		Router      Router
		User        User
		Datacenters []Datacenter
	}{
		Title:       router.Hostname + " in " + dc,
		Router:      router,
		DC:          dc,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "router", data)
}

func ShowServer(w http.ResponseWriter, r *http.Request, server Server) {
	IPs := make(map[string]string)
	if len(server.IPInternal) == 0 {
		IPs, _ = NextIPs(server.RID)
		for k, v := range IPs {
			fmt.Println("VLAN:", k, "IP:", v)
		}
	}
	data := struct {
		Title       string
		Server      Server
		User        User
		Datacenters []Datacenter
		IPs         map[string]string
	}{
		Title:       server.Hostname,
		Server:      server,
		User:        currentUser(r),
		Datacenters: Datacenters,
		IPs:         IPs,
	}
	renderTemplate(w, r, "server", data)
}

func ShowVM(w http.ResponseWriter, r *http.Request, vm VM) {
	data := struct {
		Title       string
		User        User
		VM          VM
		Datacenters []Datacenter
	}{
		Title:       "VM: " + vm.Hostname,
		User:        currentUser(r),
		VM:          vm,
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "vm", data)
}

func VMAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var v VM
		url := "/vm/all"
		objFromForm(&v, r.Form)
		var err error
		if v.ID, err = dbServer.ObjectInsert(v); err != nil {
			fmt.Println("VM ADD ERROR:", err)
			url = fmt.Sprintf("/server/edit/%d", v.SID)
		}
		redirect(w, r, url, http.StatusSeeOther)
	} else {
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) < 2 {
			notFound(w, r)
		} else {
			id, _ := strconv.ParseInt(bits[2], 0, 64)
			vm := VM{SID: id}
			data := struct {
				Title       string
				User        User
				Datacenters []Datacenter
				SID         string
				VM          VM
			}{
				Title:       "Add VM",
				User:        currentUser(r),
				Datacenters: Datacenters,
				VM:          vm,
			}
			renderTemplate(w, r, "vm", data)
		}
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
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) < 2 {
			notFound(w, r)
		} else {
			id, err := strconv.ParseInt(bits[2], 0, 64)
			if err != nil {
				fmt.Println("Bad VM ID:", err)
			}
			vm, _ := getVM("where id=?", id)
			data := struct {
				Title       string
				User        User
				VM          VM
				Datacenters []Datacenter
			}{
				Title:       "VM: " + vm.Hostname,
				User:        currentUser(r),
				VM:          vm,
				Datacenters: Datacenters,
			}
			renderTemplate(w, r, "vm", data)
		}
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
			dbServer.Add(dc)
		} else if action == "Update" {
			dbServer.Save(dc)
		} else if action == "Delete" {
			dbServer.Delete(dc)
		}
		redirect(w, r, "/dc/list", http.StatusSeeOther)
	} else {
		dc := Datacenter{}
		if len(r.URL.Path) > 0 {
			id, err := strconv.ParseInt(r.URL.Path, 0, 64)
			if err != nil {
				fmt.Println("Bad DC ID:", err)
			}
			dc.ID = id
			if err := dbServer.FindSelf(&dc); err != nil {
				fmt.Println("DC not found:", err)
			}
		}
		data := struct {
			Title       string
			User        User
			Datacenter  Datacenter
			Datacenters []Datacenter
		}{
			Title:       "DC: " + dc.Location,
			User:        currentUser(r),
			Datacenter:  dc,
			Datacenters: Datacenters,
		}
		renderTemplate(w, r, "dc", data)
	}
}

func DCList(w http.ResponseWriter, r *http.Request) {
	const query = "select id,name,location from datacenters"
	table, err := dbServer.Table(query)
	if err != nil {
		fmt.Println("dc query error", err)
	}
	table.Hide(0)
	setLinks(table, 1, "/dc/edit/%s", 0)
	data := Tabular{
		Title:       "Datacenters",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "table", data)
}

func VlanEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		var v VLAN
		fmt.Println("FORM", r.Form)
		objFromForm(&v, r.Form)
		action := r.Form.Get("action")
		if action == "Add" {
			v.Insert()
		} else if action == "Update" {
			v.Update()
		} else if action == "Delete" {
			v.Delete()
		}
		redirect(w, r, "/network/vlans", http.StatusSeeOther)
	} else {
		bits := strings.Split(r.URL.Path, "/")
		if len(bits) < 2 {
			notFound(w, r)
		} else {
			vlan, err := dcVLAN(bits[0], bits[1])
			if err != nil {
				fmt.Println("VLAN ERR", err)
				notFound(w, r)
				return
			}
			data := struct {
				Title       string
				User        User
				VLAN        VLAN
				Datacenters []Datacenter
			}{
				Title:       fmt.Sprintf("VLAN: %d (%s) ", vlan.Name, bits[0]),
				User:        currentUser(r),
				VLAN:        vlan,
				Datacenters: Datacenters,
			}
			renderTemplate(w, r, "vlan", data)
		}
	}
}

func auditPage(w http.ResponseWriter, r *http.Request) {
	const query = "select id,ts,ip,login,action,msg from audit_view order by id desc"
	table, _ := dbServer.Table(query)
	data := Tabular{
		Title:       "Audit Log",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "audit", data)
}

func ListingPage(w http.ResponseWriter, r *http.Request) {
	const query = "select id,dc,rack,ru,hostname,alias,profile,assigned,ip_ipmi,ip_internal,ip_public,note,asset_tag,vendor_sku,sn from sview"
	table, _ := dbServer.Table(query)
	data := Tabular{
		Title:       "Physical Servers",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	ShowListing(w, r, data)
}

func ServerDupes(w http.ResponseWriter, r *http.Request) {
	const query = `select a.id,a.dc,a.rack,a.ru,a.hostname,a.alias,a.profile,a.assigned,a.ip_ipmi,a.ip_internal,a.ip_public,a.asset_tag,a.vendor_sku,a.sn 
	from sview a, sview b
	where a.rid = b.rid
	  and a.ru  = b.ru
	    and a.id != b.id`
	table, _ := dbServer.Table(query)
	data := Tabular{
		Title:       "Duplicate Servers",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
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

func VlansPage(w http.ResponseWriter, r *http.Request) {
	const query = "select dc,name,gateway,route,netmask from dcvlans"
	table, err := dbServer.Table(query)
	if err != nil {
		fmt.Println("vlans query error", err)
	}
	setLinks(table, 1, "/vlan/edit/%s/%s", 0, 1)
	table.SetType("ip-address", 2, 3)
	data := Tabular{
		Title:       "VLANs",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "table", data)
}

func NetworkDevices(w http.ResponseWriter, r *http.Request) {
	const query = "select id,dc,rack,ru,hostname,make,model,note from nview"
	table, _ := dbServer.Table(query)
	data := Tabular{
		Title:       "Network Devices",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	table.Hide(0)
	setLinks(table, 4, "/network/edit/%s", 0)
	renderTemplate(w, r, "table", data)
}

func ConnectionsPage(w http.ResponseWriter, r *http.Request) {
	const columns = "id,datacenter,rack,ru,hostname,profile,ip_ipmi,ip_internal,ip_public,port_eth0,port_eth1,port_ipmi,cable_eth0,cable_eth1,cable_ipmi"
	const query = "select " + columns + " from dcview"
	table, _ := dbServer.Table(query)
	data := Tabular{
		Title:       "Physical Server Connectionss",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
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
	table, _ := dbServer.Table(query)
	data := Tabular{
		Title:       "Internal IPs",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	ShowIPs(w, r, data)
}

func IPInternalDC(w http.ResponseWriter, r *http.Request) {
	const query = "select * from ipinside where dc=?"
	dc := strings.ToUpper(r.URL.Path)
	table, _ := dbServer.Table(query, dc)
	data := Tabular{
		Title:       "Internal IPs for " + dc,
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
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
	table, _ := dbServer.Table(query)
	data := Tabular{
		Title:       "Public IPs",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
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
	table, _ := dbServer.Table(query)
	data := Tabular{
		Title:       "VMs",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	VMListLinks(w, r, data)
}

func VMOrphans(w http.ResponseWriter, r *http.Request) {
	query := "select * from vmbad"
	table, _ := dbServer.Table(query)
	table.Hide(0)
	setLinks(table, 2, "/vm/orphan/%s", 0)
	for _, row := range table.Rows {
		if len(row[2]) == 0 {
			row[2] = "*blank*"
		}
	}
	data := Tabular{
		Title:       "Orphaned VMs",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "table", data)
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
			id, err := dbServer.GetInt(q, o.DC, o.Server)
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
		if err := dbServer.ObjectLoad(&vm, "where rowid=?", id); err != nil {
			fmt.Println("ORPHAN ERR", err)
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
	table, _ := dbServer.Table(query, dc)
	data := Tabular{
		Title:       "VMs",
		Table:       table,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	VMListLinks(w, r, data)
}

func usersListPage(w http.ResponseWriter, r *http.Request) {
	Users, _ := dbServer.ObjectList(User{})
	data := struct {
		Title       string
		Users       []User
		User        User
		Datacenters []Datacenter
	}{
		Title:       "Datacenter Admins",
		Users:       Users.([]User),
		User:        currentUser(r),
		Datacenters: Datacenters,
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
				fmt.Println("Add error", err)
			}
			action = "added"
		} else {
			action = "modified"
			if err := dbServer.ObjectUpdate(u); err != nil {
				fmt.Println("update error:", err)
			}
		}
		user := currentUser(r)
		auditLog(user.ID, RemoteHost(r), action, u.Login)
		http.Redirect(w, r, "/user/list", http.StatusSeeOther)
	} else {
		var edit User
		title := "Add User"
		if len(r.URL.Path) > 0 {
			id := r.URL.Path
			edit, _ = UserByID(id)
			title = "Edit User"
		}
		data := struct {
			Title       string
			User        User
			EditUser    User
			Levels      []userLevel
			Datacenters []Datacenter
		}{
			Title:       title,
			User:        currentUser(r),
			EditUser:    edit,
			Levels:      userLevels,
			Datacenters: Datacenters,
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
			fmt.Println("RUN ERR:", err)
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func VMListing(w http.ResponseWriter, r *http.Request) {
	serverVMs := ServerVMs{}.List()
	data := struct {
		Title       string
		User        User
		Servers     []ServerVMs
		Datacenters []Datacenter
	}{
		Title:       "Server VMs",
		User:        currentUser(r),
		Servers:     serverVMs,
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "servervms", data)
}

func DatacenterPage(w http.ResponseWriter, r *http.Request) {
	dc := strings.ToUpper(r.URL.Path)
	datacenter := dcLookup[dc]
	rx, err := dbServer.ObjectListQuery(Rack{}, "where did=? order by rack", datacenter.ID)
	if err != nil {
		fmt.Println("error loading objects:", err)
	}
	racks := rx.([]Rack)
	data := DCRacks{
		Title:       "Servers in " + dcLookup[dc].Location,
		DC:          dc,
		Racks:       racks,
		Datacenters: Datacenters,
		User:        currentUser(r),
	}
	renderTemplate(w, r, "datacenter", data)
}

func pingPage(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	uptime := time.Since(start_time)
	stats := strings.Join(dbServer.Stats(), "\n")
	fmt.Fprintf(w, "status: %s\nversion: %s\nhostname: %s\nstarted:%s\nuptime: %s\ndb stats:\n%s\n", status, version, Hostname, start_time, uptime, stats)
}

func DebugPage(w http.ResponseWriter, r *http.Request) {
	what := r.URL.Path
	on, _ := strconv.ParseBool(what)
	fmt.Println("DEBUG?", what, "ON:", on)
	dbServer.Debug = true
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "db debug: %t\n", on)
}

func ErrorPage(w http.ResponseWriter, r *http.Request, errmsg string) {
	data := struct {
		Title       string
		Error       string
		User        User
		Datacenters []Datacenter
	}{
		Title:       errmsg,
		Error:       errmsg,
		User:        currentUser(r),
		Datacenters: Datacenters,
	}
	renderTemplate(w, r, "fail", data)
}

func loginFailHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("FAIL!")
	ErrorPage(w, r, "Login failed!")
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	auditLog(user.ID, RemoteHost(r), "Logout", user.Email)
	Authorized(w, false)
	Remember(w, nil)
	http.Redirect(w, r, "/", 302)
}

func ExcelPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 302)
	//w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

func Authorized(w http.ResponseWriter, yes bool) {
	c := &http.Cookie{
		Name: cfg.SAML.OKTACookie,
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
		Name: "dcuser",
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
			redirect(w, r, "/", 302)
			return
		} else {
			auditLog(0, remote_addr, "Login", "Invalid credentials for "+username)
			msg = "Invalid login credentials"
		}
	}
	data := struct{ ErrorMsg, Placeholder string }{msg, cfg.SAML.PlaceHolder}
	renderPlainTemplate(w, r, "login", data)
}

var webHandlers = []HFunc{
	{"/favicon.ico", FaviconPage},
	{"/static/", StaticPage},
	{"/login", LoginHandler},
	{"/loginfail", loginFailHandler},
	{"/logout", logoutPage},
	{"/audit/log", auditPage},
	{"/user/list", usersListPage},
	{"/user/add", UserEdit},
	{"/user/edit/", UserEdit},
	{"/user/run/", UserRun},
	{"/rack/add", RackEdit},
	{"/dc/edit/", DCEdit},
	{"/dc/racks/", DatacenterPage},
	{"/dc/list", DCList},
	{"/dc/all", ListingPage},
	{"/dc/connections", ConnectionsPage},
	{"/ip/dc/", IPInternalDC},
	{"/ip/internal/all", IPInternalAllPage},
	{"/ip/internal/list", IPInternalList},
	{"/ip/public/all", IPPublicAllPage},
	{"/rack/edit/", RackEdit},
	{"/rack/view/", RackView},
	{"/server/find", ServerFind},
	{"/server/vms", VMListing},
	{"/server/add/", ServerEdit},
	{"/server/edit/", ServerEdit},
	{"/server/audit/", ServerAudit},
	{"/server/dupes", ServerDupes},
	{"/network/devices", NetworkDevices},
	{"/network/add/", NetworkAdd},
	{"/network/edit/", NetworkEdit},
	{"/network/next/", NetworkNext},
	{"/network/audit/", NetworkAudit},
	{"/network/vlans", VlansPage},
	{"/vlan/edit/", VlanEdit},
	{"/rack/network", RackNetwork},
	{"/profile/view", ProfileView},
	{"/vm/add/", VMAdd},
	{"/vm/edit/", VMEdit},
	{"/vm/find", VMFind},
	{"/vm/all", VMAllPage},
	{"/vm/list/", VMListPage},
	{"/vm/audit/", VMAudit},
	{"/vm/orphans", VMOrphans},
	{"/vm/orphan/", VMOrphaned},
	{"/db/debug/", DebugPage},
	{"/ping", pingPage},
	{"/reload", reloadPage},
	{"/search", SearchPage},
	//{"/password", PasswordReset},
	{"/excel", ExcelPage},
	{"/data/servers.csv", ServersCSV},
	{"/data/servers.tab", ServersTab},
	{"/data/vms.csv", VMsCSV},
	{"/data/vms.tab", VMsTab},
	{"/data/upload", DataUpload},
	{"/", HomePage},
}
