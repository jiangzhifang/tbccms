package pgsql

import (
	"database/sql"
	"strings"

	"github.com/jiangzhifang/tbccms/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
	VALUES($1, $2, $3, CURRENT_TIMESTAMP(0))`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))

	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"users_uc_email\"") {
			return models.ErrDuplicateEmail
		}

	}
	return err
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte
	stmt := "SELECT id, hashed_password FROM users WHERE email = $1 AND active = TRUE"
	row := m.DB.QueryRow(stmt, email)
	err := row.Scan(&id, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	return id, err
}

func (m *UserModel) Get(id int) (*models.User, error) {
	u := &models.User{}

	stmt := `SELECT id, name, email, created, active FROM users WHERE id = $1`
	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}
	return u, nil
}
