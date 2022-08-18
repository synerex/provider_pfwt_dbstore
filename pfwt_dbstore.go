package main

import (
	// "encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	pflow "github.com/UCLabNU/proto_pflow"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	api "github.com/synerex/synerex_api"
	pbase "github.com/synerex/synerex_proto"

	sxutil "github.com/synerex/synerex_sxutil"
	//sxutil "local.packages/synerex_sxutil"

	"log"
	"sync"
)

// datastore provider provides Datastore Service.

var (
	nodesrv         = flag.String("nodesrv", "127.0.0.1:9990", "Node ID Server")
	local           = flag.String("local", "", "Local Synerex Server")
	mu              sync.Mutex
	version         = "0.01"
	baseDir         = "store"
	dataDir         string
	pcMu            *sync.Mutex = nil
	pcLoop          *bool       = nil
	ssMu            *sync.Mutex = nil
	ssLoop          *bool       = nil
	sxServerAddress string
	currentNid      uint64                  = 0 // NotifyDemand message ID
	mbusID          uint64                  = 0 // storage MBus ID
	storageID       uint64                  = 0 // storageID
	pfClient        *sxutil.SXServiceClient = nil
	pfblocks        map[string]*PFlowBlock  = map[string]*PFlowBlock{}
	holdPeriod                              = flag.Int64("holdPeriod", 720, "Flow Data Hold Time")
	db              *sql.DB
	db_host         = os.Getenv("MYSQL_HOST")
	db_name         = os.Getenv("MYSQL_DATABASE")
	db_user         = os.Getenv("MYSQL_USER")
	db_pswd         = os.Getenv("MYSQL_PASSWORD")
)

const layout = "2006-01-02T15:04:05.999999Z"
const layout_db = "2006-01-02 15:04:05.999"

func init() {
	// connect
	addr := fmt.Sprintf("%s:%s@(%s:3306)/%s", db_user, db_pswd, db_host, db_name)
	print("connecting to " + addr + "\n")
	var err error
	db, err = sql.Open("mysql", addr)
	if err != nil {
		print("connection error: ")
		print(err)
		log.Fatal("\n")
	}

	// ping
	err = db.Ping()
	if err != nil {
		print("ping error: ")
		print(err)
		log.Fatal("\n")
	}

	// create table
	_, err = db.Exec(`create table if not exists pfwt(id BIGINT unsigned not null auto_increment, time DATETIME(3) not null, src INT unsigned not null, wt_data VARCHAR(256), primary key(id))`)
	if err != nil {
		print("create table error: ")
		print(err)
		log.Fatal("\n")
	}
}

func dbStore(ts time.Time, src uint32, wt_data string) {

	// ping
	err := db.Ping()
	if err != nil {
		print("ping error: ")
		print(err)
		print("\n")
		// connect
		addr := fmt.Sprintf("%s:%s@(%s:3306)/%s", db_user, db_pswd, db_host, db_name)
		print("connecting to " + addr + "\n")
		db, err = sql.Open("mysql", addr)
		if err != nil {
			print("connection error: ")
			print(err)
			print("\n")
		}
	}

	log.Printf("Storeing %v, %s, %s", ts.Format(layout_db), src, wt_data)
	result, err := db.Exec(`insert into pfwt(time, src, wt_data) values(?, ?, ?)`, ts.Format(layout_db), src, wt_data)

	if err != nil {
		print("exec error: ")
		print(err)
		print("\n")
	} else {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			print(err)
		} else {
			print(rowsAffected)
		}
	}

}

// called for each agent data.
func supplyPFlowCallback(clt *sxutil.SXServiceClient, sp *api.Supply) {

	pf := &pflow.PFlow{}
	err := proto.Unmarshal(sp.Cdata.Entity, pf)

	if err == nil { // get PFlow
		wt := fmt.Sprintf("%d", pf.Id)
		firstPc := pf.Operation[0]
		firstTs, _ := time.Parse(layout, ptypes.TimestampString(firstPc.Timestamp))
		for _, pc := range pf.Operation {
			ts, _ := time.Parse(layout, ptypes.TimestampString(pc.Timestamp))
			wt += fmt.Sprintf(",%s,%d,%d", ts.Format(layout), pc.Sid, pc.Height)
		}
		dbStore(firstTs, firstPc.Sid, wt)
	}
}

func main() {
	flag.Parse()
	go sxutil.HandleSigInt()
	sxutil.RegisterDeferFunction(sxutil.UnRegisterNode)
	log.Printf("PFWT-dbstore(%s) built %s sha1 %s", sxutil.GitVer, sxutil.BuildTime, sxutil.Sha1Ver)

	channelTypes := []uint32{pbase.PEOPLE_WT_SVC, pbase.STORAGE_SERVICE}

	var rerr error
	sxServerAddress, rerr = sxutil.RegisterNode(*nodesrv, "PFWTdbstore", channelTypes, nil)

	if rerr != nil {
		log.Fatal("Can't register node:", rerr)
	}
	if *local != "" { // quick hack for AWS local network
		sxServerAddress = *local
	}
	log.Printf("Connecting SynerexServer at [%s]", sxServerAddress)

	wg := sync.WaitGroup{} // for syncing other goroutines

	client := sxutil.GrpcConnectServer(sxServerAddress)

	if client == nil {
		log.Fatal("Can't connect Synerex Server")
	}

	pfClient = sxutil.NewSXServiceClient(client, pbase.PEOPLE_WT_SVC, "{Client:PFWTdbStore}")

	log.Print("Subscribe PFlow Supply")
	pcMu, pcLoop = sxutil.SimpleSubscribeSupply(pfClient, supplyPFlowCallback)

	wg.Add(1)
	wg.Wait()
}
