-- Тестовые данные для Culinary Academy

-- Курс 1: Основы суши (Польский)
INSERT INTO "Course" (id, title, description, image_url, level, category, duration, lessons_count, language, is_published, instructor, stars, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'Podstawy Sushi - Kurs dla Początkujących',
  'Naucz się podstaw przygotowania sushi od prawdziwego mistrza. Ten kurs obejmuje wszystko, od przygotowania ryżu po techniki krojenia ryb.',
  'https://images.unsplash.com/photo-1579584425555-c3ce17fd4351',
  1,
  'sushi',
  120,
  5,
  'pl',
  true,
  'Chef Takeshi Nakamura',
  50,
  NOW(),
  NOW()
);

-- Курс 2: Продвинутые техники (Польский)
INSERT INTO "Course" (id, title, description, image_url, level, category, duration, lessons_count, language, is_published, instructor, stars, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'Zaawansowane Techniki Sushi',
  'Dla tych, którzy opanowali podstawy. Naucz się robić nigiri, sashimi i rolki fusion.',
  'https://images.unsplash.com/photo-1611143669185-af224c5e3252',
  5,
  'sushi',
  180,
  8,
  'pl',
  true,
  'Chef Takeshi Nakamura',
  80,
  NOW(),
  NOW()
);

-- Курс 3: Майстерність ножа (Українська)
INSERT INTO "Course" (id, title, description, image_url, level, category, duration, lessons_count, language, is_published, instructor, stars, created_at, updated_at)
VALUES (
  gen_random_uuid(),
  'Майстерність Ножа - Професійні Техніки',
  'Навчіться професійним технікам роботи з ножем для суші та сашімі.',
  'https://images.unsplash.com/photo-1617093727343-374698b1b08d',
  3,
  'knife-skills',
  90,
  6,
  'ua',
  true,
  'Chef Dmitrij Fomin',
  60,
  NOW(),
  NOW()
);

-- Уроки для курса "Podstawy Sushi"
-- Получаем ID первого курса
DO $$
DECLARE
    course_id_1 UUID;
BEGIN
    SELECT id INTO course_id_1 FROM "Course" WHERE title = 'Podstawy Sushi - Kurs dla Początkujących' LIMIT 1;

    INSERT INTO "Lesson" (id, course_id, title, description, video_url, "order", duration, content, steps, is_published, created_at, updated_at)
    VALUES 
    (gen_random_uuid(), course_id_1, 'Wprowadzenie do Sushi', 'Poznaj historię sushi i podstawowe narzędzia', 'https://youtube.com/watch?v=example1', 1, 15, 
     'W tej lekcji poznasz fascynującą historię sushi i niezbędne narzędzia do przygotowania.',
     ARRAY['Przygotuj przestrzeń roboczą', 'Sprawdź narzędzia', 'Poznaj składniki'], true, NOW(), NOW()),
    
    (gen_random_uuid(), course_id_1, 'Przygotowanie Idealnego Ryżu', 'Sekret doskonałego ryżu sushi', 'https://youtube.com/watch?v=example2', 2, 20,
     'Ryż to fundament każdego sushi. Nauczysz się, jak go idealnie przygotować.',
     ARRAY['Płukanie ryżu', 'Gotowanie ryżu', 'Dodanie octu ryżowego', 'Chłodzenie'], true, NOW(), NOW()),
    
    (gen_random_uuid(), course_id_1, 'Techniki Krojenia Ryb', 'Jak profesjonalnie kroić łososia i tuńczyka', 'https://youtube.com/watch?v=example3', 3, 25,
     'Właściwe krojenie ryby to sztuka. Poznaj techniki mistrzów.',
     ARRAY['Wybór noża', 'Kąt cięcia', 'Technika sashimi', 'Bezpieczeństwo'], true, NOW(), NOW()),
    
    (gen_random_uuid(), course_id_1, 'Tworzenie Maki Rolls', 'Zwijanie pierwszych rolek', 'https://youtube.com/watch?v=example4', 4, 30,
     'Naucz się zwijać klasyczne maki rolls krok po kroku.',
     ARRAY['Rozłóż nori', 'Nałóż ryż', 'Dodaj składniki', 'Zwiń rolkę', 'Pokrój'], true, NOW(), NOW()),
    
    (gen_random_uuid(), course_id_1, 'Prezentacja i Serwowanie', 'Jak podawać sushi jak profesjonalista', 'https://youtube.com/watch?v=example5', 5, 15,
     'Prezentacja to połowa sukcesu. Naucz się pięknie serwować sushi.',
     ARRAY['Wybór talerza', 'Układanie sushi', 'Dodatki', 'Sosy'], true, NOW(), NOW());
END $$;

-- Pytania quizowe dla kursu "Podstawy Sushi"
DO $$
DECLARE
    course_id_1 UUID;
BEGIN
    SELECT id INTO course_id_1 FROM "Course" WHERE title = 'Podstawy Sushi - Kurs dla Początkujących' LIMIT 1;

    INSERT INTO "QuizQuestion" (id, course_id, question, options, correct_answer, explanation, difficulty, language, created_at)
    VALUES 
    (gen_random_uuid(), course_id_1, 
     'Jaka temperatura jest idealna dla ryżu sushi?',
     ARRAY['10°C', '20-25°C (temperatura pokojowa)', '40°C', '60°C'],
     1,
     'Ryż sushi najlepiej smakuje w temperaturze pokojowej (20-25°C). Jest wtedy elastyczny i łatwy w formowaniu.',
     'easy', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Który nóż jest najlepszy do krojenia sashimi?',
     ARRAY['Nóż kuchenny', 'Yanagiba (nóż do sashimi)', 'Nóż chleba', 'Nóż do warzyw'],
     1,
     'Yanagiba to tradycyjny japoński nóż do krojenia sashimi. Ma długie, cienkie ostrze idealne do precyzyjnych cięć.',
     'medium', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Ile warstw nori używa się w standardowym maki roll?',
     ARRAY['Połowa arkusza', 'Jeden pełny arkusz', 'Dwa arkusze', 'Nie używa się nori'],
     0,
     'W klasycznym maki roll używa się połowy arkusza nori.',
     'easy', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Co dodaje się do ryżu po ugotowaniu?',
     ARRAY['Sól', 'Ocet ryżowy (sushi-zu)', 'Sake', 'Sos sojowy'],
     1,
     'Do ugotowanego ryżu dodaje się mieszankę octu ryżowego (sushi-zu), która nadaje charakterystyczny smak.',
     'easy', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Jaka jest najbezpieczniejsza temperatura przechowywania ryb do sushi?',
     ARRAY['-18°C (zamrażarka)', '0-4°C (lodówka)', '10°C', '25°C'],
     0,
     'Ryby do sushi powinny być zamrożone w -18°C przez minimum 24h, aby zabić pasożyty.',
     'hard', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Czym różni się nigiri od sashimi?',
     ARRAY['Nigiri to tylko ryba, sashimi to ryba z ryżem', 'Nigiri to ryba na ryżu, sashimi to sama ryba', 'To to samo', 'Nigiri jest wegańskie'],
     1,
     'Nigiri to plasterek ryby na formowanym ryżu, sashimi to samo filet ryby bez ryżu.',
     'medium', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Która strona nori powinna być na zewnątrz przy zwijaniu maki?',
     ARRAY['Błyszcząca strona', 'Szorstka strona', 'Nie ma znaczenia', 'Zależy od rodzaju rolki'],
     1,
     'Szorstka strona nori powinna być na zewnątrz, aby ryż lepiej się przylepiał.',
     'medium', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Ile gramów ryżu używa się średnio na jedno nigiri?',
     ARRAY['5g', '15-20g', '50g', '100g'],
     1,
     'Standardowe nigiri zawiera około 15-20g ryżu - idealnie do ukształtowania małej kulki.',
     'hard', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Co oznacza słowo "sushi"?',
     ARRAY['Surowa ryba', 'Kwaśny smak', 'Japońskie jedzenie', 'Ryż z wodorostami'],
     1,
     '"Sushi" pochodzi od słowa oznaczającego "kwaśny smak", nawiązując do octu w ryżu.',
     'easy', 'pl', NOW()),
    
    (gen_random_uuid(), course_id_1,
     'Jakim sosem tradycyjnie podaje się wasabi?',
     ARRAY['Majonezem', 'Sosem sojowym', 'Sosem słodko-kwaśnym', 'Ketchupem'],
     1,
     'Wasabi tradycyjnie podaje się z sosem sojowym jako dodatek do sushi.',
     'easy', 'pl', NOW());
END $$;

-- Wyświetl podsumowanie
SELECT 'Courses created:' as info, COUNT(*) as count FROM "Course";
SELECT 'Lessons created:' as info, COUNT(*) as count FROM "Lesson";
SELECT 'Quiz questions created:' as info, COUNT(*) as count FROM "QuizQuestion";
