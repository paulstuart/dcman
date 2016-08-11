package main

import (
	//"encoding/json"
	"fmt"
	//	"log"
	//	"strings"
	"time"
)

//g#o:generate stringer -type=deviceFamily,portType,ipType

type Contract struct {
	CID    int64  `sql:"cid" key:"true" table:"contracts"`
	VID    int64  `sql:"vid"`
	Policy string `sql:"policy"`
	Phone  string `sql:"phone"`
}

type DeviceType struct {
	DTI  int64  `sql:"dti" key:"true" table:"device_types"`
	Name string `sql:"name"`
}

type Device struct {
	DID      int64     `sql:"did" key:"true" table:"devices"`
	RID      int64     `sql:"rid"` // Rack ID
	KID      int64     `sql:"kid"` // SKU ID
	DTI      int64     `sql:"dti"` // Device type ID
	TID      int64     `sql:"tid"` // Tag ID
	RU       int       `sql:"ru"`
	Height   int       `sql:"height"`
	Hostname *string   `sql:"hostname"`
	Alias    *string   `sql:"alias"`
	Profile  *string   `sql:"profile"`
	SerialNo *string   `sql:"sn"`
	AssetTag *string   `sql:"asset_tag"`
	Assigned *string   `sql:"assigned"`
	Note     *string   `sql:"note"`
	UID      int       `sql:"user_id"`
	Modified time.Time `sql:"modified"`
}

// DeviceView is a more usable view of the Device record
type DeviceView struct {
	DID      int64     `sql:"did" key:"true" table:"devices_view"`
	STI      int64     `sql:"sti"` // Site ID
	KID      int64     `sql:"kid"` // SKU ID
	RID      int64     `sql:"rid"` // Rack ID
	DTI      int64     `sql:"dti"` // Device type ID
	TID      int64     `sql:"tid"` // Tag ID
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
	UID      int       `sql:"user_id"`
	Modified time.Time `sql:"modified"`
}

// DeviceIPs merges IP info into the DeviceView
type DeviceIPs struct {
	DID      int64     `sql:"did" key:"true" table:"devices_list"`
	STI      int64     `sql:"sti"` // Site ID
	RID      int64     `sql:"rid"` // Rack ID
	KID      int64     `sql:"kid"` // SKU ID
	DTI      int64     `sql:"dti"` // Device type ID
	TID      int64     `sql:"tid"` // Tag ID
	Rack     int       `sql:"rack"`
	RU       int       `sql:"ru"`
	Height   int       `sql:"height"`
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
	UID      int       `sql:"user_id"`
	Modified time.Time `sql:"modified"`
}

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
	IP32       *int32  `sql:"ip32"`
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
	IP32 *int32  `sql:"ip32"`
	IPv4 *string `sql:"ipv4"`
	Note *string `sql:"note"`
}

//id|rid|ipt|what|site|rack|ip|iptype|hostname|note
type IPsUsed struct {
	ID       int64   `sql:"id" table:"ips_list"`
	STI      int64   `sql:"sti"`
	RID      int64   `sql:"rid"`
	IPT      int64   `sql:"ipt"`
	Site     *string `sql:"site"`
	Rack     *int    `sql:"rack"`
	IP       *string `sql:"ip"`
	Type     *string `sql:"iptype"`
	Host     *string `sql:"host"`
	Hostname *string `sql:"hostname"`
	Note     *string `sql:"note"`
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

/*
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
	// TODO: just query for the damn sku id
	pl := SKU{MID: mid, PTI: tid, PartNumber: pn, Description: d}
	if mid == 0 || len(pn) == 0 {
		if err := dbObjectLoad(&pl, "where description=?", d); err != nil {
			log.Println("ADD SKU MID:", mid, "PN:", pn, "DESC:", d)
			if err := dbAdd(&pl); err != nil {
				log.Println("ADD SKU ERR:", err)
			}
		}
	} else {
		if err := dbObjectLoad(&pl, "where mid=? and part_no=?", mid, pn); err != nil {
			log.Println("ADD SKU MID:", mid, "PN:", pn, "DESC:", d)
			if err := dbAdd(&pl); err != nil {
				log.Println("ADD SKU ERR:", err)
			}
		}
	}
	return pl.KID
}

func AddDevicePart(sti, did, tid int64, manufacturer, productName, description, serialNumber, assetTag, location string) (*Part, error) {
	part := Part{
		DID:      did,
		STI:      sti,
		KID:      skuID(ManufacturerID(manufacturer), tid, productName, description),
		Serial:   &serialNumber,
		AssetTag: &assetTag,
		Location: &location,
	}
	if err := dbAdd(&part); err != nil {
		return nil, err
	}
	return &part, nil
}

func AddPart(sti int64, manufacturer, productName, description, serialNumber, assetTag, location string) (*Part, error) {
	part := Part{
		STI:      sti,
		KID:      skuID(ManufacturerID(manufacturer), 0, productName, description),
		Serial:   &serialNumber,
		AssetTag: &assetTag,
		Location: &location,
	}
	if err := dbAdd(&part); err != nil {
		return nil, err
	}
	return &part, nil
}
*/
