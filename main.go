package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/osext"
	pp "github.com/paulstuart/ping"
	"github.com/paulstuart/secrets"
	gcfg "gopkg.in/gcfg.v1"
)

var (
	insecure          bool
	version           = "1.0.1"
	sessionMinutes    = time.Duration(time.Minute * 240)
	masterMode        = true
	hostname, _       = os.Hostname()
	execDir, _        = osext.ExecutableFolder()
	startTime         = time.Now()
	sqlDir            = "sql" // dir containing sql schemas, etc
	sqlInit           = sqlDir + "/init.sql"
	dbName            = execDir + "/data.db"
	dbFile            = "file://" + dbName //+ "?cache=shared&mode=rwc"
	systemLocation, _ = time.LoadLocation("Local")
	pingTimeout       = 10
	pathPrefix        string
	bannerText        string
	cfg               = struct {
		Main    config
		Backups backupConfig
		SAML    samlConfig
	}{}
)

type config struct {
	Name     string `gcfg:"name"`
	Port     int    `gcfg:"port"`
	Prefix   string `gcfg:"prefix"`
	Uploads  string `gcfg:"uploads"`
	Banner   string `gcfg:"banner"`
	Key      string `gcfg:"key"`
	LogDir   string `gcfg:"logdir"`
	ReadOnly bool   `gcfg:"readonly"`
	PXEBoot  bool   `gcfg:"pxeboot"`
	NoKey    bool   `gcfg:"noKey"` // don't require API key for access (for testing only!!)
}

type backupConfig struct {
	Dir  string `gcfg:"dir"`
	Freq int    `gcfg:"freq"`
}

type samlConfig struct {
	URL         string `gcfg:"samlURL"`
	Cookie      string `gcfg:"cookie"`
	Login       string `gcfg:"loginURL"`
	Token       string `gcfg:"xsrfToken"`
	PlaceHolder string `gcfg:"placeholder"`
	OKTACookie  string `gcfg:"OKTACookie"`
	OKTAHash    string `gcfg:"OKTAHash"`
	Disabled    bool   `gcfg:"disabled"`
	Timeout     int    `gcfg:"timeout"`
	FakeName    string `gcfg:"fakename"`
	FakePass    string `gcfg:"fakepass"`
}

const (
	configFile = "config.gcfg"
	logLayout  = "2006-01-02 15:04:05.999"
	dateLayout = "2006-01-02"
	timeLayout = "2006-01-02 15:04:05"
)

func init() {
	flag.BoolVar(&insecure, "insecure", insecure, "ignore authentication")
	flag.StringVar(&execDir, "dir", execDir, "ignore authentication")
	flag.Parse()
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
	authCookie = cfg.SAML.OKTACookie
	bannerText += cfg.Main.Banner

	key := cfg.Main.Key
	if len(key) == 0 {
		key, _ = secrets.KeyGen()
	}
	secrets.SetKey(key)

	if cfg.SAML.Timeout > 0 {
		sessionMinutes = time.Duration(cfg.SAML.Timeout) * time.Minute
	}
}

func ping(ip string, timeout int) bool {
	return pp.Ping(ip, timeout)
}

type pingable struct {
	IP string
	OK bool
}

func bulkPing(timeout int, ips ...string) map[string]bool {
	hits := make(map[string]bool)
	c := make(chan pingable)

	for _, ip := range ips {
		go func(addr string) {
			ok := ping(addr, timeout)
			c <- pingable{addr, ok}
		}(ip)
	}
	for range ips {
		r := <-c
		hits[r.IP] = r.OK
	}
	return hits
}

func myIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !strings.HasPrefix(ipnet.String(), "127.") && strings.Index(ipnet.String(), ":") == -1 {
			return strings.Split(ipnet.String(), "/")[0]
		}
	}
	return ""
}

type found struct {
	ID   int64  `sql:"id"`
	Kind string `sql:"kind"`
	Name string `sql:"name"`
	Note string `sql:"note"`
}

func dbHits(q string, args ...interface{}) []found {
	err, recs := datastore.LoadMany(q, found{}, args...)
	if err != nil {
		log.Println("search error:", err)
		return nil
	}
	return recs.([]found)
}

func searchDB(what string) []found {
	q := "select distinct did as id, 'server' as kind, hostname, note from devices where hostname=? or sn=? or alias=? or asset_tag=? or profile=? or assigned=?"
	hits := dbHits(q, what, what, what, what, what, what)
	if len(hits) > 0 {
		return hits
	}

	q = `select distinct did as id, 'server' as kind, hostname, note from devices where hostname like ? 
			or sn like ? or alias like ? or asset_tag like ? or profile like ? or assigned like ?`
	almost := "%" + what + "%"
	hits = dbHits(q, almost, almost, almost, almost, almost, almost)
	if len(hits) > 0 {
		return hits
	}

	q = "select distinct vmi as id, 'vm' as kind, hostname, note from vms_ips where ipv4=?"
	hits = dbHits(q, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select distinct iid as id, 'ip' as kind, '* reserved *' as hostname, note from ips_reserved where ipv4=?"
	hits = dbHits(q, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select distinct vmi as id, 'vm' as kind, hostname, note from vms where hostname=?"
	hits = dbHits(q, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select distinct did as id, devtype as kind, hostname, note from devices_network where mac=? or ipv4=?"
	hits = dbHits(q, what, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select distinct did as id, 'server' as kind, hostname, note from devices where hostname like ?"
	hits = dbHits(q, "%"+what+"%")
	if len(hits) > 0 {
		return hits
	}

	q = "select distinct id, kind, hostname, note from notes where note MATCH ?"
	hits = dbHits(q, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select distinct id, kind, hostname, note from notes where note MATCH ?"
	hits = dbHits(q, what+"*")
	if len(hits) > 0 {
		return hits
	}

	q = "select distinct id, kind, hostname, note from notes where note MATCH ?"
	hits = dbHits(q, "*"+what)
	if len(hits) > 0 {
		return hits
	}

	// partial MAC addr?
	if strings.Contains(what, ":") {
		q = "select distinct did as id, devtype as kind, hostname, note from devices_network where mac like ?"
		hits = dbHits(q, "%"+what+"%")
		if len(hits) > 0 {
			return hits
		}
	}

	return hits
}

func auditLog(uid int64, ip, action, msg string) {
	//log.Println("IP:", ip)
	dbExec("insert into audit_log (uid,ip,action,msg) values(?,?,?,?)", uid, ip, strings.ToLower(action), msg)
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			log.Println("Got signal:", sig)
			// sig is a ^C, handle it
			if err := dbExec("PRAGMA wal_checkpoint(FULL)"); err != nil {
				log.Println("checkpoint error:", err)
			}
			if err := dbClose(); err != nil {
				log.Println("close error:", err)
			} else {
				log.Println("db closed")
			}
			os.Exit(1)
		}
	}()
}

func checkMAC() {
	m, e := findMAC("10.100.48.25")
	log.Println("MAC:", m, "ERR:", e)
}

func dbtest() {
	//rows, err := dbRows("select 23, atoip('10.100.32.0')")
	rows, err := dbRows("select atoip('10.100.32.0')")
	if err != nil {
		panic(err)
	}
	for _, row := range rows {
		log.Println("ROW:", row)
	}
	return
	rows, err = dbRows("select 23, atoip '10.100.32.0' ")
	if err != nil {
		panic(err)
	}
	for _, row := range rows {
		log.Println("ROW:", row)
	}
}

func main() {
	var err error

	dbPrep()
	if err != nil {
		log.Fatalln(err)
	}
	/*
		dbtest()
		return
	*/

	if cfg.Backups.Freq > 0 {
		log.Println("set up backups")
		go backups(cfg.Backups.Freq, cfg.Backups.Dir)
	}

	webServer(webHandlers)
}
