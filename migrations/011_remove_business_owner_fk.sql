-- Удаляем foreign key constraint на owner_id в таблице Business
ALTER TABLE "Business" DROP CONSTRAINT IF EXISTS "fk_Business_owner";

-- Делаем owner_id nullable
ALTER TABLE "Business" ALTER COLUMN "owner_id" DROP NOT NULL;
