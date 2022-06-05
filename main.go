package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/xo/dburl"

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
		uri := os.Getenv("MYSQL_URL")
		if uri == "" {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprint(w, "no MYSQL_URL env var")
			return
		}

		dbURL, err := dburl.Parse(uri)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "parsting MYSQL_URL")
			return
		}

		dbPassword, _ := dbURL.User.Password()
		dbName := strings.Trim(dbURL.Path, "/")
		connectionString := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=true",
			dbURL.User.Username(), dbPassword, dbURL.Hostname(), dbURL.Port(), dbName)

		db, err := sql.Open("mysql", connectionString)
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

		fmt.Fprint(w, "Successfully connected and pinged mysql.")
	})

	http.HandleFunc("/redis", func(w http.ResponseWriter, r *http.Request) {
		uri := os.Getenv("REDIS_URL")
		if uri == "" {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprint(w, "no REDIS_URL env var")
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

		fmt.Fprint(w, "Successfully connected and pinged redis.")
	})

	http.HandleFunc("/mongo", func(w http.ResponseWriter, r *http.Request) {
		uri := os.Getenv("MONGODB_URL")
		if uri == "" {
			w.WriteHeader(http.StatusNotImplemented)
			fmt.Fprint(w, "no MONGODB_URL env var")
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

		fmt.Fprint(w, "Successfully connected and pinged mongo.")
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
