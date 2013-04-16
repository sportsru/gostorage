package main

import (
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/sportsru/gostorage/gostorage"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	/*
		"os/signal"
		"runtime"
		"runtime/pprof"
	*/
)

var store *gostorage.Client
var storeCfg gostorage.Config

var cliCfg struct {
	port, address    string // don't use
	mongo, memcached string
	memPrefix        string
	verbose, debug   bool
}

func init() {
	flag.StringVar(&cliCfg.address, "address", "0.0.0.0", "service address")
	flag.StringVar(&cliCfg.port, "port", "9002", "service port")
	flag.StringVar(&cliCfg.mongo, "mongo", "127.0.0.1", "MongoDB connection string")
	flag.StringVar(&cliCfg.memcached, "memcached", "127.0.0.1", "Memcached servers list")
	flag.StringVar(&cliCfg.memPrefix, "memcached-prefix", "_s", "Memcached keys prefix")

	flag.BoolVar(&cliCfg.verbose, "verbose", false, "Log requests")
	flag.BoolVar(&cliCfg.debug, "debug", false, "Debugging")
	showHelp := flag.Bool("help", false, "")

	flag.Parse()
	if *showHelp {
		os.Stderr.WriteString(`REST service for mongo storage
Run 'gostorage-worker -h' for flags description.
`)
		os.Exit(1)
	}

	//cpuprofile := flag.String("cpuprofile", "", "Write CPU profile to file")
	//memprofile := flag.String("memprofile", "", "Write memory profile to file")

	//storeCfg.Storage = gostorage.StorageCfg{"0.0.0.0", "9002"}
	var memServers []string
	for _, serv := range strings.Split(cliCfg.memcached, ",") {
		if strings.Index(serv, ":") == -1 {
			serv = serv + ":11211"
		}
		memServers = append(memServers, serv)
	}

	storeCfg.Mongo = gostorage.MongoCfg{Url: cliCfg.mongo, Db: "default"}
	storeCfg.Memcache = gostorage.MemcacheCfg{
		Servers:   memServers, // strings.Split(cliCfg.memcached, ","),
		NameSpace: cliCfg.memPrefix,
	}
	storeCfg.Debug = cliCfg.debug
	storeCfg.Verbose = cliCfg.verbose

	store = gostorage.New(storeCfg)
}

func main() {
	_ = spew.Config
	addr := cliCfg.address + ":" + cliCfg.port
	s := &http.Server{
		Addr: addr,
		//Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	http.Handle("/version/", appHandler(versionHandler))
	http.Handle("/data/", appHandler(dataHandler))
	http.Handle("/set/", appHandler(setHandler))
	http.Handle("/setcounter/", appHandler(setCounterHandler))
	log.Print("Listen on " + addr)
	log.Fatal(s.ListenAndServe())
}

type appError struct {
	Error   error
	Message string
	Code    int
}

type appHandler func(http.ResponseWriter, *http.Request) *appError

// implement ServeHTTP for appHanlers
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println("Recovered from panic: %v \n", rec)
			http.Error(w, "internal error", 500)
			return
		}
	}()

	uid := r.FormValue("uid")
	if len(uid) == 0 {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	if cliCfg.debug {
		log.Print("uid: ", uid)
	}

	// TODO: add data struct to return values of fn 
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		fmt.Errorf("%v", e.Error)
		http.Error(w, e.Message, e.Code)
	}
}
