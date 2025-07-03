package score

import (
	"testing"
	"time"

	"github.com/alexboor/lbx-telebot/internal/model"
)

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		counts   []model.DateCount
		expected int
	}{
		{
			name:     "empty counts",
			counts:   []model.DateCount{},
			expected: 0,
		},
		{
			name: "single day with 1 message",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 1}},
			},
			expected: 1,
		},
		{
			name: "single day with 150 messages",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 150}},
			},
			expected: 2, // 2 for 101-500 messages (exclusive)
		},
		{
			name: "single day with 600 messages",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 600}},
			},
			expected: 5, // 5 for 501-1000 messages (exclusive)
		},
		{
			name: "single day with 1200 messages",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 1200}},
			},
			expected: 10, // 10 for >1000 messages
		},
		{
			name: "single day with forwards and media",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 5, Forward: 3, Media: 2}},
			},
			expected: 6, // 1 for 1-100 messages + 3 for forwards + 2 for media
		},
		{
			name: "multiple consecutive days with activity",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 10}},
				{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 20}},
				{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 30}},
			},
			expected: 3, // 1 point each day for >=1 message
		},
		{
			name: "multiple days with gaps (missing days)",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 10}},
				{Date: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 20}},
			},
			expected: 1, // 1 for day 1 + 1 for day 3 - 1 for missing day 2
		},
		{
			name: "complex scenario with all scoring elements",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 1200, Forward: 5, Media: 3}},
				{Date: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 50, Forward: 2, Media: 1}},
				{Date: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 600, Forward: 1, Media: 4}},
			},
			expected: 31, // Day 1: 10 (1200 msg) + 5 (forward) + 3 (media) = 18
			// Day 2: 1 (50 msg) + 2 (forward) + 1 (media) = 4
			// Day 3: -1 (missing day)
			// Day 4: 5 (600 msg) + 1 (forward) + 4 (media) = 10
			// Total: 18 + 4 - 1 + 10 = 31
		},
		{
			name: "zero messages but has forwards and media",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 0, Forward: 3, Media: 2}},
			},
			expected: 5, // 0 for messages + 3 for forwards + 2 for media
		},
		{
			name: "large gap between dates",
			counts: []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 10}},
				{Date: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: 20}},
			},
			expected: -6, // 1 for day 1 + 1 for day 10 - 8 for missing days (2-9)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateScore(tt.counts)
			if result != tt.expected {
				t.Errorf("calculateScore() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateScoreEdgeCases(t *testing.T) {
	t.Run("nil counts", func(t *testing.T) {
		result := calculateScore(nil)
		if result != 0 {
			t.Errorf("calculateScore(nil) = %v, want 0", result)
		}
	})

	t.Run("single day with all zero values", func(t *testing.T) {
		counts := []model.DateCount{
			{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{}},
		}
		result := calculateScore(counts)
		if result != 0 {
			t.Errorf("calculateScore() = %v, want 0", result)
		}
	})

	t.Run("exact threshold values", func(t *testing.T) {
		tests := []struct {
			messages int
			expected int
		}{
			{0, 0},
			{1, 1},
			{100, 1},
			{101, 2},
			{500, 2},
			{501, 5},
			{1000, 5},
			{1001, 10},
		}

		for _, tt := range tests {
			counts := []model.DateCount{
				{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Count: model.Count{Message: tt.messages}},
			}
			result := calculateScore(counts)
			if result != tt.expected {
				t.Errorf("calculateScore(messages=%d) = %v, want %v", tt.messages, result, tt.expected)
			}
		}
	})
}
