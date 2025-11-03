-- Тестовые достижения (achievements) для Culinary Academy

INSERT INTO "Achievement" (id, code, title, description, icon_url, category, requirement, created_at)
VALUES
-- Skill Achievements
(gen_random_uuid(), 'knife_master', 'Knife Master', 'Opanowałeś sztukę krojenia noża', 'https://img.icons8.com/color/96/knife.png', 'skill', 'Complete knife skills course', NOW()),

(gen_random_uuid(), 'fusion_pro', 'Fusion Pro', 'Mistrz kuchni fusion', 'https://img.icons8.com/color/96/star.png', 'skill', 'Create 5 fusion recipes', NOW()),

(gen_random_uuid(), 'sushi_sensei', 'Sushi Sensei', 'Prawdziwy mistrz sushi', 'https://img.icons8.com/color/96/sushi.png', 'skill', 'Reach level 10', NOW()),

-- Course Achievements
(gen_random_uuid(), 'first_course', 'First Steps', 'Ukończono pierwszy kurs', 'https://img.icons8.com/color/96/graduation-cap.png', 'course', 'Complete first course', NOW()),

(gen_random_uuid(), 'perfect_score', 'Perfect Score', '100% w teście', 'https://img.icons8.com/color/96/trophy.png', 'course', 'Get 100% on any quiz', NOW()),

(gen_random_uuid(), 'speed_learner', 'Speed Learner', 'Szybka nauka', 'https://img.icons8.com/color/96/fast-forward.png', 'course', 'Complete 3 courses in 1 week', NOW()),

-- Recipe Achievements
(gen_random_uuid(), 'first_recipe', 'First Creation', 'Pierwszy własny przepis', 'https://img.icons8.com/color/96/recipe.png', 'recipe', 'Create first recipe', NOW()),

(gen_random_uuid(), 'recipe_master', 'Recipe Master', 'Twórca 10 przepisów', 'https://img.icons8.com/color/96/cooking-book.png', 'recipe', 'Create 10 recipes', NOW()),

(gen_random_uuid(), 'bestseller', 'Bestseller', 'Przepis kupiony 50 razy', 'https://img.icons8.com/color/96/best-seller.png', 'recipe', '50 recipe purchases', NOW()),

-- Special Achievements
(gen_random_uuid(), 'early_bird', 'Early Bird', 'Jeden z pierwszych uczniów', 'https://img.icons8.com/color/96/bird.png', 'special', 'Join in first month', NOW()),

(gen_random_uuid(), 'wealthy_chef', 'Wealthy Chef', '1000 ChefToken zgromadzonych', 'https://img.icons8.com/color/96/money.png', 'special', 'Earn 1000 ChefToken', NOW()),

(gen_random_uuid(), 'social_chef', 'Social Chef', 'Pomoc innym uczniom', 'https://img.icons8.com/color/96/handshake.png', 'special', 'Help 10 students', NOW());

-- Nadaj achievement "First Creation" użytkownikowi (bo ma 1 рецепт)
INSERT INTO "UserAchievement" (id, user_id, achievement_id, unlocked_at)
SELECT 
    gen_random_uuid(),
    'ef03cd81-71fd-429f-bb5f-8be5c9172ca8'::uuid,
    id,
    NOW()
FROM "Achievement" WHERE code = 'first_recipe';

-- Nadaj achievement "Perfect Score" (bo 100% в тесте)
INSERT INTO "UserAchievement" (id, user_id, achievement_id, unlocked_at)
SELECT 
    gen_random_uuid(),
    'ef03cd81-71fd-429f-bb5f-8be5c9172ca8'::uuid,
    id,
    NOW()
FROM "Achievement" WHERE code = 'perfect_score';

SELECT 'Achievements created:', COUNT(*) FROM "Achievement";
SELECT 'User achievements:', COUNT(*) FROM "UserAchievement";
