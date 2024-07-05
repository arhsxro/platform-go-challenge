package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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
	log.Println("GET request received for user : ", userID)

	//Get filtering type
	queryParams := r.URL.Query()
	filterType := queryParams.Get("type")

	// Get pagination parameters
	pageStr := queryParams.Get("page")
	pageSizeStr := queryParams.Get("pageSize")

	// Convert pagination parameters to integers
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	log.Println("userid : " + userID + " type : " + filterType + " page : " + pageStr + " page size : " + pageSizeStr)

	assets, err := db.GetUserFavorites(userID, filterType, page, pageSize)
	if err != nil {
		log.Println("Invalid asset type or Query failed", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = WriteJSON(w, http.StatusOK, assets)
	if err != nil {
		log.Println("Error writing the json", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleAddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]

	var asset models.Asset
	err := json.NewDecoder(r.Body).Decode(&asset)
	if err != nil {
		log.Println("Invalid request payload", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Println("POST request received for user : "+userID+" and asset id : "+
		asset.ID, asset.Type, asset.Description, string(asset.Data))

	err = db.AddFavorite(userID, asset)
	if err != nil {
		log.Println("Error on executing the query", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandleRemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]
	assetID := mux.Vars(r)["asset_id"]

	log.Println("DELETE request received for user : ", userID+" with asset id : "+assetID)

	err := db.RemoveFavorite(userID, assetID)
	if err != nil {
		log.Println("Error on executing the query", err)
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
		log.Println("Invalid request payload", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Println("PUT request received for user : ", userID+" with asset id : "+assetID+" and updateDescreption : "+updatedDescription.Description)

	err = db.UpdateFavoriteDescription(userID, assetID, updatedDescription.Description)
	if err != nil {
		log.Println("Error on executing the query", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
