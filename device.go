package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

//g#o:generate stringer -type=deviceFamily,portType,ipType

type deviceFamily int

const (
	UnknownDevice deviceFamily = iota
	StandaloneServer
	Enclosure
	Blade
	NetworkSwitch
	NetworkRouter
	TerminalServer
	LoadBalancer
	PowerDistributionUnit
	CableManagement
	Console
)

type Contract struct {
	CID    int64  `sql:"cid" key:"true" table:"contracts"`
	VID    int64  `sql:"vid"`
	Policy string `sql:"policy"`
	Phone  string `sql:"phone"`
}

type Device struct {
	DID int64 `sql:"did" key:"true" table:"devices"`
	VID int64 `sql:"vid"`
	//	CID        int64      `sql:"cid"`
	RID        int64        `sql:"rid"`
	RU         int          `sql:"ru"`
	Height     int          `sql:"height"`
	Type       deviceFamily `sql:"device_type"`
	PrimaryIP  uint32       `sql:"primary_ip"`
	MgmtIP     uint32       `sql:"mgmt_ip"`
	PrimaryMac string       `sql:"primary_mac"`
	MgmtMac    string       `sql:"mgmt_mac"`
	Hostname   string       `sql:"hostname"`
	Model      string       `sql:"model"`
	AssetTag   string       `sql:"asset_tag"`
	SerialNo   string       `sql:"sn"`
	Note       string       `sql:"note"`
	// audit info
	Modified time.Time `sql:"modified"`
	UID      int       `sql:"uid"`
}

const q = `select d.*, 
    ipmi.ip_int as ipmi_ip, ipmi.mac as ipmi_mac
    eth0.ip_int as eth0ip, eth0.mac as eth0_mac
   from devices d,
   left join on ips ipmi where ipmi.did = d.did and ipmi.ip_type=?,
   left join on ips eth0 where eth0.did = d.did and eth0.ip_type=?,
`

type portType int

const (
	UnknownPort portType = iota
	ipmi
	eth0
	eth1
	eth2
	eth3
)

type Port struct {
	PID        int64    `sql:"pid" key:"true" table:"ports"`
	DID        int64    `sql:"did"`
	PortType   portType `sql:"port_type"`
	MAC        string   `sql:"mac"`
	CableTag   string   `sql:"cable_tag"`
	SwitchPort string   `sql:"switch_port"`
	// audit info
	Modified time.Time `sql:"modified"`
	UID      int       `sql:"uid"`
}

type ipType int

const (
	UnknownIP ipType = iota
	IPMI
	Internal
	Public
)

type IP struct {
	IID  int64  `sql:"iid" key:"true" table:"ips"`
	DID  int64  `sql:"did"`
	Type ipType `sql:"ip_type"`
	Int  uint32 `sql:"ip_int"`
	// audit info
	Modified time.Time `sql:"modified"`
	UID      int       `sql:"uid"`
}

func (ip *IP) FromString(in string) {
	bits := net.ParseIP(in).To4()
	if len(bits) == 4 {
		ip.Int = (uint32(bits[0])<<24 + uint32(bits[1])<<16 + uint32(bits[2])<<8 + uint32(bits[3]))
	}
}

func (ip IP) String() string {
	a := ip.Int >> 24
	b := (ip.Int >> 16) & 255
	c := (ip.Int >> 8) & 255
	d := ip.Int & 255
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

var removeWords = []string{
	"the ",
	"inc.",
	"incorporated",
	"corporation",
	"company",
}

func ManufacturerID(name string) int64 {
	aka := strings.ToLower(name)
	for _, word := range removeWords {
		aka = strings.Replace(aka, word, "", -1)
	}
	m := Manufacturer{Name: name, AKA: aka}
	if err := dbObjectLoad(&m, "where aka=?", aka); err != nil {
		if err := dbAdd(&m); err != nil {
			log.Println("mfgr add err:", err)
		}
	}
	return m.MID
}

func skuID(mid, tid int64, pn, d string) int64 {
	log.Println("ADD:", d)
	pl := SKU{MID: mid, TID: tid, PartNumber: pn, Description: d}
	if err := dbObjectLoad(&pl, "where mid=? and part_no=?", mid, pn); err != nil {
		//fmt.Println("plist load err:", err)
		dbAdd(&pl)
	}
	return pl.KID
}

func AddDevicePart(did, sid, tid int64, manufacturer, productName, description, serialNumber, assetTag, location string) (*Part, error) {
	part := Part{
		SID:      sid,
		DID:      did,
		KID:      skuID(ManufacturerID(manufacturer), tid, productName, description),
		Serial:   serialNumber,
		AssetTag: assetTag,
		Location: location,
	}
	if err := dbAdd(&part); err != nil {
		return nil, err
	}
	return &part, nil
}
