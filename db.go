package main

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

var con *sql.DB

func init() {
	dbPath := os.Getenv("DB_LOCATION")
	if dbPath == "" {
		flag.StringVar(&dbPath, "db", "./test.db", "database path (including filename)")
	}
	log.Printf("Using db path %s\n", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
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
-- 		alter table grudge 
-- 			add column guild varchar(80) null;
-- 		update grudge set guild = '523654514977406976' where guild is null;
	`
	_, err = db.Exec(initSql)
	if err != nil {
		log.Fatalf("%q: %s\n", err, initSql)
		return
	}
}

func InsertGrudge(guild string, reporter string, target string, why string) {
	stmt, err := con.Prepare("insert into grudge (guild, reporter, target, why, created) values (?, ?, ?, ?, DATETIME('now'));")
	if err != nil {
		log.Fatalf("Couldn't write to db %s\n", err)
	}
	defer stmt.Close()

	stmt.Exec(guild, reporter, target, why)
}

func DeleteGrudge(guild string, target string) {
	stmt, err := con.Prepare("delete from grudge where guild = ? and target = ?;")
	if err != nil {
		log.Fatalf("Error while preparing statement %s %s, %s\n", guild, target, err)
	}
	defer stmt.Close()
	stmt.Exec(guild, target)
}

func ListGrudges(guild string) string {
	response := ""
	var line string

	stmt, err := con.Prepare("select target || ' : ' || reporter || ' : ' || why || ' @ ' || created from grudge where guild = ? order by target, created, reporter;")
	if err != nil {
		log.Println(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(guild)
	if err != nil {
		log.Fatalf("Couldn't run query against the db %s\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&line)
		if err != nil {
			log.Println(err)
			return ""
		}
		response += line + "\n"
	}

	return response
}
