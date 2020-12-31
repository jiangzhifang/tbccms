package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

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

func main() {
	dsn := "dbname=tbccms sslmode=disable user=tbccms password=tbccms host=127.0.0.1"

	db, err := openDB(dsn)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(db)

	stmt := "INSERT INTO coursewares (course_code, course_title, created, active) VALUES ($1, $2, $3, $4)"

	now := time.Now()
	db.Exec(stmt, "TEST0004", "测试课程四", now, true)

}
