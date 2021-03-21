package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Product struct {
	Name string `json:"name"`
	Cost int `json:"cost"`
	Quantity int `json:"quantity"`
}

func main(){
	log.Printf("Server started")
	defer log.Println("Completed")

	api := http.Server{
		Addr: "localhost:8000",
		Handler: http.HandlerFunc(ListProducts),
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func(){
		log.Printf("Api listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
		case err := <- serverErrors:
			log.Fatalf("error: listenin on server %s", err)
		case <- shutdown:
			log.Println("main: Start shutdown")

		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err := api.Shutdown(ctx)
		if err != nil{
			log.Printf("main: Grateful shutdown did not complete in %v: %v", timeout, err)
			err = api.Close()
		}
		if err != nil{
			log.Fatalf("main: could not stop server gracefully: %v", err)
		}
	}
}

func ListProducts(w http.ResponseWriter, r *http.Request){
	list := []Product{
		{Name: "Gym Beam", Cost: 1100, Quantity: 8},
		{Name: "Wiskey", Cost: 1500, Quantity: 6},
		{Name: "Vodka", Cost: 700, Quantity: 5},
	}
	data, err := json.Marshal(list)
	if err != nil {
		log.Println("error marshaling", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type","application:json; charset=utf-8")
	if _, err := w.Write(data); err != nil{
		log.Println("error writing", err)
	}
}
