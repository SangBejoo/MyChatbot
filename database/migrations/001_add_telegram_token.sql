-- Migration: Add telegram_token column to users table
-- Run this if upgrading from an older version

-- Check if column exists and add if not
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'telegram_token'
    ) THEN
        ALTER TABLE users ADD COLUMN telegram_token TEXT;
        RAISE NOTICE 'Added telegram_token column to users table';
    ELSE
        RAISE NOTICE 'telegram_token column already exists';
    END IF;
END $$;

-- Verify the column was added
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'users';
