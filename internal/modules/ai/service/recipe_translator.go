package service

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RecipeTranslation represents translated recipe fields
type RecipeTranslation struct {
	NamePL        string   `json:"name_pl"`
	NameEN        string   `json:"name_en"`
	NameRU        string   `json:"name_ru"`
	DescriptionPL string   `json:"description_pl"`
	DescriptionEN string   `json:"description_en"`
	DescriptionRU string   `json:"description_ru"`
	StepsPL       []string `json:"steps_pl"`
	StepsEN       []string `json:"steps_en"`
	StepsRU       []string `json:"steps_ru"`
}

// TranslateRecipe translates recipe name, description and steps to PL/EN/RU
func (s *aiService) TranslateRecipe(name, description string, steps []string) (*RecipeTranslation, error) {
	// Build steps text
	stepsText := ""
	if len(steps) > 0 {
		stepsText = strings.Join(steps, "\n")
	}

	// Create AI prompt for translation
	prompt := fmt.Sprintf(`You are a professional culinary translator. Translate the following recipe to Polish (PL), English (EN), and Russian (RU).

RECIPE NAME:
%s

DESCRIPTION:
%s

STEPS:
%s

Return ONLY valid JSON in this exact format (no markdown, no comments):
{
  "name_pl": "translated name in Polish",
  "name_en": "translated name in English", 
  "name_ru": "translated name in Russian",
  "description_pl": "translated description in Polish",
  "description_en": "translated description in English",
  "description_ru": "translated description in Russian",
  "steps_pl": ["step 1 in Polish", "step 2 in Polish"],
  "steps_en": ["step 1 in English", "step 2 in English"],
  "steps_ru": ["step 1 in Russian", "step 2 in Russian"]
}

CRITICAL RULES:
1. Return ONLY valid JSON (no markdown blocks)
2. Keep culinary terminology accurate
3. Preserve all cooking details
4. Maintain professional tone
5. If original is empty, return empty string
6. Number of steps must match in all languages

Return the JSON now:`, name, description, stepsText)

	// Call AI
	response, err := s.groqClient.SimpleChat("", prompt)
	if err != nil {
		return nil, fmt.Errorf("AI translation failed: %w", err)
	}

	// Parse response
	parsedJSON, isJSON, parseErr := parseAIResponse(response, "translate_recipe")
	if !isJSON || parseErr != nil {
		// Try self-repair
		repairPrompt := fmt.Sprintf(`The following JSON is invalid. Fix it and return ONLY valid JSON:

%s

Return ONLY the fixed JSON (no markdown, no explanations):`, response)

		repairedResponse, repairErr := s.groqClient.SimpleChat("", repairPrompt)
		if repairErr != nil {
			return nil, fmt.Errorf("AI repair failed: %w", repairErr)
		}

		parsedJSON, isJSON, parseErr = parseAIResponse(repairedResponse, "translate_recipe")
		if !isJSON || parseErr != nil {
			return nil, fmt.Errorf("failed to parse translation: %w", parseErr)
		}
	}

	// Decode into RecipeTranslation
	var translation RecipeTranslation
	if err := json.Unmarshal([]byte(parsedJSON), &translation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal translation: %w", err)
	}

	return &translation, nil
}

// TranslateRecipeField translates a single recipe field to all 3 languages
func (s *aiService) TranslateRecipeField(fieldType, text, sourceLang string) (pl, en, ru string, err error) {
	if text == "" {
		return "", "", "", nil
	}

	prompt := fmt.Sprintf(`Translate this %s from %s to Polish (PL), English (EN), and Russian (RU).

TEXT:
%s

Return ONLY valid JSON:
{
  "pl": "translation in Polish",
  "en": "translation in English",
  "ru": "translation in Russian"
}`, fieldType, sourceLang, text)

	response, err := s.groqClient.SimpleChat("", prompt)
	if err != nil {
		return "", "", "", fmt.Errorf("AI translation failed: %w", err)
	}

	parsedJSON, isJSON, parseErr := parseAIResponse(response, "translate_field")
	if !isJSON || parseErr != nil {
		return "", "", "", fmt.Errorf("failed to parse translation: %w", parseErr)
	}

	var result struct {
		PL string `json:"pl"`
		EN string `json:"en"`
		RU string `json:"ru"`
	}

	if err := json.Unmarshal([]byte(parsedJSON), &result); err != nil {
		return "", "", "", fmt.Errorf("failed to unmarshal translation: %w", err)
	}

	return result.PL, result.EN, result.RU, nil
}
