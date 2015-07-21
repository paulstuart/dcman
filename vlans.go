package main

import (
	"fmt"
	"net"
	"strconv"
)

var (
	vlans []VLAN // kept in memory as we'll access frequently
)

type VLAN struct {
	ID      int64  `sql:"id" key:"true" table:"vlans"`
	DID     int64  `sql:"did"`
	Name    int    `sql:"name"`
	Gateway string `sql:"gateway"`
	Route   string `sql:"route"`
	Netmask string `sql:"netmask"`
	ipnet   net.IPNet
}

func (v VLAN) String() string {
	return fmt.Sprintf("VLAN: %d\nGateway: %s\nRoute: %s\nNetmask:%s", v.Name, v.Gateway, v.Route, v.Netmask)
}

func (v VLAN) DC() string {
	return dcIDs[v.DID].Name
}

func (v VLAN) Update() error {
	for i := range vlans {
		if vlans[i].ID == v.ID {
			vlans[i] = v
			break
		}
	}
	return dbObjectUpdate(v)
}

func (v VLAN) Delete() error {
	for i := range vlans {
		if vlans[i].ID == v.ID {
			vlans = append(vlans[:i], vlans[i+1:]...)
			break
		}
	}
	return dbObjectDelete(v)
}

func (v VLAN) Insert() (int64, error) {
	vlans = append(vlans, v)
	return dbObjectInsert(v)
}

func LoadVLANs() {
	v, _ := dbObjectList(VLAN{})
	vlans = v.([]VLAN)
	for i := range vlans {
		//	fmt.Println("D", vlans[i].DID, "G", vlans[i].Gateway, "M", vlans[i].Netmask)
		vlans[i].Calc()
	}
}

func (v *VLAN) Calc() {
	mask := net.ParseIP(v.Netmask).To4()
	v.ipnet.IP = net.ParseIP(v.Gateway)
	if len(mask) == 4 {
		v.ipnet.Mask = net.IPv4Mask(mask[0], mask[1], mask[2], mask[3])
	} else {
		fmt.Println("Bad mask", v.Netmask)
	}
}

func dcVLAN(dc, name string) (VLAN, error) {
	d := dcLookup[dc]
	n, _ := strconv.Atoi(name)
	for _, vlan := range vlans {
		if vlan.DID == d.ID && vlan.Name == n {
			return vlan, nil
		}
	}
	return VLAN{}, fmt.Errorf("No vlan found for dc:%s vlan:%s", dc, name)
}

func findVLAN(did int64, addr string) (VLAN, error) {
	ip := net.ParseIP(addr)
	//fmt.Println("FIND VLAN DID", did, "ADDR", addr, "IP", ip)
	for _, vlan := range vlans {
		/*
			if vlan.DID == did {
				fmt.Println("VLAN", vlan.ipnet.String())
			}
		*/
		if vlan.DID == did && vlan.ipnet.Contains(ip) {
			return vlan, nil
		}
	}
	return VLAN{}, fmt.Errorf("No vlan found for ip: %s", addr)
}
