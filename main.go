package main

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	//"sync"
	"time"

	"code.google.com/p/gcfg"
	"github.com/kardianos/osext"
	"github.com/paulstuart/secrets"
)

var (
	version           = "0.1.8"
	sessionMinutes    = time.Duration(time.Minute * 120)
	masterMode        = true
	Hostname, _       = os.Hostname()
	Basedir, _        = os.Getwd() // get abs path now, as we will be changing dirs
	execDir, _        = osext.ExecutableFolder()
	uploadDir         = filepath.Join(execDir, "uploads")
	startTime         = time.Now()
	sqlDir            = "sql" // dir containing sql schemas, etc
	sqlSchema         = sqlDir + "/schema.sql"
	dbFile            = execDir + "/data.db"
	documentDir       = execDir + "/documents"
	systemLocation, _ = time.LoadLocation("Local")
	pathPrefix        string
	bannerText        string
	cfg               = struct {
		Main    Config
		Backups BackupConfig
		Jira    JiraConfig
		SAML    SAMLConfig
	}{}
)

type Config struct {
	Name     string `gcfg:"name"`
	Port     int    `gcfg:"port"`
	Prefix   string `gcfg:"prefix"`
	Uploads  string `gcfg:"uploads"`
	Banner   string `gcfg:"banner"`
	Key      string `gcfg:"key"`
	LogDir   string `gcfg:"logdir"`
	ReadOnly bool   `gcfg:"readonly"`
	PXEBoot  bool   `gcfg:"pxeboot"`
}

type BackupConfig struct {
	Dir  string `gcfg:"dir"`
	Freq int    `gcfg:"freq"`
}

type SAMLConfig struct {
	URL         string `gcfg:"samlURL"`
	Cookie      string `gcfg:"cookie"`
	Login       string `gcfg:"loginURL"`
	Token       string `gcfg:"xsrfToken"`
	PlaceHolder string `gcfg:"placeholder"`
	OKTACookie  string `gcfg:"OKTACookie"`
	OKTAHash    string `gcfg:"OKTAHash"`
	Disabled    bool   `gcfg:"disabled"`
	Timeout     int    `gcfg:"timeout"`
}

type JiraConfig struct {
	Username string `gcfg:"username"`
	Password string `gcfg:"password"`
	URL      string `gcfg:"url"`
}

const (
	configFile = "config.gcfg"
	logLayout  = "2006-01-02 15:04:05.999"
	dateLayout = "2006-01-02"
	timeLayout = "2006-01-02 15:04:05"
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

	key := cfg.Main.Key
	if len(key) == 0 {
		key, _ = secrets.KeyGen()
	}
	secrets.SetKey(key)

	if err := os.MkdirAll(documentDir, 0755); err != nil {
		log.Panic(err)
	}

	if cfg.SAML.Timeout > 0 {
		sessionMinutes = time.Duration(cfg.SAML.Timeout) * time.Minute
	}
}

func MyIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !strings.HasPrefix(ipnet.String(), "127.") && strings.Index(ipnet.String(), ":") == -1 {
			return strings.Split(ipnet.String(), "/")[0]
		}
	}
	return ""
}

type Hit struct {
	ID   int64  `sql:"id"`
	Kind string `sql:"kind"`
	Name string `sql:"name"`
}

/*
func doSearch(c chan Hit, q string, args ...string) {
	err, found := datastore.LoadMany(q, Hit{}, args...)
	if err != nil {
		log.Printf("search error:", err)
		return nil
	}
	for _, hit := range found.([]Hit) {
		c <- hit
	}
}
*/

/*

TODO: Fix this!!!

// do parallel search for matches
func searchDB(what string) []Hit {
	hits := make([]Hit, 0, 16)
	c := make(chan Hit, 64)
	var wg, wg2 sync.WaitGroup

	wg2.Add(1)
	go func() {
		for hit := range c {
			log.Println("HIT:", hit)
			hits = append(hits, hit)
		}
		wg.Done()
	}()

	search := func(q string, args ...interface{}) {
		err, found := datastore.LoadMany(q, Hit{}, args...)
		if err != nil {
			log.Printf("search error:", err)
		} else {
			for _, hit := range found.([]Hit) {
				c <- hit
			}
		}
		wg.Done()
	}

	q := "select id, 'server' as kind, hostname from servers where hostname=? or sn=? or alias=? or asset_tag=?"
	wg.Add(1)
	go search(q, what, what, what, what)
	if ip := net.ParseIP(what); ip != nil {
		q = "select id, kind, hostname from ipmstr where ip=?"
		wg.Add(1)
		go search(q, what)
	}
	wg.Wait()
	return hits
}
*/

func dbHits(q string, args ...interface{}) []Hit {
	err, found := datastore.LoadMany(q, Hit{}, args...)
	if err != nil {
		log.Println("search error:", err)
		return nil
	}
	//log.Println("CNT:", len(found.([]Hit)))
	return found.([]Hit)
}

func searchDB(what string) []Hit {
	// TESTING ONLY!!!
	dbDebug(true)
	defer dbDebug(false)

	q := "select did as id, 'server' as kind, hostname from devices where hostname=? or sn=? or alias=? or asset_tag=? or profile=?"
	hits := dbHits(q, what, what, what, what, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select vmi as id, 'vm' as kind, hostname from vms where hostname=?"
	hits = dbHits(q, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select did as id, devtype as kind, hostname from devices_network where mac=? or ipv4=?"
	hits = dbHits(q, what, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select did as id, 'server' as kind, hostname from devices where hostname like ?"
	hits = dbHits(q, "%"+what+"%")
	if len(hits) > 0 {
		return hits
	}

	q = "select id, kind, hostname from notes where note MATCH ?"
	hits = dbHits(q, what)
	if len(hits) > 0 {
		return hits
	}

	q = "select id, kind, hostname from notes where note MATCH ?"
	hits = dbHits(q, what+"*")
	if len(hits) > 0 {
		return hits
	}

	q = "select id, kind, hostname from notes where note MATCH ?"
	hits = dbHits(q, "*"+what)
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

	webServer(webHandlers)
}
