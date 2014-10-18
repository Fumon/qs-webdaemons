package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

const servePort = ":80"

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s, %s", msg, err)
		panic(fmt.Sprintf("%s, %s", msg, err))
	}
}

func main() {
	log.Println("Life Begins")

	// Connect to DB
	db, err := sql.Open("postgres", "user=approot dbname='quantifiedSelf' sslmode=disable")
	failOnError(err, "Error connecting to database")
	defer db.Close()

	http.HandleFunc("/weighscale/validate/", func(w http.ResponseWriter, r *http.Request) {
		// Get did from URL
		vars := strings.Split(r.URL.Path, "/")
		did, err := strconv.ParseInt(vars[len(vars)-1], 10, 64)
		if err != nil {
			log.Println("Error in parsing int, ", err)
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
		// Copy did to proper db
		var newdid int
		err = db.QueryRow(`INSERT INTO aleph.weight(date, weight) SELECT date, weight FROM buffer.weight WHERE did=$1 RETURNING did`, did).Scan(&newdid)
		if err != nil {
			log.Println("Error inserting into db, ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Success!
		fmt.Fprint(w, "Success! New did: ", newdid)
	})

	log.Fatal(http.ListenAndServe(servePort, nil))

	// Die gracefully
	killchan := make(chan os.Signal)
	signal.Notify(killchan, os.Interrupt, os.Kill)

	log.Print(" [*] Waiting on messages")

	<-killchan

	log.Println("Shutting down...")
}
