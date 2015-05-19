package main

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"code.google.com/p/gcfg"
	dbu "github.com/paulstuart/dbutil"
)

var (
	version           = "1.2.0"
	Hostname, _       = os.Hostname()
	Basedir, _        = os.Getwd() // get abs path now, as we will be changing dirs
	log_layout        = "2006-01-02 15:04:05.999"
	start_time        = time.Now()
	http_port         = 8080
	assets_dir        = "assets"
	sqlDir            = "sql" // dir containing sql schemas, etc
	sqlSchema         = sqlDir + "/schema.sql"
	dbFile            = Basedir + "/inventory.db"
	dcLookup          = make(map[string]Datacenter)
	dcIDs             = make(map[int64]Datacenter)
	Datacenters       []Datacenter
	systemLocation, _ = time.LoadLocation("Local")
	dbServer          dbu.DBU
	cfg               = struct {
		Main MainConfig
		SAML SAMLConfig
	}{}
)

type MainConfig struct {
	Name string `gcfg:"name"`
}

type SAMLConfig struct {
	URL         string `gcfg:"samlURL"`
	Cookie      string `gcfg:"cookie"`
	Login       string `gcfg:"loginURL"`
	Token       string `gcfg:"xsrfToken"`
	PlaceHolder string `gcfg:"placeholder"`
}

const (
	pathPrefix     = "/dcman"
	sessionMinutes = 120
	secretKey      = "Team players only! But that's ok, we can all work together"
	configFile     = "config.gcfg"
)

func init() {
	if _, err := os.Stat(configFile); err != nil {
		log.Fatal(err)
	} else {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}
		err = gcfg.ReadStringInto(&cfg, string(data))
		if err != nil {
			log.Fatalf("Failed to parse gcfg data: %s", err)
		}
	}
}

func MyIp() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !strings.HasPrefix(ipnet.String(), "127.") && strings.Index(ipnet.String(), ":") == -1 {
			return strings.Split(ipnet.String(), "/")[0]
		}
	}
	return ""
}

func auditLog(uid int64, ip, action, msg string) {
	dbServer.Exec("insert into audit_log (uid,ip,action,msg) values(?,?,?,?)", uid, ip, strings.ToLower(action), msg)
}

// load schema if this is a new instance
func dbPrep() {
	var fresh bool
	var err error
	if _, err = os.Stat(dbFile); os.IsNotExist(err) {
		fresh = true
	}
	log.Println("FRESH:", fresh, "ERR:", err)
	db, err := dbu.Open(dbFile, true)
	if err != nil {
		panic(err)
	}
	log.Println("FRESHER:", fresh, "ERR:", err)
	if fresh {
		log.Println("READ SCHEMA")
		//db.Debug = true
		err = db.File(sqlSchema)
		log.Println("NEW ERR:", err)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	var err error
	/*
		log.Println("CONFIG:", cfg)
		return
	*/

	dbPrep()
	dbServer, err = dbu.Open(dbFile, false)
	if err != nil {
		log.Fatalln(err)
	}

	LoadVLANs()
	/*
		v, err := findVLAN(2, "10.100.61.101")
		if err != nil {
			fmt.Println("NO VLAN", err)
		}
		fmt.Println("VLAN", v)
		return
	*/
	dc, _ := dbServer.ObjectList(Datacenter{})
	Datacenters = dc.([]Datacenter)
	for _, dc := range Datacenters {
		dcLookup[dc.Name] = dc
		dcIDs[dc.ID] = dc
	}
	webServer(webHandlers)
}
