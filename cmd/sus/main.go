package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/petsk0/sus"
	bolt "go.etcd.io/bbolt"
)

func main() {
	var name string
	var port string
	var keylen int
	var ttl int
	flag.StringVar(&name, "n", "localhost", "server name")
	flag.StringVar(&port, "p", "8080", "server port")
	flag.IntVar(&keylen, "l", 5, "number of characters in shortened URL")
	flag.IntVar(&ttl, "d", 60, "cached requests time to live in seconds")
	flag.Parse()

	db, err := bolt.Open("sus.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("urls"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	srv := sus.NewShortener(name+":"+port, keylen, db)
	cr := sus.CacheRedirect(srv.HandleRedirect, time.Duration(ttl)*time.Second, db)
	routes := []sus.Route{
		sus.NewRoute("GET", "^/$", srv.HandleGet),
		sus.NewRoute("GET", "^/[a-zA-Z0-9]{"+strconv.Itoa(keylen)+"}$", cr),
		sus.NewRoute("POST", "^/$", srv.HandlePost),
	}
	finish := make(chan os.Signal, 1)
	signal.Notify(finish, os.Interrupt, syscall.SIGINT)
	httpSrv := http.Server{Addr: ":" + port}
	http.HandleFunc("/", sus.Handle(routes))
	go func() {
		e := httpSrv.ListenAndServe()
		if e != nil && e != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Printf("Serving connections at %s:%s", name, port)
	sig := <-finish
	log.Printf("Received signal %v ", sig)
	log.Print("Server shutting down")
	err = httpSrv.Shutdown(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
