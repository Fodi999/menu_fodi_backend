# 🌍 Multilingual Conflict Resolution - Frontend Integration

## Overview

When a recipe name conflict is detected (409 response), backend now returns **multilingual suggestions** in Russian, English, and Polish simultaneously.

---

## API Response Format

### Before (Single Language)
```json
{
  "success": false,
  "code": "RECIPE_NAME_EXISTS",
  "message": "Рецепт с таким названием уже существует",
  "suggestions": ["Title 1", "Title 2", "Title 3", "Title 4", "Title 5"]
}
```

### After (Multilingual)
```json
{
  "success": false,
  "code": "RECIPE_NAME_EXISTS",
  "message": "Рецепт с таким названием уже существует",
  "conflict": {
    "canonicalName": "жареный_лосось",
    "originalTitle": "жареный лосось"
  },
  "suggestions": {
    "ru": [
      "Жареный Лосось с Хрустящей Кожей",
      "Домашний Жареный Лосось с Лимоном",
      "Лосось на Сковороде с Травами",
      "Румяный Жареный Лосось с Чесноком",
      "Лосось Жареный на Гриле с Соусом"
    ],
    "en": [
      "Pan-Fried Salmon with Crispy Skin",
      "Homestyle Fried Salmon with Lemon",
      "Skillet Salmon with Herbs",
      "Golden Pan-Seared Salmon with Garlic",
      "Grilled Salmon with Special Sauce"
    ],
    "pl": [
      "Smażony Łosoś z Chrupiącą Skórką",
      "Domowy Smażony Łosoś z Cytryną",
      "Łosoś na Patelni z Ziołami",
      "Złoty Smażony Łosoś z Czosnkiem",
      "Łosoś Grillowany z Specjalnym Sosem"
    ]
  }
}
```

---

## TypeScript Types

```typescript
// Response type
interface RecipeConflictResponse {
  success: false;
  code: 'RECIPE_NAME_EXISTS';
  message: string;
  conflict: {
    canonicalName: string;  // URL-safe name: "жареный_лосось"
    originalTitle: string;  // User's input: "жареный лосось"
  };
  suggestions: {
    ru: string[];  // 5 Russian alternatives
    en: string[];  // 5 English alternatives
    pl: string[];  // 5 Polish alternatives
  };
}

// Language codes
type SupportedLanguage = 'ru' | 'en' | 'pl';

// Suggestion map
type MultilingualSuggestions = Record<SupportedLanguage, string[]>;
```

---

## Frontend Implementation Examples

### React Component (Language Tabs)

```tsx
import { useState } from 'react';

interface ConflictDialogProps {
  conflict: RecipeConflictResponse;
  onSelect: (title: string, lang: string) => void;
  onCancel: () => void;
}

export function RecipeConflictDialog({ conflict, onSelect, onCancel }: ConflictDialogProps) {
  const [activeTab, setActiveTab] = useState<'ru' | 'en' | 'pl'>('ru');
  
  const languageLabels = {
    ru: '🇷🇺 Русский',
    en: '🇬🇧 English',
    pl: '🇵🇱 Polski'
  };

  return (
    <div className="conflict-dialog">
      <h2>Recipe Name Already Exists</h2>
      <p className="conflict-info">
        "{conflict.conflict.originalTitle}" is already used.
        Choose an alternative:
      </p>

      {/* Language Tabs */}
      <div className="language-tabs">
        {(['ru', 'en', 'pl'] as const).map((lang) => (
          <button
            key={lang}
            className={activeTab === lang ? 'active' : ''}
            onClick={() => setActiveTab(lang)}
          >
            {languageLabels[lang]}
          </button>
        ))}
      </div>

      {/* Suggestions List */}
      <div className="suggestions-list">
        {conflict.suggestions[activeTab].map((title, index) => (
          <button
            key={index}
            className="suggestion-item"
            onClick={() => onSelect(title, activeTab)}
          >
            {title}
          </button>
        ))}
      </div>

      <button onClick={onCancel}>Cancel</button>
    </div>
  );
}
```

### Vue Component (Dropdown)

```vue
<template>
  <div class="conflict-dialog">
    <h2>Название рецепта уже существует</h2>
    <p>
      "{{ conflict.conflict.originalTitle }}" уже используется.
      Выберите альтернативу:
    </p>

    <!-- Language Selector -->
    <select v-model="selectedLanguage" class="language-selector">
      <option value="ru">🇷🇺 Русский</option>
      <option value="en">🇬🇧 English</option>
      <option value="pl">🇵🇱 Polski</option>
    </select>

    <!-- Suggestions -->
    <div class="suggestions">
      <button
        v-for="(title, index) in currentSuggestions"
        :key="index"
        @click="selectTitle(title)"
        class="suggestion-btn"
      >
        {{ title }}
      </button>
    </div>

    <button @click="$emit('cancel')">Отмена</button>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const props = defineProps<{
  conflict: RecipeConflictResponse;
}>();

const emit = defineEmits<{
  select: [title: string, lang: string];
  cancel: [];
}>();

const selectedLanguage = ref<'ru' | 'en' | 'pl'>('ru');

const currentSuggestions = computed(() => 
  props.conflict.suggestions[selectedLanguage.value]
);

const selectTitle = (title: string) => {
  emit('select', title, selectedLanguage.value);
};
</script>
```

### React Native (All Languages Visible)

```tsx
import { View, Text, ScrollView, TouchableOpacity } from 'react-native';

export function ConflictScreen({ conflict, onSelect, onCancel }) {
  const languages = [
    { code: 'ru', flag: '🇷🇺', name: 'Русский' },
    { code: 'en', flag: '🇬🇧', name: 'English' },
    { code: 'pl', flag: '🇵🇱', name: 'Polski' },
  ];

  return (
    <ScrollView style={styles.container}>
      <Text style={styles.title}>Recipe Name Already Exists</Text>
      <Text style={styles.subtitle}>
        "{conflict.conflict.originalTitle}" is already used.
      </Text>

      {languages.map((lang) => (
        <View key={lang.code} style={styles.section}>
          <Text style={styles.langHeader}>
            {lang.flag} {lang.name}
          </Text>
          {conflict.suggestions[lang.code].map((title, index) => (
            <TouchableOpacity
              key={index}
              style={styles.suggestionBtn}
              onPress={() => onSelect(title, lang.code)}
            >
              <Text style={styles.suggestionText}>{title}</Text>
            </TouchableOpacity>
          ))}
        </View>
      ))}

      <TouchableOpacity style={styles.cancelBtn} onPress={onCancel}>
        <Text>Cancel</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}
```

---

## API Call Handler

```typescript
async function saveRecipe(recipeData: RecipeData): Promise<SaveResult> {
  try {
    const response = await fetch('/api/admin/recipes/save', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`,
      },
      body: JSON.stringify(recipeData),
    });

    const data = await response.json();

    // Handle conflict (409)
    if (response.status === 409 && data.code === 'RECIPE_NAME_EXISTS') {
      // Show multilingual suggestions dialog
      const selectedTitle = await showConflictDialog(data);
      
      // Retry with selected title
      return saveRecipe({
        ...recipeData,
        title: selectedTitle,
      });
    }

    // Handle success (201)
    if (response.status === 201) {
      return { success: true, recipe: data.data };
    }

    // Handle other errors
    throw new Error(data.message);
  } catch (error) {
    console.error('Save recipe failed:', error);
    throw error;
  }
}
```

---

## UX Recommendations

### 1. **Default Language**
Show suggestions in user's current language first:
```typescript
const userLang = i18n.language; // 'ru' | 'en' | 'pl'
const [activeTab, setActiveTab] = useState(userLang);
```

### 2. **Language Tabs** (Recommended for Desktop)
- Most intuitive for users
- Keeps UI clean (only 5 suggestions visible at once)
- Easy to switch between languages

### 3. **Accordion** (Recommended for Mobile)
- Shows all languages collapsed initially
- User expands language of interest
- Good for limited screen space

### 4. **All Visible** (Admin Panels)
- Show all 15 suggestions (3 languages × 5 each)
- Group by language with headers
- Good when screen space available

### 5. **AI Badge**
Add badge to show suggestions are AI-generated:
```tsx
<span className="ai-badge">✨ AI-Generated</span>
```

---

## Handling Edge Cases

### 1. Missing Language

If a language is missing from response:
```typescript
const getSuggestions = (lang: string) => {
  return conflict.suggestions[lang] || [];
};

// Fallback to another language
const suggestions = getSuggestions(userLang) || 
                   getSuggestions('en') || 
                   [];
```

### 2. Empty Suggestions

```typescript
if (!conflict.suggestions || 
    Object.keys(conflict.suggestions).length === 0) {
  // Show manual input instead
  return <ManualTitleInput />;
}
```

### 3. Network Error

```typescript
try {
  const response = await fetch(...);
  // ...
} catch (error) {
  // Show error dialog with retry option
  showErrorDialog({
    message: 'Failed to get suggestions. Please try again.',
    retry: () => saveRecipe(recipeData),
  });
}
```

---

## Backend Guarantees

✅ **Always returns 3 languages:** `ru`, `en`, `pl`  
✅ **Always 5 suggestions per language**  
✅ **Suggestions are contextual** (based on original title)  
✅ **Suggestions are unique** (no duplicates within language)  
✅ **Fallback mechanism** (if AI fails, generates rule-based variants)

---

## Example CSS

```css
.conflict-dialog {
  max-width: 600px;
  padding: 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.language-tabs {
  display: flex;
  gap: 8px;
  margin: 16px 0;
  border-bottom: 2px solid #e0e0e0;
}

.language-tabs button {
  padding: 8px 16px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  transition: all 0.2s;
}

.language-tabs button.active {
  color: #2196f3;
  border-bottom: 2px solid #2196f3;
  transform: translateY(2px);
}

.suggestions-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 16px 0;
}

.suggestion-item {
  padding: 12px 16px;
  text-align: left;
  background: #f5f5f5;
  border: 2px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.suggestion-item:hover {
  background: #e3f2fd;
  border-color: #2196f3;
  transform: translateX(4px);
}
```

---

## Testing

```typescript
describe('Multilingual Conflict Resolution', () => {
  it('should display all three languages', () => {
    const mockConflict = {
      code: 'RECIPE_NAME_EXISTS',
      suggestions: {
        ru: ['Title 1', 'Title 2', 'Title 3', 'Title 4', 'Title 5'],
        en: ['Title 1', 'Title 2', 'Title 3', 'Title 4', 'Title 5'],
        pl: ['Title 1', 'Title 2', 'Title 3', 'Title 4', 'Title 5'],
      },
    };

    render(<ConflictDialog conflict={mockConflict} />);
    
    expect(screen.getByText('🇷🇺 Русский')).toBeInTheDocument();
    expect(screen.getByText('🇬🇧 English')).toBeInTheDocument();
    expect(screen.getByText('🇵🇱 Polski')).toBeInTheDocument();
  });

  it('should show 5 suggestions per language', () => {
    // ... test implementation
  });

  it('should call onSelect with correct language', async () => {
    const onSelect = jest.fn();
    render(<ConflictDialog conflict={mockConflict} onSelect={onSelect} />);
    
    // Switch to English tab
    fireEvent.click(screen.getByText('🇬🇧 English'));
    
    // Select first suggestion
    fireEvent.click(screen.getByText('Title 1'));
    
    expect(onSelect).toHaveBeenCalledWith('Title 1', 'en');
  });
});
```

---

## Migration Guide

If you have existing conflict handling code:

### Before
```typescript
// Single language array
if (error.code === 'RECIPE_NAME_EXISTS') {
  const suggestions = error.suggestions; // string[]
  showSuggestions(suggestions);
}
```

### After
```typescript
// Multilingual object
if (error.code === 'RECIPE_NAME_EXISTS') {
  const suggestions = error.suggestions; // Record<string, string[]>
  showMultilingualSuggestions(suggestions);
}
```

---

## Questions?

Contact backend team for:
- API changes
- Performance issues
- Missing suggestions
- Language support requests
