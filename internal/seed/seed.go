package seed

import (
	"time"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
)

// InitialState mirrors the frontend MVP fixture in cyber-kitchen-web/src/data/mockData.ts.
func InitialState() domain.State {
	return domain.State{
		Household: domain.Household{
			Name: "The Lee household",
			Members: []domain.HouseholdMember{
				{ID: "yl", Name: "Y. Lee", Initials: "YL"},
				{ID: "al", Name: "A. Lee", Initials: "AL"},
				{ID: "ml", Name: "M. Lee", Initials: "ML"},
			},
			Constraints: []string{"Peanut-free", "Dairy-light"},
			Goals:       []string{"High protein", "Less food waste"},
		},
		Meals: []domain.Meal{
			{
				Locale: "en",
				ID:     "miso-salmon", Name: "Miso-glazed salmon bowls",
				Description: "Caramelized salmon, sesame greens, brown rice, and a bright cucumber crunch.",
				Reason:      "Uses the salmon and cucumber that need attention first, while supporting your high-protein goal.",
				Emoji:       "🍱", Accent: "coral", Minutes: 30, Difficulty: "Easy",
				Tags: []string{"High protein", "Uses 4 pantry items"},
				Ingredients: []domain.MealIngredient{
					{InventoryID: "salmon", Name: "Salmon fillets", Amount: 2, Unit: "fillets"},
					{InventoryID: "rice", Name: "Brown rice", Amount: 1, Unit: "cup"},
					{InventoryID: "cucumber", Name: "Cucumber", Amount: 1, Unit: ""},
					{InventoryID: "spinach", Name: "Baby spinach", Amount: 3, Unit: "cups"},
					{InventoryID: "miso", Name: "White miso", Amount: 2, Unit: "tbsp"},
				},
				Steps: []string{
					"Start the brown rice according to the package instructions.",
					"Whisk miso with a splash of warm water. Brush over the salmon.",
					"Roast salmon at 220°C / 425°F for 10–12 minutes, until it flakes easily.",
					"Quick-sauté the spinach and thinly slice the cucumber.",
					"Build the bowls with rice, greens, salmon, and cucumber. Spoon over any extra glaze.",
				},
			},
			{
				Locale: "en",
				ID:     "chickpea-pasta", Name: "Creamy lemon chickpea pasta",
				Description: "Silky lemon sauce, chickpeas, spinach, and plenty of fresh herbs.",
				Reason:      "A low-effort pantry dinner with no dairy, peanuts, or complicated prep — ideal for a busy night.",
				Emoji:       "🍋", Accent: "yellow", Minutes: 22, Difficulty: "Easy",
				Tags: []string{"Plant-powered", "One pot"},
				Ingredients: []domain.MealIngredient{
					{InventoryID: "pasta", Name: "Whole-wheat pasta", Amount: 300, Unit: "g"},
					{InventoryID: "chickpeas", Name: "Chickpeas", Amount: 1, Unit: "can"},
					{InventoryID: "spinach", Name: "Baby spinach", Amount: 3, Unit: "cups"},
					{InventoryID: "lemon", Name: "Lemon", Amount: 1, Unit: ""},
				},
				Steps: []string{
					"Boil the pasta in well-salted water. Reserve a mug of cooking water.",
					"Drain and rinse the chickpeas, then warm them in the empty pot.",
					"Return pasta to the pot with spinach, lemon zest, and juice.",
					"Add pasta water a little at a time and toss until glossy and creamy.",
					"Season generously and finish with herbs or chili flakes, if you like.",
				},
			},
			{
				Locale: "en",
				ID:     "taco-tray", Name: "Smoky chicken taco tray",
				Description: "Sheet-pan chicken and peppers with warm tortillas and avocado-lime salsa.",
				Reason:      "A playful, family-style option for Friday that keeps every topping customizable at the table.",
				Emoji:       "🌮", Accent: "green", Minutes: 35, Difficulty: "Medium",
				Tags: []string{"Family favorite", "Customizable"},
				Ingredients: []domain.MealIngredient{
					{InventoryID: "chicken", Name: "Chicken breast", Amount: 450, Unit: "g"},
					{InventoryID: "peppers", Name: "Bell peppers", Amount: 2, Unit: ""},
					{InventoryID: "tortillas", Name: "Corn tortillas", Amount: 8, Unit: ""},
					{InventoryID: "avocado", Name: "Avocado", Amount: 1, Unit: ""},
					{InventoryID: "lime", Name: "Lime", Amount: 1, Unit: ""},
				},
				Steps: []string{
					"Heat the oven to 220°C / 425°F and slice the chicken and peppers.",
					"Toss chicken and peppers with oil, smoked paprika, cumin, and salt.",
					"Roast on a sheet pan for 18–20 minutes, stirring halfway.",
					"Mash avocado with lime juice and a pinch of salt.",
					"Warm the tortillas and let everyone build their own tacos.",
				},
			},
		},
		Inventory: []domain.InventoryItem{
			{ID: "salmon", Name: "Salmon fillets", Amount: 2, Unit: "fillets", Category: "Protein", LowAt: 1},
			{ID: "chicken", Name: "Chicken breast", Amount: 450, Unit: "g", Category: "Protein", LowAt: 300},
			{ID: "rice", Name: "Brown rice", Amount: 3, Unit: "cups", Category: "Pantry", LowAt: 1},
			{ID: "pasta", Name: "Whole-wheat pasta", Amount: 500, Unit: "g", Category: "Pantry", LowAt: 200},
			{ID: "chickpeas", Name: "Chickpeas", Amount: 2, Unit: "cans", Category: "Pantry", LowAt: 1},
			{ID: "miso", Name: "White miso", Amount: 8, Unit: "tbsp", Category: "Pantry", LowAt: 2},
			{ID: "cucumber", Name: "Cucumber", Amount: 1, Unit: "", Category: "Produce", LowAt: 1},
			{ID: "spinach", Name: "Baby spinach", Amount: 6, Unit: "cups", Category: "Produce", LowAt: 2},
			{ID: "lemon", Name: "Lemon", Amount: 2, Unit: "", Category: "Produce", LowAt: 1},
			{ID: "peppers", Name: "Bell peppers", Amount: 3, Unit: "", Category: "Produce", LowAt: 1},
			{ID: "avocado", Name: "Avocado", Amount: 2, Unit: "", Category: "Produce", LowAt: 1},
			{ID: "lime", Name: "Lime", Amount: 2, Unit: "", Category: "Produce", LowAt: 1},
			{ID: "tortillas", Name: "Corn tortillas", Amount: 12, Unit: "", Category: "Pantry", LowAt: 4},
		},
		History: []domain.HistoryEntry{
			{ID: "history-1", MealID: "tomato-soup", MealName: "Roasted tomato soup", Emoji: "🍅", CookedAt: mustTime("2026-09-24T18:30:00.000Z"), Rating: domain.RatingLoved, Note: "Great with extra basil."},
			{ID: "history-2", MealID: "veggie-rice", MealName: "Ginger vegetable fried rice", Emoji: "🥕", CookedAt: mustTime("2026-09-21T18:30:00.000Z"), Rating: domain.RatingOkay, Note: "Add more ginger next time."},
		},
		SelectedMealID: nil,
	}
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		panic(err)
	}
	return parsed
}
