# Database Setup Instructions

## PostgreSQL Database

### Quick Setup (New Installation)

1. Create database:
```bash
createdb -U postgres postgres
```

2. Run schema:
```bash
psql -U postgres -d postgres -f database/schema.sql
```

### Migrations (Existing Database)

Run migrations in order:

```bash
psql -U postgres -d postgres -f database/migrations/001_add_telegram_token.sql
```

### Environment Variables

Make sure your `.env` has:
```env
DATABASE_URL=postgres://postgres:your_password@localhost:5432/postgres?sslmode=disable
```

### Tables Overview

| Table | Description |
|-------|-------------|
| `users` | User accounts with roles, limits, tokens |
| `bot_config` | Bot configuration (per-tenant) |
| `menus` | Dynamic bot menus and buttons |
| `dynamic_tables` | Registry of imported CSV tables |
| `products` | Legacy product catalog |
| `message_usage` | Daily message tracking |
| `conversation_logs` | Chat history (optional) |
