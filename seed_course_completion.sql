-- Создаём UserProgress для завершённого курса
INSERT INTO "UserProgress" (
    user_id,
    course_id,
    completed_lessons,
    total_lessons,
    quiz_score,
    stars_earned,
    is_completed,
    completed_at
) VALUES (
    'ef03cd81-71fd-429f-bb5f-8be5c9172ca8',
    'e37cb669-9bc3-4688-b723-5af965a57f20',
    5,  -- все 5 уроков завершены
    5,
    100,  -- 100% результат теста
    5,    -- 5 звёзд
    true,
    NOW()
);
