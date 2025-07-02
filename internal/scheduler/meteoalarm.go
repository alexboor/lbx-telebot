package scheduler

import (
	"encoding/json"
	"fmt"

	"github.com/alexboor/lbx-telebot/internal"
	"github.com/alexboor/lbx-telebot/internal/meteoalarm"
)

func (s *Schedule) MeteoalarmTask() {

	today, tomorrow, err := meteoalarm.Extract()
	if err != nil {
		fmt.Printf("error extracting meteoalarm data: %s\n", err)
		return
	}

	todayJSON, err := json.Marshal(today)
	if err != nil {
		fmt.Printf("error marshalling today slice to JSON: %s\n", err)
		return
	}

	s.Memory.Set(internal.MemkeyMeteoalarmToday, todayJSON)

	tomorrowJSON, err := json.Marshal(tomorrow)
	if err != nil {
		fmt.Printf("error marshalling tomorrow slice to JSON: %s\n", err)
		return
	}

	s.Memory.Set(internal.MemkeyMeteoalarmTomorrow, tomorrowJSON)

	return
}
