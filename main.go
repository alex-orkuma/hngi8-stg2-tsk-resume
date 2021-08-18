package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"text/template"
)

func home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	files := []string{
		"./index.page.tmpl",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	err = ts.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

	}

}

func main() {
	addr := flag.String("addr", ":4000", "Network address")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	erroLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)
	//initialize a new severmux
	mux := http.NewServeMux()

	// register the home handler
	mux.HandleFunc("/", home)

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: erroLog,
		Handler:  mux,
	}

	infoLog.Printf("Starting a server on %s", *addr)

	// start a new webserver on port :4000
	err := srv.ListenAndServe()
	erroLog.Fatal(err)
}
