package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	//"github.com/sportsru/gostorage/gostorage"
	"log"
	"net/http"
	"strings"
	//"time"
)

func dataHandler(w http.ResponseWriter, r *http.Request) *appError {
	log.Print("proccess Data")
	var out string
	if r.FormValue("counter") == "" {
		out = store.GetDataJSON(r.FormValue("uid"))
	} else {
		// get tags here
		out = store.GetDataJSON(r.FormValue("uid"))
	}

	// check err instead of out
	if out == "" {
		return &appError{Message: "data not found", Code: 404}
	}

	fmt.Fprintf(w, out)
	return nil
}

// GET
// FIXME: в оригинальном стораджере в случае не найденного документа в кеш писалась -1
// и она же возвращалась в JSON
func versionHandler(w http.ResponseWriter, r *http.Request) *appError {
	log.Print("proccess Version")

	ver := store.GetVersion(r.FormValue("uid"))
	if len(ver) == 0 {
		return &appError{Message: "version not found", Code: 404}
	}

	fmt.Fprintf(w, "{version: "+ver+"}\n")
	return nil
}

func setCounterHandler(w http.ResponseWriter, r *http.Request) *appError {
	uid := r.FormValue("uid")
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
	_ = store.SetTags(uid, tagFields)
	return nil
}

func setHandler(w http.ResponseWriter, r *http.Request) *appError {
	processSetReq(r)

	//fields := make(map[string]interface{})
	fmt.Fprintf(w, "")
	//"Hi there, it's set handler!"

	//log.Print("proccess Set")
	return nil
}

// TODO: add return values
func processSetReq(r *http.Request) {
	//log.Print("Method: " + r.Method)
	uid := r.FormValue("uid")

	fields := make(map[string]interface{})
	for paramName := range r.Form {
		if paramName == "uid" {
			fields[paramName] = r.FormValue(paramName)
			continue
		}
		log.Print("Param: ", paramName)
		spew.Dump(r.FormValue(paramName))
		fields["data."+paramName] = r.FormValue(paramName)
	}

	if cliCfg.debug {
		fmt.Print("fields in handler: ")
		spew.Dump(fields)
	}

	_ = store.SetData(uid, fields)
}
