package service

import (
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/database"
	"gorm.io/gorm"
)

// MetaService handles metadata and reference data
type MetaService interface {
	GetCountries() ([]CountryInfo, error)
	GetCuisines() ([]CuisineInfo, error)
	GetCategories() ([]CategoryInfo, error)
	GetDifficulties() ([]DifficultyInfo, error)
}

type metaService struct {
	db *gorm.DB
}

// NewMetaService creates a new metadata service
func NewMetaService() MetaService {
	return &metaService{db: database.GetDB()}
}

// CountryInfo represents country metadata
type CountryInfo struct {
	Code       string `json:"code"`       // ISO 3166-1 alpha-2 (e.g., "PL", "IT", "GR")
	Name       string `json:"name"`       // English name
	NameLocal  string `json:"nameLocal"`  // Local name (e.g., "Polska", "Italia", "Ελλάδα")
	NamePL     string `json:"namePL"`     // Polish translation
	NameRU     string `json:"nameRU"`     // Russian translation
	RecipeCount int   `json:"recipeCount"` // Number of recipes from this country
	Flag       string `json:"flag"`       // Unicode flag emoji
}

// CuisineInfo represents cuisine type metadata (synonym for country-based cuisine)
type CuisineInfo struct {
	ID          string `json:"id"`          // Cuisine identifier (e.g., "polish", "italian", "greek")
	Name        string `json:"name"`        // English name
	NamePL      string `json:"namePL"`      // Polish translation
	NameRU      string `json:"nameRU"`      // Russian translation
	Country     string `json:"country"`     // Associated country
	RecipeCount int    `json:"recipeCount"` // Number of recipes
	Icon        string `json:"icon"`        // Emoji icon
}

// CategoryInfo represents recipe category metadata
type CategoryInfo struct {
	ID          string `json:"id"`          // Category ID (e.g., "main", "soup", "dessert")
	Name        string `json:"name"`        // English name
	NamePL      string `json:"namePL"`      // Polish translation
	NameRU      string `json:"nameRU"`      // Russian translation
	RecipeCount int    `json:"recipeCount"` // Number of recipes in this category
	Icon        string `json:"icon"`        // Emoji icon
	Order       int    `json:"order"`       // Display order
}

// DifficultyInfo represents difficulty level metadata
type DifficultyInfo struct {
	ID          string `json:"id"`          // Difficulty ID (e.g., "easy", "medium", "hard")
	Name        string `json:"name"`        // English name
	NamePL      string `json:"namePL"`      // Polish translation
	NameRU      string `json:"nameRU"`      // Russian translation
	RecipeCount int    `json:"recipeCount"` // Number of recipes with this difficulty
	Icon        string `json:"icon"`        // Emoji icon
	Level       int    `json:"level"`       // Numeric level (1=easy, 2=medium, 3=hard)
}

// GetCountries returns list of all countries with recipes
func (s *metaService) GetCountries() ([]CountryInfo, error) {
	type CountryResult struct {
		Country string
		Count   int
	}

	var results []CountryResult
	err := s.db.
		Raw(`SELECT country, COUNT(*) as count FROM "Recipe" GROUP BY country ORDER BY country`).
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}

	// Map countries to full info with translations
	countryMap := map[string]CountryInfo{
		"Poland": {
			Code:       "PL",
			Name:       "Poland",
			NameLocal:  "Polska",
			NamePL:     "Polska",
			NameRU:     "Польша",
			Flag:       "🇵🇱",
		},
		"Italy": {
			Code:       "IT",
			Name:       "Italy",
			NameLocal:  "Italia",
			NamePL:     "Włochy",
			NameRU:     "Италия",
			Flag:       "🇮🇹",
		},
		"Greece": {
			Code:       "GR",
			Name:       "Greece",
			NameLocal:  "Ελλάδα",
			NamePL:     "Grecja",
			NameRU:     "Греция",
			Flag:       "🇬🇷",
		},
		"France": {
			Code:       "FR",
			Name:       "France",
			NameLocal:  "France",
			NamePL:     "Francja",
			NameRU:     "Франция",
			Flag:       "🇫🇷",
		},
		"Spain": {
			Code:       "ES",
			Name:       "Spain",
			NameLocal:  "España",
			NamePL:     "Hiszpania",
			NameRU:     "Испания",
			Flag:       "🇪🇸",
		},
		"Germany": {
			Code:       "DE",
			Name:       "Germany",
			NameLocal:  "Deutschland",
			NamePL:     "Niemcy",
			NameRU:     "Германия",
			Flag:       "🇩🇪",
		},
		"Ukraine": {
			Code:       "UA",
			Name:       "Ukraine",
			NameLocal:  "Україна",
			NamePL:     "Ukraina",
			NameRU:     "Украина",
			Flag:       "🇺🇦",
		},
		"Japan": {
			Code:       "JP",
			Name:       "Japan",
			NameLocal:  "日本",
			NamePL:     "Japonia",
			NameRU:     "Япония",
			Flag:       "🇯🇵",
		},
		"China": {
			Code:       "CN",
			Name:       "China",
			NameLocal:  "中国",
			NamePL:     "Chiny",
			NameRU:     "Китай",
			Flag:       "🇨🇳",
		},
		"India": {
			Code:       "IN",
			Name:       "India",
			NameLocal:  "भारत",
			NamePL:     "Indie",
			NameRU:     "Индия",
			Flag:       "🇮🇳",
		},
		"Mexico": {
			Code:       "MX",
			Name:       "Mexico",
			NameLocal:  "México",
			NamePL:     "Meksyk",
			NameRU:     "Мексика",
			Flag:       "🇲🇽",
		},
		"Thailand": {
			Code:       "TH",
			Name:       "Thailand",
			NameLocal:  "ประเทศไทย",
			NamePL:     "Tajlandia",
			NameRU:     "Таиланд",
			Flag:       "🇹🇭",
		},
	}

	countries := make([]CountryInfo, 0, len(results))
	for _, r := range results {
		if info, ok := countryMap[r.Country]; ok {
			info.RecipeCount = r.Count
			countries = append(countries, info)
		}
	}

	return countries, nil
}

// GetCuisines returns list of cuisines (country-based)
func (s *metaService) GetCuisines() ([]CuisineInfo, error) {
	countries, err := s.GetCountries()
	if err != nil {
		return nil, err
	}

	cuisines := make([]CuisineInfo, 0, len(countries))
	for _, country := range countries {
		cuisines = append(cuisines, CuisineInfo{
			ID:          strings.ToLower(country.Name),
			Name:        country.Name + " Cuisine",
			NamePL:      "Kuchnia " + country.NameLocal,
			NameRU:      country.NameRU + "ская кухня",
			Country:     country.Name,
			RecipeCount: country.RecipeCount,
			Icon:        country.Flag,
		})
	}

	return cuisines, nil
}

// GetCategories returns list of recipe categories
func (s *metaService) GetCategories() ([]CategoryInfo, error) {
	type CategoryResult struct {
		Category string
		Count    int
	}

	var results []CategoryResult
	err := s.db.
		Raw(`SELECT category, COUNT(*) as count FROM "Recipe" GROUP BY category ORDER BY category`).
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}

	categoryMap := map[string]CategoryInfo{
		"appetizer": {ID: "appetizer", Name: "Appetizer", NamePL: "Przystawka", NameRU: "Закуска", Icon: "🥗", Order: 1},
		"soup":      {ID: "soup", Name: "Soup", NamePL: "Zupa", NameRU: "Суп", Icon: "🍲", Order: 2},
		"salad":     {ID: "salad", Name: "Salad", NamePL: "Sałatka", NameRU: "Салат", Icon: "🥗", Order: 3},
		"main":      {ID: "main", Name: "Main Course", NamePL: "Danie główne", NameRU: "Основное блюдо", Icon: "🍽️", Order: 4},
		"side":      {ID: "side", Name: "Side Dish", NamePL: "Dodatek", NameRU: "Гарнир", Icon: "🥔", Order: 5},
		"dessert":   {ID: "dessert", Name: "Dessert", NamePL: "Deser", NameRU: "Десерт", Icon: "🍰", Order: 6},
		"beverage":  {ID: "beverage", Name: "Beverage", NamePL: "Napój", NameRU: "Напиток", Icon: "🥤", Order: 7},
	}

	categories := make([]CategoryInfo, 0, len(results))
	for _, r := range results {
		if info, ok := categoryMap[r.Category]; ok {
			info.RecipeCount = r.Count
			categories = append(categories, info)
		}
	}

	return categories, nil
}

// GetDifficulties returns list of difficulty levels
func (s *metaService) GetDifficulties() ([]DifficultyInfo, error) {
	type DifficultyResult struct {
		Difficulty string
		Count      int
	}

	var results []DifficultyResult
	err := s.db.
		Raw(`SELECT difficulty, COUNT(*) as count FROM "Recipe" GROUP BY difficulty ORDER BY difficulty`).
		Scan(&results).Error
	
	if err != nil {
		return nil, err
	}

	difficultyMap := map[string]DifficultyInfo{
		"easy":   {ID: "easy", Name: "Easy", NamePL: "Łatwy", NameRU: "Легко", Icon: "😊", Level: 1},
		"medium": {ID: "medium", Name: "Medium", NamePL: "Średni", NameRU: "Средне", Icon: "😐", Level: 2},
		"hard":   {ID: "hard", Name: "Hard", NamePL: "Trudny", NameRU: "Сложно", Icon: "😰", Level: 3},
	}

	difficulties := make([]DifficultyInfo, 0, len(results))
	for _, r := range results {
		if info, ok := difficultyMap[r.Difficulty]; ok {
			info.RecipeCount = r.Count
			difficulties = append(difficulties, info)
		}
	}

	return difficulties, nil
}
