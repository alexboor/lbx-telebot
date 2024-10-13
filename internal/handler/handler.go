package handler

import (
	"github.com/alexboor/lbx-telebot/internal/cfg"
	"github.com/alexboor/lbx-telebot/internal/storage"
	"github.com/alexboor/lbx-telebot/internal/storage/memory"
)

type Handler struct {
	Config  *cfg.Cfg
	Storage storage.Storage
	Memory  *memory.InMemoryStorage
}

// New create and return the handler instance
func New(storage storage.Storage, mem *memory.InMemoryStorage, cfg *cfg.Cfg) *Handler {
	return &Handler{
		Storage: storage,
		Memory:  mem,
		Config:  cfg,
	}
}
