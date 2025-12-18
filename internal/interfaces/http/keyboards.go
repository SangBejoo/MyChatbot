package http

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CreateCategoryKeyboard creates inline keyboard buttons for product categories
func CreateCategoryKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 General", "cat_general"),
			tgbotapi.NewInlineKeyboardButtonData("👑 Luxury", "cat_luxury"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Export", "type_export"),
			tgbotapi.NewInlineKeyboardButtonData("📥 Import", "type_import"),
		),
	)
}

// CreateTypeKeyboard creates inline keyboard for export/import selection
func CreateTypeKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Export Products", "type_export"),
			tgbotapi.NewInlineKeyboardButtonData("📥 Import Products", "type_import"),
		),
	)
}

// CreateFollowUpMenu creates menu buttons after AI response
func CreateFollowUpMenu() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧮 Calculate Price", "action_calculate"),
			tgbotapi.NewInlineKeyboardButtonData("❓ Ask More", "action_ask"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Back to Menu", "action_menu"),
		),
	)
}
