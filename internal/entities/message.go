package entities

type Message struct {
	ID         string
	From       string
	To         string
	Content    string
	Platform   string // e.g., "whatsapp", "web", "telegram"
	AIContext  string // Context from CSV/data for RAG
	IsCallback bool   // Whether this is from a button callback
}

type Response struct {
	Content string
}