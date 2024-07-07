package api

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arhsxro/platform-go-challenge/models"
	"github.com/arhsxro/platform-go-challenge/storage"

	"github.com/gorilla/mux"
)

var db storage.Store // Use the interface

func Init(dbInstance storage.Store) {
	db = dbInstance
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func HandleGetFavorites(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

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

	var assets []models.Asset
	err = retryWithExponentialBackoff(ctx, func() error {
		var err error
		assets, err = db.GetUserFavorites(ctx, userID, filterType, page, pageSize)
		return err
	})
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Println("Request timed out:", err)
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else if strings.Contains(err.Error(), "invalid asset type") {
			log.Println("Invalid asset type", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			log.Println("Error on executing the query ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
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
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID := mux.Vars(r)["user_id"]
	log.Println("POST request received to add a single asset for user : ", userID)

	var asset models.Asset
	err := json.NewDecoder(r.Body).Decode(&asset)
	if err != nil {
		log.Println("Invalid request payload ", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Println("Asset to be added -> asset id : "+
		asset.ID, asset.Type, asset.Description, string(asset.Data))

	err = retryWithExponentialBackoff(ctx, func() error {
		return db.AddFavorite(ctx, userID, asset)
	})

	if err != nil {
		if err == context.DeadlineExceeded {
			log.Println("Request timed out: ", err)
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			log.Println("Error on executing the query ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandleAddMultipleFavorites(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID := mux.Vars(r)["user_id"]
	log.Println("POST request received to add multiple assets for user : ", userID)

	var assets []models.Asset
	err := json.NewDecoder(r.Body).Decode(&assets)
	if err != nil {
		log.Println("Invalid request payload ", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	errCh := make(chan *models.AssetError, len(assets))

	// Use a goroutine to add each asset concurrently
	for _, asset := range assets {
		log.Println("Asset to be added -> asset id : "+
			asset.ID, asset.Type, asset.Description, string(asset.Data))
		wg.Add(1)
		go func(asset models.Asset) {
			defer wg.Done()
			err = retryWithExponentialBackoff(ctx, func() error {
				return db.AddFavorite(ctx, userID, asset)
			})
			if err != nil {

				errCh <- &models.AssetError{Asset: asset, Err: err}
			}
		}(asset)
	}

	// Wait for all goroutines to finish and close the error channel
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Check for errors in goroutines
	for assetErr := range errCh {
		if assetErr.Err == context.DeadlineExceeded {
			log.Println("Request timed out for asset ID ", assetErr.Asset.ID+" and error message: ", assetErr.Err)
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			log.Println("Error on executing the query for asset ID ", assetErr.Asset.ID+" and error message: ", assetErr.Err)
			http.Error(w, assetErr.Err.Error(), http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func HandleRemoveFavorite(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := mux.Vars(r)["user_id"]
	assetID := mux.Vars(r)["asset_id"]

	log.Println("DELETE request received for user : ", userID+" with asset id : "+assetID)

	err := retryWithExponentialBackoff(ctx, func() error {
		return db.RemoveFavorite(ctx, userID, assetID)
	})

	if err != nil {
		if err == context.DeadlineExceeded {
			log.Println("Request timed out: ", err)
			http.Error(w, "Request timed out ", http.StatusGatewayTimeout)
		} else {
			log.Println("Error on executing the query ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleEditDescription(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	userID := mux.Vars(r)["user_id"]
	assetID := mux.Vars(r)["asset_id"]

	log.Println("PUT request received for user : ", userID)

	var updatedDescription struct {
		Description string `json:"description"`
	}

	err := json.NewDecoder(r.Body).Decode(&updatedDescription)
	if err != nil {
		log.Println("Invalid request payload", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Println("Asset to be edited--> asset id : " + assetID + " and updateDescreption : " + updatedDescription.Description)

	err = retryWithExponentialBackoff(ctx, func() error {
		return db.UpdateDescription(ctx, userID, assetID, updatedDescription.Description)
	})

	if err != nil {
		if err == context.DeadlineExceeded {
			log.Println("Request timed out: ", err)
			http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		} else {
			log.Println("Error on executing the query ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Retry function with exponential backoff
func retryWithExponentialBackoff(ctx context.Context, operation func() error) error {
	const maxAttempts = 3
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		backoff := time.Duration(attempt) * time.Second
		backoff += time.Duration(rand.Intn(1000)) * time.Millisecond

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}
	log.Println("Reached all the retry attemps ")
	return err
}
