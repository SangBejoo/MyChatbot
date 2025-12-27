-- PostgreSQL Schema for Project MasAde
-- Bot Management Platform with Multi-Tenant Support

-- =====================================================
-- USERS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    schema_name VARCHAR(100),           -- Tenant schema name
    is_active BOOLEAN DEFAULT true,     -- Account enabled
    wa_enabled BOOLEAN DEFAULT true,    -- WhatsApp enabled
    telegram_token TEXT,                -- User's Telegram bot token
    daily_limit INTEGER DEFAULT 200,    -- Max messages per day
    monthly_limit INTEGER DEFAULT 5000, -- Max messages per month
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_schema ON users(schema_name);

-- =====================================================
-- BOT CONFIGURATIONS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS bot_config (
    id SERIAL PRIMARY KEY,
    schema_name VARCHAR(100) DEFAULT 'public',
    key VARCHAR(100) NOT NULL,
    value TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(schema_name, key)
);

-- Default configs
INSERT INTO bot_config (key, value) VALUES 
    ('welcome_message', 'Welcome! 👋 How can I help you?'),
    ('default_reply', 'I did not understand that. Please try again.'),
    ('ai_system_prompt', 'You are a helpful assistant. Answer based on the provided context only.')
ON CONFLICT (schema_name, key) DO NOTHING;

-- =====================================================
-- MENUS TABLE (Dynamic Bot Menus)
-- =====================================================
CREATE TABLE IF NOT EXISTS menus (
    id SERIAL PRIMARY KEY,
    schema_name VARCHAR(100) DEFAULT 'public',
    slug VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    items JSONB DEFAULT '[]',           -- Array of menu items
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(schema_name, slug)
);

-- Index for menu lookups
CREATE INDEX IF NOT EXISTS idx_menus_schema_slug ON menus(schema_name, slug);

-- =====================================================
-- DYNAMIC TABLES REGISTRY
-- =====================================================
CREATE TABLE IF NOT EXISTS dynamic_tables (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(255) NOT NULL,   -- Actual table name (dt_xxx_timestamp)
    display_name VARCHAR(255) NOT NULL, -- User-friendly name
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_name)
);

-- Index for table lookups
CREATE INDEX IF NOT EXISTS idx_dynamic_tables_display ON dynamic_tables(display_name);

-- =====================================================
-- PRODUCTS TABLE - DEPRECATED
-- Use dynamic_tables instead for product data
-- Import products via CSV in Data Manager
-- =====================================================
-- (Removed - use dynamic datasets)

-- =====================================================
-- MESSAGE USAGE TRACKING
-- =====================================================
CREATE TABLE IF NOT EXISTS message_usage (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    message_count INTEGER DEFAULT 0,
    UNIQUE(user_id, date)
);

-- Index for usage lookups
CREATE INDEX IF NOT EXISTS idx_usage_user_date ON message_usage(user_id, date);

-- =====================================================
-- CONVERSATION LOGS (Optional - for analytics)
-- =====================================================
CREATE TABLE IF NOT EXISTS conversation_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    platform VARCHAR(20), -- 'telegram', 'whatsapp', 'web'
    chat_id VARCHAR(100),
    message_type VARCHAR(20), -- 'incoming', 'outgoing'
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for log queries
CREATE INDEX IF NOT EXISTS idx_logs_user ON conversation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_logs_created ON conversation_logs(created_at);
