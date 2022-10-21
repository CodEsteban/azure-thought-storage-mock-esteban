package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

type Thought struct {
	Preview string `json:"preview"`
	Title   string `json:"title"`
	History string `json:"history"`
}

func handleError(err error) {
	if err != nil {
		println("-------------------")
		println("azure-thought\nstorage-mock died.")
		println("-------------------")
		fmt.Print(err)
		log.Panic()
	}
}

func main() {
	// Get Environment Variables
	port := os.Getenv("PORT")
	if port == "" {
		println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		println("!!! please provide env variables. !!!")
		println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		log.Panic()
	}

	// Creates context, acts as fake storage
	thoughts := &[]Thought{}
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	contx := context.WithValue(ctx, Thought{}, thoughts)

	// Basic http server
	mux := http.NewServeMux()
	mux.HandleFunc("/readThoughts", readThoughts)
	mux.HandleFunc("/createThought", createThought)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return contx
		},
	}

	println("------------------")
	println("azure-storage-mock\nis up and running!")
	println("------------------")
	server.ListenAndServe()
}

func readThoughts(w http.ResponseWriter, r *http.Request) {
	// Gets Thoughts List from context
	ctx := r.Context()
	thoughtList := ctx.Value(Thought{}).(*[]Thought)

	// Converts thoughts to JSON
	jsonData, err := json.Marshal(thoughtList)
	handleError(err)

	w.Write(jsonData)
}

func badRequest(err error, w *http.ResponseWriter) bool {
	if err != nil {
		fmt.Println(err)
		http.Error(*w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return true
	}
	return false
}

func createThought(w http.ResponseWriter, r *http.Request) {
	// Gets Thoughts List from context
	ctx := r.Context()
	thoughtList := ctx.Value(Thought{}).(*[]Thought)

	body, err := io.ReadAll(r.Body)
	if badRequest(err, &w) {
		return
	}

	fmt.Println(string(body))

	// Parses JSON into Struct
	var data *Thought
	err = json.Unmarshal(body, &data)
	if badRequest(err, &w) {
		return
	}

	// Inserts Thought Struct into slice
	*thoughtList = append(*thoughtList, *data)
}
