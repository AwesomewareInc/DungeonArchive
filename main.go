package main

import (
	"embed"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template" // due to markdown and wanting better code we cannot use html/template lol
	"time"
)

//go:embed pages/*.*
var pages embed.FS
var tmpl *template.Template

var Values []string // todo: maybe don't make this global since this could change while the user is doing things if accessed fast enough

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
		log.Fatalln(err)
	}
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	// How are we trying to access the site?
	switch r.Method {
	case http.MethodGet, http.MethodHead: // These methods are allowed. continue.
	default: // Send them an error for other ones.
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// Password
	if Config.Password != "" {
		_, pass, ok := r.BasicAuth()
		if !ok || pass != Config.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Get the pagename.
	pagename, values := getPagename(r.URL.EscapedPath())

	var internal bool
	var filename string

	var file *os.File
	var err error

	// Check if it could refer to an internal page
	if file, err = os.Open("pages/" + pagename + ".html"); err == nil {
		filename = "pages/" + pagename + ".html"
		internal = true
		// Otherwise, check if it could refer to a regular file.
	} else {
		if file, err = os.Open("./" + pagename); err == nil {
			filename = "./" + pagename
		} else {
			// If all else fails, send a 404.
			http.Error(w, err.Error(), 404)
			return
		}
	}

	// get the mime-type.
	contentType, err := GetContentType(file)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Name", filename)
	w.WriteHeader(200)

	var Info struct {
		Values []string
		Query  url.Values
	}
	Info.Values = values
	Info.Query = r.URL.Query()

	// Serve the file differently based on whether it's an internal page or not.
	if internal {
		if err := tmpl.ExecuteTemplate(w, pagename+".html", Info); err != nil {
			http.Error(w, err.Error(), 500)
		}
	} else {
		page, err := os.ReadFile(filename)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(page)
	}
}

func getPagename(fullpagename string) (string, []string) {
	// Split the pagename into sections
	if fullpagename[0] == '/' && len(fullpagename) > 1 {
		fullpagename = fullpagename[1:]
	}
	values := strings.Split(fullpagename, "/")

	// Then try and get the relevant pagename from that, accounting for many specifics.
	pagename := values[0]
	// If it's blank, set it to the default page.
	if pagename == "" {
		return "index", values
	}
	// If the first part is resources, then treat the rest of the url normally
	if pagename == "resources" {
		return fullpagename, values
	}
	// If the URL has two parts, then the second part should be an internal page
	// prefixed with campaign
	if len(values) > 2 {
		if values[2] == "" {
			return "campaign", values
		} else {
			return "campaign_" + values[2], values
		}

	}
	return pagename, values
}

func GetContentType(output *os.File) (string, error) {
	ext := filepath.Ext(output.Name())
	file := make([]byte, 1024)
	switch ext {
	case ".htm", ".html":
		return "text/html", nil
	case ".css":
		return "text/css", nil
	case ".js":
		return "application/javascript", nil
	default:
		_, err := output.Read(file)
		if err != nil {
			return "", err
		}
		return http.DetectContentType(file), nil
	}
}
