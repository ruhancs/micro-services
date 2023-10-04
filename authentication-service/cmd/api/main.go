package main

import (
	"auth-service/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq" //driver postgres
)

const webPort = "80"

var counts int64

type Config struct {
	DB *sql.DB
	Models data.Models
}

func main() {
	log.Println("start auth service")

	conn := connectToDB()
	if conn == nil {
		log.Panic("Can't connect to DB!")
	}

	app := Config{
		DB: conn,
		Models: data.New(conn),
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB,error) {
	db,err := sql.Open("postgres",dsn)
	if err != nil {
		return nil,err
	}
	
	err = db.Ping()
	if err != nil {
		return nil,err
	}

	return db,nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	
	for {
		conn,err := openDB(dsn)
		if err != nil {
			log.Println("DB not yet ready ..")
			counts++
		} else {
			log.Println("Connected to DB")
			return conn
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("backing off for two seconds...")
		time.Sleep(2 * time.Second)
		continue
	}
}