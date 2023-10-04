package model

import (
	"gopkg.in/telebot.v3"
	"strings"
)

type Profile struct {
	Id        int64
	FirstName string
	LastName  string
	UserName  string
	Position  int
	Count     Count
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

func (p Profile) GetCountTotal() int {
	return p.Count.Word + p.Count.Media + p.Count.Reply + p.Count.Sticker + p.Count.Forward
}
