package main

import (
	// "encoding/json"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	pflow "github.com/UCLabNU/proto_pflow"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/jackc/pgx/v4/pgxpool"
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
	db              *pgxpool.Pool
	db_host         = os.Getenv("POSTGRES_HOST")
	db_name         = os.Getenv("POSTGRES_DB")
	db_user         = os.Getenv("POSTGRES_USER")
	db_pswd         = os.Getenv("POSTGRES_PASSWORD")
)

const layout = "2006-01-02T15:04:05.999999Z"
const layout_db = "2006-01-02 15:04:05.999"

func init() {
	// connect
	ctx := context.Background()
	addr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", db_user, db_pswd, db_host, db_name)
	print("connecting to " + addr + "\n")
	var err error
	db, err = pgxpool.Connect(ctx, addr)
	if err != nil {
		print("connection error: ")
		log.Println(err)
		log.Fatal("\n")
	}
	defer db.Close()

	// ping
	err = db.Ping(ctx)
	if err != nil {
		print("ping error: ")
		log.Println(err)
		log.Fatal("\n")
	}

	// create table
	_, err = db.Exec(ctx, `create table if not exists pfwt(id BIGSERIAL NOT NULL, time TIMESTAMP not null, src INT not null, wt_data VARCHAR(256), primary key(id))`)
	if err != nil {
		print("create table error: ")
		log.Println(err)
		log.Fatal("\n")
	}
}

func dbStore(ts time.Time, src uint32, wt_data string) {

	// ping
	ctx := context.Background()
	err := db.Ping(ctx)
	if err != nil {
		print("ping error: ")
		log.Println(err)
		print("\n")
		// connect
		addr := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", db_user, db_pswd, db_host, db_name)
		print("connecting to " + addr + "\n")
		db, err = pgxpool.Connect(ctx, addr)
		if err != nil {
			print("connection error: ")
			log.Println(err)
			print("\n")
		}
	}

	log.Printf("Storeing %v, %s, %s", ts.Format(layout_db), src, wt_data)
	result, err := db.Exec(ctx, `insert into pfwt(time, src, wt_data) values($1, $2, $3)`, ts.Format(layout_db), src, wt_data)

	if err != nil {
		print("exec error: ")
		log.Println(err)
		print("\n")
	} else {
		rowsAffected := result.RowsAffected()
		if err != nil {
			log.Println(err)
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
		src, _ := strconv.Atoi(pf.Area)
		dbStore(firstTs, uint32(src), wt)
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
