package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gomodule/redigo/redis"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! you've requested %s\n", r.URL.Path)
	})

	http.HandleFunc("/mysql", func(w http.ResponseWriter, r *http.Request) {
		uri := os.Getenv("DATABASE_URL")
		if uri == "" {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprint(w, "no DATABASE_URL env var")
			return
		}

		db, err := sql.Open("mysql", uri)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "error connecting to the database: "+err.Error())
			return
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "error connecting to the database: "+err.Error())
			return
		}

		fmt.Fprint(w, "Successfully connected and pinged.")
	})

	http.HandleFunc("/redis", func(w http.ResponseWriter, r *http.Request) {
		uri := os.Getenv("DATABASE_URL")
		if uri == "" {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprint(w, "no DATABASE_URL env var")
			return
		}

		pool := &redis.Pool{
			MaxIdle:     10,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.DialURL(uri)
			},
		}

		c := pool.Get()
		defer c.Close()

		_, err := c.Do("PING")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "error connecting to the database: "+err.Error())
			return
		}

		fmt.Fprint(w, "Successfully connected and pinged.")
	})

	http.HandleFunc("/mongo", func(w http.ResponseWriter, r *http.Request) {
		uri := os.Getenv("DATABASE_URL")
		if uri == "" {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprint(w, "no DATABASE_URL env var")
			return
		}

		// Create a new client and connect to the server
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "error connecting to the database: "+err.Error())
			return
		}
		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				fmt.Print(err)
			}
		}()

		// Ping the primary
		if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "error ping to the database: "+err.Error())
			return
		}

		fmt.Fprint(w, "Successfully connected and pinged.")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	bindAddr := fmt.Sprintf(":%s", port)

	fmt.Printf("==> Server listening at %s 🚀\n", bindAddr)

	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}