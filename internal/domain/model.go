package domain

import "time"

type HouseholdMember struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Initials string   `json:"initials"`
	HeightCm *float64 `json:"heightCm"`
	Notes    []string `json:"notes"`
}

type Household struct {
	Name        string            `json:"name"`
	Members     []HouseholdMember `json:"members"`
	Constraints []string          `json:"constraints"`
	Goals       []string          `json:"goals"`
	Preferences []string          `json:"preferences"`
}

type InventoryItem struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Amount     float64  `json:"amount"`
	Unit       string   `json:"unit"`
	Category   string   `json:"category"`
	LowAt      float64  `json:"lowAt"`
	Count      *float64 `json:"count"`
	CountUnit  string   `json:"countUnit"`
	Storage    string   `json:"storage"`
	RecordedOn string   `json:"recordedOn"`
	Notes      string   `json:"notes"`
}

type MealIngredient struct {
	InventoryID string  `json:"inventoryId"`
	Name        string  `json:"name"`
	Amount      float64 `json:"amount"`
	Unit        string  `json:"unit"`
	Optional    bool    `json:"optional,omitempty"`
}

type Meal struct {
	Locale      string           `json:"locale"`
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Reason      string           `json:"reason"`
	Emoji       string           `json:"emoji"`
	Accent      string           `json:"accent"`
	Minutes     int              `json:"minutes"`
	Difficulty  string           `json:"difficulty"`
	Tags        []string         `json:"tags"`
	Ingredients []MealIngredient `json:"ingredients"`
	Steps       []string         `json:"steps"`
}

type Rating string

const (
	RatingLoved    Rating = "loved"
	RatingOkay     Rating = "okay"
	RatingNotForUs Rating = "not-for-us"
)

func (r Rating) Valid() bool {
	return r == RatingLoved || r == RatingOkay || r == RatingNotForUs
}

type HistoryEntry struct {
	ID       string    `json:"id"`
	MealID   string    `json:"mealId"`
	MealName string    `json:"mealName"`
	Emoji    string    `json:"emoji"`
	CookedAt time.Time `json:"cookedAt"`
	Rating   Rating    `json:"rating"`
	Note     string    `json:"note"`
}

type State struct {
	Household      Household       `json:"household"`
	Inventory      []InventoryItem `json:"inventory"`
	Meals          []Meal          `json:"meals"`
	History        []HistoryEntry  `json:"history"`
	SelectedMealID *string         `json:"selectedMealId"`
}
