# Recipe Card UI - API Response Format
**Date:** 2025-12-21  
**Endpoint:** `POST /api/recipes/recommendations`

## Эталонная карточка рецепта (UI Target)

```
Pizza Margherita

⏱ 120 min   👥 2 porcje   🌍 Italy   medium

Dopasowanie: 50% (2 z 4 składników)

Z lodówki (2):
• Pomidor – 150 g
• Oliwa z oliwek – 20 ml

Do dokupienia (2):
• Mąka pszenna – 300 g
• Mozzarella – 200 g

Ekonomia:
• Wartość z lodówki: 1.20 PLN
• Szacunkowy koszt zakupów: ~8.50 PLN

[🛒 Dodaj do listy zakupów]
[⭐ Zapisz przepis]
[🔄 Pokaż inny przepis]
```

---

## API Response Format

### Request
```typescript
POST /api/recipes/recommendations?testUserID={userId}

{
  "mode": "fridge",
  "limit": 10,
  "excludeRecipeIds": []  // Optional: exclude previously shown recipes
}
```

### Response (Success)
```typescript
{
  "success": true,
  "data": {
    "recipe": {
      "id": "uuid",
      "canonicalName": "Pizza Margherita",
      "localName": "Pizza Margherita",
      "country": "Italy",
      "category": "main",
      "difficulty": "medium",        // ✅ easy | medium | hard
      "timeMinutes": 120,            // ✅ для "⏱ 120 min"
      "servings": 2,                 // ✅ для "👥 2 porcje"
      "steps": [
        "1. Zagnieć ciasto z mąki",
        "2. Rozwałkuj ciasto",
        "3. Posmaruj sosem",
        "4. Dodaj ser i pomidory",
        "5. Piecz 15 minut"
      ],
      "allergens": ["Gluten", "Mleko"],
      "dietTags": ["Vegetarian"]
    },
    "match": {
      "canCookNow": false,           // ✅ false = trzeba dokupić
      "missingRequired": [           // ✅ "Do dokupienia (2)"
        {
          "ingredientId": "uuid1",
          "name": "Mąka pszenna",
          "quantity": 300,
          "unit": "g",
          "estimatedCost": 5.50      // ✅ dla "Szacunkowy koszt"
        },
        {
          "ingredientId": "uuid2",
          "name": "Mozzarella",
          "quantity": 200,
          "unit": "g",
          "estimatedCost": 3.00
        }
      ],
      "usedIngredients": [           // ✅ "Z lodówki (2)"
        {
          "ingredientId": "uuid3",
          "name": "Pomidor",
          "quantity": 150,           // ✅ Potrzeba w przepisie
          "unit": "g",
          "available": 600,          // ✅ Jest w lodówce
          "isExpiringSoon": false
        },
        {
          "ingredientId": "uuid4",
          "name": "Oliwa z oliwek",
          "quantity": 20,
          "unit": "ml",
          "available": 250,
          "isExpiringSoon": true     // ✅ Można pokazać ostrzeżenie
        }
      ]
    },
    "economy": {
      "usedFromFridge": 1.20,        // ✅ "Wartość z lodówki: 1.20 PLN"
      "saved": 1.20                  // ✅ Same value (money saved)
    }
  }
}
```

### Response (No Recipes Available)
```typescript
{
  "success": true,
  "data": null,
  "message": "Brak dostępnych przepisów. Dodaj produkty do lodówki lub sprawdź później."
}
```

### Response (Error)
```typescript
{
  "success": false,
  "error": "Invalid request format",
  "message": "Mode must be 'fridge'"
}
```

---

## Frontend Mapping Guide

### 1. Header Section
```typescript
// Nazwa przepisu
recipe.localName  // "Pizza Margherita"

// Czas
`⏱ ${recipe.timeMinutes} min`  // "⏱ 120 min"

// Porcje
`👥 ${recipe.servings} ${recipe.servings === 1 ? 'porcja' : 'porcje'}`  // "👥 2 porcje"

// Kraj
`🌍 ${recipe.country}`  // "🌍 Italy"

// Trudność
recipe.difficulty  // "medium" → "średni" (translate)
```

### 2. Matching Score (Dopasowanie)
```typescript
const total = match.usedIngredients.length + match.missingRequired.length;
const available = match.usedIngredients.length;
const percentage = total > 0 ? Math.round((available / total) * 100) : 0;

// Display
`Dopasowanie: ${percentage}% (${available} z ${total} składników)`
// "Dopasowanie: 50% (2 z 4 składników)"
```

### 3. Ingredients from Fridge (Z lodówki)
```typescript
{match.usedIngredients.length > 0 && (
  <>
    <h4>Z lodówki ({match.usedIngredients.length}):</h4>
    <ul>
      {match.usedIngredients.map(ing => (
        <li key={ing.ingredientId}>
          {ing.name} – {ing.quantity} {ing.unit}
          {ing.isExpiringSoon && <span>⚠️ Wkrótce traci ważność</span>}
        </li>
      ))}
    </ul>
  </>
)}
```

### 4. Missing Ingredients (Do dokupienia)
```typescript
{match.missingRequired.length > 0 && (
  <>
    <h4>Do dokupienia ({match.missingRequired.length}):</h4>
    <ul>
      {match.missingRequired.map(ing => (
        <li key={ing.ingredientId}>
          {ing.name} – {ing.quantity} {ing.unit}
          <span className="cost">~{ing.estimatedCost.toFixed(2)} PLN</span>
        </li>
      ))}
    </ul>
  </>
)}
```

### 5. Economy Section (Ekonomia)
```typescript
const totalShoppingCost = match.missingRequired.reduce(
  (sum, ing) => sum + ing.estimatedCost, 
  0
);

<div className="economy">
  <p>• Wartość z lodówki: {economy.usedFromFridge.toFixed(2)} PLN</p>
  <p>• Szacunkowy koszt zakupów: ~{totalShoppingCost.toFixed(2)} PLN</p>
</div>
```

### 6. Action Buttons
```typescript
// 1. Add to Shopping List
<button onClick={() => addMissingToShoppingList(match.missingRequired)}>
  🛒 Dodaj do listy zakupów
</button>

// 2. Save Recipe
<button onClick={() => saveRecipe(recipe.id)}>
  ⭐ Zapisz przepis
</button>

// 3. Show Another Recipe (Exclude current)
<button onClick={() => getNextRecommendation([recipe.id])}>
  🔄 Pokaż inny przepis
</button>

async function getNextRecommendation(excludeIds: string[]) {
  const response = await fetch('/api/recipes/recommendations', {
    method: 'POST',
    body: JSON.stringify({
      mode: 'fridge',
      limit: 10,
      excludeRecipeIds: excludeIds  // ✅ Exclude already shown
    })
  });
  // Update UI with new recipe
}
```

---

## TypeScript Interfaces

```typescript
// Request
interface RecommendationRequest {
  mode: 'fridge';
  limit?: number;
  excludeRecipeIds?: string[];
}

// Response
interface RecommendationResponse {
  success: boolean;
  data?: {
    recipe: RecipeInfo;
    match: MatchInfo;
    economy: EconomyInfo;
  };
  message?: string;
  error?: string;
}

interface RecipeInfo {
  id: string;
  canonicalName: string;
  localName: string;
  country: string;
  category: string;
  difficulty: 'easy' | 'medium' | 'hard';
  timeMinutes: number;
  servings: number;
  steps: string[];
  allergens?: string[];
  dietTags?: string[];
}

interface MatchInfo {
  canCookNow: boolean;
  missingRequired: MissingIngredient[];
  usedIngredients: UsedIngredient[];
}

interface MissingIngredient {
  ingredientId: string;
  name: string;
  quantity: number;
  unit: string;
  estimatedCost: number;  // PLN
}

interface UsedIngredient {
  ingredientId: string;
  name: string;
  quantity: number;
  unit: string;
  available: number;
  isExpiringSoon: boolean;
}

interface EconomyInfo {
  usedFromFridge: number;  // PLN
  saved: number;           // PLN (same as usedFromFridge)
}
```

---

## Translation Map (EN → PL)

```typescript
const difficultyMap: Record<string, string> = {
  'easy': 'łatwy',
  'medium': 'średni',
  'hard': 'trudny'
};

const categoryMap: Record<string, string> = {
  'appetizer': 'Przystawka',
  'soup': 'Zupa',
  'salad': 'Sałatka',
  'main': 'Danie główne',
  'side': 'Dodatek',
  'dessert': 'Deser',
  'beverage': 'Napój'
};

const countryMap: Record<string, string> = {
  'Poland': 'Polska',
  'Italy': 'Włochy',
  'France': 'Francja',
  'Greece': 'Grecja',
  // Add more as needed
};
```

---

## Complete Card Component Example

```typescript
import React from 'react';
import { RecommendationResponse } from './types';

interface RecipeCardProps {
  data: RecommendationResponse['data'];
  onAddToShoppingList: (ingredients: MissingIngredient[]) => void;
  onSaveRecipe: (recipeId: string) => void;
  onShowNext: (excludeIds: string[]) => void;
}

export const RecipeCard: React.FC<RecipeCardProps> = ({
  data,
  onAddToShoppingList,
  onSaveRecipe,
  onShowNext
}) => {
  if (!data) {
    return <div>Brak dostępnych przepisów</div>;
  }

  const { recipe, match, economy } = data;
  
  // Calculate matching percentage
  const total = match.usedIngredients.length + match.missingRequired.length;
  const available = match.usedIngredients.length;
  const matchPercentage = total > 0 ? Math.round((available / total) * 100) : 0;
  
  // Calculate total shopping cost
  const shoppingCost = match.missingRequired.reduce(
    (sum, ing) => sum + ing.estimatedCost, 
    0
  );

  return (
    <div className="recipe-card">
      {/* Header */}
      <h2>{recipe.localName}</h2>
      <div className="meta">
        <span>⏱ {recipe.timeMinutes} min</span>
        <span>👥 {recipe.servings} {recipe.servings === 1 ? 'porcja' : 'porcje'}</span>
        <span>🌍 {recipe.country}</span>
        <span>{difficultyMap[recipe.difficulty]}</span>
      </div>

      {/* Matching Score */}
      <div className="matching">
        Dopasowanie: {matchPercentage}% ({available} z {total} składników)
      </div>

      {/* Used from Fridge */}
      {match.usedIngredients.length > 0 && (
        <div className="section">
          <h4>Z lodówki ({match.usedIngredients.length}):</h4>
          <ul>
            {match.usedIngredients.map(ing => (
              <li key={ing.ingredientId}>
                • {ing.name} – {ing.quantity} {ing.unit}
                {ing.isExpiringSoon && (
                  <span className="warning"> ⚠️ Wkrótce traci ważność</span>
                )}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Missing Ingredients */}
      {match.missingRequired.length > 0 && (
        <div className="section">
          <h4>Do dokupienia ({match.missingRequired.length}):</h4>
          <ul>
            {match.missingRequired.map(ing => (
              <li key={ing.ingredientId}>
                • {ing.name} – {ing.quantity} {ing.unit}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Economy */}
      <div className="economy">
        <h4>Ekonomia:</h4>
        <p>• Wartość z lodówki: {economy.usedFromFridge.toFixed(2)} PLN</p>
        {shoppingCost > 0 && (
          <p>• Szacunkowy koszt zakupów: ~{shoppingCost.toFixed(2)} PLN</p>
        )}
      </div>

      {/* Actions */}
      <div className="actions">
        {match.missingRequired.length > 0 && (
          <button onClick={() => onAddToShoppingList(match.missingRequired)}>
            🛒 Dodaj do listy zakupów
          </button>
        )}
        
        <button onClick={() => onSaveRecipe(recipe.id)}>
          ⭐ Zapisz przepis
        </button>
        
        <button onClick={() => onShowNext([recipe.id])}>
          🔄 Pokaż inny przepis
        </button>
      </div>
    </div>
  );
};
```

---

## API Status Checks

### Check 1: Recipe Count
```bash
curl "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/match?testUserID=407582be-59d5-4d21-873b-1a72d31b0d42&limit=100" | jq '.data.count'

Expected: >= 10
```

### Check 2: Get Recommendation
```bash
curl -X POST "https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/recipes/recommendations?testUserID=407582be-59d5-4d21-873b-1a72d31b0d42" \
  -H "Content-Type: application/json" \
  -d '{"mode": "fridge", "limit": 10}' | jq '.'

Expected: { "success": true, "data": {...} }
```

### Check 3: Sequential Recommendations
```bash
# Get first
RECIPE1=$(curl -s -X POST "..." -d '{"mode": "fridge"}' | jq -r '.data.recipe.id')

# Get second (exclude first)
curl -X POST "..." -d "{\"mode\": \"fridge\", \"excludeRecipeIds\": [\"$RECIPE1\"]}" | jq '.data.recipe.localName'

Expected: Different recipe name
```

---

## Summary

✅ **API Already Provides All Needed Data**
- Recipe metadata (name, time, servings, difficulty, country)
- Matching score (usedIngredients, missingRequired)
- Economy calculations (usedFromFridge, estimatedCost per ingredient)
- Exclude functionality for "next recipe" button

✅ **Frontend Just Needs to Map**
- Calculate percentage: `(available / total) * 100`
- Translate difficulty/category/country
- Sum up shopping costs from `missingRequired[].estimatedCost`
- Show expiring soon warnings from `isExpiringSoon`

✅ **User Flow**
1. User clicks "Analizuj lodówkę" → GET recommendation
2. Card displays with all details
3. User clicks "Pokaż inny" → GET recommendation with excludeRecipeIds
4. Repeat for browsing multiple options

🎯 **All data is ready. Frontend can build the card exactly as designed.**
