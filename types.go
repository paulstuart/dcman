package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	dbu "github.com/paulstuart/dbutil"
)

var (
	noNumbers = regexp.MustCompile("[^0-9]*")
	noRange   = regexp.MustCompile("-.*")
)

//go:generate dbgen

type User struct {
	ID       int64  `sql:"id" key:"true" table:"users"`
	RealID   int64  // when emulating another user, retain real identity
	Login    string `sql:"login"`
	First    string `sql:"firstname"`
	Last     string `sql:"lastname"`
	Email    string `sql:"email"`
	Password string `sql:"password"`
	Level    int    `sql:"admin"`
}

func (u User) Admin() bool {
	return u.Level > 1
}

func (u User) Editor() bool {
	return u.Level > 0
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
	Location   string    `sql:"location"`
	Modified   time.Time `sql:"modified"`
	RemoteAddr string    `sql:"remote_addr"`
	UID        int64     `sql:"uid"`
}

func (d Datacenter) Count() int {
	c, err := dbServer.GetInt("select count(*) from rackunits where dc=?", d.Name)
	if err != nil {
		fmt.Println("ERR!", err)
	}
	return c
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
	IPMI     string `sql:"ipmi"`
	Internal string `sql:"internal"`
}

func (r Rack) Units() ([]RackUnit, error) {
	RUs, err := dbServer.ObjectListQuery(RackUnit{}, "where rid=? order by ru asc", r.ID)
	return RUs.([]RackUnit), err
}

func (r Rack) PDUs() ([]PDU, error) {
	if r.ID == 0 {
		return []PDU{}, nil
	}
	pdus, err := dbServer.ObjectListQuery(PDU{}, "where rid=?", r.ID)
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
			//fmt.Println("ZERO RU:",units[idx].RU)
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
	vms, _ := dbServer.ObjectList(ServerVMs{})
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
	Modified   time.Time `sql:"modified"`
	RemoteAddr string    `sql:"remote_addr"`
	UID        int64     `sql:"uid"`
}

func (s Server) InternalVLAN() string {
	v, err := findVLAN(s.DID(), s.IPInternal)
	if err != nil {
		return "vlan error:" + err.Error()
	}
	return v.String()
}

func deleteServerFromRack(rid, ru string) error {
	s := Server{}
	query := fmt.Sprintf("delete from %s where rid=? and ru=?", s.TableName())
	_, err := dbServer.Exec(query, rid, ru)
	return err
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
	Table    dbu.Table
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
func RackTable(args ...string) (Rack, *dbu.Table, error) {
	if len(args) == 0 {
		return Rack{}, nil, fmt.Errorf("No datacenter or rack number provided\n")
	}
	query := "select id,dc,rack,ru,hostname,alias,profile,assigned,ip_ipmi,ip_internal,ip_public,asset_tag,vendor_sku,sn from sview"
	if len(args) == 1 {
		query += " where dc=? order by dc,rack,ru desc"
		table, err := dbServer.Table(query, args[0])
		return Rack{}, table, err
	}
	dc := dcLookup[args[0]]
	rack, err := getRack("where did=? and rack=?", dc.ID, args[1])
	if err != nil {
		return Rack{}, nil, err
	}
	query += " where rid=? order by dc,rack,ru desc"
	table, err := dbServer.Table(query, rack.ID)
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
	rn, _ := dbServer.ObjectListQuery(RackNet{}, "where rid=? order by vid", r.ID)
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
	_, err := dbServer.Exec("delete from vmdetail where rowid=?", o.ID)
	if err != nil {
		fmt.Println("Orphan delete error", err)
	}
	return err
}

// for loading an object from an http post
func objFromForm(obj interface{}, values map[string][]string) {
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
			default:
				fmt.Println("unhandled field type for:", f.Name, "type:", b.Type())
			}
		}
	}
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
	columns := dbu.GetColumns(Server{})
	for _, word := range words {
		if key, ok := columns[word]; !ok {
			// we will use these to calculate rack id
			if word == "dc" || word == "rack" {
				continue
			}
			valid := []string{"dc", "rack"}
			for k, v := range columns {
				if !v && k != "id" && k != "rid" {
					valid = append(valid, k)
				}
			}
			return fmt.Errorf("invalid column: %s\nValid columns: %s", word, strings.Join(valid, ","))
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
			fmt.Println("bad rack number for rack: %s (%s): %s", rack, dc, err)
		}
		rid, err = dbServer.ObjectInsert(Rack{DID: d.ID, Label: num})
		if err != nil {
			return fmt.Errorf("can't create rack: %s (%s): %s", rack, dc, err)
		}
	}
	args = append(args, fmt.Sprintf("%d", rid))
	params = append(params, "rid")
	query := fmt.Sprintf("replace into servers (%s) values (%s)", strings.Join(params, ","), dbu.Placeholders(len(args)))
	_, err := dbServer.Insert(query, args...)
	return err
}

// an array of tab-delimited lines
func LoadServers(data []string) error {
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
	s, _ := dbServer.ObjectListQuery(Server{}, where, args...)
	return s.([]Server)
}

func getServer(where string, args ...interface{}) (Server, error) {
	var s Server
	return s, dbServer.ObjectLoad(&s, where, args...)
}

func getRouter(where string, args ...interface{}) (Router, error) {
	var s Router
	return s, dbServer.ObjectLoad(&s, where, args...)
}

func getVM(where string, args ...interface{}) (VM, error) {
	var v VM
	return v, dbServer.ObjectLoad(&v, where, args...)
}

func getVMs(serverID int64) []VM {
	v, _ := dbServer.ObjectListQuery(VM{}, "where sid=?", serverID)
	return v.([]VM)
}

func vmsByQuery(where string, args ...interface{}) []VM {
	r, _ := dbServer.ObjectListQuery(VM{}, where, args...)
	return r.([]VM)
}

func getRack(where string, args ...interface{}) (Rack, error) {
	var r Rack
	return r, dbServer.ObjectLoad(&r, where, args...)
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
	return dbServer.ObjectDelete(v)
}

func (s Server) Delete() error {
	for _, vm := range getVMs(s.ID) {
		if err := vm.Delete(); err != nil {
			return err
		}
	}
	return dbServer.ObjectDelete(s)
}

func (s Server) Update() error {
	return dbServer.ObjectUpdate(s)
}

func (s Server) VMs() []VM {
	v, _ := dbServer.ObjectListQuery(VM{}, "where sid=?", s.ID)
	return v.([]VM)
}

func (s Server) DID() int64 {
	q := "where id=?"
	if r, err := getRack(q, s.RID); err == nil {
		return r.DID
	}
	return 0
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
	return dbServer.ObjectInsert(s)
}

func (s Router) Delete() error {
	return dbServer.ObjectDelete(s)
}

func (s Router) Update() error {
	return dbServer.ObjectUpdate(s)
}

func (s VM) Insert() (int64, error) {
	return dbServer.ObjectInsert(s)
}

func (s VM) Update() error {
	return dbServer.ObjectUpdate(s)
}

func getUser(where string, args ...interface{}) (User, error) {
	u := User{}
	err := dbServer.ObjectLoad(&u, where, args...)
	return u, err
}

func userUpdate(user User) error {
	return dbServer.ObjectUpdate(user)
}

func userAdd(user User) (int64, error) {
	if len(user.Password) == 0 {
		user.Password = "n/a" // TODO: no longer using stored password , need to update schema
	}
	return dbServer.ObjectInsert(user)
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
	list, err := dbServer.Rows(query)
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

	data, err := dbServer.ObjectListQuery(RackNet{}, "where rid=? order by min_ip", rid)
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
