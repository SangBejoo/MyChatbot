# Chatbot Backend with Clean Architecture

This is a Golang backend for a multi-platform chatbot using Clean Architecture. It integrates with Telegram Bot API for free messaging and uses Gemini API for AI responses.

## Features
- Clean Architecture: Entities, Use Cases, Interfaces, Infrastructure
- Multi-platform: Telegram (free), Web (easily extensible to WhatsApp, Discord)
- Async processing for fast responses
- AI integration with Google Gemini

## Setup
1. Clone and `go mod tidy`
2. Get Telegram Bot Token: Talk to @BotFather on Telegram, create a bot, get the token.
3. Create a `.env` file in the root directory with:
   ```
   GEMINI_API_KEY=your_gemini_api_key_here
   ```
4. Run `go run cmd/main.go`
5. For web integration: POST to `/webhook/web` with JSON { "from": "user", "content": "message" }

## Usage
- Telegram: Start a chat with your bot (@wwg_adeBot), send messages.
- Web: POST to `/webhook/web` with JSON { "from": "user", "content": "message" }

## Extending
Add new messengers by implementing the Messenger interface.