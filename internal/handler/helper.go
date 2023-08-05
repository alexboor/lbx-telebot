package handler

import (
	"github.com/alexboor/lbx-telebot/internal/model"
	"strconv"
	"strings"
	"time"
)

func parseProfilePayload(payload string) (model.Option, bool) {
	if len(payload) == 0 {
		return model.Option{}, false
	}

	opts := strings.Split(payload, " ")

	result := model.Option{}
	if len(opts) == 1 {
		indexName := findIndexName(opts)
		if indexName != -1 {
			result.Profile = opts[indexName]
		} else {
			result.Date = parseDuration(opts[0])
			if result.Date.IsZero() {
				return model.Option{}, false
			}
		}
	} else if len(opts) == 2 {
		var indexDuration int
		indexName := findIndexName(opts)

		if indexName != -1 {
			result.Profile = opts[indexName]
			if indexName == 0 {
				indexDuration = 1
			}

			result.Date = parseDuration(opts[indexDuration])
			if result.Date.IsZero() {
				return model.Option{}, false
			}
		} else {
			return model.Option{}, false
		}
	} else if len(opts) >= 3 {
		return model.Option{}, false
	}

	return result, true
}

func parseTopAndBottomPayload(payload string) (model.Option, bool) {
	if len(payload) == 0 {
		return model.Option{}, false
	}

	opts := strings.Split(payload, " ")

	if len(opts) == 1 {
		return parseSingleOpt(opts[0])
	} else if len(opts) == 2 {
		return parseTwoOpts(opts)
	}

	return model.Option{}, false
}

func parseSingleOpt(opt string) (model.Option, bool) {
	result := model.Option{}
	result.Limit = parseInt(opt)
	if result.Limit == 0 {
		result.Date = parseDuration(opt)
		if result.Date.IsZero() {
			return model.Option{}, false
		}
	}
	return result, true
}

func parseTwoOpts(opts []string) (model.Option, bool) {
	result := model.Option{}

	result.Limit = parseInt(opts[0])
	if result.Limit == 0 {
		result.Limit = parseInt(opts[1])
		if result.Limit == 0 {
			return model.Option{}, false
		}
		result.Date = parseDuration(opts[0])
		if result.Date.IsZero() {
			return model.Option{}, false
		}
	} else {
		result.Date = parseDuration(opts[1])
		if result.Date.IsZero() {
			return model.Option{}, false
		}
	}

	return result, true
}

func parseDuration(str string) time.Time {
	var result time.Time

	var period time.Duration
	var err error
	if strings.Contains(str, "d") && strings.Count(str, "d") == 1 {
		str = strings.ReplaceAll(str, "d", "")
		days, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return time.Time{}
		}
		period = time.Hour * 24 * time.Duration(days)
	} else {
		period, err = time.ParseDuration(str)
		if err != nil {
			return time.Time{}
		}
	}

	result = time.Now().Add(-1 * period).Truncate(time.Hour * 24)
	return result
}

func parseInt(str string) int {
	result, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return int(result)
}

func findIndexName(opts []string) int {
	for i, o := range opts {
		if parseDuration(o).IsZero() && parseInt(o) == 0 {
			return i
		}
	}
	return -1
}
