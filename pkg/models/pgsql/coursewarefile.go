package pgsql

import (
	"database/sql"
	"fmt"

	"github.com/jiangzhifang/tbccms/pkg/models"
)

// SnippetModel Define a SnippetModel type which wraps a sql.DB connection pool.
type CoursewareFileModel struct {
	DB *sql.DB
}

func (m *CoursewareFileModel) Insert(courseCode, coursewareFileName string) error {
	stmt := `INSERT INTO courseware_files (course_code, courseware_filename)
	VALUES ($1, $2)`

	_, err := m.DB.Exec(stmt, courseCode, coursewareFileName)

	if err != nil {
		fmt.Println("error for insert")
	}
	return nil
}

func (m *CoursewareFileModel) Get(courseCode string) (*models.CoursewareFile, error) {
	stmt := `SELECT course_code, courseware_filename FROM courseware_files
	WHERE course_code = $1`

	cf := &models.CoursewareFile{}
	err := m.DB.QueryRow(stmt, courseCode).Scan(&cf.CourseCode, &cf.CoursewareFileName)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return cf, nil
}
