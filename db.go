package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var con *sql.DB


func init() {
	db, err := sql.Open("sqlite3", "./test.db")
	if (err != nil) {
		log.Fatal(err)
		return
	}
	con = db

	initSql := `
		create table if not exists grudge (
			reporter varchar(80) not null,
			target varchar(80) not null,
			created datetime
		);
	`
	_, err = db.Exec(initSql)
	if (err != nil) {
		log.Printf("%q: %s\n" ,err, initSql)
		return
	}
}

func InsertGrudge(reporter string, target string) {
	stmt, err := con.Prepare("insert into grudge (reporter, target, created) values (?, ?, DATETIME('now'));")
	if (err != nil) {
		log.Printf("Couldn't write to db %s", err)
	}
	defer stmt.Close()

	stmt.Exec(reporter, target)
}

func DeleteGrudge(target string) {
	stmt, err := con.Prepare("delete from grudge where target = ?;")
	if (err != nil) {
		log.Println(err)
	}
	defer stmt.Close()

	stmt.Exec(target)
}

func ListGrudges() (string) {
	stmt, err := con.Prepare("select target, reporter, created from grudge order by target, reporter;")
	if (err != nil) {
		log.Println(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if (err != nil) {
		log.Println(err)
	}
	defer rows.Close()

	var response string
	var target string
	var reporter string
	var when string
	for (rows.Next()) {
		err := rows.Scan(&target, &reporter, &when)
		if (err != nil) {
			log.Println(err)
			return ""
		}
		response += target + " reported by " + reporter + " on " + when + "\n"
	}

	return response
}