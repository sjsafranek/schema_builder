package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	PORT string = "8081"
)

var (
	port string
)

// Index
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("tmpl/schema_builder_template.html"))
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	err := t.ExecuteTemplate(w, "schema_builder_template.html", nil)
	log.Println(r.RemoteAddr, "GET / [200]")
	if err != nil {
		log.Println(err)
		log.Println(r.RemoteAddr, "GET / [500]")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Ping
func PingHandler(w http.ResponseWriter, r *http.Request) {
	type Ping struct {
		Message    string    `json:"message"`
		Registered time.Time `json:"registered"`
		Runtime    float64   `json:"runtime"`
	}
	log.Printf("%s something is happening...", r.RemoteAddr)
	resp := Ping{Message: "Pong", Registered: start_time, Runtime: time.Since(start_time).Seconds()}
	log.Println(r.RemoteAddr, "GET /ping [200]")
	json.NewEncoder(w).Encode(resp)
}

// Create Schema from CSV File
func CreateSchemaHandler(w http.ResponseWriter, r *http.Request) {
	// MAX BYTES READER!!!!
	// LIMIT FILE SIZE

	// upload logic
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		tempfile := "./tmp/" + fmt.Sprintf("%v_"+handler.Filename, time.Now().Unix())

		f, err := os.OpenFile(tempfile, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		results, err := processCsvFile(tempfile)

		if nil != err {
			log.Println(tempfile, err)
			log.Println(r.RemoteAddr, "POST /create_schema [500]")
			// keep tmp file for examination
			w.Write([]byte(`{"status":"error","error":"` + err.Error() + `"}`))
		} else {
			log.Println(r.RemoteAddr, "POST /create_schema [200]")
			// remove tmp file
			os.Remove(tempfile)
			w.Write([]byte(`{"status":"ok","data":` + string(results) + `}`))
		}

	}
}

func Server() {

	// Static Files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Main Routes
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/create_schema", CreateSchemaHandler)
	http.HandleFunc("/ping", PingHandler)

	// Start app
	log.Printf("Magic happens on port %s...", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}

}
