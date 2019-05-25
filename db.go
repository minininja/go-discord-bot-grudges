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

	doMigrations(con)
}

func Grudge(guild string, reporter string, target string, why string) {
	stmt, err := con.Prepare("insert into grudge (guild, reporter, target, why, created) values (?, ?, ?, ?, DATETIME('now'));")
	if err != nil {
		log.Fatalf("Couldn't write to db %s\n", err)
	}
	defer stmt.Close()

	stmt.Exec(guild, reporter, target, why)
}

func Ungrudge(guild string, target string) {
	stmt, err := con.Prepare("delete from grudge where guild = ? and target = ?;")
	if err != nil {
		log.Fatalf("Error while preparing statement %s %s, %s\n", guild, target, err)
	}
	defer stmt.Close()
	stmt.Exec(guild, target)
}

func Grudges(guild string) string {
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

func Ally(guild string, ally string, status string) {
	stmt, err := con.Prepare("insert into ally (guild, ally, status, created) values (?, ?, ?, DATETIME('now')) on conflict(guild, ally) do update set status = ?")
	if nil != err {
		log.Fatalf("Could not prepare query to insert ally: " + err.Error())
	}
	defer stmt.Close()

	stmt.Exec(guild, ally, status, status)
}

func Unally(guild string, ally string) {
	stmt, err := con.Prepare("delete from ally where guild = ? and ally = ?")
	if nil != err {
		log.Fatalf("Could not prepare query to remove ally")
	}
	defer stmt.Close()

	stmt.Exec(guild, ally)
}

func Allies(guild string) string {
	stmt, err := con.Prepare("select ally || ' : ' || status || ' @ ' || created from ally where guild = ? order by status, ally;")
	if nil != err {
		log.Fatalf("Could not prepare query to search for allies: " + err.Error())
	}
	defer stmt.Close()

	rows, err := stmt.Query(guild)
	if nil != err {
		log.Fatalf("Query for allies failed")
	}

	response := ""
	var line string
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

var migrations = []struct {
	ver int
	sql string
}{
	{1, "create table migrations (ver int null); insert into migrations values(0);"},
	{2, "create table if not exists grudge ( guild varchar(80) not null, reporter varchar(80) not null, target varchar(80) not null, why varchar(255) null, created datetime);"},
	{3, "create index if not exists grudge_idx on grudge(guild, target);"},
	{4, "create table if not exists ally (guild varchar(80) not null, ally varchar(80) not null, status varchar(80), created datetime );"},
	{5, "create unique index if not exists ally_idx on ally(guild);"},
	{6, "create table if not exists roe ( guild varchar(80) not null, roe varchar(1024) not null, created datetime );"},
	{7, "create index if not exists roe_idx on row(guild);"},
	{8, "drop index ally_idx;"},
	{9, "create unique index if not exists ally_idx on ally(guild, ally);"},
	{10, "drop table roe;"},

}

func doMigrations(con *sql.DB) {
	var ver int
	row := con.QueryRow("select ver from migrations")
	row.Scan(&ver)

	for _, migration := range migrations {
		if (migration.ver > ver) {
			con.Exec(migration.sql)
			con.Exec("update migrations set ver = ?", migration.ver)
			ver = migration.ver
		}
	}
}

