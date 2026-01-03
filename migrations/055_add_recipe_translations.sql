-- ========================================
-- Migration 055: Add Recipe Translations
-- ========================================
-- Date: 2026-01-03
-- Purpose: Унифицировать модель Recipe с Ingredient (добавить поля переводов)
-- Why: Сейчас Recipe использует localName (смешанный язык), 
--      а Ingredient имеет name_ru, name_en, name_pl
--      Нужно привести к единому стандарту для мультиязычности

-- ========================================
-- ПРОБЛЕМА (до миграции)
-- ========================================
/*
Recipe:
- canonicalName: "Scrambled Eggs" (EN, system name)
- localName: "Яичница" (RU/PL/EN - неконсистентно!)
- description: "Классическая яичница..." (RU/EN - неконсистентно!)

Ingredient (правильно):
- name: "Jaja" (canonical, system name)
- name_pl: "Jaja"
- name_en: "Egg"
- name_ru: "яйцо"

❌ Frontend не знает, какой язык в localName
❌ Backend не может выбрать правильный перевод по Accept-Language
❌ Нет консистентности между Recipe и Ingredient
*/

-- ========================================
-- РЕШЕНИЕ (после миграции)
-- ========================================
/*
Recipe:
- canonicalName: "Scrambled Eggs" (EN, system name) ✅ остаётся
- localName: "Scrambled Eggs" (deprecated, для совместимости) ⚠️
- name_pl: "Jajecznica"
- name_en: "Scrambled Eggs"
- name_ru: "Яичница"
- description_pl: "Klasyczna jajecznica..."
- description_en: "Classic scrambled eggs..."
- description_ru: "Классическая яичница..."

✅ Backend выбирает язык по Accept-Language header
✅ API отдаёт одно поле "title" (нужный перевод)
✅ Frontend просто отображает, не думает о языках
✅ Консистентность с Ingredient моделью
*/

-- ========================================
-- STEP 1: Add translation columns
-- ========================================

ALTER TABLE "Recipe" 
ADD COLUMN IF NOT EXISTS name_pl VARCHAR(255),
ADD COLUMN IF NOT EXISTS name_en VARCHAR(255),
ADD COLUMN IF NOT EXISTS name_ru VARCHAR(255),
ADD COLUMN IF NOT EXISTS description_pl TEXT,
ADD COLUMN IF NOT EXISTS description_en TEXT,
ADD COLUMN IF NOT EXISTS description_ru TEXT;

-- ========================================
-- STEP 2: Migrate existing data
-- ========================================

-- 2.1: Scrambled Eggs (единственный с русским переводом)
UPDATE "Recipe"
SET 
  name_en = 'Scrambled Eggs',
  name_ru = 'Яичница',
  name_pl = 'Jajecznica',
  description_en = 'Classic scrambled eggs — simple and quick breakfast. Perfect for one person.',
  description_ru = 'Классическая яичница — простой и быстрый завтрак. Идеально для одного человека.',
  description_pl = 'Klasyczna jajecznica — prosty i szybki śniadanie. Idealny dla jednej osoby.'
WHERE "canonicalName" = 'Scrambled Eggs';

-- 2.2: Greek Salad (польское название уже есть)
UPDATE "Recipe"
SET 
  name_en = 'Greek Salad',
  name_ru = 'Греческий салат',
  name_pl = 'Sałatka grecka',
  description_en = 'Fresh Greek salad with tomatoes, cucumbers, olives and feta cheese.',
  description_ru = 'Свежий греческий салат с помидорами, огурцами, оливками и сыром фета.',
  description_pl = 'Świeża sałatka grecka z pomidorami, ogórkami, oliwkami i serem feta.'
WHERE "canonicalName" = 'Greek Salad';

-- 2.3: Polish Chicken Soup (Rosół)
UPDATE "Recipe"
SET 
  name_en = 'Polish Chicken Soup',
  name_ru = 'Польский куриный бульон',
  name_pl = 'Rosół',
  description_en = 'Traditional Polish chicken broth with vegetables and noodles.',
  description_ru = 'Традиционный польский куриный бульон с овощами и лапшой.',
  description_pl = 'Tradycyjny polski rosół z warzywami i makaronem.'
WHERE "canonicalName" = 'Polish Chicken Soup';

-- 2.4: Polish Hunters Stew (Bigos)
UPDATE "Recipe"
SET 
  name_en = 'Polish Hunters Stew',
  name_ru = 'Бигос (охотничье рагу)',
  name_pl = 'Bigos myśliwski',
  description_en = 'Traditional Polish stew with sauerkraut, meat and mushrooms.',
  description_ru = 'Традиционное польское рагу с квашеной капустой, мясом и грибами.',
  description_pl = 'Tradycyjny polski bigos z kapustą kiszoną, mięsem i grzybami.'
WHERE "canonicalName" = 'Polish Hunters Stew';

-- 2.5: Polish Meat Dumplings (Pierogi z mięsem)
UPDATE "Recipe"
SET 
  name_en = 'Polish Meat Dumplings',
  name_ru = 'Пельмени с мясом',
  name_pl = 'Pierogi z mięsem',
  description_en = 'Traditional Polish dumplings filled with seasoned meat.',
  description_ru = 'Традиционные польские пельмени с мясной начинкой.',
  description_pl = 'Tradycyjne polskie pierogi z mięsem.'
WHERE "canonicalName" = 'Polish Meat Dumplings';

-- 2.6: Pierogi Ruskie
UPDATE "Recipe"
SET 
  name_en = 'Pierogi Ruskie',
  name_ru = 'Пироги русские',
  name_pl = 'Pierogi ruskie',
  description_en = 'Polish dumplings with potato and cheese filling.',
  description_ru = 'Польские пироги с картофелем и сыром.',
  description_pl = 'Pierogi z ziemniakami i serem.'
WHERE "canonicalName" = 'Pierogi Ruskie';

-- 2.7: Pizza Margherita
UPDATE "Recipe"
SET 
  name_en = 'Pizza Margherita',
  name_ru = 'Пицца Маргарита',
  name_pl = 'Pizza Margherita',
  description_en = 'Classic Italian pizza with tomato sauce, mozzarella and basil.',
  description_ru = 'Классическая итальянская пицца с томатным соусом, моцареллой и базиликом.',
  description_pl = 'Klasyczna włoska pizza z sosem pomidorowym, mozzarellą i bazylią.'
WHERE "canonicalName" = 'Pizza Margherita';

-- 2.8: Polish Breaded Pork Chop (Kotlet schabowy)
UPDATE "Recipe"
SET 
  name_en = 'Polish Breaded Pork Chop',
  name_ru = 'Польский шницель',
  name_pl = 'Kotlet schabowy',
  description_en = 'Breaded pork chop, Polish-style schnitzel.',
  description_ru = 'Отбивная в панировке, польский шницель.',
  description_pl = 'Panierowany kotlet schabowy po polsku.'
WHERE "canonicalName" = 'Polish Breaded Pork Chop';

-- 2.9: Polish Potato Pancakes (Placki ziemniaczane)
UPDATE "Recipe"
SET 
  name_en = 'Polish Potato Pancakes',
  name_ru = 'Драники',
  name_pl = 'Placki ziemniaczane',
  description_en = 'Traditional Polish potato pancakes served with sour cream.',
  description_ru = 'Традиционные польские драники, подаются со сметаной.',
  description_pl = 'Tradycyjne polskie placki ziemniaczane podawane ze śmietaną.'
WHERE "canonicalName" = 'Polish Potato Pancakes';

-- 2.10: Spaghetti Carbonara
UPDATE "Recipe"
SET 
  name_en = 'Spaghetti Carbonara',
  name_ru = 'Спагетти Карбонара',
  name_pl = 'Spaghetti alla Carbonara',
  description_en = 'Classic Italian pasta with eggs, bacon and parmesan cheese.',
  description_ru = 'Классическая итальянская паста с яйцами, беконом и пармезаном.',
  description_pl = 'Klasyczny włoski makaron z jajkami, boczkiem i parmezanem.'
WHERE "canonicalName" = 'Spaghetti Carbonara';

-- ========================================
-- STEP 3: Create indexes for performance
-- ========================================

CREATE INDEX IF NOT EXISTS idx_recipe_name_ru ON "Recipe"(name_ru) WHERE name_ru IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_recipe_name_en ON "Recipe"(name_en) WHERE name_en IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_recipe_name_pl ON "Recipe"(name_pl) WHERE name_pl IS NOT NULL;

-- ========================================
-- STEP 4: Add comments for documentation
-- ========================================

COMMENT ON COLUMN "Recipe".name_pl IS 'Polish name (used when Accept-Language: pl)';
COMMENT ON COLUMN "Recipe".name_en IS 'English name (used when Accept-Language: en, default)';
COMMENT ON COLUMN "Recipe".name_ru IS 'Russian name (used when Accept-Language: ru)';
COMMENT ON COLUMN "Recipe".description_pl IS 'Polish description';
COMMENT ON COLUMN "Recipe".description_en IS 'English description';
COMMENT ON COLUMN "Recipe".description_ru IS 'Russian description';
COMMENT ON COLUMN "Recipe"."localName" IS 'DEPRECATED: Use name_pl/name_en/name_ru instead. Kept for backward compatibility.';

-- ========================================
-- VERIFICATION QUERIES
-- ========================================

-- 1. Check all recipes have translations
SELECT 
  "canonicalName",
  name_en IS NOT NULL as has_en,
  name_ru IS NOT NULL as has_ru,
  name_pl IS NOT NULL as has_pl,
  description_en IS NOT NULL as desc_en,
  description_ru IS NOT NULL as desc_ru,
  description_pl IS NOT NULL as desc_pl
FROM "Recipe"
ORDER BY "canonicalName";

-- Expected: All should have TRUE for all columns

-- 2. Sample translations for one recipe
SELECT 
  "canonicalName",
  name_en as "English",
  name_ru as "Russian",
  name_pl as "Polish"
FROM "Recipe"
WHERE "canonicalName" = 'Scrambled Eggs';

-- Expected:
-- English: Scrambled Eggs
-- Russian: Яичница
-- Polish: Jajecznica

-- 3. Count complete translations
SELECT 
  COUNT(*) FILTER (WHERE name_en IS NOT NULL) as has_english,
  COUNT(*) FILTER (WHERE name_ru IS NOT NULL) as has_russian,
  COUNT(*) FILTER (WHERE name_pl IS NOT NULL) as has_polish,
  COUNT(*) as total_recipes
FROM "Recipe";

-- Expected: 10/10/10/10 (all recipes translated)

-- ========================================
-- NOTES FOR BACKEND IMPLEMENTATION
-- ========================================
/*
1. Update RecipeCatalog model:
   type RecipeCatalog struct {
     CanonicalName string `json:"canonicalName"`
     LocalName     string `json:"localName"` // DEPRECATED
     NamePl        *string `json:"namePl,omitempty"`
     NameEn        *string `json:"nameEn,omitempty"`
     NameRu        *string `json:"nameRu,omitempty"`
     DescriptionPl *string `json:"descriptionPl,omitempty"`
     DescriptionEn *string `json:"descriptionEn,omitempty"`
     DescriptionRu *string `json:"descriptionRu,omitempty"`
   }

2. Add helper method to get localized name:
   func (r *RecipeCatalog) GetLocalizedName(lang string) string {
     switch lang {
     case "ru":
       if r.NameRu != nil { return *r.NameRu }
     case "pl":
       if r.NamePl != nil { return *r.NamePl }
     default:
       if r.NameEn != nil { return *r.NameEn }
     }
     return r.CanonicalName
   }

3. Update API responses to use single "title" field:
   {
     "id": "...",
     "title": recipe.GetLocalizedName(userLang),
     "description": recipe.GetLocalizedDescription(userLang),
     "language": userLang
   }

4. Frontend просто отображает "title", не думает о языках ✅
*/

-- ========================================
-- ROLLBACK (if needed)
-- ========================================
/*
ALTER TABLE "Recipe"
DROP COLUMN IF EXISTS name_pl,
DROP COLUMN IF EXISTS name_en,
DROP COLUMN IF EXISTS name_ru,
DROP COLUMN IF EXISTS description_pl,
DROP COLUMN IF EXISTS description_en,
DROP COLUMN IF EXISTS description_ru;

DROP INDEX IF EXISTS idx_recipe_name_ru;
DROP INDEX IF EXISTS idx_recipe_name_en;
DROP INDEX IF EXISTS idx_recipe_name_pl;
*/
