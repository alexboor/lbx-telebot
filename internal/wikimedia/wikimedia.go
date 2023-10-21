package wikimedia

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	Holiday = iota
	Birthday
	Event

	wikimediaUrl = "https://api.wikimedia.org/feed/v1/wikipedia/en/onthisday/all"
)

var onThisDayList map[time.Time][]OnThisDay

type (
	onThisDayRaw struct {
		Births []struct {
			Text string `json:"text"`
			Year int    `json:"year"`
		} `json:"births"`
		Events []struct {
			Text string `json:"text"`
			Year int    `json:"year"`
		} `json:"events"`
		Holidays []struct {
			Text string `json:"text"`
		} `json:"holidays"`
	}

	OnThisDay struct {
		Text string
		Year int
		Type int
	}
)

func GetOnThisDay() (OnThisDay, error) {
	today := time.Now().Truncate(time.Hour * 24)
	events, ok := onThisDayList[today]
	if !ok {
		if err := setOnThisDayList(); err != nil {
			return OnThisDay{}, err
		}
		events = onThisDayList[today]
	}

	if len(events) == 0 {
		return OnThisDay{}, errors.New("failed to get events list")
	}

	event := events[getInt(len(events))]
	return event, nil
}

func setOnThisDayList() error {
	day := time.Now().Format("02")
	month := time.Now().Format("01")
	u := fmt.Sprintf("%s/%s/%s", wikimediaUrl, month, day)
	resp, err := resty.New().R().Get(u)
	if err != nil {
		return fmt.Errorf("failed to make query to endpoint=%s: %w", u, err)
	} else if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to get response from endpoint=%s, status_code=%d, body=%v",
			u, resp.StatusCode(), string(resp.Body()))
	}

	var raw onThisDayRaw
	if err := json.Unmarshal(resp.Body(), &raw); err != nil {
		return fmt.Errorf("failed to parse data from endpoint=%s: %w", u, err)
	}

	today := time.Now().Truncate(time.Hour * 24)
	onThisDayList = map[time.Time][]OnThisDay{today: {}}

	for _, v := range raw.Holidays {
		otd := OnThisDay{Text: v.Text, Type: Holiday}
		onThisDayList[today] = append(onThisDayList[today], otd)
	}

	for _, v := range raw.Births {
		otd := OnThisDay{Text: v.Text, Year: v.Year, Type: Birthday}
		onThisDayList[today] = append(onThisDayList[today], otd)
	}

	for _, v := range raw.Events {
		otd := OnThisDay{Text: v.Text, Year: v.Year, Type: Event}
		onThisDayList[today] = append(onThisDayList[today], otd)
	}

	return nil
}
