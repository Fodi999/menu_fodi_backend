-- ===================================================================
-- PRICE FLOW DIAGNOSTIC SQL QUERIES
-- Run these in Neon.tech SQL Editor to diagnose economy calculation
-- ===================================================================

-- ===================================================================
-- STEP 1: Check if ANY prices exist in database
-- ===================================================================
SELECT 
  COUNT(*) as total_products,
  COUNT(current_price_per_unit) as products_with_price,
  ROUND(COUNT(current_price_per_unit)::numeric / NULLIF(COUNT(*), 0) * 100, 2) as percent_with_price
FROM user_fridge_items;

-- Expected: At least some products should have prices
-- ❌ If products_with_price = 0 → NO PRICES IN DATABASE → Frontend not sending prices


-- ===================================================================
-- STEP 2: View existing products with prices (verify normalization)
-- ===================================================================
SELECT
  ufi.user_id,
  u.email as user_email,
  i.name as ingredient_name,
  ufi.quantity,
  ufi.unit,
  ufi.current_price_per_unit,  -- Should be normalized (PLN/g or PLN/ml)
  ufi.current_price_currency,
  ufi.price_updated_at,
  -- Calculate expected display price (for verification)
  CASE 
    WHEN ufi.unit = 'g' THEN ROUND((ufi.current_price_per_unit * 1000)::numeric, 2) || ' PLN/kg'
    WHEN ufi.unit = 'ml' THEN ROUND((ufi.current_price_per_unit * 1000)::numeric, 2) || ' PLN/l'
    ELSE ROUND(ufi.current_price_per_unit::numeric, 4) || ' PLN/' || ufi.unit
  END as display_price
FROM user_fridge_items ufi
JOIN "Ingredient" i ON ufi.ingredient_id = i.id
JOIN "User" u ON ufi.user_id = u.id
WHERE ufi.current_price_per_unit IS NOT NULL
ORDER BY u.email, i.name
LIMIT 20;

-- Expected normalization:
--   Input: 20.56 PLN/kg → Stored: 0.02056 PLN/g
--   Input: 3.24 PLN/l   → Stored: 0.00324 PLN/ml
-- ❌ If table empty → NO PRICES SAVED


-- ===================================================================
-- STEP 3: Check specific user's fridge (for recipe generation test)
-- ===================================================================
-- Replace USER_EMAIL with actual user
SELECT
  i.name as ingredient,
  ufi.quantity,
  ufi.unit,
  ufi.current_price_per_unit,
  CASE 
    WHEN ufi.current_price_per_unit IS NULL THEN '❌ NO PRICE'
    ELSE '✅ ' || ROUND((ufi.quantity * ufi.current_price_per_unit)::numeric, 2) || ' PLN'
  END as item_value,
  CASE 
    WHEN ufi.expires_at < NOW() THEN '❌ EXPIRED'
    WHEN ufi.expires_at < NOW() + INTERVAL '2 days' THEN '⚠️ CRITICAL'
    WHEN ufi.expires_at < NOW() + INTERVAL '5 days' THEN '⚠️ WARNING'
    ELSE '✅ OK'
  END as status
FROM user_fridge_items ufi
JOIN "Ingredient" i ON ufi.ingredient_id = i.id
JOIN "User" u ON ufi.user_id = u.id
WHERE u.email = 'YOUR_EMAIL_HERE'  -- ← CHANGE THIS
ORDER BY ufi.expires_at NULLS LAST;

-- Expected: At least some items with prices and not expired
-- ❌ If all "NO PRICE" → User hasn't entered prices


-- ===================================================================
-- STEP 4: Simulate economy calculation for a user
-- ===================================================================
WITH user_products AS (
  SELECT
    i.name,
    ufi.quantity,
    ufi.current_price_per_unit,
    CASE 
      WHEN ufi.expires_at < NOW() THEN 0  -- expired = not used
      WHEN ufi.expires_at < NOW() + INTERVAL '5 days' THEN 1  -- priority
      ELSE 2  -- low priority
    END as priority,
    ROUND((ufi.quantity * COALESCE(ufi.current_price_per_unit, 0))::numeric, 2) as item_cost
  FROM user_fridge_items ufi
  JOIN "Ingredient" i ON ufi.ingredient_id = i.id
  JOIN "User" u ON ufi.user_id = u.id
  WHERE u.email = 'YOUR_EMAIL_HERE'  -- ← CHANGE THIS
    AND ufi.expires_at >= NOW()  -- not expired
    AND ufi.quantity > 0
)
SELECT
  COUNT(*) as total_items,
  COUNT(CASE WHEN current_price_per_unit > 0 THEN 1 END) as items_with_price,
  SUM(item_cost) as total_value_pln,
  json_agg(
    json_build_object(
      'name', name,
      'quantity', quantity,
      'pricePerUnit', current_price_per_unit,
      'cost', item_cost,
      'priority', priority
    ) ORDER BY priority, item_cost DESC
  ) as items_detail
FROM user_products;

-- Expected: total_value_pln > 0 if prices exist
-- ❌ If total_value_pln = 0 → Either no prices OR all prices = 0


-- ===================================================================
-- STEP 5: INSERT TEST DATA (if database is empty)
-- ===================================================================
-- First, get or create a test user
DO $$
DECLARE
  test_user_id UUID;
  ingredient_salt_id UUID;
  ingredient_milk_id UUID;
BEGIN
  -- Get test user ID (or use your real user)
  SELECT id INTO test_user_id 
  FROM "User" 
  WHERE email = 'pricetest@fodi.app'
  LIMIT 1;
  
  IF test_user_id IS NULL THEN
    RAISE NOTICE 'User not found. Please register first or change email.';
    RETURN;
  END IF;
  
  RAISE NOTICE 'Using user: %', test_user_id;
  
  -- Get ingredient IDs
  SELECT id INTO ingredient_salt_id FROM "Ingredient" WHERE LOWER(name) = 'sól' LIMIT 1;
  SELECT id INTO ingredient_milk_id FROM "Ingredient" WHERE LOWER(name) LIKE 'mleko%' LIMIT 1;
  
  -- Insert Sól (200g @ 2.50 PLN/kg = 0.0025 PLN/g)
  IF ingredient_salt_id IS NOT NULL THEN
    INSERT INTO user_fridge_items (
      id, user_id, ingredient_id, quantity, unit,
      current_price_per_unit, current_price_currency,
      price_updated_at, expires_at, created_at, updated_at
    ) VALUES (
      gen_random_uuid(),
      test_user_id,
      ingredient_salt_id,
      200,  -- 200g
      'g',
      0.0025,  -- 2.50 PLN/kg normalized to PLN/g
      'PLN',
      NOW(),
      NOW() + INTERVAL '365 days',
      NOW(),
      NOW()
    )
    ON CONFLICT (user_id, ingredient_id) 
    DO UPDATE SET
      quantity = user_fridge_items.quantity + 200,
      current_price_per_unit = 0.0025,
      price_updated_at = NOW(),
      updated_at = NOW();
    
    RAISE NOTICE '✅ Added/Updated: Sól (200g, 0.0025 PLN/g)';
  ELSE
    RAISE NOTICE '❌ Ingredient not found: Sól';
  END IF;
  
  -- Insert Mleko (1000ml @ 3.24 PLN/l = 0.00324 PLN/ml)
  IF ingredient_milk_id IS NOT NULL THEN
    INSERT INTO user_fridge_items (
      id, user_id, ingredient_id, quantity, unit,
      current_price_per_unit, current_price_currency,
      price_updated_at, expires_at, created_at, updated_at
    ) VALUES (
      gen_random_uuid(),
      test_user_id,
      ingredient_milk_id,
      1000,  -- 1000ml
      'ml',
      0.00324,  -- 3.24 PLN/l normalized to PLN/ml
      'PLN',
      NOW(),
      NOW() + INTERVAL '5 days',
      NOW(),
      NOW()
    )
    ON CONFLICT (user_id, ingredient_id)
    DO UPDATE SET
      quantity = user_fridge_items.quantity + 1000,
      current_price_per_unit = 0.00324,
      price_updated_at = NOW(),
      updated_at = NOW();
    
    RAISE NOTICE '✅ Added/Updated: Mleko (1000ml, 0.00324 PLN/ml)';
  ELSE
    RAISE NOTICE '❌ Ingredient not found: Mleko';
  END IF;
  
END $$;

-- Verify insert
SELECT
  i.name,
  ufi.quantity,
  ufi.unit,
  ufi.current_price_per_unit,
  ufi.current_price_currency
FROM user_fridge_items ufi
JOIN "Ingredient" i ON ufi.ingredient_id = i.id
JOIN "User" u ON ufi.user_id = u.id
WHERE u.email = 'pricetest@fodi.app'  -- ← CHANGE THIS IF NEEDED
ORDER BY i.name;


-- ===================================================================
-- STEP 6: Cleanup test data (optional)
-- ===================================================================
/*
DELETE FROM user_fridge_items
WHERE user_id = (SELECT id FROM "User" WHERE email = 'pricetest@fodi.app');
*/
