package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"flag"
	"os"
)

var con *sql.DB


func init() {
	dbPath := os.Getenv("DB_LOCATION")
	if (dbPath == "") {
		flag.StringVar(&dbPath, "db", "./test.db", "database path (including filename)")
	}
	log.Printf("Using db path %s\n", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if (err != nil) {
		log.Fatal(err)
		return
	}
	con = db

	initSql := `
		create table if not exists grudge (
			reporter varchar(80) not null,
			target varchar(80) not null,
			why varchar(255) null,
			created datetime
		);
	`
	_, err = db.Exec(initSql)
	if (err != nil) {
		log.Fatalf("%q: %s\n" ,err, initSql)
		return
	}
}

func InsertGrudge(reporter string, target string, why string) {
	stmt, err := con.Prepare("insert into grudge (reporter, target, why, created) values (?, ?, ?, DATETIME('now'));")
	if (err != nil) {
		log.Fatalf("Couldn't write to db %s\n", err)
	}
	defer stmt.Close()

	stmt.Exec(reporter, target, why)
}

func DeleteGrudge(target string) {
	stmt, err := con.Prepare("delete from grudge where target = ?;")
	if (err != nil) {
		log.Fatalf("Delete failed for %s, %s\n", target, err)
	}
	defer stmt.Close()

	stmt.Exec(target)
}

func ListGrudges() (string) {
	response := ""
	var line string

	stmt, err := con.Prepare("select target || ' reported by ' || reporter || ' because ' || why || ' on ' || created from grudge order by target, reporter;")
	if (err != nil) {
		log.Println(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if (err != nil) {
		log.Fatalf("Couldn't run query against the db %s\n", err)
	}
	defer rows.Close()

	for (rows.Next()) {
		err := rows.Scan(&line)
		if (err != nil) {
			log.Println(err)
			return ""
		}
		response += line + "\n"
	}

	return response
}