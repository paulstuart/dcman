package main

import (
	"database/sql/driver"
	"fmt"
	"net"
	"time"
)

//go:generate dbgen

type jsonDate time.Time

func (d jsonDate) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	if t.IsZero() {
		return []byte(`""`), nil
	}
	stamp := fmt.Sprintf(`"%s"`, t.Format("2006-01-02"))
	return []byte(stamp), nil
}

func (d *jsonDate) UnmarshalJSON(in []byte) error {
	s := string(in)
	fmt.Printf("\nPARSE THIS: (%d) %s\n\n", len(s), s)
	if len(in) < 3 {
		return nil
	}
	if d == nil {
		d = new(jsonDate)
	}
	const longform = `"2006-01-02T15:04:05.000Z"`
	if len(s) == len(longform) {
		t, err := time.Parse(longform, s)
		*d = jsonDate(t)
		return err
	}
	t, err := time.Parse(`"2006-1-2"`, s)
	if err != nil {
		t, err = time.Parse(`"2006/1/2"`, s)
	}
	if err == nil {
		*d = jsonDate(t)
	}
	return err
}

// Scan implements the Scanner interface.
func (d *jsonDate) Scan(value interface{}) error {
	*d = jsonDate(value.(time.Time))
	return nil
}

// Value implements the driver Valuer interface.
func (d *jsonDate) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return time.Time(*d), nil
}

type summary struct {
	ID      int64   `sql:"sti" key:"true" table:"summary"`
	Site    *string `sql:"site"`
	Servers *string `sql:"servers"`
	VMs     *string `sql:"vms"`
}

type user struct {
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
type fullUser struct {
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

type vendor struct {
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

type ipType struct {
	IPT   int64   `sql:"ipt" key:"true" table:"ip_types"`
	Name  *string `sql:"name"`
	Mgmt  bool    `sql:"mgmt"`
	Multi bool    `sql:"multi"`
}

type rma struct {
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
	Shipped   *jsonDate `sql:"date_shipped"`
	Received  *jsonDate `sql:"date_received"`
	Closed    *jsonDate `sql:"date_closed"`
	Created   *jsonDate `sql:"date_created"`
	USR       int64     `sql:"usr"`
}

type rmaView struct {
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
	Shipped     *jsonDate `sql:"date_shipped"`
	Received    *jsonDate `sql:"date_received"`
	Closed      *jsonDate `sql:"date_closed"`
	Created     *jsonDate `sql:"date_created"`
	USR         int64     `sql:"usr"`
}

type manufacturer struct {
	MID      int64     `sql:"mid" key:"true" table:"mfgrs"`
	Name     string    `sql:"name"`
	Note     *string   `sql:"note"`
	AKA      *string   `sql:"aka"`
	URL      *string   `sql:"url"`
	USR      int64     `sql:"usr"  audit:"user"`
	Modified time.Time `sql:"ts" audit:"time"`
}

type partType struct {
	PTI      int64     `sql:"pti" key:"true" table:"part_types"`
	Name     string    `sql:"name"`
	USR      int64     `sql:"usr"  audit:"user"`
	Modified time.Time `sql:"ts" audit:"time"`
}

type sku struct {
	KID         int64     `sql:"kid" key:"true" table:"skus"`
	MID         *int64    `sql:"mid"`
	PTI         *int64    `sql:"pti"`
	PartNumber  *string   `sql:"part_no"`
	Description *string   `sql:"description"`
	SKU         *string   `sql:"sku"`
	USR         int64     `sql:"usr"  audit:"user"`
	Modified    time.Time `sql:"ts" audit:"time"`
}

type part struct {
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

type partView struct {
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

func (u user) Admin() bool {
	return u.Level > 1
}

func (u user) Editor() bool {
	return u.Level > 0 && !cfg.Main.ReadOnly
}

func (u user) Access() string {
	switch {
	case u.Level == 2:
		return "Admin"
	case u.Level == 1:
		return "Editor"
	default:
		return "User"
	}
}

type site struct {
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

type tag struct {
	TID  int64  `sql:"tid" key:"true" table:"tags"`
	Name string `sql:"tag"`
}

type rack struct {
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

type rackView struct {
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

type vm struct {
	VMI      int64     `sql:"vmi" key:"true" table:"vms"`
	DID      int64     `sql:"did"`
	Hostname *string   `sql:"hostname"`
	Profile  *string   `sql:"profile"`
	Note     *string   `sql:"note"`
	USR      int64     `sql:"usr"`
	Modified time.Time `sql:"ts"`
}

type vmView struct {
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

func getUser(where string, args ...interface{}) (user, error) {
	u := user{}
	err := dbObjectLoad(&u, where, args...)
	return u, err
}

func userUpdate(user user) error {
	return dbObjectUpdate(user)
}

func userAdd(user user) (int64, error) {
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

type inventory struct {
	STI         int64    `sql:"sti" key:"true" table:"inventory" json:",omitempty"`
	KID         *int64   `sql:"kid" json:",omitempty"`
	PTI         *int64   `sql:"pti" json:",omitempty"`
	Qty         *int64   `sql:"qty" json:",omitempty"`
	Site        *string  `sql:"site" json:",omitempty"`
	Mfgr        *string  `sql:"mfgr" json:",omitempty"`
	PartNumber  *string  `sql:"part_no" json:",omitempty"`
	PartType    *string  `sql:"part_type" json:",omitempty"`
	Description *string  `sql:"description" json:",omitempty"`
	Cents       *int     `sql:"cents" json:",omitempty"`
	Price       *float32 `sql:"price" json:",omitempty"`
}

type contract struct {
	CID    int64   `sql:"cid" key:"true" table:"contracts"`
	VID    *int64  `sql:"vid"`
	Policy *string `sql:"policy"`
	Phone  *string `sql:"phone"`
}

type deviceType struct {
	DTI  int64  `sql:"dti" key:"true" table:"device_types"`
	Name string `sql:"name"`
}

type device struct {
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

// deviceView is a more usable view of the Device record
type deviceView struct {
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

// deviceIPs merges IP info into the DeviceView
type deviceIPs struct {
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

// special view with triggers for resizing/moving unit
type deviceAdjust struct {
	DID    int64 `sql:"did" key:"true" table:"devices_adjust"`
	RID    int64 `sql:"rid"`
	RU     int   `sql:"ru"`
	Height int   `sql:"height"`
}

type iface struct {
	IFD        int64   `sql:"ifd" key:"true" table:"interfaces"`
	DID        int64   `sql:"did"`
	Mgmt       bool    `sql:"mgmt"`
	Port       int     `sql:"port"`
	MAC        *string `sql:"mac"`
	CableTag   *string `sql:"cable_tag"`
	SwitchPort *string `sql:"switch_port"`
}

type ifaceView struct {
	IFD        int64   `sql:"ifd" key:"true" table:"interfaces_view"`
	DID        int64   `sql:"did"`
	IID        *int64  `sql:"iid"`
	IPT        *int64  `sql:"ipt"`
	IP32       *uint32 `sql:"ip32"`
	Mgmt       bool    `sql:"mgmt"`
	Port       int     `sql:"port"`
	IP         *string `sql:"ipv4"`
	ipType     *string `sql:"iptype"`
	MAC        *string `sql:"mac"`
	CableTag   *string `sql:"cable_tag"`
	SwitchPort *string `sql:"switch_port"`
}

type ipAddr struct {
	IID  int64   `sql:"iid" key:"true" table:"ips"`
	IFD  *int64  `sql:"ifd"`
	VMI  *int64  `sql:"vmi"`
	VLI  *int64  `sql:"vli"` // reserved IPs link to their respective vlan
	IPT  *int64  `sql:"ipt"`
	IP32 *uint32 `sql:"ip32"`
	IPv4 *string `sql:"ipv4"`
	Note *string `sql:"note"`
}

type ipsUsed struct {
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

type provider struct {
	PRI     int64   `sql:"pri" key:"true" table:"providers"`
	Name    *string `sql:"name"`
	Contact *string `sql:"provider"`
	Phone   *string `sql:"a_side_xcon"`
	EMail   *string `sql:"a_side_handoff"`
	URL     *string `sql:"z_side_xcon"`
	Note    *string `sql:"note"`
}

type circuit struct {
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

type circuitView struct {
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

type subCircuit struct {
	SCI   int64   `sql:"sci" key:"true" table:"sub_circuits"`
	CID   *int64  `sql:"cid"`
	SubID *string `sql:"sub_circuit_id"`
	Note  *string `sql:"note"`
}

// CircuitList is a list of circuits and their sub-circuits
type circuitList struct {
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

type vlanProfile struct {
	VPID int64 `sql:"vpid" key:"true" table:"vlan_profiles"`
	Name int   `sql:"name"`
}

type vlan struct {
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

type vlanView struct {
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

type pxeDevice struct {
	DID      int64   `sql:"did" key:"true" table:"pxedevice"`
	STI      int64   `sql:"sti"`
	RID      int64   `sql:"rid"`
	Site     *string `sql:"site"`
	Rack     int     `sql:"rack"`
	RU       int     `sql:"ru"`
	Hostname *string `sql:"hostname"`
	Profile  *string `sql:"profile"`
	MAC      *string `sql:"mac"`
	IP       *string `sql:"ip"`
	IPMI     *string `sql:"ipmi"`
	Note     *string `sql:"note"`
}
