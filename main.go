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
	"github.com/paulstuart/dbutil"
	"github.com/paulstuart/secrets"
)

var (
	version           = "1.3.2"
	Hostname, _       = os.Hostname()
	Basedir, _        = os.Getwd() // get abs path now, as we will be changing dirs
	execDir, _        = osext.ExecutableFolder()
	log_layout        = "2006-01-02 15:04:05.999"
	start_time        = time.Now()
	sqlDir            = "sql" // dir containing sql schemas, etc
	sqlSchema         = sqlDir + "/schema.sql"
	dbFile            = execDir + "/inventory.db"
	dcLookup          = make(map[string]Datacenter)
	dcIDs             = make(map[int64]Datacenter)
	Datacenters       []Datacenter
	systemLocation, _ = time.LoadLocation("Local")
	dbServer          dbutil.DBU
	pathPrefix        string
	bannerText        string
	cfg               = struct {
		Main    MainConfig
		Backups BackupConfig
		SAML    SAMLConfig
	}{}
)

type MainConfig struct {
	Name   string `gcfg:"name"`
	Port   int    `gcfg:"port"`
	Prefix string `gcfg:"prefix"`
	Banner string `gcfg:"banner"`
	//BackupDir  string `gcfg:"backup_dir"`
	//BackupFreq int    `gcfg:"backup_freq"`
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
	authCookie = cfg.SAML.OKTACookie
	bannerText = cfg.Main.Banner
	key, _ := secrets.KeyGen()
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
	dbServer.Exec("insert into audit_log (uid,ip,action,msg) values(?,?,?,?)", uid, ip, strings.ToLower(action), msg)
}

// load schema if this is a new instance
func dbPrep() {
	var fresh bool
	var err error
	//log.Println("DBFILE:", dbFile)
	if _, err = os.Stat(dbFile); os.IsNotExist(err) {
		fresh = true
	}
	db, err := dbutil.Open(dbFile, true)
	if err != nil {
		panic(err)
	}
	if fresh {
		err = db.File(sqlSchema)
		if err != nil {
			panic(err)
		}
	}
}

func Backups(freq int, to string) {
	if _, err := os.Stat(to); err != nil {
		to = filepath.Join(execDir, to)
		if _, err := os.Stat(to); err != nil {
			log.Fatal(err)
		}
	}
	layout := "2006-01-02_15-04-05"
	t := time.NewTicker(time.Minute * time.Duration(freq))
	for {
		select {
		case now := <-t.C:
			/*
				// affected, lastid, err
				_, _, err := dbServer.Cmd("PRAGMA main.wal_checkpoint(FULL);")
				if err != nil {
					log.Println(err)
				}
				time.Sleep(time.Second)
			*/
			//v, _ := dbServer.Version()
			//log.Println("VERSION", v, "BACKED UP", dbServer.BackedUp)
			//if dbServer.Changed() {
			to := filepath.Join(to, now.Format(layout)+".db")
			dbServer.Backup(to)
			//}
		}
	}

}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for sig := range c {
			log.Println("Got signal:", sig)
			// sig is a ^C, handle it
			err := dbServer.Close()
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
	dbServer, err = dbutil.Open(dbFile, false)
	if err != nil {
		log.Fatalln(err)
	}
	if cfg.Backups.Freq > 0 {
		go Backups(cfg.Backups.Freq, cfg.Backups.Dir)
	}

	getColumns()
	LoadVLANs()

	dc, _ := dbServer.ObjectList(Datacenter{})
	Datacenters = dc.([]Datacenter)
	for _, dc := range Datacenters {
		dcLookup[dc.Name] = dc
		dcIDs[dc.ID] = dc
	}
	webServer(webHandlers)
}
