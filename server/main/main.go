package main

import (
	"fmt"
	"log"
	"net/http"

	"server/api"
	"server/model"
)

func main() {
	data := model.DummyData()
	mux := api.ServeMux(data)

	addr := ":8080"
	fmt.Printf("Listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
