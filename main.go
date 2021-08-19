package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/bmizerany/pat"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Contact struct {
	ID      int
	Name    string
	Email   string
	Content string
	Created time.Time
}

func home(w http.ResponseWriter, r *http.Request) {
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

func createContact(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
	}

	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")
	content := r.PostForm.Get("content")

	_, err = addContact(Contact{
		Name:    name,
		Email:   email,
		Content: content,
		Created: time.Now(),
	})

	if err != nil {
		fmt.Printf("addContact: %v", err)
	}
}

func main() {
	addr := flag.String("addr", ":4000", "Network address")
	dsn := flag.String("dsn", "alex:Or$kumashnd417@/contactbox?parseTime=True", "MYSQL data source name")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	erroLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Llongfile)

	var err error
	db, err = openDB(*dsn)
	if err != nil {
		erroLog.Fatal(err)
	}

	defer db.Close()
	//initialize a new severmux
	mux := pat.New()

	// register the home handler
	mux.Get("/", http.HandlerFunc(home))
	mux.Post("/contact", http.HandlerFunc(createContact))

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: erroLog,
		Handler:  mux,
	}

	infoLog.Printf("Starting a server on %s", *addr)

	// start a new webserver on port :4000
	err = srv.ListenAndServe()
	erroLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func addContact(cont Contact) (int64, error) {
	result, err := db.Exec("INSERT INTO contacts (name, email, content, created) VALUES( ?, ?, ?, ?)", cont.Name, cont.Email, cont.Content, cont.Created)

	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
