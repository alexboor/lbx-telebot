package handler

import (
	"github.com/alexboor/lbx-telebot/internal/cfg"
	"github.com/alexboor/lbx-telebot/internal/storage"
)

type Handler struct {
	Config  *cfg.Cfg
	Storage storage.Storage
}

// New create and return the handler instance
func New(storage storage.Storage, cfg *cfg.Cfg) *Handler {
	return &Handler{
		Storage: storage,
		Config:  cfg,
	}
}
