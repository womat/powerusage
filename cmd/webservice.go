package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/womat/debug"

	"powerusage/global"
)

func init() {
	InitWebService()
}

func InitWebService() (err error) {
	for pattern, f := range map[string]func(http.ResponseWriter, *http.Request){
		"version":     httpGetVersion,
		"currentdata": httpReadCurrentData,
	} {
		if set, ok := global.Config.Webserver.Webservices[pattern]; ok && set {
			http.HandleFunc("/"+pattern, f)
		}
	}

	port := ":" + strconv.Itoa(global.Config.Webserver.Port)
	go http.ListenAndServe(port, nil)
	return
}

// httpGetVersion prints the SW Version
func httpGetVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(global.VERSION)); err != nil {
		debug.ErrorLog.Println(err)
		return
	}
}

// httpReadCurrentData supplies the data of all meters
func httpReadCurrentData(w http.ResponseWriter, r *http.Request) {
	var j []byte
	var err error

	if j, err = json.MarshalIndent(&global.Measurements, "", "  "); err != nil {
		debug.ErrorLog.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(j); err != nil {
		debug.ErrorLog.Println(err)
		return
	}
}
