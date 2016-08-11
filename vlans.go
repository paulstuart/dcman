package main

import (
	//"fmt"
	//	"net"
	"time"
)

type VProfile struct {
	VPID int64 `sql:"vpid" key:"true" table:"vlan_profiles"`
	Name int   `sql:"name"`
}

type VLAN struct {
	VLI      int64     `sql:"vli" key:"true" table:"vlans"`
	STI      int64     `sql:"sti"`
	Name     int       `sql:"name"`
	Profile  string    `sql:"profile"`
	Gateway  string    `sql:"gateway"`
	Route    string    `sql:"route"`
	Netmask  string    `sql:"netmask"`
	Note     *string   `sql:"note"`
	Modified time.Time `sql:"modified" audit:"time"`
	UID      int64     `sql:"user_id"  audit:"user"`
}

type VLANView struct {
	VLI      int64     `sql:"vli" key:"true" table:"vlans_view"`
	STI      int64     `sql:"sti"`
	Name     int       `sql:"name"`
	Site     string    `sql:"site"`
	Profile  string    `sql:"profile"`
	Gateway  string    `sql:"gateway"`
	Route    string    `sql:"route"`
	Netmask  string    `sql:"netmask"`
	Note     *string   `sql:"note"`
	Modified time.Time `sql:"modified" audit:"time"`
	UID      int64     `sql:"user_id"  audit:"user"`
}

/*
func (v *VLAN) Calc() error {
	mask := net.ParseIP(v.Netmask).To4()
	v.ipnet.IP = net.ParseIP(v.Gateway)
	if len(mask) == 4 {
		v.ipnet.Mask = net.IPv4Mask(mask[0], mask[1], mask[2], mask[3])
	} else {
		return fmt.Errorf("Bad mask: %s", v.Netmask)
	}
	return nil
}

func dcVLAN(dc, name string) (VLAN, error) {
	d := dcLookup[dc]
	n, _ := strconv.Atoi(name)
	for _, vlan := range vlans {
		if vlan.DCD == d.DCD && vlan.Name == n {
			return vlan, nil
		}
	}
	return VLAN{}, ErrNoVLAN
}

func ipVLAN(addr string) (VLAN, error) {
	ip := net.ParseIP(addr)
	for _, vlan := range vlans {
		if vlan.ipnet.Contains(ip) {
			return vlan, nil
		}
	}
	return VLAN{}, ErrNoVLAN
}

func findVLAN(did int64, addr string) (VLAN, error) {
	ip := net.ParseIP(addr)
	for _, vlan := range vlans {
		if vlan.DCD == did && vlan.ipnet.Contains(ip) {
			return vlan, nil
		}
	}
	return VLAN{}, ErrNoVLAN
}
*/
