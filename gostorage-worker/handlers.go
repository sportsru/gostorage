package main

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	//"io/ioutil"
	//"github.com/sportsru/gostorage/gostorage"
	"log"
	"net/http"
	"strings"
	//"time"
)

func dataHandler(hd *HandlerData, w http.ResponseWriter, r *http.Request) *appError {
	log.Print("proccess Data")
	var out string
	if r.FormValue("counter") == "" {
		out = store.GetDataJSON(hd.uid)
	} else {
		// get tags here
		out = store.GetDataJSON(hd.uid)
	}

	// check err instead of out
	if out == "" {
		return &appError{Message: "data not found", Code: 404}
	}

	fmt.Fprintf(w, out)
	return nil
}

func versionHandler(hd *HandlerData, w http.ResponseWriter, r *http.Request) *appError {
	//log.Print("proccess Version")
	ver := store.GetVersion(hd.uid)
	if len(ver) == 0 {
		return &appError{Message: "version not found", Code: 404}
	}

	fmt.Fprintf(w, "{\"version\": "+ver+"}\n")
	return nil
}

func setCounterHandler(hd *HandlerData, w http.ResponseWriter, r *http.Request) *appError {
	tagStr := r.FormValue("tg")
	if tagStr == "" {
		fmt.Fprintf(w, "")
		return nil
	}

	tagFields := make(map[string]interface{})
	for _, tag := range strings.Split(tagStr, ".") {
		tagFields["tags."+tag] = int32(1)
	}

	if cliCfg.debug {
		fmt.Print("tags: ")
		spew.Dump(tagFields)
	}

	// TODO: process errors
	_ = store.SetTags(hd.uid, tagFields)
	fmt.Fprintf(w, "OK")
	return nil
}

func setHandler(hd *HandlerData, w http.ResponseWriter, r *http.Request) *appError {
	if !hd.isPost {
		return nil
	}

	fields := make(map[string]interface{})
	if err := json.Unmarshal(hd.body, &fields); err != nil {
		panic("PARSE ERROR: " + string(err.Error()) +
			"\n" + string(hd.body))
	}

	fSet := make(map[string]interface{})
	for key, value := range fields {
		fSet["data."+key] = value
	}
	fSet["uid"] = hd.uid
	if cliCfg.debug {
		fmt.Print("fields in SET handler: ")
		spew.Dump(fSet)
	}

	_ = store.SetData(hd.uid, fSet)
	fmt.Fprintf(w, "OK")
	return nil
}
