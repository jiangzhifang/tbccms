package pgsql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jiangzhifang/tbccms/pkg/models"
)

// SnippetModel Define a SnippetModel type which wraps a sql.DB connection pool.
type CoursewareModel struct {
	DB *sql.DB
}

func (m *CoursewareModel) Insert(courseCode, courseTitle string, active bool) error {
	stmt := `INSERT INTO coursewares (course_code, course_title, created, active)
	VALUES ($1, $2, $3, $4)`

	now := time.Now()

	_, err := m.DB.Exec(stmt, courseCode, courseTitle, now, active)

	if err != nil {
		fmt.Println("error for insert")
	}
	return nil
}

func (m *CoursewareModel) Get(courseCode string) (*models.Courseware, error) {
	stmt := `SELECT course_code, course_title, created, active FROM coursewares
	WHERE course_code = $1`

	c := &models.Courseware{}
	err := m.DB.QueryRow(stmt, courseCode).Scan(&c.CourseCode, &c.CourseTitle, &c.Created, &c.Active)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return c, nil
}

// This will return the 10 most recently created snippets.
func (m *CoursewareModel) Latest() ([]*models.Courseware, error) {
	stmt := `SELECT course_code, course_title, created, active FROM coursewares
	WHERE active = true ORDER BY created DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	coursewares := []*models.Courseware{}
	for rows.Next() {
		c := &models.Courseware{}
		err = rows.Scan(&c.CourseCode, &c.CourseTitle, &c.Created, &c.Active)
		if err != nil {
			return nil, err
		}
		coursewares = append(coursewares, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return coursewares, nil
}
