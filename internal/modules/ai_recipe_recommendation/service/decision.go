package service

// ============================================================================
// DECISION ENGINE - принимает финальное решение о рекомендациях
// Принцип: анализирует результаты matching и формирует ответ для пользователя
// ============================================================================

// DecisionEngine - сервис для принятия решений о рекомендациях
type DecisionEngine struct {
	matcher *RecipeMatcher
}

// NewDecisionEngine - конструктор
func NewDecisionEngine(matcher *RecipeMatcher) *DecisionEngine {
	return &DecisionEngine{matcher: matcher}
}

// MakeRecommendation - анализирует холодильник и выдает рекомендацию
func (d *DecisionEngine) MakeRecommendation(req RecipeMatchRequest) (*RecipeMatchResponse, error) {
	// Устанавливаем default значения
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Language == "" {
		req.Language = "pl"
	}

	// Получаем результаты matching от Rules Engine
	matches, err := d.matcher.MatchRecipes(req.UserID, req.Language, req.Limit)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return &RecipeMatchResponse{
			Decision:     DecisionNeedMore,
			Summary:      getLocalizedSummary(DecisionNeedMore, req.Language),
			TotalMatches: 0,
			Recipes:      []RecipeMatchResult{},
		}, nil
	}

	// Анализируем общую ситуацию
	decision := analyzeOverallDecision(matches)
	summary := getLocalizedSummary(decision, req.Language)

	return &RecipeMatchResponse{
		Decision:     decision,
		Summary:      summary,
		TotalMatches: len(matches),
		Recipes:      matches,
	}, nil
}

// analyzeOverallDecision - анализирует общую ситуацию по всем рецептам
func analyzeOverallDecision(matches []RecipeMatchResult) string {
	if len(matches) == 0 {
		return DecisionNeedMore
	}

	// Проверяем лучший результат (первый, т.к. отсортировано по match_percent)
	best := matches[0]

	switch best.MatchStatus {
	case StatusReady:
		return DecisionReady // 🟢 Есть рецепты которые можно готовить сейчас
	case StatusAlmostReady:
		return DecisionAlmostReady // 🟡 Почти готово (не хватает 1-2 ингредиента)
	default:
		return DecisionNeedMore // 🔴 Нужно больше продуктов
	}
}

// getLocalizedSummary - возвращает локализованное резюме
func getLocalizedSummary(decision string, lang string) string {
	summaries := map[string]map[string]string{
		DecisionReady: {
			"pl": "Świetnie! Możesz przygotować kilka przepisów już teraz.",
			"en": "Great! You can cook several recipes right now.",
			"ru": "Отлично! Вы можете приготовить несколько рецептов прямо сейчас.",
		},
		DecisionAlmostReady: {
			"pl": "Prawie gotowe! Brakuje tylko kilku składników.",
			"en": "Almost ready! You're just missing a few ingredients.",
			"ru": "Почти готово! Не хватает всего нескольких ингредиентов.",
		},
		DecisionNeedMore: {
			"pl": "Potrzebujesz więcej składników, aby przygotować te przepisy.",
			"en": "You need more ingredients to cook these recipes.",
			"ru": "Вам нужно больше продуктов, чтобы приготовить эти рецепты.",
		},
	}

	if langMap, ok := summaries[decision]; ok {
		if summary, ok := langMap[lang]; ok {
			return summary
		}
		// Fallback to English
		return langMap["en"]
	}

	return "No recipes found" // Fallback
}
