package message

import (
	"fmt"
	"strings"

	"github.com/alexboor/lbx-telebot/internal/model"
)

// CreateRating returns message with information about given profiles
func CreateRating(profiles []model.Profile) string {
	var msg strings.Builder
	for _, profile := range profiles {
		if msg.Len() != 0 {
			msg.WriteString("\n")
		}
		msg.WriteString(fmt.Sprintf("%v. %v: %v",
			profile.Position, getName(profile), profile.Count))
	}

	return msg.String()
}

// CreateUserCount returns information about given profile
func CreateUserCount(profile model.Profile) string {
	return fmt.Sprintf("%v: %v", getName(profile), profile.Count)
}

// getName checks given profile and returns name
//
//	if there is no first name or last name returns only username
//	if there is no username, but there is first name or last name returns first name + last name
//	otherwise returns 'unknown' + id
func getName(profile model.Profile) string {
	var name string

	switch {

	case len(profile.UserName) != 0 && (len(profile.FirstName) == 0 || len(profile.LastName) == 0):
		name = fmt.Sprintf("@%v", profile.UserName)

	case len(profile.UserName) != 0 && len(profile.FirstName) != 0 && len(profile.LastName) != 0:
		name = strings.TrimSpace(fmt.Sprintf("%v @%v %v", profile.FirstName, profile.UserName, profile.LastName))

	case len(profile.UserName) == 0 && (len(profile.FirstName) != 0 || len(profile.LastName) != 0):
		name = strings.TrimSpace(fmt.Sprintf("%v %v", profile.FirstName, profile.LastName))

	case len(profile.UserName) == 0 && len(profile.FirstName) == 0 && len(profile.LastName) == 0:
		name = fmt.Sprintf("unknown %v", profile.Id)
	}

	return name
}
