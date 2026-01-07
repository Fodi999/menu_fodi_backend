package dto

// CreateIngredientAIRequest - запрос на создание ингредиента через AI
// Backend принимает ТОЛЬКО название, всё остальное определяет AI автоматически
type CreateIngredientAIRequest struct {
	InputName string `json:"inputName" validate:"required,min=2"` // Название на ЛЮБОМ языке
}

// CreateIngredientAIResponse - ответ после AI-классификации
type CreateIngredientAIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		ID              string `json:"id"`
		NamePL          string `json:"namePl"`          // AI определяет перевод на польский
		NameEN          string `json:"nameEn"`          // AI определяет перевод на английский
		NameRU          string `json:"nameRu"`          // AI определяет перевод на русский
		Category        string `json:"category"`        // AI определяет категорию
		Unit            string `json:"unit"`            // AI определяет единицу измерения
		NormalizedValue string `json:"normalizedValue"` // AI создает normalized value для дубликатов
		AutoTranslated  bool   `json:"autoTranslated"`  // Всегда true для AI-созданных
	} `json:"data"`
}

// IngredientCategory - допустимые категории (AI выбирает из этого списка)
type IngredientCategory string

const (
	CategoryProtein   IngredientCategory = "protein"   // Белки: мясо, рыба, яйца
	CategoryVegetable IngredientCategory = "vegetable" // Овощи
	CategoryFruit     IngredientCategory = "fruit"     // Фрукты и ягоды
	CategoryDairy     IngredientCategory = "dairy"     // Молочные продукты
	CategoryGrain     IngredientCategory = "grain"     // Крупы, макароны, хлеб
	CategoryCondiment IngredientCategory = "condiment" // Специи, соусы, масла
	CategoryOther     IngredientCategory = "other"     // Прочее
)

// IngredientUnit - допустимые единицы измерения (AI выбирает из этого списка)
type IngredientUnit string

const (
	UnitGrams      IngredientUnit = "g"   // Граммы (для твердых продуктов)
	UnitMilliliter IngredientUnit = "ml"  // Миллилитры (для жидкостей)
	UnitPieces     IngredientUnit = "pcs" // Штуки (для целых единиц)
)

// AIClassificationExample - примеры работы AI
/*
Примеры входных данных и результатов классификации:

Input: "Соль каменная"
Output: {
  namePl: "sól kamienia",
  nameEn: "rock salt",
  nameRu: "соль каменная",
  category: "condiment",
  unit: "g",
  normalizedValue: "salt"
}

Input: "Fresh Eggs"
Output: {
  namePl: "świeże jajka",
  nameEn: "fresh eggs",
  nameRu: "свежие яйца",
  category: "protein",
  unit: "pcs",
  normalizedValue: "egg"
}

Input: "Pomidor"
Output: {
  namePl: "pomidor",
  nameEn: "tomato",
  nameRu: "помидор",
  category: "vegetable",
  unit: "g",
  normalizedValue: "tomato"
}

Input: "Молоко"
Output: {
  namePl: "mleko",
  nameEn: "milk",
  nameRu: "молоко",
  category: "dairy",
  unit: "ml",
  normalizedValue: "milk"
}
*/
