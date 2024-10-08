package handler

import (
	"fmt"
	"github.com/alexboor/lbx-telebot/internal"
	tele "gopkg.in/telebot.v3"
	"strings"
)

func (h Handler) MeteoAlarm(c tele.Context) error {

	alertToday, exists := h.Memory.Get(internal.MemkeyMeteoalarmToday)
	if !exists {
		fmt.Println("No data for today")
		return nil
	}

	alertTomorrow, exists := h.Memory.Get(internal.MemkeyMeteoalarmTomorrow)
	if !exists {
		fmt.Println("No data for tomorrow")
		return nil
	}

	var today strings.Builder
	today.WriteString("Weather alert for today: ")

	if alertTodayBytes, ok := alertToday.([]byte); ok {
		today.Write(alertTodayBytes)
	} else {
		fmt.Println("Invalid data type for alertToday")
		return nil
	}

	var tomorrow strings.Builder
	tomorrow.WriteString("Weather alert for tomorrow: ")

	if alertTomorrowBytes, ok := alertTomorrow.([]byte); ok {
		tomorrow.Write(alertTomorrowBytes)
	} else {
		fmt.Println("Invalid data type for alertTomorrow")
		return nil
	}

	c.Reply(today.String())
	c.Reply(tomorrow.String())

	return nil
}
