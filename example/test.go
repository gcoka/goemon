package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf8")
	w.WriteHeader(200)

	fmt.Println("Access from", r.UserAgent())

	fmt.Fprintf(w, "hello %v", "world")
}

func startHTTPServer() *http.Server {
	srv := &http.Server{Addr: ":8080"}

	http.HandleFunc("/", handler)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	return srv
}

func main() {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh,
		os.Interrupt,
		os.Kill,
	)

	srv := startHTTPServer()
	fmt.Println("test server started on localhost:8080")

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	sig := <-quit

	fmt.Println("Shutdown Server with Signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalln("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
