package ai

import (
	"log"
	"strings"
)

// MentorChat AI-наставник для кулинарных вопросов
type MentorChat struct {
	client          *GroqClient
	lang            string
	conversationLog []GroqMessage // история диалога
}

// NewMentorChat создаёт нового AI наставника
func NewMentorChat(lang string) *MentorChat {
	if lang == "" {
		lang = "pl"
	}
	return &MentorChat{
		client:          NewGroqClient(),
		lang:            lang,
		conversationLog: []GroqMessage{},
	}
}

// Ask задать вопрос наставнику
func (mc *MentorChat) Ask(question string) (string, error) {
	log.Printf("[MENTOR] 💬 Question (%s): %s", mc.lang, truncate(question, 50))

	// Добавляем system prompt если это первый вопрос
	if len(mc.conversationLog) == 0 {
		mc.conversationLog = append(mc.conversationLog, GroqMessage{
			Role:    "system",
			Content: mc.getSystemPrompt(),
		})
	}

	// Добавляем вопрос пользователя
	mc.conversationLog = append(mc.conversationLog, GroqMessage{
		Role:    "user",
		Content: question,
	})

	// Отправляем в Groq
	resp, err := mc.client.Chat(mc.conversationLog, 0.8, 512)
	if err != nil {
		log.Printf("[MENTOR] ❌ Groq API error: %v", err)
		return mc.getFallbackAnswer(question), nil
	}

	if len(resp.Choices) == 0 {
		return mc.getFallbackAnswer(question), nil
	}

	answer := resp.Choices[0].Message.Content

	// Добавляем ответ в историю
	mc.conversationLog = append(mc.conversationLog, GroqMessage{
		Role:    "assistant",
		Content: answer,
	})

	log.Printf("[MENTOR] ✅ Answer: %s", truncate(answer, 100))
	return answer, nil
}

// ClearHistory очистить историю диалога
func (mc *MentorChat) ClearHistory() {
	mc.conversationLog = []GroqMessage{}
	log.Printf("[MENTOR] 🗑️ Conversation history cleared")
}

// getSystemPrompt возвращает system prompt для наставника
func (mc *MentorChat) getSystemPrompt() string {
	prompts := map[string]string{
		"pl": `Jesteś doświadczonym szefem kuchni i mentorem kulinarnym.
Pomagasz uczniom w nauce gotowania, odpowiadasz na pytania o techniki, składniki i receptury.
Bądź przyjazny, profesjonalny i konkretny. Używaj prostego języka.
Zawsze dawaj praktyczne porady. Jeśli nie znasz odpowiedzi, bądź szczery.
Twoja specjalizacja: kuchnia japońska i polska.`,

		"ua": `Ти досвідчений шеф-кухар і кулінарний наставник.
Допомагаєш учням вивчати кулінарію, відповідаєш на питання про техніки, інгредієнти та рецепти.
Будь дружелюбним, професійним і конкретним. Використовуй просту мову.
Завжди давай практичні поради. Якщо не знаєш відповіді, будь чесним.
Твоя спеціалізація: японська та українська кухня.`,

		"en": `You are an experienced chef and culinary mentor.
Help students learn cooking, answer questions about techniques, ingredients, and recipes.
Be friendly, professional, and specific. Use simple language.
Always give practical advice. If you don't know the answer, be honest.
Your specialization: Japanese and Polish cuisine.`,
	}

	if prompt, ok := prompts[mc.lang]; ok {
		return prompt
	}
	return prompts["pl"]
}

// getFallbackAnswer возвращает базовый ответ если AI недоступен
func (mc *MentorChat) getFallbackAnswer(question string) string {
	questionLower := strings.ToLower(question)

	// Простой pattern matching для популярных вопросов
	fallbacks := map[string]map[string]string{
		"pl": {
			"ryż":      "Używaj proporcji 1:1.2 (ryż:woda). Gotuj 15 min, potem odstaw na 10 min.",
			"sushi":    "Kluczem do dobrych sushi jest świeży ryż i dobrej jakości ryba. Ćwicz formowanie.",
			"nóż":      "Trzymaj nóż pewnie, używaj ruchu kołysania. Ostrz nóż regularnie.",
			"default":  "To ciekawe pytanie! Najlepiej skonsultuj się z doświadczonym szefem lub sprawdź w specjalistycznych źródłach.",
		},
		"ua": {
			"рис":      "Використовуй пропорцію 1:1.2 (рис:вода). Вари 15 хв, потім відстав на 10 хв.",
			"суші":     "Ключ до хороших суші - свіжий рис і якісна риба. Практикуй формування.",
			"ніж":      "Тримай ніж впевнено, використовуй рух гойдання. Гостри ніж регулярно.",
			"default":  "Це цікаве питання! Найкраще проконсультуйся з досвідченим шефом або перевір у спеціалізованих джерелах.",
		},
		"en": {
			"rice":     "Use 1:1.2 ratio (rice:water). Cook 15 min, then let stand 10 min.",
			"sushi":    "The key to good sushi is fresh rice and quality fish. Practice shaping.",
			"knife":    "Hold knife firmly, use rocking motion. Sharpen regularly.",
			"default":  "That's an interesting question! Best to consult with an experienced chef or check specialized sources.",
		},
	}

	langFallbacks := fallbacks[mc.lang]
	if langFallbacks == nil {
		langFallbacks = fallbacks["pl"]
	}

	// Проверяем ключевые слова
	for keyword, answer := range langFallbacks {
		if keyword != "default" && strings.Contains(questionLower, keyword) {
			return answer
		}
	}

	return langFallbacks["default"]
}

// truncate обрезает строку для логов
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
