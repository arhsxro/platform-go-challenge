package models

import "encoding/json"

type User struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id" db:"user_id"`
	Favourites []Asset `json:"favourites"`
}

type AssetType string

const (
	ChartType    AssetType = "Chart"
	InsightType  AssetType = "Insight"
	AudienceType AssetType = "Audience"
)

var ValidAssetTypes = []AssetType{ChartType, InsightType, AudienceType}

type Asset struct {
	ID          string          `json:"id" db:"asset_id"`
	Type        AssetType       `json:"type" db:"type"`
	Description string          `json:"description" db:"description"`
	Data        json.RawMessage `json:"data" db:"data"`
}

type Chart struct {
	Title     string `json:"title"`
	AxisTitle string `json:"axisTitle"`
	Data      []int  `json:"data"`
}

type Insight struct {
	Text string `json:"text"`
}

type Audience struct {
	Gender             string `json:"gender"`
	BirthCountry       string `json:"birthCountry"`
	AgeGroup           string `json:"ageGroup"`
	SocialMediaHours   int    `json:"socialMediaHours"`
	PurchasesLastMonth int    `json:"purchasesLastMonth"`
}
