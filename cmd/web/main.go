package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jiangzhifang/tbccms/pkg/models/pgsql"

	"github.com/golangcollege/sessions"
	_ "github.com/lib/pq"
)

type contextKey string

var contextKeyIsAuthenticated = contextKey("isAuthenticated")

type application struct {
	errorLog        *log.Logger
	infoLog         *log.Logger
	session         *sessions.Session
	coursewares     *pgsql.CoursewareModel
	coursewareFiles *pgsql.CoursewareFileModel
	templateCache   map[string]*template.Template
	users           *pgsql.UserModel
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := "dbname=tbccms sslmode=disable user=tbccms password=tbccms host=127.0.0.1"
	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
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

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode

	app := &application{
		errorLog:        errorLog,
		infoLog:         infoLog,
		session:         session,
		coursewares:     &pgsql.CoursewareModel{DB: db},
		coursewareFiles: &pgsql.CoursewareFileModel{DB: db},
		templateCache:   templateCache,
		users:           &pgsql.UserModel{DB: db},
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
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
