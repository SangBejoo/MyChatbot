package http

import (
	"project_masAde/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GenerateDynamicKeyboard creates keyboard from dynamic menu items
func GenerateDynamicKeyboard(items []repository.MenuItem) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	// Create 2 buttons per row for better mobile layout
	var row []tgbotapi.InlineKeyboardButton
	for i, item := range items {
		// Callback data format: "action_" + Action + "_" + Payload (max 64 chars)
		// To save space, let's use: "a:" + Action + ":" + Payload? 
		// Or if Action is "view_table", payload is "dt_...". 
		// Our Callback handler expects "action_", "cat_", "type_".
		// We need to update main.go handler to be generic. 
		// For now, let's format data as "dyn:" + item.Action + ":" + item.Payload
		
		btnData := "dyn:" + item.Action + ":" + item.Payload
		
		btn := tgbotapi.NewInlineKeyboardButtonData(item.Label, btnData)
		row = append(row, btn)

		if (i+1)%2 == 0 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

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
