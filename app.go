package main

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	hn "github.com/johnwarden/hn"
)

// FrontPageData contains the data to populate the front page template.
type FrontPageData struct {
	Stories []hn.Item
}

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

func main() {
	fmt.Println("In main")

	sqliteDataDir := os.Getenv("SQLITE_DATA_DIR")
	if sqliteDataDir == "" {
		panic("SQLITE_DATA_DIR not set")
	}

	frontpageDatabaseFilename := fmt.Sprintf("%s/frontpage.sqlite", sqliteDataDir)
	fmt.Println("Database file", frontpageDatabaseFilename)

	db, err := sql.Open("sqlite3", frontpageDatabaseFilename)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	t := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	c := hn.NewClient(&http.Client{
		Timeout:   time.Duration(60 * time.Second),
		Transport: t,
	})

	go storiesCrawler(db, c)
	go rankCrawler(db, c)

	httpServer()

}

func httpServer() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"

	}
	http.HandleFunc("/", frontpageHandler())

	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func runCrawler() {
	sqliteDataDir := os.Getenv("SQLITE_DATA_DIR")

	if sqliteDataDir == "" {
		panic("SQLITE_DATA_DIR not set")
	}

	rankDatasetDatabaseFilename := fmt.Sprintf("%s/dataset.sqlite", sqliteDataDir)

	fmt.Println("Database file", rankDatasetDatabaseFilename)
	db, err := sql.Open("sqlite3", rankDatasetDatabaseFilename)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	sqlQuery := "select id, gain from dataset limit 2"

	rows, err := db.Query(sqlQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var gain int
		err = rows.Scan(&id, &gain)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Got Row", id, gain)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully executed select query")
}
