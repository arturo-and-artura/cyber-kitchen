package domain

import "time"

type HouseholdMember struct {
	ID       string
	Name     string
	Initials string
}

type Household struct {
	Name        string
	Members     []HouseholdMember
	Constraints []string
	Goals       []string
}

type InventoryItem struct {
	ID       string
	Name     string
	Amount   float64
	Unit     string
	Category string
	LowAt    float64
}

type MealIngredient struct {
	InventoryID string
	Name        string
	Amount      float64
	Unit        string
	Optional    bool
}

type Meal struct {
	ID          string
	Name        string
	Description string
	Reason      string
	Emoji       string
	Accent      string
	Minutes     int
	Difficulty  string
	Tags        []string
	Ingredients []MealIngredient
	Steps       []string
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
	ID       string
	MealID   string
	MealName string
	Emoji    string
	CookedAt time.Time
	Rating   Rating
	Note     string
}

type State struct {
	Household      Household
	Inventory      []InventoryItem
	Meals          []Meal
	History        []HistoryEntry
	SelectedMealID *string
}
