package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/arhsxro/platform-go-challenge/api"
	"github.com/arhsxro/platform-go-challenge/config"
	"github.com/arhsxro/platform-go-challenge/storage"
	"github.com/gorilla/mux"
)

var db *storage.PostgresStore

func main() {

	cfg := config.LoadConfig()

	// Initialize database
	maxAttempts := 5
	var err error
	var dbInstance *storage.PostgresStore
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		dbInstance, err = storage.NewPostgresStore(cfg)
		if err == nil {
			break
		}
		log.Println("Attempt "+strconv.Itoa(attempt)+": Failed to connect to database: ", err)
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbInstance.Close()

	db = dbInstance

	// Initialize API with the database instance
	api.Init(db)

	router := mux.NewRouter()
	router.HandleFunc("/favorites/{user_id}", api.HandleGetFavorites).Methods("GET")
	router.HandleFunc("/favorites/{user_id}", api.HandleAddFavorite).Methods("POST")
	router.HandleFunc("/multiple/favorites/{user_id}", api.HandleAddMultipleFavorites).Methods("POST")
	router.HandleFunc("/favorites/{user_id}/{asset_id}", api.HandleRemoveFavorite).Methods("DELETE")
	router.HandleFunc("/favorites/{user_id}/{asset_id}", api.HandleEditDescription).Methods("PUT")

	http.ListenAndServe(":8080", router)
}
