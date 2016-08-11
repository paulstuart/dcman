package main

import (
	"database/sql/driver"
	"fmt"
	//"log"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/paulstuart/dbutil"
)

var (
	noNumbers = regexp.MustCompile("[^0-9]*")
	noRange   = regexp.MustCompile("-.*")
	//ipAddr    = regexp.MustCompile("[0-9+]\\.[0-9+]\\.[0-9+]\\.[0-9+]")
)

//go:generate dbgen

type JSONDate time.Time

func (d JSONDate) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte(`""`), nil
	}
	stamp := fmt.Sprintf(`"%s"`, t.Format("2006-01-02"))
	return []byte(stamp), nil
}

func (d *JSONDate) UnmarshalJSON(in []byte) error {
	s := string(in)
	fmt.Printf("\nPARSE THIS: (%d) %s\n\n", len(s), s)
	if len(in) < 3 {
		return nil
	}
	if d == nil {
		d = new(JSONDate)
	}
	//const xx =       "2016-07-27T18:26:49.037Z"
	const longform = `"2006-01-02T15:04:05.000Z"`
	if len(s) == len(longform) {
		t, err := time.Parse(longform, s)
		*d = JSONDate(t)
		return err
	}
	t, err := time.Parse(`"2006-1-2"`, s)
	if err != nil {
		t, err = time.Parse(`"2006/1/2"`, s)
	}
	if err == nil {
		*d = JSONDate(t)
	}
	return err
}

// Scan implements the Scanner interface.
func (d *JSONDate) Scan(value interface{}) error {
	//*d = value.(JSONDate) //(time.Time)
	*d = JSONDate(value.(time.Time))
	return nil
}

// Value implements the driver Valuer interface.
func (d *JSONDate) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return time.Time(*d), nil
}

type Summary struct {
	ID      int64   `sql:"sti" key:"true" table:"summary"`
	Site    *string `sql:"site"`
	Servers *string `sql:"servers"`
	VMs     *string `sql:"vms"`
}

type User struct {
	ID     int64  `sql:"id" key:"true" table:"users"`
	RealID int64  // when emulating another user, retain real identity
	Login  string `sql:"login"`
	First  string `sql:"firstname"`
	Last   string `sql:"lastname"`
	Email  string `sql:"email"`
	APIKey string `sql:"apikey"`
	Level  int    `sql:"admin"`
}

// FullUser has *all* user fields exposed
type FullUser struct {
	ID       int64  `sql:"id" key:"true" table:"users"`
	RealID   int64  // when emulating another user, retain real identity
	Login    string `sql:"login"`
	First    string `sql:"firstname"`
	Last     string `sql:"lastname"`
	Email    string `sql:"email"`
	APIKey   string `sql:"apikey"`
	Password string `sql:"pw_hash"`
	Salt     string `sql:"pw_salt"`
	Level    int    `sql:"admin"`
}

type Vendor struct {
	VID      int64     `sql:"vid" key:"true" table:"vendors"`
	Name     string    `sql:"name"`
	WWW      string    `sql:"www"`
	Phone    string    `sql:"phone"`
	Address  string    `sql:"address"`
	City     string    `sql:"city"`
	State    string    `sql:"state"`
	Country  string    `sql:"country"`
	Postal   string    `sql:"postal"`
	Note     string    `sql:"note"`
	UID      int64     `sql:"user_id"  audit:"user"`
	Modified time.Time `sql:"modified" audit:"time"`
}

type IPType struct {
	IPT  int64   `sql:"ipt" key:"true" table:"ip_types"`
	Name *string `sql:"name"`
}

type RMA struct {
	RMAID     int64     `sql:"rma_id" key:"true" table:"rmas"`
	STI       int64     `sql:"sti"`
	DID       int64     `sql:"did"`
	VID       int64     `sql:"vid"`
	OldPID    int64     `sql:"old_pid"`
	NewPID    int64     `sql:"new_pid"`
	VendorRMA *string   `sql:"vendor_rma"`
	Jira      *string   `sql:"jira"`
	ShipTrack *string   `sql:"ship_tracking"`
	RecvTrack *string   `sql:"recv_tracking"`
	DCTicket  *string   `sql:"dc_ticket"`
	Receiving *string   `sql:"dc_receiving"`
	Note      *string   `sql:"note"`
	Shipped   *JSONDate `sql:"date_shipped"`
	Received  *JSONDate `sql:"date_received"`
	Closed    *JSONDate `sql:"date_closed"`
	Created   *JSONDate `sql:"date_created"`
	UID       int64     `sql:"user_id"`
}

type RMAView struct {
	RMAID       int64     `sql:"rma_id" key:"true" table:"rmaview"`
	STI         int64     `sql:"sti"`
	DID         int64     `sql:"did"`
	VID         int64     `sql:"vid"`
	OldPID      int64     `sql:"old_pid"`
	NewPID      int64     `sql:"new_pid"`
	Site        *string   `sql:"site"`
	Hostname    *string   `sql:"hostname"`
	ServerSN    *string   `sql:"server_sn"`
	Description *string   `sql:"description"`
	PartSN      *string   `sql:"part_sn"`
	PartNumber  *string   `sql:"part_no"`
	VendorRMA   *string   `sql:"vendor_rma"`
	Jira        *string   `sql:"jira"`
	ShipTrack   *string   `sql:"ship_tracking"`
	RecvTrack   *string   `sql:"recv_tracking"`
	DCTicket    *string   `sql:"dc_ticket"`
	Receiving   *string   `sql:"dc_receiving"`
	Note        *string   `sql:"note"`
	Shipped     *JSONDate `sql:"date_shipped"`
	Received    *JSONDate `sql:"date_received"`
	Closed      *JSONDate `sql:"date_closed"`
	Created     *JSONDate `sql:"date_created"`
	UID         int64     `sql:"user_id"`
}

type Carrier struct {
	CarrierID int64     `sql:"cr_id" key:"true" table:"carriers"`
	Name      string    `sql:"name"`
	URL       string    `sql:"tracking_url"`
	UID       int64     `sql:"user_id"`
	Modified  time.Time `sql:"modified"`
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

type Received struct {
	RMAID int64     `sql:"rma_id" table:"rma_received"`
	PID   int64     `sql:"pid"`
	UID   int64     `sql:"user_id"`
	TS    time.Time `sql:"date_received"`
}

type Manufacturer struct {
	MID      int64     `sql:"mid" key:"true" table:"mfgr"`
	Name     string    `sql:"name"`
	AKA      string    `sql:"aka"`
	URL      string    `sql:"url"`
	UID      int64     `sql:"user_id"  audit:"user"`
	Modified time.Time `sql:"modified" audit:"time"`
}

type PartType struct {
	PTI      int64     `sql:"pti" key:"true" table:"part_types"`
	Name     string    `sql:"name"`
	UID      int64     `sql:"user_id"  audit:"user"`
	Modified time.Time `sql:"modified" audit:"time"`
}

type SKU struct {
	KID         int64     `sql:"kid" key:"true" table:"skus"`
	MID         int64     `sql:"mid"`
	PTI         int64     `sql:"pti"`
	PartNumber  string    `sql:"part_no"`
	Description string    `sql:"description"`
	UID         int64     `sql:"user_id"  audit:"user"`
	Modified    time.Time `sql:"modified" audit:"time"`
}

type Part struct {
	PID      int64     `sql:"pid" key:"true" table:"parts"`
	KID      *int64    `sql:"kid"` // vendor sku id
	VID      *int64    `sql:"vid"` // vendor id
	DID      *int64    `sql:"did"` // server id
	STI      *int64    `sql:"sti"` // site id
	Location *string   `sql:"location"`
	Serial   *string   `sql:"serial_no"`
	AssetTag *string   `sql:"asset_tag"`
	Unused   bool      `sql:"unused"` // BOOL
	Bad      bool      `sql:"bad"`    // BOOL
	UID      int64     `sql:"user_id"  audit:"user"`
	Modified time.Time `sql:"modified" audit:"time"`
}

type PartView struct {
	PID         int64   `sql:"pid" key:"true" table:"parts_view"`
	KID         *int64  `sql:"kid"`    // vendor sku id
	VID         *int64  `sql:"vid"`    // vendor id
	DID         *int64  `sql:"did"`    // device id
	STI         *int64  `sql:"sti"`    // site id
	RMAID       *int64  `sql:"rma_id"` // rma id
	Site        *string `sql:"site"`
	Hostname    *string `sql:"hostname"`
	Location    *string `sql:"location"`
	Serial      *string `sql:"serial_no"`
	AssetTag    *string `sql:"asset_tag"`
	PartType    *string `sql:"part_type"`
	PartNumber  *string `sql:"part_no"`
	Description *string `sql:"description"`
	Mfgr        *string `sql:"mfgr"`
	Unused      bool    `sql:"unused"` // BOOL
	Bad         bool    `sql:"bad"`    // BOOL
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

type Site struct {
	STI      int64     `sql:"sti" key:"true" table:"sites"`
	Name     string    `sql:"name"`
	Address  string    `sql:"address"`
	City     string    `sql:"city"`
	State    string    `sql:"state"`
	Phone    string    `sql:"phone"`
	Web      string    `sql:"web"`
	DCMan    string    `sql:"dcman"`
	UID      int64     `sql:"user_id"  audit:"user"`
	Modified time.Time `sql:"modified" audit:"time"`
}

type Tag struct {
	TID  int64  `sql:"tid" key:"true" table:"tags"`
	Name string `sql:"tag"`
}

type Rack struct {
	RID      int64        `sql:"rid" key:"true" table:"racks"`
	STI      int64        `sql:"sti"`
	RUs      int          `sql:"rackunits"`
	Label    int          `sql:"rack"`
	VendorID string       `sql:"vendor_id"`
	XPos     string       `sql:"x_pos"`
	YPos     string       `sql:"y_pos"`
	UID      int          `sql:"uid"`
	TS       *time.Time   `sql:"ts" update:"false"`
	Table    dbutil.Table `json:"-"`
}

type RackView struct {
	RID      int64        `sql:"rid" key:"true" table:"racks_view"`
	STI      int64        `sql:"sti"`
	RUs      int          `sql:"rackunits"`
	Label    int          `sql:"rack"`
	Site     string       `sql:"site"`
	VendorID string       `sql:"vendor_id"`
	XPos     string       `sql:"x_pos"`
	YPos     string       `sql:"y_pos"`
	UID      int          `sql:"uid"`
	TS       *time.Time   `sql:"ts" update:"false"`
	Table    dbutil.Table `json:"-"`
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
	rn, _ := dbObjectListQuery(RackNet{}, "where rid=? order by vid", r.RID)
	return rn.([]RackNet)
}

type RackNets []RackNet

func (rn RackNets) Len() int           { return len(rn) }
func (rn RackNets) Swap(i, j int)      { rn[i], rn[j] = rn[j], rn[i] }
func (rn RackNets) Less(i, j int) bool { return rn[i].MinIP < rn[j].MinIP }

type VM struct {
	VMI      int64     `sql:"vmi" key:"true" table:"vms"`
	DID      int64     `sql:"did"`
	Hostname *string   `sql:"hostname"`
	Profile  *string   `sql:"profile"`
	Note     *string   `sql:"note"`
	Modified time.Time `sql:"modified"`
	UID      int64     `sql:"user_id"`
}

type VMView struct {
	VMI      int64     `sql:"vmi" key:"true" table:"vms_view"`
	DID      int64     `sql:"did"`
	RID      *int64    `sql:"rid"`
	STI      *int64    `sql:"sti"`
	Rack     *int      `sql:"rack"`
	Site     *string   `sql:"site"`
	Server   *string   `sql:"server"`
	Hostname *string   `sql:"hostname"`
	Profile  *string   `sql:"profile"`
	Note     *string   `sql:"note"`
	Modified time.Time `sql:"modified"`
	UID      int64     `sql:"user_id"`
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

/*
func PartsAdd(dcd int64, columns, words []string) error {
	var item, mfgr, asset, partNo, sn string
	qty := 1
	log.Println("PARTS COLS:", columns)
	var err error
	for i, col := range columns {
		word := strings.TrimSpace(words[i])
		//	log.Println("COL:", col, "WORD:", word)
		switch {
		case col == "qty":
			qty, err = strconv.Atoi(word)
			if err != nil {
				return err
			}
		case col == "asset_tag":
			asset = word
		case col == "item":
			item = word
		case col == "mfgr":
			mfgr = word
		case col == "part_no":
			partNo = word
		case col == "sn":
			sn = word
		default:
			return fmt.Errorf("unknown column: " + col)
		}
	}
	log.Println("ADDING MFGR:", mfgr, "PN:", partNo, "DESC:", item)
	for i := 0; i < qty; i++ {
		if _, err = AddPart(dcd, mfgr, partNo, item, sn, asset, ""); err != nil {
			return err
		}
	}
	return err
}
*/

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
	fmt.Println("NEXTIPS RID:", rid)
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

type Audit struct {
	Hostname string `sql:"hostname" table:"auditing"`
	IP       string `sql:"remote_addr"`
	FQDN     string `sql:"fqdn"`
	IPs      string `sql:"ips"`
	Eth0     string `sql:"eth0"`
	Eth1     string `sql:"eth1"`
	SN       string `sql:"sn"`
	Asset    string `sql:"asset"`
	IpmiIP   string `sql:"ipmi_ip"`
	IpmiMac  string `sql:"ipmi_mac"`
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

/*
func typeID(name string) int64 {
	pt := PartType{}
	if err := dbObjectLoad(&pt, "where name=?", name); err != nil {
		log.Println("type id lookup failed:", err)
		return 0
	}
	return pt.TID
}
*/

type DiskInfo struct {
	Hostname, IP string
	Disks        []DiskData
}

type DiskData struct {
	Size, Location, Manufacturer, PartNumber, SerialNumber string
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

type Inventory struct {
	STI         *int64  `sql:"sti" table:"inventory" json:",omitempty"`
	KID         *int64  `sql:"kid"			json:",omitempty"`
	PTI         *int64  `sql:"pti"			json:",omitempty"`
	Qty         *int64  `sql:"qty"			json:",omitempty"`
	Site        *string `sql:"site"			json:",omitempty"`
	Mfgr        *string `sql:"mfgr"		    json:",omitempty"`
	PartNumber  *string `sql:"part_no"		json:",omitempty"`
	PartType    *string `sql:"part_type"		json:",omitempty"`
	Description *string `sql:"description"	json:",omitempty"`
}
