package models

import (
	"time"

	"github.com/scylladb/gocqlx/table"
)

type User struct {
	Email     string    `db:"email"`
	PhNumber  string    `db:"ph_number"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Gender    string    `db:"gender"`
	Dob       time.Time `db:"dob"`
	Access    string    `db:"access"`
}

var UserMetadata = table.Metadata{
	Name:    "catalog.users",
	Columns: []string{"email", "ph_number", "first_name", "last_name", "gender", "dob", "access"},
	PartKey: []string{"email"},
}
