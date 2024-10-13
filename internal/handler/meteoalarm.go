package handler

import (
	"encoding/json"
	"fmt"
	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/message"
	"github.com/alexboor/lbx-telebot/internal/meteoalarm"
	tele "gopkg.in/telebot.v3"
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

	var today []meteoalarm.Alert
	var tomorrow []meteoalarm.Alert

	if alertTodayBytes, ok := alertToday.([]byte); ok {
		if err := json.Unmarshal(alertTodayBytes, &today); err != nil {
			fmt.Printf("error unmarshalling today data: %s\n", err)
			return nil
		}
	} else {
		fmt.Println("Invalid data type for alertToday")
		return nil
	}

	if alertTomorrowBytes, ok := alertTomorrow.([]byte); ok {
		if err := json.Unmarshal(alertTomorrowBytes, &tomorrow); err != nil {
			fmt.Printf("error unmarshalling tomorrow data: %s\n", err)
			return nil
		}
	} else {
		fmt.Println("Invalid data type for alertTomorrow")
		return nil
	}

	d0, d0alarm, d1, d1alarm := message.GetMeteoAlarm(today, tomorrow)

	if d0alarm {
		_, err := c.Bot().Send(c.Sender(), d0, internal.MarkdownOpt)
		if err != nil {
			return err
		}
	}

	if d1alarm {
		_, err := c.Bot().Send(c.Sender(), d1, internal.MarkdownOpt)
		if err != nil {
			return err
		}
	}

	return nil
}
