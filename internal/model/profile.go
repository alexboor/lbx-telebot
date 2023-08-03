package model

import "strings"

type Profile struct {
	Id        int64
	FirstName string
	LastName  string
	UserName  string
	Position  int
	Count     int
}

// NewProfile creates new Profile by given id, username, first name and last name
func NewProfile(id int64, user, first, last string) Profile {
	return Profile{
		Id:        id,
		FirstName: first,
		LastName:  last,
		UserName:  strings.ToLower(user),
	}
}
