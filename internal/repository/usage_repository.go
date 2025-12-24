package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsageRepository struct {
	db *pgxpool.Pool
}

type DailyUsage struct {
	Date             time.Time `json:"date"`
	MessagesSent     int       `json:"messages_sent"`
	MessagesReceived int       `json:"messages_received"`
}

type UserQuotaStatus struct {
	DailyLimit      int `json:"daily_limit"`
	MonthlyLimit    int `json:"monthly_limit"`
	TodaySent       int `json:"today_sent"`
	MonthSent       int `json:"month_sent"`
	DailyRemaining  int `json:"daily_remaining"`
	MonthlyRemaining int `json:"monthly_remaining"`
	DailyPercent    int `json:"daily_percent"`
	MonthlyPercent  int `json:"monthly_percent"`
}

func NewUsageRepository(db *pgxpool.Pool) *UsageRepository {
	return &UsageRepository{db: db}
}

// IncrementSent increments messages_sent for today
func (r *UsageRepository) IncrementSent(userID int) error {
	today := time.Now().Format("2006-01-02")
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO message_usage (user_id, date, messages_sent, messages_received)
		VALUES ($1, $2, 1, 0)
		ON CONFLICT (user_id, date) 
		DO UPDATE SET messages_sent = message_usage.messages_sent + 1
	`, userID, today)
	return err
}

// IncrementReceived increments messages_received for today
func (r *UsageRepository) IncrementReceived(userID int) error {
	today := time.Now().Format("2006-01-02")
	_, err := r.db.Exec(context.Background(), `
		INSERT INTO message_usage (user_id, date, messages_sent, messages_received)
		VALUES ($1, $2, 0, 1)
		ON CONFLICT (user_id, date) 
		DO UPDATE SET messages_received = message_usage.messages_received + 1
	`, userID, today)
	return err
}

// GetTodayUsage returns today's message count
func (r *UsageRepository) GetTodayUsage(userID int) (sent, received int, err error) {
	today := time.Now().Format("2006-01-02")
	err = r.db.QueryRow(context.Background(), `
		SELECT COALESCE(messages_sent, 0), COALESCE(messages_received, 0) 
		FROM message_usage WHERE user_id = $1 AND date = $2
	`, userID, today).Scan(&sent, &received)
	if err != nil {
		return 0, 0, nil // No record means 0 usage
	}
	return sent, received, nil
}

// GetMonthUsage returns this month's total message count
func (r *UsageRepository) GetMonthUsage(userID int) (sent, received int, err error) {
	firstOfMonth := time.Now().Format("2006-01") + "-01"
	err = r.db.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(messages_sent), 0), COALESCE(SUM(messages_received), 0) 
		FROM message_usage WHERE user_id = $1 AND date >= $2
	`, userID, firstOfMonth).Scan(&sent, &received)
	return sent, received, err
}

// GetUsageHistory returns last N days of usage
func (r *UsageRepository) GetUsageHistory(userID int, days int) ([]DailyUsage, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	rows, err := r.db.Query(context.Background(), `
		SELECT date, messages_sent, messages_received 
		FROM message_usage 
		WHERE user_id = $1 AND date >= $2
		ORDER BY date ASC
	`, userID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usage := []DailyUsage{}
	for rows.Next() {
		var u DailyUsage
		if err := rows.Scan(&u.Date, &u.MessagesSent, &u.MessagesReceived); err != nil {
			return nil, err
		}
		usage = append(usage, u)
	}
	return usage, nil
}

// GetQuotaStatus returns comprehensive quota status for a user
func (r *UsageRepository) GetQuotaStatus(userID int, dailyLimit, monthlyLimit int) (*UserQuotaStatus, error) {
	todaySent, _, _ := r.GetTodayUsage(userID)
	monthSent, _, _ := r.GetMonthUsage(userID)

	status := &UserQuotaStatus{
		DailyLimit:   dailyLimit,
		MonthlyLimit: monthlyLimit,
		TodaySent:    todaySent,
		MonthSent:    monthSent,
	}

	// Calculate remaining
	if dailyLimit > 0 {
		status.DailyRemaining = dailyLimit - todaySent
		if status.DailyRemaining < 0 {
			status.DailyRemaining = 0
		}
		status.DailyPercent = (todaySent * 100) / dailyLimit
		if status.DailyPercent > 100 {
			status.DailyPercent = 100
		}
	} else {
		status.DailyRemaining = -1 // Unlimited
		status.DailyPercent = 0
	}

	if monthlyLimit > 0 {
		status.MonthlyRemaining = monthlyLimit - monthSent
		if status.MonthlyRemaining < 0 {
			status.MonthlyRemaining = 0
		}
		status.MonthlyPercent = (monthSent * 100) / monthlyLimit
		if status.MonthlyPercent > 100 {
			status.MonthlyPercent = 100
		}
	} else {
		status.MonthlyRemaining = -1 // Unlimited
		status.MonthlyPercent = 0
	}

	return status, nil
}

// CanSendMessage checks if user can send a message based on quotas
func (r *UsageRepository) CanSendMessage(userID int, dailyLimit, monthlyLimit int) (bool, string) {
	todaySent, _, _ := r.GetTodayUsage(userID)
	monthSent, _, _ := r.GetMonthUsage(userID)

	if dailyLimit > 0 && todaySent >= dailyLimit {
		return false, "Daily message limit reached"
	}
	if monthlyLimit > 0 && monthSent >= monthlyLimit {
		return false, "Monthly message limit reached"
	}
	return true, ""
}
