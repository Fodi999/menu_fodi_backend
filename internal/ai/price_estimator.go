package ai

import (
	"fmt"
	"log"
	"strings"
)

// PriceEstimation оценка стоимости блюда
type PriceEstimation struct {
	EstimatedCost float64 `json:"estimatedCost"` // себестоимость
	SuggestedPrice float64 `json:"suggestedPrice"` // рекомендуемая цена продажи
	Margin        float64 `json:"margin"`         // маржа (%)
	PriceCategory string  `json:"priceCategory"`  // budget/mid/premium
	Explanation   string  `json:"explanation"`    // объяснение
}

// PriceEstimator AI-оценщик стоимости блюд
type PriceEstimator struct {
	client *GroqClient
	lang   string
}

// NewPriceEstimator создаёт новый оценщик
func NewPriceEstimator(lang string) *PriceEstimator {
	if lang == "" {
		lang = "pl"
	}
	return &PriceEstimator{
		client: NewGroqClient(),
		lang:   lang,
	}
}

// EstimatePrice оценивает стоимость блюда через AI
func (pe *PriceEstimator) EstimatePrice(recipeName, ingredients string, portionSize int) (*PriceEstimation, error) {
	systemPrompt := pe.getSystemPrompt()
	userMessage := fmt.Sprintf(`Oceń koszt przepisu:
Nazwa: %s
Składniki: %s
Porcje: %d

Zwróć oszacowanie w PLN.`, recipeName, ingredients, portionSize)

	log.Printf("[PRICE] 💰 Estimating price for: %s", recipeName)

	response, err := pe.client.SimpleChat(systemPrompt, userMessage)
	if err != nil {
		log.Printf("[PRICE] ❌ Groq API error: %v", err)
		return pe.getFallbackEstimation(recipeName, ingredients), nil
	}

	estimation, err := pe.parseEstimation(response)
	if err != nil {
		log.Printf("[PRICE] ⚠️ Failed to parse estimation: %v", err)
		return pe.getFallbackEstimation(recipeName, ingredients), nil
	}

	log.Printf("[PRICE] ✅ Estimated: cost=%.2f PLN, price=%.2f PLN (margin: %.1f%%)",
		estimation.EstimatedCost, estimation.SuggestedPrice, estimation.Margin)

	return estimation, nil
}

// getSystemPrompt возвращает system prompt для оценки цен
func (pe *PriceEstimator) getSystemPrompt() string {
	return `Jesteś ekspertem w wycenie kosztów potraw.
Analizujesz składniki i szacujesz:
1. Koszt składników (estimatedCost)
2. Cenę sprzedaży z marżą (suggestedPrice)
3. Procent marży (margin)
4. Kategorię cenową (priceCategory: budget/mid/premium)
5. Krótkie wyjaśnienie (explanation)

Odpowiadaj w formacie: "Cost: 25.50 PLN, Price: 45.00 PLN, Margin: 76%, Category: mid, Explanation: Fresh ingredients and medium complexity"`
}

// parseEstimation parsuje odpowiedź AI
func (pe *PriceEstimator) parseEstimation(response string) (*PriceEstimation, error) {
	// Простой парсинг формата "Cost: X, Price: Y, Margin: Z%, Category: W, Explanation: ..."
	estimation := &PriceEstimation{}

	// Cost
	if cost, err := pe.extractFloat(response, "Cost:"); err == nil {
		estimation.EstimatedCost = cost
	}

	// Price
	if price, err := pe.extractFloat(response, "Price:"); err == nil {
		estimation.SuggestedPrice = price
	}

	// Margin
	if margin, err := pe.extractFloat(response, "Margin:"); err == nil {
		estimation.Margin = margin
	}

	// Category
	if strings.Contains(strings.ToLower(response), "budget") {
		estimation.PriceCategory = "budget"
	} else if strings.Contains(strings.ToLower(response), "premium") {
		estimation.PriceCategory = "premium"
	} else {
		estimation.PriceCategory = "mid"
	}

	// Explanation
	if idx := strings.Index(response, "Explanation:"); idx != -1 {
		estimation.Explanation = strings.TrimSpace(response[idx+12:])
	}

	return estimation, nil
}

// extractFloat извлекает число из строки после указанного префикса
func (pe *PriceEstimator) extractFloat(text, prefix string) (float64, error) {
	idx := strings.Index(text, prefix)
	if idx == -1 {
		return 0, fmt.Errorf("prefix not found")
	}

	// Ищем число после префикса
	text = text[idx+len(prefix):]
	var value float64
	_, err := fmt.Sscanf(text, "%f", &value)
	return value, err
}

// getFallbackEstimation возвращает базовую оценку
func (pe *PriceEstimator) getFallbackEstimation(recipeName, ingredients string) *PriceEstimation {
	// Простая эвристика на основе ключевых слов
	cost := 20.0
	ingredientsLower := strings.ToLower(ingredients)

	// Повышаем стоимость для премиум-ингредиентов
	premiumKeywords := []string{"łosoś", "tuńczyk", "krewetki", "awokado", "truffle"}
	for _, keyword := range premiumKeywords {
		if strings.Contains(ingredientsLower, keyword) {
			cost += 15.0
		}
	}

	margin := 80.0 // 80% маржа
	price := cost * (1 + margin/100)

	category := "mid"
	if price > 60 {
		category = "premium"
	} else if price < 30 {
		category = "budget"
	}

	return &PriceEstimation{
		EstimatedCost:  cost,
		SuggestedPrice: price,
		Margin:         margin,
		PriceCategory:  category,
		Explanation:    "Базова оцінка на основі інгредієнтів",
	}
}
