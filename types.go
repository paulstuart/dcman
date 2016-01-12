package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/paulstuart/dbutil"
	"github.com/paulstuart/sshclient"
)

var (
	noNumbers = regexp.MustCompile("[^0-9]*")
	noRange   = regexp.MustCompile("-.*")
)

//go:generate dbgen

type User struct {
	ID     int64  `sql:"id" key:"true" table:"users"`
	RealID int64  // when emulating another user, retain real identity
	Login  string `sql:"login"`
	First  string `sql:"firstname"`
	Last   string `sql:"lastname"`
	Email  string `sql:"email"`
	Level  int    `sql:"admin"`
}

type Document struct {
	ID         int64     `sql:"id" key:"true" table:"documents"`
	DID        int64     `sql:"did"`
	Filename   string    `sql:"filename"`
	Title      string    `sql:"title"`
	RemoteAddr string    `sql:"remote_addr"`
	UID        int64     `sql:"user_id"`
	Modified   time.Time `sql:"modified"`
}

func (d Document) Fullpath() string {
	return path.Join(documentDir, d.Filename)
}

type Vendor struct {
	VID        int64     `sql:"vid" key:"true" table:"vendors"`
	Name       string    `sql:"name"`
	WWW        string    `sql:"www"`
	Phone      string    `sql:"phone"`
	Address    string    `sql:"address"`
	City       string    `sql:"city"`
	State      string    `sql:"state"`
	Country    string    `sql:"country"`
	Postal     string    `sql:"postal"`
	Note       string    `sql:"note"`
	RemoteAddr string    `sql:"remote_addr"`
	UID        int64     `sql:"user_id"  audit:"user"`
	Modified   time.Time `sql:"modified" audit:"time"`
}

type RMA struct {
	ID        int64     `sql:"rma_id" key:"true" table:"rmas"`
	DID       int64     `sql:"did"`
	VID       int64     `sql:"vid"`
	UID       int64     `sql:"user_id"`
	Number    string    `sql:"rma_no"`
	Note      string    `sql:"note"`
	Jira      string    `sql:"jira"`
	DCTicket  string    `sql:"dc_ticket"`
	Receiving string    `sql:"dc_receiving"`
	Opened    time.Time `sql:"date_opened"`
	Closed    time.Time `sql:"date_closed"`
}

func notReturned(s string) bool {
	return s == "return"
}

func (r RMA) Table() (*dbutil.Table, error) {
	const q = "select * from rma_action where rma_id=?"
	log.Println("RMA ID:", r.ID)
	table, err := dbTable(q, r.ID)
	if err != nil {
		return nil, err
	}
	table.Hide(0, 1, 2, 3, 4)
	setLinks(table, 7, "/server/parts/%s", 3)
	setLinksWhen(table, notReturned, 10, "/rma/return/add/%s/%s", 2, 1)
	return table, nil
}

func (r RMA) Returns() (*dbutil.Table, error) {
	const q = "select return_id,tracking_no from rma_returns where rma_id=?"
	table, err := dbTable(q, r.ID)
	if err != nil {
		return nil, err
	}
	table.Hide(0)
	setLinks(table, 1, "/rma/return/%s", 0)
	return table, nil
}

type Carrier struct {
	CarrierID int64     `sql:"cr_id" key:"true" table:"carriers"`
	Name      string    `sql:"name"`
	URL       string    `sql:"tracking_url"`
	UID       int64     `sql:"user_id"`
	Modified  time.Time `sql:"modified"`
}

func carriers() []Carrier {
	c, err := dbObjectList(Carrier{})
	if err != nil || c == nil {
		log.Println("no carriers:", err)
		return []Carrier{}
	}
	return c.([]Carrier)
}

type Return struct {
	ReturnID  int64     `sql:"return_id" key:"true" table:"rma_returns"`
	RMAID     int64     `sql:"rma_id"`
	CarrierID int64     `sql:"cr_id"`
	Tracking  string    `sql:"tracking_no"`
	UID       int64     `sql:"user_id"`
	Sent      time.Time `sql:"date_sent"`
}

type Sent struct {
	ReturnID int64 `sql:"return_id" table:"rma_sent"`
	PID      int64 `sql:"pid"`
}

func (s Sent) Unsent() bool {
	const q = "select count(*) from rma_sent where return_id=? and pid=?"
	cnt, _ := dbGetInt(q, s.ReturnID, s.PID)
	return cnt < 1
}

type Received struct {
	RMAID int64     `sql:"rma_id" table:"rma_received"`
	PID   int64     `sql:"pid"`
	UID   int64     `sql:"user_id"`
	TS    time.Time `sql:"date_received"`
}

func (r RMA) Parts() []Part {
	p, err := dbObjectListQuery(Part{}, "where rmaid=?", r.ID)
	if err != nil {
		log.Println("parts err:", err)
	}
	return p.([]Part)
}

func (r RMA) DC() string {
	if dc, ok := dcIDs[r.DID]; ok {
		return dc.Name
	}
	return ""
}

type Manufacturer struct {
	MID      int64     `sql:"mid" key:"true" table:"mfgr"`
	Name     string    `sql:"name"`
	AKA      string    `sql:"aka"`
	URL      string    `sql:"url"`
	UID      int64     `sql:"user_id"  audit:"user"`
	Modified time.Time `sql:"modified" audit:"time"`
}

func (m *Manufacturer) PageData(r *http.Request) (interface{}, error) {
	if len(r.URL.Path) > 0 {
		if err := dbFindByID(m, r.URL.Path); err != nil {
			return nil, err
		}
	}
	return struct {
		Common
		Manufacturer *Manufacturer
	}{
		Common:       NewCommon(r, "Edit"),
		Manufacturer: m,
	}, nil
}

type PartType struct {
	TID      int64     `sql:"tid" key:"true" table:"part_types"`
	Name     string    `sql:"name"`
	UID      int64     `sql:"user_id"  audit:"user"`
	Modified time.Time `sql:"modified" audit:"time"`
}

func (p *PartType) PageData(r *http.Request) (interface{}, error) {
	if len(r.URL.Path) > 0 {
		if err := dbFindByID(p, r.URL.Path); err != nil {
			return nil, err
		}
	}
	return struct {
		Common
		PartType *PartType
	}{
		Common:   NewCommon(r, "Edit"),
		PartType: p,
	}, nil
}

func partTypes() []PartType {
	pt, _ := dbObjectList(PartType{})
	return pt.([]PartType)
}

type SKU struct {
	KID         int64     `sql:"kid" key:"true" table:"skus"`
	MID         int64     `sql:"mid"`
	TID         int64     `sql:"tid"`
	PartNumber  string    `sql:"part_no"`
	Description string    `sql:"description"`
	UID         int64     `sql:"user_id"  audit:"user"`
	Modified    time.Time `sql:"modified" audit:"time"`
}

func (p *SKU) Manufacturer() Manufacturer {
	m := Manufacturer{MID: p.MID}
	dbFindSelf(&m)
	return m
}

func (s *SKU) PartType() *PartType {
	if s != nil {
		pt := &PartType{}
		if err := dbFindByID(pt, s.TID); err == nil {
			return pt
		}
	}
	return nil
}

func (p *SKU) PageData(r *http.Request) (interface{}, error) {
	if len(r.URL.Path) > 0 {
		if err := dbFindByID(p, r.URL.Path); err != nil {
			return nil, err
		}
	}
	return struct {
		Common
		SKU   *SKU
		Types []PartType
	}{
		Common: NewCommon(r, "Edit"),
		SKU:    p,
		Types:  partTypes(),
	}, nil
}

type Part struct {
	PID      int64     `sql:"pid" key:"true" table:"parts"`
	KID      int64     `sql:"kid"`    // part type lookup id
	SID      int64     `sql:"sid"`    // server id
	DID      int64     `sql:"did"`    // dc id
	RMAID    int64     `sql:"rma_id"` // rma id
	Location string    `sql:"location"`
	Serial   string    `sql:"serial_no"`
	AssetTag string    `sql:"asset_tag"`
	UID      int64     `sql:"user_id"  audit:"user"`
	Modified time.Time `sql:"modified" audit:"time"`
}

func (p *Part) SKU() *SKU {
	pl := &SKU{}
	if nil == dbFindByID(pl, p.KID) {
		return pl
	}
	return nil
}

func (p *Part) PartType() *PartType {
	if s := p.SKU(); s != nil {
		if nil == dbFindByID(s, p.KID) {
			return s.PartType()
		}
	}
	return nil
}

func (p *Part) PartNumber() string {
	if pl := p.SKU(); pl != nil {
		return pl.PartNumber
	}
	return ""
}

func (p *Part) Manufacturer() *Manufacturer {
	m := p.SKU().Manufacturer()
	return &m
}

func (p *Part) Server() *Server {
	if p != nil {
		s := Server{}
		if nil == dbFindByID(&s, p.SID) {
			return &s
		}
	}
	return nil
}

func (p *Part) Datacenter() *Datacenter {
	if p == nil {
		return nil
	}
	if p.DID > 0 {
		if dc, ok := dcIDs[p.DID]; ok {
			return &dc
		}
		return nil
	}
	if s := p.Server(); s != nil {
		return s.Datacenter()
	}
	return nil
}

func (p *Part) PageData(r *http.Request) (interface{}, error) {
	if len(r.URL.Path) > 0 {
		if err := dbFindByID(p, r.URL.Path); err != nil {
			return nil, err
		}
	}
	return struct {
		Common
		Part *Part
	}{
		Common: NewCommon(r, "Edit"),
		Part:   p,
	}, nil
}

func (p *Part) Log(action string, u User) {
	const q = "insert into part_log (pid, action, ts, user_id) values (?,?,?,?)"
	dbExec(q, p.PID, action, time.Now(), u.ID)
}

func (u User) Admin() bool {
	return u.Level > 1
}

func (u User) Editor() bool {
	return u.Level > 0 && !cfg.Main.ReadOnly
}

func (u User) Access() string {
	switch {
	case u.Level == 2:
		return "Admin"
	case u.Level == 1:
		return "Editor"
	default:
		return "User"
	}
}

type Datacenter struct {
	ID         int64     `sql:"id" key:"true" table:"datacenters"`
	Name       string    `sql:"name"`
	Address    string    `sql:"address"`
	City       string    `sql:"city"`
	State      string    `sql:"state"`
	Phone      string    `sql:"phone"`
	Web        string    `sql:"web"`
	DCMan      string    `sql:"dcman"`
	PXEHost    string    `sql:"pxehost"`
	PXEUser    string    `sql:"pxeuser"`
	PXEPass    string    `sql:"pxepass"`
	PXEKey     string    `sql:"pxekey"`
	RemoteAddr string    `sql:"remote_addr"`
	UID        int64     `sql:"user_id"  audit:"user"`
	Modified   time.Time `sql:"modified" audit:"time"`
}

func sshCmd(host, username, password, cmd string, timeout int) error {
	rc, _, _, err := sshclient.Exec(host+":22", username, password, cmd, timeout)
	if err != nil {
		return err
	}
	if rc > 0 {
		return ErrExecFailed
	}
	return nil
}

func sshTest(host, username, password string, timeout int) error {
	rc, _, _, err := sshclient.Exec(host+":22", username, password, "exit", timeout)
	if err != nil {
		return err
	}
	if rc > 0 {
		return ErrExecFailed
	}
	return nil
}

func (dc Datacenter) Remote(cmd string, timeout int) (int, string, string, error) {
	return sshclient.Exec(dc.PXEHost+":22", dc.PXEUser, dc.PXEPass, cmd, timeout)
}

func (d Datacenter) Count() int {
	c, err := dbGetInt("select count(*) from rackunits where dc=?", d.Name)
	if err != nil {
		fmt.Println("ERR!", err)
	}
	return c
}

func (d Datacenter) Selected() template.HTML {
	if thisDC.ID == d.ID {
		return template.HTML("selected")
	}
	return template.HTML("")
}

func (d Datacenter) Current() bool {
	return thisDC.ID == d.ID
}

func (d Datacenter) Racks() []Rack {
	r, err := dbObjectListQuery(Rack{}, "where did=? order by rack", d.ID)
	if err != nil {
		fmt.Println("racks err:", err)
	}
	return r.([]Rack)
}

type DCView struct {
	ID          int64     `sql:"id" key:"true" table:"dcview"`
	DID         int64     `sql:"datacenter"`
	Hostname    string    `sql:"hostname"`
	AssetNumber string    `sql:"asset_number"`
	CPU         string    `sql:"cpu_id"`
	CPU_Speed   int       `sql:"cpu_speed"`
	MemoryMB    int       `sql:"memory"`
	Created     time.Time `sql:"created" update:"false"`
}

type ServerVMs struct {
	ID       int64          `sql:"id" key:"true" table:"servervms"`
	DC       string         `sql:"dc"`
	Hostname string         `sql:"hostname"`
	VMList   sql.NullString `sql:"vms"`
	IDList   sql.NullString `sql:"ids"`
}

type VMPair struct {
	ID       int
	Hostname string
}

type RackUnit struct {
	DC       string `sql:"dc" table:"rackunits"`
	Rack     int    `sql:"rack"`
	NID      int64  `sql:"nid"`
	SID      int64  `sql:"sid"`
	RID      int64  `sql:"rid"`
	RU       int    `sql:"ru"`
	Height   int    `sql:"height"`
	Hostname string `sql:"hostname"`
	Alias    string `sql:"alias"`
	AssetTag string `sql:"asset_tag"`
	SerialNo string `sql:"sn"`
	IPMI     string `sql:"ipmi"`
	Internal string `sql:"internal"`
	Note     string `sql:"note"`
}

func (r Rack) Units() ([]RackUnit, error) {
	RUs, err := dbObjectListQuery(RackUnit{}, "where rid=? order by ru asc", r.ID)
	return RUs.([]RackUnit), err
}

func (r Rack) PDUs() ([]PDU, error) {
	if r.ID == 0 {
		return []PDU{}, nil
	}
	pdus, err := dbObjectListQuery(PDU{}, "where rid=?", r.ID)
	return pdus.([]PDU), err
}

func (r Rack) RackUnits() []RackUnit {
	size := r.RUs
	if size == 0 {
		size = 45
	}
	units := make([]RackUnit, size)
	for i := range units {
		units[size-(i+1)].RU = i + 1
		units[size-(i+1)].Height = 1
	}
	RUs, err := r.Units()
	if err != nil {
		fmt.Println("RackUnits error:", err)
	} else {
		for _, unit := range RUs {
			units[size-unit.RU] = unit
		}
	}
	// clear heights above 1U
	zero := 0
	for i := range units {
		idx := size - (i + 1)
		if units[idx].Height > 1 {
			zero = units[idx].Height - 1
			continue
		}
		if zero > 0 {
			units[idx].Height = 0
			zero--
		}

	}
	return units
}

func (s ServerVMs) VMs() []VMPair {
	vms := strings.Split(s.VMList.String, ",")
	m := make([]VMPair, len(vms))
	for i, id := range strings.Split(s.IDList.String, ",") {
		m[i].ID, _ = strconv.Atoi(id)
		m[i].Hostname = vms[i]
	}
	return m
}

func (s ServerVMs) List() []ServerVMs {
	vms, _ := dbObjectList(ServerVMs{})
	return vms.([]ServerVMs)
}

type Server struct {
	ID         int64     `sql:"id" key:"true" table:"servers"`
	RID        int64     `sql:"rid"`
	RU         int       `sql:"ru"`
	Height     int       `sql:"height"`
	Hostname   string    `sql:"hostname"`
	Alias      string    `sql:"alias"`
	Profile    string    `sql:"profile"`
	Assigned   string    `sql:"assigned"`
	Note       string    `sql:"note"`
	AssetTag   string    `sql:"asset_tag"`
	PartNo     string    `sql:"vendor_sku"`
	SerialNo   string    `sql:"sn"`
	IPInternal string    `sql:"ip_internal"`
	IPPublic   string    `sql:"ip_public"`
	IPIpmi     string    `sql:"ip_ipmi"`
	PortEth0   string    `sql:"port_eth0"`
	PortEth1   string    `sql:"port_eth1"`
	PortIpmi   string    `sql:"port_ipmi"`
	CableEth0  string    `sql:"cable_eth0"`
	CableEth1  string    `sql:"cable_eth1"`
	CableIpmi  string    `sql:"cable_ipmi"`
	MacPort0   string    `sql:"mac_eth0"`
	MacPort1   string    `sql:"mac_eth1"`
	MacIPMI    string    `sql:"mac_ipmi"`
	CPU        string    `sql:"cpu"`
	RemoteAddr string    `sql:"remote_addr"`
	Modified   time.Time `sql:"modified" audit:"time"`
	UID        int64     `sql:"uid"      audit:"user"`
}

func (s Server) InternalVLAN() string {
	v, err := findVLAN(s.DID(), s.IPInternal)
	if err != nil {
		return "vlan error:" + err.Error()
	}
	return v.String()
}

func (s Server) RunScript(script string) error {
	cmd := fmt.Sprintf(`curl -s "%sapi/script/%s" | bash`, baseURL, script)
	log.Println("RUN SCRIPT:", cmd)
	return sshCmd(s.IPInternal, cfg.SSH.Username, cfg.SSH.Password, cmd, 60)
}

func deleteServerFromRack(rid, ru string) error {
	s := Server{}
	query := fmt.Sprintf("delete from %s where rid=? and ru=?", s.TableName())
	return dbExec(query, rid, ru)
}

type Router struct {
	ID         int64     `sql:"id" key:"true" table:"routers"`
	RID        int64     `sql:"rid"`
	Height     int       `sql:"height"`
	RU         int       `sql:"ru"`
	Hostname   string    `sql:"hostname"`
	Make       string    `sql:"make"`
	Model      string    `sql:"model"`
	Note       string    `sql:"note"`
	AssetTag   string    `sql:"asset_tag"`
	MgmtIP     string    `sql:"ip_mgmt"`
	PartNo     string    `sql:"sku"`
	SerialNo   string    `sql:"sn"`
	RemoteAddr string    `sql:"remote_addr"`
	Modified   time.Time `sql:"modified"`
	UID        int       `sql:"uid"`
}

type Rack struct {
	ID       int64      `sql:"id" key:"true" table:"racks"`
	DID      int64      `sql:"did"`
	RUs      int        `sql:"rackunits"`
	Label    int        `sql:"rack"`
	VendorID string     `sql:"vendor_id"`
	XPos     string     `sql:"x_pos"`
	YPos     string     `sql:"y_pos"`
	UID      int        `sql:"uid"`
	TS       *time.Time `sql:"ts" update:"false"`
	Table    dbutil.Table
}

func (r Rack) Datacenter() Datacenter {
	dc := dcIDs[r.DID]
	return dc
}

func (r Rack) DC() string {
	dc := dcIDs[r.DID]
	return dc.Name
}

func (r Router) DC() string {
	q := "where id=?"
	rack, _ := getRack(q, r.RID)
	dc := dcIDs[rack.DID]
	return dc.Name
}

func (r Rack) String() string {
	return fmt.Sprintf("rack: %d dc: %s", r.Label, r.DC())
}

func (s Server) String() string {
	return fmt.Sprintf("server: %s dc: %s", s.Hostname, s.DC())
}

func (r Router) String() string {
	return fmt.Sprintf("router: %s dc: %s", r.Hostname, r.DC())
}

// arg 1 is dc, arg 2 is rack number
func RackTable(args ...string) (Rack, *dbutil.Table, error) {
	if len(args) == 0 {
		return Rack{}, nil, fmt.Errorf("No datacenter or rack number provided\n")
	}
	query := "select id,dc,rack,ru,hostname,alias,profile,assigned,ip_ipmi,ip_internal,ip_public,asset_tag,vendor_sku,sn from sview"
	if len(args) == 1 {
		query += " where dc=? order by dc,rack,ru desc"
		table, err := dbTable(query, args[0])
		return Rack{}, table, err
	}
	dc := dcLookup[args[0]]
	rack, err := getRack("where did=? and rack=?", dc.ID, args[1])
	if err != nil {
		return Rack{}, nil, err
	}
	query += " where rid=? order by dc,rack,ru desc"
	table, err := dbTable(query, rack.ID)
	return rack, table, err
}

type RackNet struct {
	RID     int64  `sql:"rid" table:"racknet"`
	VID     uint32 `sql:"vid"`
	CIDR    string `sql:"cidr"`
	Actual  string `sql:"actual"`
	Subnet  int    `sql:"subnet"`
	MinIP   uint32 `sql:"min_ip"`
	MaxIP   uint32 `sql:"max_ip"`
	FirstIP string `sql:"first_ip"`
	LastIP  string `sql:"last_ip"`
	next    uint32 // next free IP
	used    bool
}

func (r RackNet) String() string {
	return fmt.Sprintf("rid:%d vid:%d first:%s last:%s", r.RID, r.VID, r.FirstIP, r.LastIP)
}

func (r Rack) RackNets() []RackNet {
	rn, _ := dbObjectListQuery(RackNet{}, "where rid=? order by vid", r.ID)
	return rn.([]RackNet)
}

type RackNets []RackNet

func (a RackNets) Len() int           { return len(a) }
func (a RackNets) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RackNets) Less(i, j int) bool { return a[i].MinIP < a[j].MinIP }

type VM struct {
	ID         int64     `sql:"id" key:"true" table:"vms"`
	SID        int64     `sql:"sid"`
	Hostname   string    `sql:"hostname"`
	Private    string    `sql:"private"`
	Public     string    `sql:"public"`
	VIP        string    `sql:"vip"`
	Profile    string    `sql:"profile"`
	Note       string    `sql:"note"`
	Modified   time.Time `sql:"modified"`
	RemoteAddr string    `sql:"remote_addr"`
	UID        int64     `sql:"uid"`
}

type Orphan struct {
	ID       int64  `sql:"rowid" key:"true" table:"vmbad"`
	DC       string `sql:"dc"`
	Hostname string `sql:"hostname"`
	Private  string `sql:"private"`
	Public   string `sql:"public"`
	VIP      string `sql:"vip"`
	Note     string `sql:"note"`
	Profile  string
	Server   string
	Error    string
}

func (o Orphan) Delete() error {
	err := dbExec("delete from vmdetail where rowid=?", o.ID)
	if err != nil {
		fmt.Println("Orphan delete error", err)
	}
	return err
}

func normalColumns(words []string) {
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	for i, word := range words {
		word = strings.TrimSpace(word)
		word = strings.ToLower(word)
		word = re.ReplaceAllString(word, "_")
		words[i] = word
	}
}

func ServerColumns(words []string) error {
	columns := dbutil.GetColumns(Server{})
	for _, word := range words {
		if key, ok := columns[word]; !ok {
			// we will use these to calculate rack id
			if word == "dc" || word == "rack" {
				continue
			}
			vakid := []string{"dc", "rack"}
			for k, v := range columns {
				if !v && k != "id" && k != "rid" {
					vakid = append(vakid, k)
				}
			}
			return fmt.Errorf("invalid column: %s\nValid columns: %s", word, strings.Join(vakid, ","))
		} else if key {
			return fmt.Errorf("invalid column: %s - it is a key field and is internal only", word)
		}
	}
	return nil
}

func ServerAdd(columns, words []string) error {
	var dc, rack, ru, hostname string
	args := []interface{}{}
	params := []string{}
	for i, col := range columns {
		switch {
		case col == "dc":
			dc = strings.ToUpper(words[i])
		case col == "rack":
			rack = noNumbers.ReplaceAllLiteralString(words[i], "")
		case col == "ru":
			ru = noRange.ReplaceAllLiteralString(words[i], "")
			ru = noNumbers.ReplaceAllLiteralString(ru, "")
			args = append(args, ru)
			params = append(params, col)
		case col == "hostname":
			hostname = strings.ToLower(words[i])
			args = append(args, hostname)
			params = append(params, col)
		default:
			args = append(args, words[i])
			params = append(params, col)
		}
	}
	if len(dc) == 0 {
		return fmt.Errorf("no datacenter specified")
	}
	if len(rack) == 0 {
		return fmt.Errorf("no rack specified")
	}
	d, ok := dcLookup[dc]
	if !ok {
		return fmt.Errorf("invalid datacenter: %s", dc)
	}
	rid := RackID(d.ID, rack)
	if rid == 0 {
		var err error
		num, err := strconv.Atoi(rack)
		if err != nil {
			fmt.Printf("bad rack number for rack: %s (%s): %s\n", rack, dc, err)
		}
		rid, err = dbObjectInsert(Rack{DID: d.ID, Label: num})
		if err != nil {
			return fmt.Errorf("can't create rack: %s (%s): %s", rack, dc, err)
		}
		log.Println("added rid:", rid)
	}
	args = append(args, fmt.Sprintf("%d", rid))
	params = append(params, "rid")
	dbDebug(true)
	query := fmt.Sprintf("replace into servers (%s) values (%s)", strings.Join(params, ","), dbutil.Placeholders(len(args)))
	_, err := dbInsert(query, args...)
	dbDebug(false)
	return err
}

// an array of tab-delimited lines
func LoadServers(data []string) error {
	log.Println("Loading!")
	var columns []string
	for i, line := range data {
		if i == 0 {
			columns = strings.Split(line, "\t")
			normalColumns(columns)
			if err := ServerColumns(columns); err != nil {
				return err
			}
			continue
		}
		words := strings.Split(line, "\t")
		if err := ServerAdd(columns, words); err != nil {
			return err
		}
	}
	return nil
}

func serversByQuery(where string, args ...interface{}) []Server {
	s, _ := dbObjectListQuery(Server{}, where, args...)
	return s.([]Server)
}

func getServer(where string, args ...interface{}) (Server, error) {
	var s Server
	return s, dbObjectLoad(&s, where, args...)
}

func serverReimage(id, jira, email, menu string) error {
	if err := JiraAssigned(jira, email); err != nil {
		return err
	}
	s, err := getServer("where id=?", id)
	if err != nil {
		return err
	}
	dc := s.Datacenter()
	timeout := 30

	// pxemenu command looks for list of ips on stdin
	// it gets confused when running over ssh, so just
	// pipe the single IP so it uses stdin
	cmd := fmt.Sprintf("echo %s | pxemenu %s", s.IPIpmi, menu)
	rc, stdout, stderr, err := dc.Remote(cmd, timeout)
	if err != nil {
		return err
	}
	if rc != 0 {
		return fmt.Errorf("RC:%d OUT:%s ERR:%s", rc, stdout, stderr)
	}
	username, password, err := GetCredentials(s.IPIpmi)
	if err != nil {
		return err
	}
	if err = ipmigo(s.IPIpmi, username, password, "chassis bootdev pxe"); err != nil {
		return err
	}
	return ipmigo(s.IPIpmi, username, password, "chassis power cycle")
}

func getRouter(where string, args ...interface{}) (Router, error) {
	var s Router
	return s, dbObjectLoad(&s, where, args...)
}

func getVM(where string, args ...interface{}) (VM, error) {
	var v VM
	return v, dbObjectLoad(&v, where, args...)
}

func getVMs(serverID int64) []VM {
	v, _ := dbObjectListQuery(VM{}, "where sid=?", serverID)
	return v.([]VM)
}

func vmsByQuery(where string, args ...interface{}) []VM {
	r, _ := dbObjectListQuery(VM{}, where, args...)
	return r.([]VM)
}

func getRack(where string, args ...interface{}) (Rack, error) {
	var r Rack
	return r, dbObjectLoad(&r, where, args...)
}

func RackID(dc int64, rack string) int64 {
	q := "where did=? and rack=?"
	r, err := getRack(q, dc, rack)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rack did (%d) id (%s) error: %s\n", dc, rack, err)
	}
	return r.ID
}

func (v VM) Server() Server {
	s, _ := getServer("where id=?", v.SID)
	return s
}

func (v VM) Delete() error {
	return dbObjectDelete(v)
}

func (s Server) Delete() error {
	for _, vm := range getVMs(s.ID) {
		if err := vm.Delete(); err != nil {
			return err
		}
	}
	return dbObjectDelete(s)
}

func (s Server) Update() error {
	return dbObjectUpdate(s)
}

func (s Server) VMs() []VM {
	v, _ := dbObjectListQuery(VM{}, "where sid=?", s.ID)
	return v.([]VM)
}

func (s Server) DID() int64 {
	q := "where id=?"
	if r, err := getRack(q, s.RID); err == nil {
		return r.DID
	}
	return 0
}

func (s Server) Datacenter() *Datacenter {
	q := "where id=?"
	r, _ := getRack(q, s.RID)
	if dc, ok := dcIDs[r.DID]; ok {
		return &dc
	}
	return nil
}

func (s Server) DC() string {
	q := "where id=?"
	r, _ := getRack(q, s.RID)
	if dc, ok := dcIDs[r.DID]; ok {
		return dc.Name
	}
	return ""
}

func (s Server) Rack() int {
	r, err := getRack("where id=?", s.RID)
	if err != nil {
		fmt.Println("Server.Rack() rid:", s.RID, "error:", err)
	}
	return r.Label
}

func (r Router) Rack() int {
	rack, err := getRack("where id=?", r.RID)
	if err != nil {
		fmt.Println("Router.Rack() rid:", r.RID, "error:", err)
	}
	return rack.Label
}

func (s Router) Insert() (int64, error) {
	return dbObjectInsert(s)
}

func (s Router) Delete() error {
	return dbObjectDelete(s)
}

func (s Router) Update() error {
	return dbObjectUpdate(s)
}

func (s VM) Insert() (int64, error) {
	return dbObjectInsert(s)
}

func (s VM) Update() error {
	return dbObjectUpdate(s)
}

func getUser(where string, args ...interface{}) (User, error) {
	u := User{}
	err := dbObjectLoad(&u, where, args...)
	return u, err
}

func userUpdate(user User) error {
	return dbObjectUpdate(user)
}

func userAdd(user User) (int64, error) {
	return dbObjectInsert(user)
}

func ipFromString(in string) uint32 {
	ip := net.ParseIP(in).To4()
	if len(ip) < 4 {
		return 0
	}
	return (uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3]))
}

func ipToString(in uint32) string {
	a := in >> 24
	b := (in >> 16) & 255
	c := (in >> 8) & 255
	d := in & 255
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

type IPList []uint32

func (a IPList) Len() int           { return len(a) }
func (a IPList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a IPList) Less(i, j int) bool { return a[i] < a[j] }

func (rn RackNets) Used(ip uint32) {
	for i, r := range rn {
		if ip >= r.MinIP || ip <= r.MaxIP {
			rn[i].used = true
			return
		}
	}
}

func (rn RackNets) Done() bool {
	for _, r := range rn {
		if r.next == 0 {
			return false
		}
	}
	return true
}

func InternalIPs() IPList {
	const query = "select * from ippool"
	list, err := dbRows(query)
	if err != nil {
		fmt.Println("internal ips error", err)
	}
	sorted := make(IPList, 0, len(list))
	for _, ipv4 := range list {
		ip := ipFromString(ipv4)
		if ip > 0 {
			sorted = append(sorted, ip)
		}
	}
	sort.Sort(sorted)
	return sorted
}

func NextIPs(rid int64) (map[string]string, error) {
	next := map[string]string{}

	data, err := dbObjectListQuery(RackNet{}, "where rid=? order by min_ip", rid)
	if err != nil {
		fmt.Println("RACKNET ERR 1:", err)
		return next, err
	}
	racknets := RackNets(data.([]RackNet))

	var prior uint32
	for _, ip := range InternalIPs() {
		if prior == 0 {
			prior = ip
			continue
		}
		racknets.Used(ip)
		// gap in IPs?
		gap := prior + 1
		if gap < ip {
			for i, rn := range racknets {
				if gap < rn.MinIP || gap > rn.MaxIP {
					continue
				}
				if rn.next > 0 {
					break
				}
				// is gap stating in the middle?
				if prior < rn.MinIP || prior > rn.MaxIP {
					racknets[i].next = rn.MinIP
					//fmt.Println("SKIPPED:", rn.VID, "PRIOR:", ipToString(prior), "GAP:", ipToString(gap), "IP:", ipToString(ip))
				} else {
					//fmt.Println("NEXT   :", rn.VID, "PRIOR:", ipToString(prior), "GAP:", ipToString(gap), "IP:", ipToString(ip))
					racknets[i].next = gap
				}
				break
			}
		}
		prior = ip
	}
	for _, rn := range racknets {
		if rn.next == 0 {
			rn.next = rn.MinIP
		}
		next[strconv.Itoa(int(rn.VID))] = ipToString(rn.next)
	}
	return next, err
}

func (s Server) FixMac() {
	var err error
	if s.MacPort0, err = FindMAC(s.IPIpmi); err == nil {
		dbSave(&s)
	}
}

type Audit struct {
	Hostname string `sql:"hostname" table:"auditing"`
	IP       string `sql:"remote_addr"`
	FQDN     string `sql:"fqdn"`
	IPs      string `sql:"ips"`
	Eth0     string `sql:"eth0"`
	Eth1     string `sql:"eth1"`
	SN       string `sql:"sn"`
	Asset    string `sql:"asset"`
	IPMI_IP  string `sql:"ipmi_ip"`
	IPMI_MAC string `sql:"ipmi_mac"`
	CPU      string `sql:"cpu"`
	Mem      string `sql:"mem"`
	VMs      string `sql:"vms"`
	Kernel   string `sql:"kernel"`
	Release  string `sql:"release"`
}

type PDU struct {
	ID       int64  `sql:"id" key:"true" table:"pdus"`
	RID      int64  `sql:"rid"`
	Hostname string `sql:"hostname"`
	IP       string `sql:"ip_address"`
	Netmask  string `sql:"netmask"`
	Gateway  string `sql:"gateway"`
	DNS      string `sql:"dns"`
	AssetTag string `sql:"asset_tag"`
}

func typeID(name string) int64 {
	pt := PartType{}
	if err := dbObjectLoad(&pt, "where name=?", name); err != nil {
		log.Println("type id lookup failed:", err)
		return 0
	}
	return pt.TID
}

type DiskInfo struct {
	Size, Location, Manufacturer, PartNumber, SerialNumber string
}

func ServerImportDMI(sid int64, r io.Reader) error {
	s := Server{}
	if err := dbFindByID(&s, sid); err != nil {
		return err
	}
	dc := s.Datacenter()
	did := dc.ID
	tid := typeID("memory")

	sys := ParseDMI(r)
	if sys == nil {
		return fmt.Errorf("dmi parse failed")
	}
	for _, m := range sys.Memory {
		if len(m.Size) == 0 {
			continue
		}
		if m.PartNumber == "NO DIMM" || m.SerialNumber == "NO DIMM" {
			continue
		}
		desc := m.Size + " " + m.Speed
		if _, err := AddDevicePart(did, sid, tid, m.Manufacturer, m.PartNumber, desc, m.SerialNumber, m.AssetTag, m.Locator); err != nil {
			log.Println("add error:", err)
			return err
		}
	}
	tid = typeID("motherboard")
	b := sys.Motherboard
	if _, err := AddDevicePart(did, sid, tid, b.Manufacturer, b.ProductName, b.Type, b.SerialNumber, b.AssetTag, b.LocationInChassis); err != nil {
		log.Println("add error:", err)
		return err
	}
	return nil
}

func ServerImportDisks(sid interface{}, disks []DiskInfo) error {
	//log.Println("IMPORTING DISK SID:", sid, "RECORDS:", len(disks))
	s := Server{}
	if err := dbFindByID(&s, sid); err != nil {
		return err
	}
	dc := s.Datacenter()
	did := dc.ID
	tid := typeID("disk")

	for _, disk := range disks {
		desc := disk.Size
		if _, err := AddDevicePart(did, s.ID, tid, disk.Manufacturer, disk.PartNumber, desc, disk.SerialNumber, "", disk.Location); err != nil {
			log.Println("disk add error:", err)
			return err
		}
	}
	return nil
}
