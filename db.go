package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
	/*
		"database/sql"
		"net"
		"reflect"
		"regexp"
		"sort"
		"strconv"
		"strings"
	*/

	"github.com/paulstuart/dbutil"
)

var (
	datastore   dbutil.DBU
	ErrNoDB     = fmt.Errorf("no database")
	ErrReadOnly = fmt.Errorf("database is read only")
)

/*
dbServer.Add
dbServer.BackedUp
dbServer.Backup
dbServer.Changed
dbServer.Close
dbServer.Cmd
dbServer.Debug
dbServer.Delete
dbServer.Exec
dbServer.FindSelf
dbServer.GetInt
dbServer.Insert
dbServer.ObjectDelete
dbServer.ObjectInsert
dbServer.ObjectList
dbServer.ObjectListQuery
dbServer.ObjectLoad
dbServer.ObjectUpdate
dbServer.Replace
dbServer.Rows
dbServer.Save
dbServer.Stats
dbServer.StreamCSV
dbServer.StreamJSON
dbServer.StreamTab
dbServer.Table
dbServer.Version
*/

func readable() error {
	if datastore.DB == nil {
		return ErrNoDB
	}
	return nil
}

func writable() error {
	if datastore.DB == nil {
		return ErrNoDB
	}
	if cfg.Main.ReadOnly {
		return ErrReadOnly
	}
	return nil
}

func dbAdd(o dbutil.DBObject) error {
	if err := writable(); err != nil {
		return err
	}
	return datastore.Add(o)
}
func dbBackedUp() {}
func dbBackup(to string) error {
	if datastore.DB == nil {
		return ErrNoDB
	}
	return datastore.Backup(to)
}
func dbChanged() {}
func dbClose() error {
	if datastore.DB == nil {
		return ErrNoDB
	}
	return datastore.Close()
}
func dbCmd() {}
func dbDelete(o dbutil.DBObject) error {
	if err := writable(); err != nil {
		return err
	}
	return datastore.Delete(o)
}
func dbDebug(enable bool) {
	if datastore.DB == nil {
		return
	}
	datastore.Debug = enable
}
func dbExec(query string, args ...interface{}) error {
	if err := writable(); err != nil {
		return err
	}
	_, err := datastore.Exec(query, args...)
	return err
}
func dbFindSelf(o dbutil.DBObject) error {
	if err := readable(); err != nil {
		return err
	}
	return datastore.FindSelf(o)
}
func dbGetInt(q string, args ...interface{}) (int, error) {
	if err := readable(); err != nil {
		return -1, err
	}
	return dbGetInt(q, args...)
}

func dbInsert(q string, args ...interface{}) (i int64, e error) {
	if err := writable(); err != nil {
		return -1, err
	}
	return datastore.Insert(q, args...)
}
func dbObjectDelete(obj interface{}) error {
	if err := writable(); err != nil {
		return err
	}
	return datastore.ObjectDelete(obj)
}
func dbObjectInsert(obj interface{}) (int64, error) {
	if err := writable(); err != nil {
		return -1, err
	}
	return datastore.ObjectInsert(obj)
}
func dbObjectList(kind interface{}) (interface{}, error) {
	if err := readable(); err != nil {
		return nil, err
	}
	return datastore.ObjectList(kind)
}
func dbObjectListQuery(kind interface{}, extra string, args ...interface{}) (interface{}, error) {
	if err := readable(); err != nil {
		return nil, err
	}
	return datastore.ObjectListQuery(kind, extra, args...)
}

func dbObjectLoad(obj interface{}, extra string, args ...interface{}) error {
	if datastore.DB == nil {
		return ErrNoDB
	}
	return datastore.ObjectLoad(obj, extra, args...)
}
func dbObjectUpdate(obj interface{}) error {
	if err := writable(); err != nil {
		return err
	}
	return datastore.ObjectUpdate(obj)
}

func dbReplace(o dbutil.DBObject) error {
	if err := writable(); err != nil {
		return err
	}
	return datastore.Replace(o)
}
func dbRows(q string, args ...interface{}) (results []string, err error) {
	if err := readable(); err != nil {
		return []string{}, err
	}
	return datastore.Rows(q, args...)
}
func dbSave(o dbutil.DBObject) error {
	if err := writable(); err != nil {
		return err
	}
	return datastore.Save(o)
}
func dbStats() []string {
	if err := readable(); err != nil {
		return []string{}
	}
	return datastore.Stats()
}

func dbStreamJSON(w io.Writer, query string, args ...interface{}) error {
	if err := readable(); err != nil {
		return err
	}
	return datastore.StreamJSON(w, query, args...)
}
func dbStreamCSV(w io.Writer, query string, args ...interface{}) error {
	if err := readable(); err != nil {
		return err
	}
	return datastore.StreamCSV(w, query, args...)
}

func dbStreamTab(w io.Writer, query string, args ...interface{}) error {
	if err := readable(); err != nil {
		return err
	}
	return datastore.StreamTab(w, query, args...)
}

func dbTable(query string, args ...interface{}) (*dbutil.Table, error) {
	if err := readable(); err != nil {
		return &dbutil.Table{}, err
	}
	return datastore.Table(query, args...)
}
func dbVersion() {}

func dbPrep() {
	var fresh bool
	var err error
	//log.Println("DBFILE:", dbFile)
	if _, err = os.Stat(dbFile); os.IsNotExist(err) {
		fresh = true
	}
	datastore, err = dbutil.Open(dbFile, true)
	if err != nil {
		panic(err)
	}
	if fresh {
		err = datastore.File(sqlSchema)
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
			dbBackup(to)
			//}
		}
	}

}
