package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/jiangzhifang/tbccms/pkg/models/pgsql"

	_ "github.com/lib/pq"
)

type application struct {
	errorLog        *log.Logger
	infoLog         *log.Logger
	coursewares     *pgsql.CoursewareModel
	coursewareFiles *pgsql.CoursewareFileModel
	templateCache   map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := "dbname=tbccms sslmode=disable user=tbccms password=tbccms host=127.0.0.1"
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	app := &application{
		errorLog:        errorLog,
		infoLog:         infoLog,
		coursewares:     &pgsql.CoursewareModel{DB: db},
		coursewareFiles: &pgsql.CoursewareFileModel{DB: db},
		templateCache:   templateCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
