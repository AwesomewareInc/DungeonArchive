package main

import (
	"net/http"
	"net/url"
	"text/template" // due to markdown and wanting better code we cannot use html/template lol
	"embed"
	"log"
	"time"
	"os"

	// todo: wait fuck we don't need this
	"github.com/gabriel-vasile/mimetype" 
)

//go:embed pages/*.*
var pages embed.FS
var tmpl *template.Template

var queryValues url.Values

func main() {
	// initialize the template shit
	tmpl = template.New("")
	tmpl.Funcs(FuncMap)
	_, err := tmpl.ParseFS(pages, "pages/*")
	if err != nil {
		log.Println(err)
	}

	// initialize the main server
	s := &http.Server{
		Addr:           ":8081",
		Handler:        http.HandlerFunc(handlerFunc),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatalln(err);
	}
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	// How are we trying to access the site?
	switch r.Method {
		case http.MethodGet, http.MethodHead: 	// These methods are allowed. continue.
		default:								// Send them an error for other ones.
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
	}

	// Get the pagename.
	pagename := r.URL.EscapedPath()

	// Cache the query values for the template page
	queryValues = r.URL.Query()

	// Then try and get the file name; they could either be trying to access an internal page in the files folder,
	// or a file anywhere.
	if(pagename[0] == '/') {
		pagename = pagename[1:]
	} 
	if(pagename == "") {
		pagename = "index"
	}

	var internal bool
	var filename string

	// Check if it could refer to an internal page
	if _, err := os.Open("pages/"+pagename+".html"); err == nil { 
		filename = "pages/"+pagename+".html"
		internal = true
	// Otherwise, check if it could refer to a regular file.
	} else {
		if _, err := os.Open("./"+pagename); err == nil {
			filename = "./"+pagename
		} else {
			// If all else fails, send a 404.
			http.Error(w, err.Error(), 404) 
			return
		}
	}

	// get the mime-type.
	contentType, err := mimetype.DetectFile(filename)
	if(err != nil) {
		http.Error(w, err.Error(), 500) 
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", contentType.String())
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Name", filename)

	// Serve the file differently based on whether it's an internal page or not.
	if(internal) {
		if err := tmpl.ExecuteTemplate(w, pagename+".html", nil); err != nil {
			http.Error(w, err.Error(), 500) 
		}
	} else {
		page, err := os.ReadFile(filename)
		if(err != nil) {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(page)
	}
}

// Template friendly function for getting a query
func GetQuery(value string) (string) {
	return queryValues.Get(value)
}