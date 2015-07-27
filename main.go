package main

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"code.google.com/p/gcfg"
	"github.com/kardianos/osext"
	"github.com/paulstuart/secrets"
)

var (
	version           = "1.3.4"
	masterMode        = true
	Hostname, _       = os.Hostname()
	Basedir, _        = os.Getwd() // get abs path now, as we will be changing dirs
	execDir, _        = osext.ExecutableFolder()
	uploadDir         = filepath.Join(execDir, "uploads")
	log_layout        = "2006-01-02 15:04:05.999"
	start_time        = time.Now()
	sqlDir            = "sql" // dir containing sql schemas, etc
	sqlSchema         = sqlDir + "/schema.sql"
	dbFile            = execDir + "/inventory.db"
	dcLookup          = make(map[string]Datacenter)
	dcIDs             = make(map[int64]Datacenter)
	thisDC            Datacenter
	Datacenters       []Datacenter
	systemLocation, _ = time.LoadLocation("Local")
	pathPrefix        string
	bannerText        string
	cfg               = struct {
		Main    MainConfig
		Backups BackupConfig
		SAML    SAMLConfig
		SSH     SSHConfig
	}{}
)

type MainConfig struct {
	Name     string `gcfg:"name"`
	Port     int    `gcfg:"port"`
	Prefix   string `gcfg:"prefix"`
	Uploads  string `gcfg:"uploads"`
	Banner   string `gcfg:"banner"`
	Key      string `gcfg:"key"`
	ReadOnly bool   `gcfg:"readonly"`
	//BackupDir  string `gcfg:"backup_dir"`
	//BackupFreq int    `gcfg:"backup_freq"`
}

type BackupConfig struct {
	Dir  string `gcfg:"dir"`
	Freq int    `gcfg:"freq"`
}

type SSHConfig struct {
	Username string `gcfg:"username"`
	Password string `gcfg:"password"`
	Host     string `gcfg:"host"`
}

type SAMLConfig struct {
	URL         string `gcfg:"samlURL"`
	Cookie      string `gcfg:"cookie"`
	Login       string `gcfg:"loginURL"`
	Token       string `gcfg:"xsrfToken"`
	PlaceHolder string `gcfg:"placeholder"`
	OKTACookie  string `gcfg:"OKTACookie"`
	OKTAHash    string `gcfg:"OKTAHash"`
}

const (
	sessionMinutes = 120
	configFile     = "config.gcfg"
)

func init() {
	f := configFile
	if _, err := os.Stat(configFile); err != nil {
		f = filepath.Join(execDir, configFile)
		if _, err := os.Stat(f); err != nil {
			log.Fatal(err)
		}
	}
	a := assetDir
	if _, err := os.Stat(a); err != nil {
		a = filepath.Join(execDir, assetDir)
		if _, err := os.Stat(a); err != nil {
			log.Fatal(err)
		}
		assetDir = a
		sqlDir = filepath.Join(execDir, sqlDir)
	}
	tdir = filepath.Join(assetDir, "templates")

	data, err := ioutil.ReadFile(f)
	if err != nil {
		log.Fatal(err)
	}
	err = gcfg.ReadStringInto(&cfg, string(data))
	if err != nil {
		log.Fatalf("Failed to parse gcfg data: %s", err)
	}
	if len(cfg.Main.Prefix) > 0 {
		pathPrefix = cfg.Main.Prefix
	}
	if len(cfg.Main.Uploads) > 0 {
		uploadDir = cfg.Main.Uploads
	}
	authCookie = cfg.SAML.OKTACookie
	bannerText += cfg.Main.Banner

	var key string
	if len(cfg.Main.Key) > 0 {
		key = cfg.Main.Key
	} else {
		key, _ = secrets.KeyGen()
	}
	log.Println("KEY:", key)
	secrets.SetKey(key)
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
	dbExec("insert into audit_log (uid,ip,action,msg) values(?,?,?,?)", uid, ip, strings.ToLower(action), msg)
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			log.Println("Got signal:", sig)
			// sig is a ^C, handle it
			err := dbClose()
			if err != nil {
				log.Println("CLOSE ERROR:", err)
			}
			os.Exit(1)
		}
	}()
}

func main() {
	var err error

	dbPrep()
	if err != nil {
		log.Fatalln(err)
	}
	if cfg.Backups.Freq > 0 {
		go Backups(cfg.Backups.Freq, cfg.Backups.Dir)
	}

	getColumns()
	LoadVLANs()

	dc, _ := dbObjectList(Datacenter{})
	Datacenters = dc.([]Datacenter)
	for _, dc := range Datacenters {
		dcLookup[dc.Name] = dc
		dcIDs[dc.ID] = dc
	}
	if vlan, err := ipVLAN(MyIp()); err == nil {
		thisDC = dcIDs[vlan.DID]
	}
	webServer(webHandlers)
}
