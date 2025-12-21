># Recipe AI Adaptation - Quick Reference

## 🎯 Концепция: AI = Adapter, NOT Generator

**Ключевое отличие:**
- ❌ **Старый подход**: AI создает рецепт с нуля → непредсказуемо, тратит токены
- ✅ **Новый подход**: AI адаптирует существующий рецепт → предсказуемо, эффективно

---

## 🔄 When to Use Adaptation

### Use Catalog Match (No AI):
```
User fridge: Makaron, Jajko, Boczek, Parmezan
Match: Carbonara - 100% match
→ Return catalog recipe as-is
```

### Use AI Adaptation:
```
User fridge: Makaron, Jajko, Kurczak (no Boczek, no Parmezan)
Match: Carbonara - 50% match
→ AI adapts: "Carbonara z kurczakiem" (substitute bacon → chicken)
```

---

## 📋 API Contract

### Endpoint
```http
POST /api/recipes/{id}/adapt
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json
```

### Request Body
```json
{
  "recipeId": "550e8400-e29b-41d4-a716-446655440000",
  "fridgeSnapshot": [
    {
      "ingredientId": "uuid",
      "name": "Makaron spaghetti",
      "quantity": 400,
      "unit": "g",
      "isExpiringSoon": false
    },
    {
      "ingredientId": "uuid",
      "name": "Jajko",
      "quantity": 4,
      "unit": "pcs",
      "isExpiringSoon": false
    },
    {
      "ingredientId": "uuid",
      "name": "Kurczak",
      "quantity": 200,
      "unit": "g",
      "isExpiringSoon": true
    }
  ],
  "missingIngredients": ["Boczek", "Parmezan"],
  "userPreferences": {
    "allowSubstitutions": true,
    "preferExpiring": true,
    "simplifySteps": false,
    "avoidAllergens": [],
    "reduceServings": null
  },
  "language": "pl"
}
```

### Response
```json
{
  "success": true,
  "data": {
    "originalRecipeId": "550e8400-e29b-41d4-a716-446655440000",
    "originalName": "Spaghetti Carbonara",
    "originalServings": 4,
    
    "adaptedName": "Spaghetti Carbonara z kurczakiem",
    "adaptedServings": 4,
    
    "adaptedIngredients": [
      {
        "originalName": "Makaron spaghetti",
        "substitutedWith": null,
        "quantity": 400,
        "unit": "g",
        "isAvailable": true
      },
      {
        "originalName": "Jajko",
        "substitutedWith": null,
        "quantity": 4,
        "unit": "pcs",
        "isAvailable": true
      },
      {
        "originalName": "Boczek",
        "substitutedWith": "Kurczak",
        "quantity": 200,
        "unit": "g",
        "isAvailable": true,
        "reason": "Kurczak jest dostępny i jest podobnym białkiem"
      },
      {
        "originalName": "Parmezan",
        "substitutedWith": null,
        "quantity": 0,
        "unit": "g",
        "isAvailable": false,
        "reason": "Opcjonalny składnik, można pominąć"
      }
    ],
    
    "adaptedSteps": [
      {
        "step": 1,
        "instruction": "Gotuj makaron spaghetti al dente w osolonej wodzie"
      },
      {
        "step": 2,
        "instruction": "Pokrój kurczaka w kostkę i podsmaż na patelni do złocistości"
      },
      {
        "step": 3,
        "instruction": "Rozbij jajka do miski i wymieszaj"
      },
      {
        "step": 4,
        "instruction": "Odcedź makaron, zachowaj szklankę wody"
      },
      {
        "step": 5,
        "instruction": "Wymieszaj gorący makaron z jajkami poza ogniem"
      },
      {
        "step": 6,
        "instruction": "Dodaj podsmażony kurczak, wymieszaj"
      },
      {
        "step": 7,
        "instruction": "Podawaj natychmiast z pieprzem"
      }
    ],
    
    "adaptations": [
      {
        "type": "substitution",
        "description": "Boczek zastąpiony kurczakiem - zachowano białko",
        "impact": "moderate"
      },
      {
        "type": "ingredient_removed",
        "description": "Pominięto parmezan - był opcjonalny",
        "impact": "minor"
      }
    ],
    
    "canCookNow": true,
    "difficultyChange": "same",
    "timeChange": 0,
    "adaptedAt": "2025-12-21T18:30:00Z"
  }
}
```

---

## 🧠 AI Prompt Strategy

### ✅ What AI CAN Do:

1. **Substitute Ingredients**
   ```
   Original: Boczek (bacon)
   Available: Kurczak (chicken)
   AI: Replace bacon with chicken, adjust cooking method
   ```

2. **Reduce Portions**
   ```
   Original: 4 servings
   Available: Only half ingredients
   AI: Scale recipe to 2 servings
   ```

3. **Simplify Steps**
   ```
   Original: "Make roux, add stock gradually, whisk continuously..."
   Simplified: "Mix flour with cold stock, heat while stirring"
   ```

4. **Remove Optional Ingredients**
   ```
   Original: Garnish with parsley
   Available: No parsley
   AI: Skip garnish, recipe still works
   ```

5. **Adjust Cooking Time**
   ```
   Original: Bake chicken 45 min
   Available: Chicken breast (not whole)
   AI: Reduce to 25 min
   ```

---

### ❌ What AI CANNOT Do:

1. **Change Dish Type**
   ```
   ❌ Carbonara → Ramen
   ❌ Pizza → Pasta
   ❌ Soup → Salad
   ```

2. **Invent New Recipe**
   ```
   ❌ "Create Asian fusion dish"
   ✅ "Adapt Carbonara with available ingredients"
   ```

3. **Add Ingredients NOT in Fridge**
   ```
   ❌ "Add saffron for better taste" (not in fridge)
   ✅ "Use available turmeric instead"
   ```

4. **Completely Rename**
   ```
   ❌ "Spaghetti Carbonara" → "Pasta Supreme"
   ✅ "Spaghetti Carbonara" → "Spaghetti Carbonara z kurczakiem"
   ```

---

## 🎨 Frontend Integration

### Step 1: User sees match result

```typescript
// User clicks on recipe card with 50% match
const recipe = {
  recipeId: "uuid",
  canonicalName: "Spaghetti Carbonara",
  score: 50,
  canCookNow: false,
  missingIngredients: [
    { name: "Boczek", quantity: 150, unit: "g" },
    { name: "Parmezan", quantity: 100, unit: "g" }
  ]
};
```

### Step 2: Show "Adapt" button

```typescript
<RecipeCard recipe={recipe}>
  {recipe.canCookNow ? (
    <Button color="green">🍳 Gotuj teraz</Button>
  ) : (
    <div>
      <Button color="yellow" onClick={() => addMissing(recipe)}>
        ➕ Dodaj {recipe.missingIngredients.length} składników
      </Button>
      <Button color="blue" onClick={() => adaptRecipe(recipe)}>
        🤖 Dostosuj przepis do lodówki
      </Button>
    </div>
  )}
</RecipeCard>
```

### Step 3: Call adapt endpoint

```typescript
const adaptRecipe = async (recipe: RecipeMatchItem) => {
  setLoading(true);
  
  try {
    const fridgeSnapshot = await fridgeApi.getAll();
    
    const response = await fetch(`/api/recipes/${recipe.recipeId}/adapt`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        recipeId: recipe.recipeId,
        fridgeSnapshot: fridgeSnapshot.map(item => ({
          ingredientId: item.ingredientId,
          name: item.name,
          quantity: item.quantity,
          unit: item.unit,
          isExpiringSoon: item.isExpiringSoon,
        })),
        missingIngredients: recipe.missingIngredients.map(i => i.name),
        userPreferences: {
          allowSubstitutions: true,
          preferExpiring: true,
          simplifySteps: false,
        },
        language: 'pl',
      }),
    });
    
    const data = await response.json();
    
    if (data.success) {
      showAdaptedRecipe(data.data);
    }
  } finally {
    setLoading(false);
  }
};
```

### Step 4: Show adaptation result

```typescript
const AdaptedRecipeModal = ({ adapted }: { adapted: AdaptedRecipeData }) => {
  return (
    <Modal>
      <ModalHeader>
        <Title>{adapted.adaptedName}</Title>
        <Badge color="blue">Dostosowany</Badge>
      </ModalHeader>
      
      <ModalBody>
        {/* Show what changed */}
        <Alert color="info">
          <Text>Dokonano {adapted.adaptations.length} zmian:</Text>
          <List>
            {adapted.adaptations.map((adapt, i) => (
              <ListItem key={i}>
                <Badge color={getImpactColor(adapt.impact)}>
                  {adapt.impact}
                </Badge>
                {adapt.description}
              </ListItem>
            ))}
          </List>
        </Alert>
        
        {/* Show ingredients with substitutions */}
        <Section>
          <Heading>Składniki</Heading>
          <List>
            {adapted.adaptedIngredients.map((ing, i) => (
              <ListItem key={i}>
                {ing.substitutedWith ? (
                  <div>
                    <Text strikethrough>{ing.originalName}</Text>
                    <Text color="green">→ {ing.substitutedWith}</Text>
                    <Text small>{ing.reason}</Text>
                  </div>
                ) : (
                  <Text>{ing.originalName}: {ing.quantity} {ing.unit}</Text>
                )}
              </ListItem>
            ))}
          </List>
        </Section>
        
        {/* Show adapted steps */}
        <Section>
          <Heading>Kroki przygotowania</Heading>
          <OrderedList>
            {adapted.adaptedSteps.map(step => (
              <ListItem key={step.step}>
                {step.instruction}
              </ListItem>
            ))}
          </OrderedList>
        </Section>
        
        {/* Difficulty & time changes */}
        <Flex gap={2}>
          {adapted.difficultyChange !== 'same' && (
            <Badge>
              Trudność: {adapted.difficultyChange === 'easier' ? '⬇️ Łatwiej' : '⬆️ Trudniej'}
            </Badge>
          )}
          {adapted.timeChange !== 0 && (
            <Badge>
              Czas: {adapted.timeChange > 0 ? '+' : ''}{adapted.timeChange} min
            </Badge>
          )}
        </Flex>
      </ModalBody>
      
      <ModalFooter>
        <Button color="green" onClick={() => startCooking(adapted)}>
          🍳 Gotuj teraz
        </Button>
        <Button color="gray" onClick={() => saveAdapted(adapted)}>
          💾 Zapisz przepis
        </Button>
      </ModalFooter>
    </Modal>
  );
};
```

---

## 🔍 Validation Rules

Backend validates AI output to ensure quality:

```go
func (s *RecipeAdapterService) ValidateAdaptation(
    original models.RecipeCatalog,
    adapted *dto.AdaptedRecipeData,
) error {
    // 1. Name similarity check
    if !strings.Contains(adapted.AdaptedName, extractMainDishName(original.CanonicalName)) {
        return errors.New("adapted name too different")
    }
    
    // 2. Reasonable portion size
    if adapted.AdaptedServings < 1 || adapted.AdaptedServings > original.Servings * 2 {
        return errors.New("invalid servings")
    }
    
    // 3. Has cooking steps
    if len(adapted.AdaptedSteps) == 0 {
        return errors.New("no steps")
    }
    
    return nil
}
```

---

## 📊 Success Metrics

Track adaptation quality:

```typescript
// Analytics
analytics.track('recipe_adapted', {
  originalRecipe: adapted.originalName,
  adaptedRecipe: adapted.adaptedName,
  adaptationsCount: adapted.adaptations.length,
  canCookNow: adapted.canCookNow,
  userStartedCooking: false, // Track later
});
```

**Key metrics:**
- Adaptation success rate (% valid responses)
- User acceptance rate (% users who cook adapted recipe)
- Average adaptations per recipe (2-3 ideal)
- AI response time (<2 seconds target)

---

## 🚀 Competitive Advantage

**Why this is killer feature:**

1. **Reduces Waste**: Use expiring ingredients
2. **Saves Money**: No need to buy missing items
3. **Personalization**: Adapts to YOUR fridge
4. **Predictable**: Based on real recipes, not AI hallucinations
5. **Educational**: Shows substitution reasoning

**vs Competitors:**
- ❌ Allrecipes, Epicurious: Static recipes
- ❌ ChatGPT: Generates from scratch (unreliable)
- ✅ FodiFOOD: Adapts real recipes to your fridge

---

## ✅ Implementation Checklist

### Backend
- [x] Create DTO (AdaptRecipeRequest/Response)
- [x] Create RecipeAdapterService
- [x] Create AI prompt template
- [x] Add validation logic
- [x] Create POST /api/recipes/:id/adapt endpoint
- [ ] Add error handling & retry logic
- [ ] Add adaptation caching (same fridge → same result)

### Frontend
- [ ] Add "Dostosuj przepis" button
- [ ] Create adaptRecipe() API call
- [ ] Create AdaptedRecipeModal component
- [ ] Show substitutions with reasoning
- [ ] Add "Save adapted recipe" feature
- [ ] Track analytics

### Testing
- [ ] Test with missing 1 ingredient
- [ ] Test with missing 50% ingredients
- [ ] Test with no allergen ingredients
- [ ] Test portion reduction
- [ ] Test step simplification
- [ ] Measure AI response quality

---

## 🎯 Next Steps

1. **Integrate Groq client** in adapter service
2. **Register routes** in main.go
3. **Test adaptation** with real recipes
4. **Frontend integration** (modal + button)
5. **A/B test**: Catalog vs Adapted vs Generated
