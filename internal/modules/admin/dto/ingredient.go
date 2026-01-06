package dto

// CreateIngredientRequest - запрос на создание ингредиента с автопереводом
// Frontend отправляет только inputName и inputLang
// Backend автоматически переводит на все 3 языка через AI
type CreateIngredientRequest struct {
	InputName string `json:"inputName" validate:"required,min=1"`        // Название на любом языке
	InputLang string `json:"inputLang" validate:"required,oneof=pl en ru"` // Язык ввода: pl | en | ru
	Category  string `json:"category" validate:"required,oneof=protein vegetable dairy grain condiment other"`
	Unit      string `json:"unit" validate:"required,oneof=g ml pcs kg l"`
}

// CreateIngredientResponse - ответ после создания ингредиента
type CreateIngredientResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    *struct {
		ID             string `json:"id"`
		NamePL         string `json:"namePl"`
		NameEN         string `json:"nameEn"`
		NameRU         string `json:"nameRu"`
		Category       string `json:"category"`
		Unit           string `json:"unit"`
		AutoTranslated bool   `json:"autoTranslated"`
	} `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
