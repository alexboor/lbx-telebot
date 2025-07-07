package score

import (
	"context"
	"fmt"
	"time"

	"github.com/alexboor/lbx-telebot/internal/message"
	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/alexboor/lbx-telebot/internal/storage"
)

// ShowScore shows a scoreboard with 10 top users ordered by score descending
func ShowScores10(s storage.Storage) string {
	ctx := context.Background()
	scores, err := s.GetAllScores(ctx)
	if err != nil {
		fmt.Printf("Error getting scoreboard: %v", err)
		return "Opps, something wrong"
	}

	return message.GetScores(scores[:10])
}

// CalculateAllScore recalculates all score for all users in the group
func CalculateAllScore(s storage.Storage) string {

	fmt.Println("Calculating all score")

	ctx := context.Background()

	users, err := s.GetAllIds(ctx, -1001328533803) //TODO: get chat id from config
	if err != nil {
		fmt.Println("Error getting all ids", err)
	}

	var (
		inserted int
		errors   int
	)

	for _, user := range users {
		counts, err := s.GetAllCountsByUser(ctx, -1001328533803, user)
		if err != nil {
			fmt.Println("Error getting all counts by user", err)
		}

		score := calculateScore(counts)

		if err := s.StoreScore(ctx, user, score); err != nil {
			errors++
		} else {
			inserted++
		}
	}

	return fmt.Sprintf("Inserted: %d, Errors: %d", inserted, errors)
}

// calculateScore calculates the score for a user by given stats
// it recalculates all over all days stored in DB
// TODO: later add transaction to calculation
func calculateScore(counts []model.DateCount) int {
	if len(counts) == 0 {
		return 0
	}

	// Find all days between first and last date
	firstDate := counts[0].Date
	lastDate := counts[len(counts)-1].Date

	// Create a map to quickly check if a date exists in counts
	countsMap := make(map[time.Time]model.Count)
	for _, count := range counts {
		countsMap[count.Date] = count.Count
	}

	totalScore := 0

	// Iterate through each day from first to last
	for currentDate := firstDate; !currentDate.After(lastDate); currentDate = currentDate.AddDate(0, 0, 1) {
		if count, exists := countsMap[currentDate]; exists {
			// User has activity on this day
			dayScore := 0

			// 2.1. if user has at least 1 Count.Message, he gets 1 point
			if count.Message >= 1 && count.Message <= 100 {
				dayScore += 1
			}

			// 2.2. if user has more that 100 Count.Message, he gets 2 points
			if count.Message > 100 && count.Message <= 500 {
				dayScore += 2
			}

			// 2.3. if user has more than 500 Count.Message, he gets 5 points
			if count.Message > 500 && count.Message <= 1000 {
				dayScore += 5
			}

			// 2.3. if user has more than 1000 Count.Message, he gets 10 points
			if count.Message > 1000 {
				dayScore += 10
			}

			// 2.4. if user has Count.Forward, he gets equal number of points
			dayScore += count.Forward

			// 2.5. if user has Count.Media, he gets equal number of points
			dayScore += count.Media

			totalScore += dayScore
		} else {
			// 3. if there is no day in DateCount then user substract 1 point
			totalScore -= 1
		}
	}

	return totalScore
}
