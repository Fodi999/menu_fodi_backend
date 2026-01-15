package service

import (
	"fmt"
	"strings"

	"github.com/dmitrijfomin/menu-fodifood/backend/internal/models"
	"github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/ai_core"
)

// AINotificationGenerator генератор AI-текстов для уведомлений
type AINotificationGenerator struct {
	groqClient *ai_core.GroqClient
}

func NewAINotificationGenerator() *AINotificationGenerator {
	return &AINotificationGenerator{
		groqClient: ai_core.NewGroqClient(),
	}
}

// GenerateExpiryMessage генерирует персонализированное сообщение о истечении срока
func (g *AINotificationGenerator) GenerateExpiryMessage(item *models.FridgeItem, ingredient *models.Ingredient, daysLeft int) (string, error) {
	ingredientName := ingredient.Name
	if ingredient.NamePL != nil && *ingredient.NamePL != "" {
		ingredientName = *ingredient.NamePL
	}

	// Формируем prompt в зависимости от срочности
	var systemPrompt string
	var userPrompt string

	switch {
	case daysLeft < 0:
		// Просрочено
		systemPrompt = `You are a helpful kitchen assistant. Generate a short, empathetic message about expired food.
Keep it under 100 characters. Mention the money loss. Language: Polish.`
		
		userPrompt = fmt.Sprintf(`Ingredient: "%s", quantity: %.1f %s
Status: EXPIRED (%d days ago)
Value lost: %.2f PLN

Generate a short notification message.`, ingredientName, item.Quantity, item.Unit, -daysLeft, item.PriceTotal)

	case daysLeft == 0:
		// Истекает сегодня
		systemPrompt = `You are a motivating kitchen assistant. Generate an urgent but friendly message to use food TODAY.
Keep it under 100 characters. Emphasize the urgency. Language: Polish.`
		
		userPrompt = fmt.Sprintf(`Ingredient: "%s", quantity: %.1f %s
Status: EXPIRES TODAY
Value: %.2f PLN

Generate an urgent message encouraging to cook it today.`, ingredientName, item.Quantity, item.Unit, item.PriceTotal)

	case daysLeft == 1:
		// Истекает завтра
		systemPrompt = `You are a friendly kitchen assistant. Generate a warning message to use food TOMORROW.
Keep it under 120 characters. Be motivating. Language: Polish.`
		
		userPrompt = fmt.Sprintf(`Ingredient: "%s", quantity: %.1f %s
Status: expires in 1 day
Value: %.2f PLN

Generate a motivating message to use it tomorrow.`, ingredientName, item.Quantity, item.Unit, item.PriceTotal)

	default:
		// 2-3 дня
		systemPrompt = `You are a helpful kitchen assistant. Generate a friendly reminder about food expiring soon.
Keep it under 100 characters. Be casual and helpful. Language: Polish.`
		
		userPrompt = fmt.Sprintf(`Ingredient: "%s", quantity: %.1f %s
Expires in: %d days
Value: %.2f PLN

Generate a friendly reminder message.`, ingredientName, item.Quantity, item.Unit, daysLeft, item.PriceTotal)
	}

	// Вызываем AI
	response, err := g.groqClient.SimpleChat(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	// Очищаем ответ
	message := strings.TrimSpace(response)
	message = strings.Trim(message, `"`)

	return message, nil
}

// GenerateRecipeSuggestion генерирует рекомендацию рецепта для использования продукта
func (g *AINotificationGenerator) GenerateRecipeSuggestion(item *models.FridgeItem, ingredient *models.Ingredient) (string, error) {
	ingredientName := ingredient.Name
	if ingredient.NamePL != nil && *ingredient.NamePL != "" {
		ingredientName = *ingredient.NamePL
	}

	systemPrompt := `You are a creative chef. Suggest a simple recipe idea using the given ingredient.
Keep it under 80 characters. Be inspiring. Language: Polish.`

	userPrompt := fmt.Sprintf(`Ingredient: "%s" (%.1f %s)
Category: %s

Suggest a quick recipe idea.`, ingredientName, item.Quantity, item.Unit, ingredient.Category)

	response, err := g.groqClient.SimpleChat(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("AI recipe suggestion failed: %w", err)
	}

	suggestion := strings.TrimSpace(response)
	suggestion = strings.Trim(suggestion, `"`)

	return suggestion, nil
}
