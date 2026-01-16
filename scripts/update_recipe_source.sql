-- Обновляем source.type для рецепта "Яичница"
-- Философия ChefOS: Админ утверждает → становится professional

UPDATE "Recipe"
SET "source" = jsonb_set(
    "source",
    '{type}',
    '"professional"'
)
WHERE "id" = '859d8c56-338e-4da0-8e5c-9ef5412b22ab';

-- Проверяем результат
SELECT 
    "id",
    "title",
    "source"->>'type' as source_type,
    "source"->>'generator' as generator,
    "source"->>'authorId' as author_id
FROM "Recipe"
WHERE "id" = '859d8c56-338e-4da0-8e5c-9ef5412b22ab';
