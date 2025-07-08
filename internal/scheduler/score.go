package scheduler

import (
	"log/slog"

	"github.com/alexboor/lbx-telebot/internal/score"
)

func (s *Schedule) CleanupProfile() {
	msg := score.CleanupProfile(s.Storage, s.Config.ScoreTargetChat, s.Bot)
	slog.Info("cleanup profile job scheduled", "msg", msg)
}

func (s *Schedule) RecalculateScore() {
	msg := score.CalculateAllScore(s.Storage, s.Config.ScoreTargetChat)
	slog.Info("recalculated score job scheduled", "msg", msg)
}
