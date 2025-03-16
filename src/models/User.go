package models

import (
	"time"

	userspb "2k4sm/grpc-crud/proto/users"

	"github.com/scylladb/gocqlx/table"
)

func AccessStrToAccess(access string) userspb.Access {
	res := userspb.Access_UNBLOCKED

	switch access {
	case "BLOCKED":
		res = userspb.Access_BLOCKED
	case "UNBLOCKED":
		res = userspb.Access_UNBLOCKED
	}
	return res
}

func GenderStrToGender(gender string) userspb.Gender {
	res := userspb.Gender_MALE
	switch gender {
	case "MALE":
		res = userspb.Gender_MALE
	case "FEMALE":
		res = userspb.Gender_FEMALE
	}
	return res
}

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
