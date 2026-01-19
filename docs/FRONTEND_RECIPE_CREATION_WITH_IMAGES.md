# Frontend: Создание рецепта с фото - Полное руководство

## 📋 Содержание
1. [Общий процесс](#общий-процесс)
2. [Шаг 1: AI-генерация рецепта](#шаг-1-ai-генерация-рецепта)
3. [Шаг 2: Загрузка изображения](#шаг-2-загрузка-изображения)
4. [Полный пример React компонента](#полный-пример-react-компонента)
5. [TypeScript типы](#typescript-типы)
6. [Обработка ошибок](#обработка-ошибок)

---

## 🔄 Общий процесс

```
1. Пользователь заполняет форму рецепта
   ↓
2. Frontend отправляет POST /api/admin/recipes/create-ai
   ↓
3. Backend через AI генерирует полный рецепт и сохраняет в БД
   ↓
4. Backend возвращает созданный рецепт с ID
   ↓
5. Frontend загружает фото через POST /api/admin/recipes/{id}/image
   ↓
6. Backend сохраняет фото в Cloudinary и обновляет БД
   ↓
7. Готово! Рецепт с фото создан ✅
```

---

## 🤖 Шаг 1: AI-генерация рецепта

### Endpoint
```
POST /api/admin/recipes/create-ai
```

### Требования
- **Аутентификация:** JWT токен с ролью `admin`
- **Content-Type:** `application/json`
- **Accept-Language:** `pl`, `en`, или `ru` (опционально)

### Request Body

```typescript
interface CreateRecipeRequest {
  title: string;              // "Жареный лосось с рисом"
  language?: string;          // "pl" | "en" | "ru" (опционально)
  ingredients: Array<{
    ingredientId: string;     // UUID ингредиента из базы
    quantity?: number;        // 150
    amount?: number;          // 150 (альтернативное поле)
    unit: string;             // "g", "ml", "pieces"
  }>;
  rawCookingText: string;     // "Рыбу замариновать, обжарить. Рис отварить."
}
```

### Пример запроса

```javascript
const createRecipe = async (adminToken, recipeData) => {
  const response = await fetch(
    'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/create-ai',
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${adminToken}`,
        'Content-Type': 'application/json',
        'Accept-Language': 'pl',  // Язык ответа AI
      },
      body: JSON.stringify({
        title: "Smażony łosoś z ryżem",
        language: "pl",
        ingredients: [
          {
            ingredientId: "123e4567-e89b-12d3-a456-426614174000",
            quantity: 200,
            unit: "g"
          },
          {
            ingredientId: "123e4567-e89b-12d3-a456-426614174001",
            quantity: 150,
            unit: "g"
          }
        ],
        rawCookingText: "Łosoś zamarynować w sosie teriyaki przez 15 minut. Smażyć na patelni 5-7 minut z każdej strony. Ryż ugotować według instrukcji na opakowaniu."
      })
    }
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to create recipe');
  }

  const result = await response.json();
  return result.data; // Возвращает созданный рецепт с ID
};
```

### Response (201 Created)

```json
{
  "success": true,
  "message": "Recipe created via AI",
  "data": {
    "id": "6b8628ef-ef1e-42eb-a166-924566bb9b7b",
    "canonicalName": "fried_salmon_with_rice",
    "title": "Smażony łosoś z ryżem",
    "namePl": "Smażony łosoś z ryżem",
    "nameEn": "Fried Salmon with Rice",
    "nameRu": "Жареный лосось с рисом",
    "descriptionPl": "Soczysty łosoś z aromatycznym ryżem",
    "country": "Poland",
    "category": "main",
    "difficulty": "easy",
    "timeMinutes": 30,
    "servings": 2,
    "calories": 520,
    "ingredients": [...],
    "steps": [...],
    "createdAt": "2026-01-19T10:30:00Z",
    "updatedAt": "2026-01-19T10:30:00Z"
  }
}
```

### Возможные ошибки

| Код | Описание | Причина |
|-----|----------|---------|
| 401 | Unauthorized | Нет JWT токена или не admin |
| 400 | Bad Request | Невалидный JSON или отсутствуют обязательные поля |
| 409 | Conflict | Рецепт с похожим названием уже существует |
| 422 | Unprocessable Entity | AI не смог обработать рецепт |
| 500 | Internal Server Error | Ошибка сервера |

---

## 📸 Шаг 2: Загрузка изображения

### Endpoint
```
POST /api/admin/recipes/{recipeId}/image
```

### Требования
- **Аутентификация:** JWT токен с ролью `admin`
- **Content-Type:** `multipart/form-data`
- **Макс. размер файла:** 5MB
- **Форматы:** JPEG, PNG, WebP

### Request

```javascript
const uploadRecipeImage = async (adminToken, recipeId, imageFile) => {
  const formData = new FormData();
  formData.append('image', imageFile);

  const response = await fetch(
    `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/${recipeId}/image`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${adminToken}`,
        // НЕ указываем Content-Type - браузер сам установит boundary для multipart
      },
      body: formData
    }
  );

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to upload image');
  }

  return await response.json();
};
```

### Response (200 OK)

```json
{
  "success": true,
  "data": {
    "imageUrl": "https://res.cloudinary.com/dwrn0ohbp/image/upload/v1768818751/recipes/recipe_6b8628ef-ef1e-42eb-a166-924566bb9b7b.webp",
    "imagePublicId": "recipes/recipe_6b8628ef-ef1e-42eb-a166-924566bb9b7b"
  },
  "message": "Image uploaded successfully"
}
```

### Автоматические преобразования Cloudinary

- ✅ **Формат:** Конвертируется в WebP для оптимизации
- ✅ **Размер:** 1200x800px (соотношение 16:9)
- ✅ **Обрезка:** `c_fill` (заполняет рамку, обрезая если нужно)
- ✅ **Качество:** Автоматическая оптимизация
- ✅ **Замена:** При загрузке нового фото старое удаляется автоматически

---

## ⚛️ Полный пример React компонента

```tsx
import React, { useState } from 'react';

// ========================================
// TypeScript Types
// ========================================

interface Ingredient {
  ingredientId: string;
  quantity: number;
  unit: string;
  name?: string; // Для отображения в UI
}

interface CreateRecipeFormData {
  title: string;
  language: 'pl' | 'en' | 'ru';
  ingredients: Ingredient[];
  rawCookingText: string;
}

interface CreatedRecipe {
  id: string;
  canonicalName: string;
  title: string;
  imageUrl?: string;
  // ... другие поля
}

// ========================================
// Component
// ========================================

export function CreateRecipeWithImage() {
  const [formData, setFormData] = useState<CreateRecipeFormData>({
    title: '',
    language: 'pl',
    ingredients: [],
    rawCookingText: '',
  });

  const [imageFile, setImageFile] = useState<File | null>(null);
  const [imagePreview, setImagePreview] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdRecipe, setCreatedRecipe] = useState<CreatedRecipe | null>(null);

  // Токен админа (получаем из auth context или localStorage)
  const adminToken = localStorage.getItem('adminToken');

  // ========================================
  // Валидация изображения
  // ========================================
  const validateImage = (file: File): string | null => {
    const MAX_SIZE = 5 * 1024 * 1024; // 5MB
    const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/webp'];

    if (file.size > MAX_SIZE) {
      return 'Размер файла не должен превышать 5MB';
    }

    if (!ALLOWED_TYPES.includes(file.type)) {
      return 'Разрешены только форматы JPEG, PNG и WebP';
    }

    return null;
  };

  // ========================================
  // Обработчик выбора файла
  // ========================================
  const handleImageChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const validationError = validateImage(file);
    if (validationError) {
      setError(validationError);
      return;
    }

    setImageFile(file);
    setError(null);

    // Создаём preview
    const reader = new FileReader();
    reader.onloadend = () => {
      setImagePreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  // ========================================
  // Добавление ингредиента
  // ========================================
  const addIngredient = () => {
    setFormData({
      ...formData,
      ingredients: [
        ...formData.ingredients,
        { ingredientId: '', quantity: 0, unit: 'g' }
      ]
    });
  };

  const updateIngredient = (index: number, field: keyof Ingredient, value: any) => {
    const newIngredients = [...formData.ingredients];
    newIngredients[index] = { ...newIngredients[index], [field]: value };
    setFormData({ ...formData, ingredients: newIngredients });
  };

  const removeIngredient = (index: number) => {
    const newIngredients = formData.ingredients.filter((_, i) => i !== index);
    setFormData({ ...formData, ingredients: newIngredients });
  };

  // ========================================
  // Создание рецепта
  // ========================================
  const createRecipe = async (): Promise<CreatedRecipe> => {
    const response = await fetch(
      'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/create-ai',
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${adminToken}`,
          'Content-Type': 'application/json',
          'Accept-Language': formData.language,
        },
        body: JSON.stringify({
          title: formData.title,
          language: formData.language,
          ingredients: formData.ingredients,
          rawCookingText: formData.rawCookingText,
        }),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Не удалось создать рецепт');
    }

    const result = await response.json();
    return result.data;
  };

  // ========================================
  // Загрузка изображения
  // ========================================
  const uploadImage = async (recipeId: string): Promise<string> => {
    if (!imageFile) {
      throw new Error('Файл изображения не выбран');
    }

    const formData = new FormData();
    formData.append('image', imageFile);

    const response = await fetch(
      `https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/admin/recipes/${recipeId}/image`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${adminToken}`,
        },
        body: formData,
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Не удалось загрузить изображение');
    }

    const result = await response.json();
    return result.data.imageUrl;
  };

  // ========================================
  // Отправка формы
  // ========================================
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // Валидация
      if (!formData.title.trim()) {
        throw new Error('Название рецепта обязательно');
      }
      if (formData.ingredients.length === 0) {
        throw new Error('Добавьте хотя бы один ингредиент');
      }
      if (!formData.rawCookingText.trim()) {
        throw new Error('Опишите процесс приготовления');
      }

      console.log('🚀 Шаг 1: Создание рецепта через AI...');
      const recipe = await createRecipe();
      console.log('✅ Рецепт создан:', recipe.id);

      setCreatedRecipe(recipe);

      // Если есть изображение - загружаем его
      if (imageFile) {
        console.log('📸 Шаг 2: Загрузка изображения...');
        const imageUrl = await uploadImage(recipe.id);
        console.log('✅ Изображение загружено:', imageUrl);

        // Обновляем созданный рецепт с imageUrl
        setCreatedRecipe({ ...recipe, imageUrl });
      }

      // Успех!
      alert('✅ Рецепт успешно создан!');
      
      // Можно перенаправить на страницу рецепта
      // navigate(`/recipes/${recipe.id}`);

    } catch (err) {
      console.error('❌ Ошибка:', err);
      setError(err instanceof Error ? err.message : 'Произошла ошибка');
    } finally {
      setLoading(false);
    }
  };

  // ========================================
  // Render
  // ========================================
  return (
    <div className="create-recipe-container">
      <h1>Создать новый рецепт</h1>

      {error && (
        <div className="error-message" style={{ color: 'red', marginBottom: '1rem' }}>
          {error}
        </div>
      )}

      {createdRecipe && (
        <div className="success-message" style={{ color: 'green', marginBottom: '1rem' }}>
          ✅ Рецепт "{createdRecipe.title}" успешно создан!
          {createdRecipe.imageUrl && <div>📸 С фото: {createdRecipe.imageUrl}</div>}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        {/* Название */}
        <div className="form-group">
          <label>Название рецепта *</label>
          <input
            type="text"
            value={formData.title}
            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
            placeholder="Smażony łosoś z ryżem"
            required
          />
        </div>

        {/* Язык */}
        <div className="form-group">
          <label>Язык рецепта</label>
          <select
            value={formData.language}
            onChange={(e) => setFormData({ ...formData, language: e.target.value as any })}
          >
            <option value="pl">Polski (Polish)</option>
            <option value="en">English</option>
            <option value="ru">Русский (Russian)</option>
          </select>
        </div>

        {/* Ингредиенты */}
        <div className="form-group">
          <label>Ингредиенты *</label>
          {formData.ingredients.map((ing, index) => (
            <div key={index} style={{ display: 'flex', gap: '0.5rem', marginBottom: '0.5rem' }}>
              <input
                type="text"
                placeholder="ID ингредиента (UUID)"
                value={ing.ingredientId}
                onChange={(e) => updateIngredient(index, 'ingredientId', e.target.value)}
                style={{ flex: 2 }}
              />
              <input
                type="number"
                placeholder="Количество"
                value={ing.quantity || ''}
                onChange={(e) => updateIngredient(index, 'quantity', parseFloat(e.target.value))}
                style={{ flex: 1 }}
              />
              <select
                value={ing.unit}
                onChange={(e) => updateIngredient(index, 'unit', e.target.value)}
                style={{ flex: 1 }}
              >
                <option value="g">граммы (g)</option>
                <option value="ml">миллилитры (ml)</option>
                <option value="pieces">штуки</option>
                <option value="tbsp">ст. ложки</option>
                <option value="tsp">ч. ложки</option>
              </select>
              <button type="button" onClick={() => removeIngredient(index)}>
                Удалить
              </button>
            </div>
          ))}
          <button type="button" onClick={addIngredient}>
            + Добавить ингредиент
          </button>
        </div>

        {/* Инструкции по приготовлению */}
        <div className="form-group">
          <label>Инструкции по приготовлению *</label>
          <textarea
            value={formData.rawCookingText}
            onChange={(e) => setFormData({ ...formData, rawCookingText: e.target.value })}
            placeholder="Опишите процесс приготовления в свободной форме. AI создаст структурированные шаги."
            rows={6}
            required
          />
          <small style={{ color: '#666' }}>
            Пример: "Łosoś zamarynować w sosie teriyaki przez 15 minut. Smażyć na patelni 5-7 minut z każdej strony..."
          </small>
        </div>

        {/* Загрузка изображения */}
        <div className="form-group">
          <label>Фото рецепта (опционально)</label>
          <input
            type="file"
            accept="image/jpeg,image/png,image/webp"
            onChange={handleImageChange}
            disabled={loading}
          />
          <small style={{ color: '#666' }}>
            Макс. размер: 5MB. Форматы: JPEG, PNG, WebP.
          </small>

          {imagePreview && (
            <div style={{ marginTop: '1rem' }}>
              <img
                src={imagePreview}
                alt="Preview"
                style={{ maxWidth: '300px', borderRadius: '8px' }}
              />
            </div>
          )}
        </div>

        {/* Кнопка отправки */}
        <button type="submit" disabled={loading}>
          {loading ? '⏳ Создание рецепта...' : '🚀 Создать рецепт'}
        </button>
      </form>
    </div>
  );
}
```

---

## 📝 TypeScript типы (полный набор)

```typescript
// types/recipe.ts

// ========================================
// Request Types
// ========================================

export interface CreateRecipeRequest {
  title: string;
  language?: 'pl' | 'en' | 'ru';
  ingredients: RecipeIngredientInput[];
  rawCookingText: string;
}

export interface RecipeIngredientInput {
  ingredientId: string;
  quantity?: number;
  amount?: number;  // Альтернативное поле
  unit: string;
}

// ========================================
// Response Types
// ========================================

export interface Recipe {
  id: string;
  canonicalName: string;
  title: string;
  
  // Multilingual names
  namePl?: string;
  nameEn?: string;
  nameRu?: string;
  
  // Multilingual descriptions
  descriptionPl?: string;
  descriptionEn?: string;
  descriptionRu?: string;
  
  // Image
  imageUrl?: string;
  imagePublicId?: string;
  
  // Details
  country: string;
  category: 'appetizer' | 'main' | 'dessert' | 'soup' | 'salad' | 'breakfast' | 'snack';
  difficulty: 'easy' | 'medium' | 'hard';
  timeMinutes: number;
  servings: number;
  calories?: number;
  
  // Relations
  ingredients: RecipeIngredient[];
  steps: RecipeStep[];
  
  // Timestamps
  createdAt: string;
  updatedAt: string;
}

export interface RecipeIngredient {
  id: string;
  recipeId: string;
  ingredientId: string;
  quantity: number;
  unit: string;
  optional: boolean;
  
  // Ingredient details (if preloaded)
  ingredient?: {
    id: string;
    namePl?: string;
    nameEn?: string;
    nameRu?: string;
  };
}

export interface RecipeStep {
  order: number;
  text: string;
  time?: number;
}

// ========================================
// API Response Wrappers
// ========================================

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
}

export interface ApiError {
  success: false;
  message: string;
  code?: string;
}

export interface ImageUploadResponse {
  imageUrl: string;
  imagePublicId: string;
}

// ========================================
// Form State Types
// ========================================

export type RecipeFormData = Omit<CreateRecipeRequest, 'ingredients'> & {
  ingredients: Array<RecipeIngredientInput & { name?: string }>;
};

export interface RecipeCreationState {
  loading: boolean;
  error: string | null;
  recipe: Recipe | null;
  imageUploading: boolean;
}
```

---

## 🔧 API Service (для переиспользования)

```typescript
// services/recipeApi.ts

const API_BASE_URL = 'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api';

class RecipeApiService {
  private getAuthHeaders(token: string, language?: string) {
    return {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      ...(language && { 'Accept-Language': language }),
    };
  }

  /**
   * Создать рецепт через AI
   */
  async createRecipeWithAI(
    token: string,
    data: CreateRecipeRequest
  ): Promise<Recipe> {
    const response = await fetch(`${API_BASE_URL}/admin/recipes/create-ai`, {
      method: 'POST',
      headers: this.getAuthHeaders(token, data.language),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to create recipe');
    }

    const result: ApiResponse<Recipe> = await response.json();
    return result.data;
  }

  /**
   * Загрузить изображение рецепта
   */
  async uploadRecipeImage(
    token: string,
    recipeId: string,
    imageFile: File
  ): Promise<ImageUploadResponse> {
    const formData = new FormData();
    formData.append('image', imageFile);

    const response = await fetch(
      `${API_BASE_URL}/admin/recipes/${recipeId}/image`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to upload image');
    }

    const result: ApiResponse<ImageUploadResponse> = await response.json();
    return result.data;
  }

  /**
   * Удалить изображение рецепта
   */
  async deleteRecipeImage(
    token: string,
    recipeId: string
  ): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/admin/recipes/${recipeId}/image`,
      {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to delete image');
    }
  }

  /**
   * Получить список рецептов
   */
  async getRecipes(filters?: {
    country?: string;
    category?: string;
    difficulty?: string;
    maxTime?: number;
  }): Promise<Recipe[]> {
    const params = new URLSearchParams();
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          params.append(key, String(value));
        }
      });
    }

    const response = await fetch(
      `${API_BASE_URL}/recipes?${params.toString()}`
    );

    if (!response.ok) {
      throw new Error('Failed to fetch recipes');
    }

    const result = await response.json();
    return result.data.recipes;
  }
}

export const recipeApi = new RecipeApiService();
```

---

## ⚠️ Обработка ошибок

### Типичные ошибки и решения

```typescript
const handleRecipeCreation = async (data: CreateRecipeRequest) => {
  try {
    const recipe = await recipeApi.createRecipeWithAI(adminToken, data);
    return recipe;
  } catch (error) {
    if (error instanceof Error) {
      // Обработка специфичных ошибок
      if (error.message.includes('already exists')) {
        alert('Рецепт с таким названием уже существует. Выберите другое название.');
      } else if (error.message.includes('Unauthorized')) {
        alert('Требуется авторизация как администратор');
        // Перенаправить на логин
      } else if (error.message.includes('AI could not process')) {
        alert('AI не смог обработать рецепт. Попробуйте упростить описание.');
      } else {
        alert(`Ошибка: ${error.message}`);
      }
    }
    throw error;
  }
};
```

### Валидация на фронтенде

```typescript
const validateRecipeForm = (data: CreateRecipeRequest): string | null => {
  if (!data.title.trim()) {
    return 'Название рецепта обязательно';
  }

  if (data.title.length < 3) {
    return 'Название должно содержать минимум 3 символа';
  }

  if (data.ingredients.length === 0) {
    return 'Добавьте хотя бы один ингредиент';
  }

  for (const ing of data.ingredients) {
    if (!ing.ingredientId) {
      return 'Все ингредиенты должны иметь ID';
    }
    if ((ing.quantity || ing.amount || 0) <= 0) {
      return 'Количество ингредиента должно быть больше 0';
    }
    if (!ing.unit) {
      return 'Укажите единицу измерения для всех ингредиентов';
    }
  }

  if (!data.rawCookingText.trim()) {
    return 'Опишите процесс приготовления';
  }

  if (data.rawCookingText.length < 20) {
    return 'Описание приготовления слишком короткое (минимум 20 символов)';
  }

  return null; // Всё ОК
};
```

---

## 🎨 Стилизация компонента

```css
/* styles/CreateRecipe.css */

.create-recipe-container {
  max-width: 800px;
  margin: 2rem auto;
  padding: 2rem;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 600;
  color: #333;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 1rem;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #4CAF50;
}

.form-group small {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.875rem;
}

button {
  padding: 0.75rem 1.5rem;
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.3s;
}

button:hover:not(:disabled) {
  background: #45a049;
}

button:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.error-message {
  padding: 1rem;
  background: #ffebee;
  border-left: 4px solid #f44336;
  border-radius: 4px;
}

.success-message {
  padding: 1rem;
  background: #e8f5e9;
  border-left: 4px solid #4CAF50;
  border-radius: 4px;
}
```

---

## ✅ Чек-лист разработки

### Backend (готово ✅)
- [x] POST /api/admin/recipes/create-ai - создание рецепта через AI
- [x] POST /api/admin/recipes/{id}/image - загрузка фото
- [x] DELETE /api/admin/recipes/{id}/image - удаление фото
- [x] GET /api/recipes - возвращает imageUrl
- [x] Cloudinary интеграция
- [x] Транзакционная целостность
- [x] Валидация (5MB, JPEG/PNG/WebP)

### Frontend (TODO)
- [ ] Форма создания рецепта
- [ ] Выбор ингредиентов (autocomplete с поиском по базе)
- [ ] Текстовое поле для инструкций
- [ ] Загрузка фото с preview
- [ ] Валидация формы
- [ ] Обработка ошибок
- [ ] Loading states
- [ ] Success feedback
- [ ] Отображение созданного рецепта
- [ ] Адаптивная вёрстка

---

## 📚 Полезные ссылки

- **API Docs:** `docs/CLOUDINARY_IMAGE_UPLOAD_GUIDE.md`
- **Cloudinary transformations:** https://cloudinary.com/documentation/image_transformations
- **React Hook Form:** https://react-hook-form.com/ (рекомендуется для форм)
- **TypeScript:** https://www.typescriptlang.org/docs/

---

## 🧪 Тестовые данные

### Тестовый админ
```
Email: admin@example.com
Password: admin_password_123
```

### Пример рецепта для теста
```json
{
  "title": "Smażony łosoś z ryżem",
  "language": "pl",
  "ingredients": [
    {
      "ingredientId": "РЕАЛЬНЫЙ_UUID_ИЗ_БАЗЫ",
      "quantity": 200,
      "unit": "g"
    }
  ],
  "rawCookingText": "Łosoś zamarynować w sosie teriyaki przez 15 minut. Smażyć na patelni 5-7 minut z każdej strony. Ryż ugotować według instrukcji na opakowaniu. Podawać razem."
}
```

---

**Готово! 🎉** Теперь у вас есть полное руководство по созданию рецептов с фото на фронтенде.
