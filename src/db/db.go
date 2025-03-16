package db

import (
	"log"
	"os"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
)

func InitDb() *gocqlx.Session {
	cluster := gocql.NewCluster(os.Getenv("SDB_URI"))
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		log.Fatal(err)
	}

	err = session.ExecStmt(`CREATE KEYSPACE IF NOT EXISTS catalog
    	WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1}`)

	if err != nil {
		log.Fatal("Failed to create keyspace:", err)
	}

	err = session.ExecStmt(`CREATE TABLE IF NOT EXISTS catalog.users (
		first_name text,
		last_name text,
		gender text,
		dob date,
		ph_number text,
		email text,
		access text,
	    PRIMARY KEY (email)
	   )`)

	if err != nil {
		log.Fatal("Failed to create table", err.Error())
	}

	err = session.ExecStmt(`CREATE INDEX IF NOT EXISTS ON catalog.users (ph_number);`)
	if err != nil {
		log.Fatal("Error creating index:", err.Error())
	}

	log.Println("Connected to sycalladb at: ", os.Getenv("SDB_URI"))
	return &session
}
