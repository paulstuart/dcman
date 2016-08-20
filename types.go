package main

import (
	"database/sql/driver"
	"fmt"
	//"log"
	"net"
	//	"regexp"
	//"sort"
	//"strconv"
	//	"strings"
	"time"
	//	"github.com/paulstuart/dbutil"
)

/*
var (
	noNumbers = regexp.MustCompile("[^0-9]*")
	noRange   = regexp.MustCompile("-.*")
	//ipAddr    = regexp.MustCompile("[0-9+]\\.[0-9+]\\.[0-9+]\\.[0-9+]")
)
*/

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
	USR    int64  `sql:"usr" key:"true" table:"users"`
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
	USR      int64  `sql:"usr" key:"true" table:"users"`
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
	WWW      *string   `sql:"www"`
	Phone    *string   `sql:"phone"`
	Address  *string   `sql:"address"`
	City     *string   `sql:"city"`
	State    *string   `sql:"state"`
	Country  *string   `sql:"country"`
	Postal   *string   `sql:"postal"`
	Note     *string   `sql:"note"`
	USR      int64     `sql:"usr"  audit:"user"`
	Modified time.Time `sql:"ts" audit:"time"`
}

type IPType struct {
	IPT   int64   `sql:"ipt" key:"true" table:"ip_types"`
	Name  *string `sql:"name"`
	Multi bool    `sql:"multi"`
}

type RMA struct {
	RMD       int64     `sql:"rmd" key:"true" table:"rmas"`
	STI       *int64    `sql:"sti"`
	DID       *int64    `sql:"did"`
	VID       *int64    `sql:"vid"`
	OldPID    *int64    `sql:"old_pid"`
	NewPID    *int64    `sql:"new_pid"`
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
	USR       int64     `sql:"usr"`
}

type RMAView struct {
	RMD         int64     `sql:"rmd" key:"true" table:"rmas_view"`
	STI         *int64    `sql:"sti"`
	DID         *int64    `sql:"did"`
	VID         *int64    `sql:"vid"`
	OldPID      *int64    `sql:"old_pid"`
	NewPID      *int64    `sql:"new_pid"`
	Site        *string   `sql:"site"`
	Hostname    *string   `sql:"hostname"`
	DeviceSN    *string   `sql:"device_sn"`
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
	USR         int64     `sql:"usr"`
}

/*
type Carrier struct {
	CarrierID int64     `sql:"cr_id" key:"true" table:"carriers"`
	Name      string    `sql:"name"`
	URL       string    `sql:"tracking_url"`
	USR       int64    `sql:"usr"`
	Modified  time.Time `sql:"ts"`
}

type Return struct {
	ReturnID  int64     `sql:"return_id" key:"true" table:"rma_returns"`
	RMD       int64     `sql:"rmd"`
	CarrierID int64     `sql:"cr_id"`
	Tracking  string    `sql:"tracking_no"`
	USR       int64    `sql:"usr"`
	Sent      time.Time `sql:"date_sent"`
}

type Sent struct {
	ReturnID int64 `sql:"return_id" table:"rma_sent"`
	PID      int64 `sql:"pid"`
}

type Received struct {
	RMD int64     `sql:"rmd" table:"rma_received"`
	PID int64     `sql:"pid"`
	UID *int64    `sql:"usr"`
	TS  time.Time `sql:"date_received"`
}
*/

type Manufacturer struct {
	MID      int64     `sql:"mid" key:"true" table:"mfgrs"`
	Name     string    `sql:"name"`
	Note     *string   `sql:"note"`
	AKA      *string   `sql:"aka"`
	URL      *string   `sql:"url"`
	USR      int64     `sql:"usr"  audit:"user"`
	Modified time.Time `sql:"ts" audit:"time"`
}

type PartType struct {
	PTI      int64     `sql:"pti" key:"true" table:"part_types"`
	Name     string    `sql:"name"`
	USR      int64     `sql:"usr"  audit:"user"`
	Modified time.Time `sql:"ts" audit:"time"`
}

type SKU struct {
	KID         int64     `sql:"kid" key:"true" table:"skus"`
	MID         *int64    `sql:"mid"`
	PTI         *int64    `sql:"pti"`
	PartNumber  *string   `sql:"part_no"`
	Description *string   `sql:"description"`
	SKU         *string   `sql:"sku"`
	USR         int64     `sql:"usr"  audit:"user"`
	Modified    time.Time `sql:"ts" audit:"time"`
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
	Unused   bool      `sql:"unused"`
	Bad      bool      `sql:"bad"`
	Cents    *int      `sql:"cents"` // in cents to avoid floating point issues
	USR      int64     `sql:"usr"  audit:"user"`
	Modified time.Time `sql:"ts" audit:"time"`
}

type PartView struct {
	PID         int64    `sql:"pid" key:"true" table:"parts_view"`
	KID         *int64   `sql:"kid"` // vendor sku id
	VID         *int64   `sql:"vid"` // vendor id
	DID         *int64   `sql:"did"` // device id
	STI         *int64   `sql:"sti"` // site id
	RMD         *int64   `sql:"rmd"` // rma id
	Site        *string  `sql:"site"`
	Hostname    *string  `sql:"hostname"`
	Location    *string  `sql:"location"`
	Serial      *string  `sql:"serial_no"`
	AssetTag    *string  `sql:"asset_tag"`
	PartType    *string  `sql:"part_type"`
	PartNumber  *string  `sql:"part_no"`
	SKU         *string  `sql:"sku"`
	Description *string  `sql:"description"`
	Mfgr        *string  `sql:"mfgr"`
	Vendor      *string  `sql:"vendor"`
	Cents       *int     `sql:"cents"` // in cents
	Price       *float32 `sql:"price"`
	Unused      bool     `sql:"unused"`
	Bad         bool     `sql:"bad"`
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
	Name     *string   `sql:"name"`
	Address  *string   `sql:"address"`
	City     *string   `sql:"city"`
	State    *string   `sql:"state"`
	Postal   *string   `sql:"postal"`
	Country  *string   `sql:"country"`
	Phone    *string   `sql:"phone"`
	Web      *string   `sql:"web"`
	USR      int64     `sql:"usr"  audit:"user"`
	Modified time.Time `sql:"ts" audit:"time"`
}

type Tag struct {
	TID  int64  `sql:"tid" key:"true" table:"tags"`
	Name string `sql:"tag"`
}

type Rack struct {
	RID      int64      `sql:"rid" key:"true" table:"racks"`
	STI      int64      `sql:"sti"`
	RUs      int        `sql:"rackunits"`
	Label    int        `sql:"rack"`
	VendorID *string    `sql:"vendor_id"`
	XPos     *string    `sql:"x_pos"`
	YPos     *string    `sql:"y_pos"`
	Note     *string    `sql:"note"`
	USR      int64      `sql:"usr"`
	TS       *time.Time `sql:"ts" update:"false"`
}

type RackView struct {
	RID      int64      `sql:"rid" key:"true" table:"racks_view"`
	STI      int64      `sql:"sti"`
	RUs      int        `sql:"rackunits"`
	Label    int        `sql:"rack"`
	Site     *string    `sql:"site"`
	VendorID *string    `sql:"vendor_id"`
	XPos     *string    `sql:"x_pos"`
	YPos     *string    `sql:"y_pos"`
	Note     *string    `sql:"note"`
	USR      int64      `sql:"usr"`
	TS       *time.Time `sql:"ts" update:"false"`
}

type VM struct {
	VMI      int64     `sql:"vmi" key:"true" table:"vms"`
	DID      int64     `sql:"did"`
	Hostname *string   `sql:"hostname"`
	Profile  *string   `sql:"profile"`
	Note     *string   `sql:"note"`
	USR      int64     `sql:"usr"`
	Modified time.Time `sql:"ts"`
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
	USR      int64     `sql:"usr"`
	Modified time.Time `sql:"ts"`
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
	ID       int64   `sql:"id" key:"true" table:"pdus"`
	RID      *int64  `sql:"rid"`
	Hostname *string `sql:"hostname"`
	IP       *string `sql:"ip_address"`
	Netmask  *string `sql:"netmask"`
	Gateway  *string `sql:"gateway"`
	DNS      *string `sql:"dns"`
	AssetTag *string `sql:"asset_tag"`
}

/*
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
*/

type Inventory struct {
	STI         int64    `sql:"sti" key:"true" table:"inventory" json:",omitempty"`
	KID         *int64   `sql:"kid"			json:",omitempty"`
	PTI         *int64   `sql:"pti"			json:",omitempty"`
	Qty         *int64   `sql:"qty"			json:",omitempty"`
	Site        *string  `sql:"site"		json:",omitempty"`
	Mfgr        *string  `sql:"mfgr"		json:",omitempty"`
	PartNumber  *string  `sql:"part_no"		json:",omitempty"`
	PartType    *string  `sql:"part_type"	json:",omitempty"`
	Description *string  `sql:"description"	json:",omitempty"`
	Cents       *int     `sql:"cents"       json:",omitempty"`
	Price       *float32 `sql:"price"       json:",omitempty"`
}

type Contract struct {
	CID    int64   `sql:"cid" key:"true" table:"contracts"`
	VID    *int64  `sql:"vid"`
	Policy *string `sql:"policy"`
	Phone  *string `sql:"phone"`
}

type DeviceType struct {
	DTI  int64  `sql:"dti" key:"true" table:"device_types"`
	Name string `sql:"name"`
}

type Device struct {
	DID      int64     `sql:"did" key:"true" table:"devices"`
	RID      *int64    `sql:"rid"` // Rack ID
	KID      *int64    `sql:"kid"` // SKU ID
	DTI      *int64    `sql:"dti"` // Device type ID
	TID      *int64    `sql:"tid"` // Tag ID
	RU       int       `sql:"ru"`
	Height   int       `sql:"height"`
	Hostname *string   `sql:"hostname"`
	Alias    *string   `sql:"alias"`
	Profile  *string   `sql:"profile"`
	SerialNo *string   `sql:"sn"`
	AssetTag *string   `sql:"asset_tag"`
	Assigned *string   `sql:"assigned"`
	Note     *string   `sql:"note"`
	USR      int64     `sql:"usr"`
	Modified time.Time `sql:"ts"`
}

// DeviceView is a more usable view of the Device record
type DeviceView struct {
	DID      int64     `sql:"did" key:"true" table:"devices_view"`
	STI      *int64    `sql:"sti"` // Site ID
	KID      *int64    `sql:"kid"` // SKU ID
	RID      *int64    `sql:"rid"` // Rack ID
	DTI      *int64    `sql:"dti"` // Device type ID
	TID      *int64    `sql:"tid"` // Tag ID
	Rack     int       `sql:"rack"`
	RU       int       `sql:"ru"`
	Height   int       `sql:"height"`
	Hostname *string   `sql:"hostname"`
	Alias    *string   `sql:"alias"`
	Profile  *string   `sql:"profile"`
	SerialNo *string   `sql:"sn"`
	AssetTag *string   `sql:"asset_tag"`
	Assigned *string   `sql:"assigned"`
	Note     *string   `sql:"note"`
	DevType  *string   `sql:"devtype"`
	Tag      *string   `sql:"tag"`
	Site     *string   `sql:"site"`
	USR      int64     `sql:"usr"`
	Modified time.Time `sql:"ts"`
}

// DeviceIPs merges IP info into the DeviceView
type DeviceIPs struct {
	DID      int64     `sql:"did" key:"true" table:"devices_list"`
	STI      *int64    `sql:"sti"` // Site ID
	RID      *int64    `sql:"rid"` // Rack ID
	KID      *int64    `sql:"kid"` // SKU ID
	DTI      *int64    `sql:"dti"` // Device type ID
	TID      *int64    `sql:"tid"` // Tag ID
	Rack     *int      `sql:"rack"`
	RU       *int      `sql:"ru"`
	Height   *int      `sql:"height"`
	Hostname *string   `sql:"hostname"`
	IPs      *string   `sql:"ips"`
	Mgmt     *string   `sql:"mgmt"`
	Alias    *string   `sql:"alias"`
	Profile  *string   `sql:"profile"`
	SerialNo *string   `sql:"sn"`
	AssetTag *string   `sql:"asset_tag"`
	Assigned *string   `sql:"assigned"`
	Tag      *string   `sql:"tag"`
	Note     *string   `sql:"note"`
	DevType  *string   `sql:"devtype"`
	Site     *string   `sql:"site"`
	USR      int64     `sql:"usr"`
	Modified time.Time `sql:"ts"`
}

/*
// DeviceNetwork merges network info with DeviceView
// It will usually result in multiple device "instances" due to
// the one to many relationship of device and network data
// sti|site|rack|did|rid|dti|kid|tid|ru|height|hostname|alias|asset_tag|sn|profile|assigned|note|usr|ts|devtype|tag|ifd|did:1|mgmt|port|mac|cable_tag|switch_port|iid|ipt|ip32|ipv4|iptype
//2|SFO|11|1418|26|1||0|7|2|sfo1cs01|||||||0|2016-08-17 20:16:34|Server||1|1418|1|IPMI|0c:c4:7a:1b:60:d4|||1|1|174339100|10.100.52.28|IPMI

type DeviceNetwork struct {
	IID      int64     `sql:"iid" key:"true" table:"devices_network"`
	STI      *int64    `sql:"sti"` // Site ID
	RID      *int64    `sql:"rid"` // Rack ID
	KID      *int64    `sql:"kid"` // SKU ID
	DTI      *int64    `sql:"dti"` // Device type ID
	TID      *int64    `sql:"tid"` // Tag ID
	IFD      *int64    `sql:"ifd"` // Interface ID
	DID      *int64    `sql:"did"` // Device ID
	Rack     *int      `sql:"rack"`
	RU       *int      `sql:"ru"`
	Height   *int      `sql:"height"`
	Hostname *string   `sql:"hostname"`
	Alias    *string   `sql:"alias"`
	Profile  *string   `sql:"profile"`
	SerialNo *string   `sql:"sn"`
	AssetTag *string   `sql:"asset_tag"`
	Assigned *string   `sql:"assigned"`
	Tag      *string   `sql:"tag"`
	Note     *string   `sql:"note"`
	DevType  *string   `sql:"devtype"`
	Site     *string   `sql:"site"`
	IPv4     *string   `sql:"ipv4"`
	IP32     *int      `sql:"ip32"`
	USR      int64     `sql:"usr"`
	Modified time.Time `sql:"ts"`
}
*/

type DeviceAdjust struct {
	DID    int64 `sql:"did" key:"true" table:"devices_adjust"`
	RID    int64 `sql:"rid"`
	RU     int   `sql:"ru"`
	Height int   `sql:"height"`
}

/*
type IPInfo struct {
	DID        int64   `sql:"did" key:"true" table:"devnet"`
	Mgmt       bool    `sql:"mgmt"`
	Port       int     `sql:"port"`
	Mac        string  `sql:"mac"`
	CableTag   string  `sql:"cable_tag"`
	SwitchPort string  `sql:"switch_port"`
	IPv4       *string `sql:"ipv4"`
}

func (i IPInfo) Interface() string {
	if i.Mgmt {
		return "Mgmt"
	}
	return fmt.Sprintf("eth%d", i.Port)
}
*/

type IFace struct {
	IFD        int64   `sql:"ifd" key:"true" table:"interfaces"`
	DID        int64   `sql:"did"`
	Mgmt       bool    `sql:"mgmt"`
	Port       *string `sql:"port"`
	MAC        *string `sql:"mac"`
	CableTag   *string `sql:"cable_tag"`
	SwitchPort *string `sql:"switch_port"`
}

type IFaceView struct {
	IFD        int64   `sql:"ifd" key:"true" table:"interfaces_view"`
	DID        int64   `sql:"did"`
	IID        *int64  `sql:"iid"`
	IPT        *int64  `sql:"ipt"`
	IP32       *uint32 `sql:"ip32"`
	Mgmt       bool    `sql:"mgmt"`
	Port       *string `sql:"port"`
	IP         *string `sql:"ipv4"`
	IPType     *string `sql:"iptype"`
	MAC        *string `sql:"mac"`
	CableTag   *string `sql:"cable_tag"`
	SwitchPort *string `sql:"switch_port"`
}

type IPAddr struct {
	IID  int64   `sql:"iid" key:"true" table:"ips"`
	IFD  *int64  `sql:"ifd"`
	VMI  *int64  `sql:"vmi"`
	VLI  *int64  `sql:"vli"` // reserved IPs link to their respective vlan
	IPT  *int64  `sql:"ipt"`
	IP32 *uint32 `sql:"ip32"`
	IPv4 *string `sql:"ipv4"`
	Note *string `sql:"note"`
}

type IPsUsed struct {
	ID       int64   `sql:"id" table:"ips_list"`
	STI      *int64  `sql:"sti"`
	RID      *int64  `sql:"rid"`
	IPT      *int64  `sql:"ipt"`
	Site     *string `sql:"site"`
	Rack     *int    `sql:"rack"`
	IP       *string `sql:"ip"`
	Type     *string `sql:"iptype"`
	Host     *string `sql:"host"`
	Hostname *string `sql:"hostname"`
	Note     *string `sql:"note"`
}

type Provider struct {
	PRI     int64   `sql:"pri" key:"true" table:"providers"`
	Name    *string `sql:"name"`
	Contact *string `sql:"provider"`
	Phone   *string `sql:"a_side_xcon"`
	EMail   *string `sql:"a_side_handoff"`
	URL     *string `sql:"z_side_xcon"`
	Note    *string `sql:"note"`
}

type Circuit struct {
	CID          int64   `sql:"cid" key:"true" table:"circuits"`
	STI          *int64  `sql:"site"`
	PRI          *int64  `sql:"pri"`
	CircuitID    *string `sql:"circuit_id"`
	ASideXConn   *string `sql:"a_side_xcon"`
	ASideHandoff *string `sql:"a_side_handoff"`
	ZSideXConn   *string `sql:"z_side_xcon"`
	ZSideHandoff *string `sql:"z_side_handoff"`
	Note         *string `sql:"note"`
}

type CircuitView struct {
	CID          int64   `sql:"cid" key:"true" table:"circuits"`
	STI          *int64  `sql:"site"`
	PRI          *int64  `sql:"pri"`
	Site         *string `sql:"site"`
	Provider     *string `sql:"provider"`
	CircuitID    *string `sql:"circuit_id"`
	ASideXConn   *string `sql:"a_side_xcon"`
	ASideHandoff *string `sql:"a_side_handoff"`
	ZSideXConn   *string `sql:"z_side_xcon"`
	ZSideHandoff *string `sql:"z_side_handoff"`
	Note         *string `sql:"note"`
}

type SubCircuit struct {
	SCI   int64   `sql:"sci" key:"true" table:"sub_circuits"`
	CID   *int64  `sql:"cid"`
	SubID *string `sql:"sub_circuit_id"`
	Note  *string `sql:"note"`
}

// CircuitList is a list of circuits and their sub-circuits
type CircuitList struct {
	CID          int64   `sql:"cid" key:"true" table:"circuits"`
	STI          *int64  `sql:"site"`
	PRI          *int64  `sql:"pri"`
	Site         *string `sql:"site"`
	Provider     *string `sql:"provider"`
	CircuitID    *string `sql:"circuit_id"`
	SubID        *string `sql:"sub_circuit_id"`
	ASideXConn   *string `sql:"a_side_xcon"`
	ASideHandoff *string `sql:"a_side_handoff"`
	ZSideXConn   *string `sql:"z_side_xcon"`
	ZSideHandoff *string `sql:"z_side_handoff"`
	Note         *string `sql:"note"`
	SubNote      *string `sql:"sub_note"`
}

/*
func (ip *IP) FromString(in string) {
	bits := net.ParseIP(in).To4()
	if len(bits) == 4 {
		ip.IP32 = (uint32(bits[0])<<24 + uint32(bits[1])<<16 + uint32(bits[2])<<8 + uint32(bits[3]))
	}
}

func (ip IP) String() string {
	a := ip.IP32 >> 24
	b := (ip.IP32 >> 16) & 255
	c := (ip.IP32 >> 8) & 255
	d := ip.IP32 & 255
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}
*/

var removeWords = []string{
	" ",
	"the",
	"inc.",
	"incorporated",
	"corporation",
	"company",
}

type VProfile struct {
	VPID int64 `sql:"vpid" key:"true" table:"vlan_profiles"`
	Name int   `sql:"name"`
}

type VLAN struct {
	VLI      int64     `sql:"vli" key:"true" table:"vlans"`
	STI      int64     `sql:"sti"`
	Name     int       `sql:"name"`
	Profile  *string   `sql:"profile"`
	Gateway  *string   `sql:"gateway"`
	Route    *string   `sql:"route"`
	Netmask  *string   `sql:"netmask"`
	Note     *string   `sql:"note"`
	Modified time.Time `sql:"ts" audit:"time"`
	USR      int64     `sql:"usr"  audit:"user"`
}

type VLANView struct {
	VLI      int64     `sql:"vli" key:"true" table:"vlans_view"`
	STI      int64     `sql:"sti"`
	Name     int       `sql:"name"`
	Site     *string   `sql:"site"`
	Profile  *string   `sql:"profile"`
	Gateway  *string   `sql:"gateway"`
	Route    *string   `sql:"route"`
	Netmask  *string   `sql:"netmask"`
	Note     *string   `sql:"note"`
	Modified time.Time `sql:"ts" audit:"time"`
	USR      int64     `sql:"usr"  audit:"user"`
}
