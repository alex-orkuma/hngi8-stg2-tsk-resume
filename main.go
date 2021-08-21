package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/bmizerany/pat"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var EmailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

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

	errors := make(map[string]string)

	if strings.TrimSpace(name) == "" {
		errors["name"] = "This field cannot be blank"
	} else if utf8.RuneCountInString(name) > 100 {
		errors["name"] = "This field is too long (maximum is 100 characters)"
	}

	if strings.TrimSpace(content) == "" {
		errors["content"] = "This field cannot be blank"
	}

	if strings.TrimSpace(email) == "" {
		errors["email"] = "This field cannot be blank"
	} else if !EmailRegex.MatchString(email) {
		errors["email"] = "Please enter a valid email address"
	}

	if len(errors) > 0 {
		fmt.Fprint(w, errors)
		return
	}

	id, err := addContact(Contact{
		Name:    name,
		Email:   email,
		Content: content,
		Created: time.Now(),
	})

	if err != nil {
		fmt.Printf("addContact: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/contact/%d", id), http.StatusSeeOther)
}

func getContact(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	cont, err := readContact(int64(id))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "Thank you %v for contacting me, I will get back to you soon", cont.Name)
}

func main() {
	addr := flag.String("addr", ":9990", "Network address")
	dsn := flag.String("dsn", "ubuntu:Or$kumashnd417@/contactbox?parseTime=True", "MYSQL data source name")

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
	mux.Get("/contact/:id", http.HandlerFunc(getContact))

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

func readContact(id int64) (Contact, error) {
	var cont Contact

	row := db.QueryRow("SELECT * FROM contacts WHERE id = ?", id)

	if err := row.Scan(&cont.ID, &cont.Name, &cont.Email, &cont.Content, &cont.Created); err != nil {
		if err == sql.ErrNoRows {
			return cont, fmt.Errorf("contactById %d: no such message", id)
		}
		return cont, fmt.Errorf("contactById %d: %v", id, err)
	}

	return cont, nil
}
