package main

import (
	"log"
	"net/http"

	"github.com/arhsxro/platform-go-challenge/api"
	"github.com/arhsxro/platform-go-challenge/config"
	"github.com/arhsxro/platform-go-challenge/storage"
	"github.com/gorilla/mux"
)

var db *storage.PostgresStore

func main() {

	cfg := config.LoadConfig()
	var err error
	db, err = storage.NewPostgresStore(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	api.Init(db)

	router := mux.NewRouter()
	router.HandleFunc("/favorites/{user_id}", api.HandleGetFavorites).Methods("GET")
	router.HandleFunc("/favorites/{user_id}", api.HandleAddFavorite).Methods("POST")
	router.HandleFunc("/favorites/{user_id}/{asset_id}", api.HandleRemoveFavorite).Methods("DELETE")
	router.HandleFunc("/favorites/{user_id}/{asset_id}", api.HandleEditDescription).Methods("PUT")

	http.ListenAndServe(":8080", router)
}
