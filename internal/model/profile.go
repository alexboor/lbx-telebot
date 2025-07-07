package model

import (
	"strings"
	"time"

	"gopkg.in/telebot.v3"
)

type Profile struct {
	Id        int64
	FirstName string
	LastName  string
	UserName  string
	Position  int
	Count     Count
}

type ProfileWithScore struct {
	Profile
	Score      int
	LastUpdate time.Time
}

// NewProfile creates new Profile by given id, username, first name and last name
func NewProfile(user *telebot.User) Profile {
	return Profile{
		Id:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		UserName:  strings.ToLower(user.Username),
	}
}
