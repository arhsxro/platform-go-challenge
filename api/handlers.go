package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/arhsxro/platform-go-challenge/models"
	"github.com/arhsxro/platform-go-challenge/storage"

	"github.com/gorilla/mux"
)

var db *storage.PostgresStore

func Init(dbInstance *storage.PostgresStore) {
	db = dbInstance
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func HandleGetFavorites(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]

	assets, err := db.GetUserFavorites(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = WriteJSON(w, http.StatusOK, assets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleAddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]

	var asset models.Asset
	err := json.NewDecoder(r.Body).Decode(&asset)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = db.AddFavorite(userID, asset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("error with adding to db")
		return
	}
	log.Printf("Received request to add a new favorite asset for user %s: Type: %s, Description: %s, Data: %s", userID, asset.Type, asset.Description, string(asset.Data))
	w.WriteHeader(http.StatusCreated)
}

func HandleRemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]
	assetID := mux.Vars(r)["asset_id"]

	err := db.RemoveFavorite(userID, assetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleEditDescription(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]
	assetID := mux.Vars(r)["asset_id"]

	var updatedDescription struct {
		Description string `json:"description"`
	}

	err := json.NewDecoder(r.Body).Decode(&updatedDescription)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = db.UpdateFavoriteDescription(userID, assetID, updatedDescription.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
