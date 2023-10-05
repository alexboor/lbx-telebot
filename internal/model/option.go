package model

import (
	"strconv"
	"strings"
	"time"
)

type Option struct {
	Date    time.Time
	Limit   int
	Profile string
}

func NewProfileOption(pld string) (Option, bool) {
	if len(pld) == 0 {
		return Option{}, false
	}

	result := Option{}
	opts := strings.Split(pld, " ")
	if len(opts) == 1 {
		indexName := findIndexName(opts)
		if indexName != -1 {
			result.Profile = opts[indexName]
		} else {
			result.Date = parseDuration(opts[0])
			if result.Date.IsZero() {
				return Option{}, false
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
				return Option{}, false
			}
		} else {
			return Option{}, false
		}
	} else if len(opts) >= 3 {
		return Option{}, false
	}

	return result, true
}

func NewRatingOption(pld string) (Option, bool) {
	var result Option

	if len(pld) == 0 {
		return result, false
	}

	opts := strings.Split(pld, " ")

	if len(opts) == 1 {
		return parseSingleOpt(opts[0])
	} else if len(opts) == 2 {
		return parseTwoOpts(opts)
	}

	return result, false
}

func parseSingleOpt(opt string) (Option, bool) {
	result := Option{}
	result.Limit = parseInt(opt)
	if result.Limit == 0 {
		result.Date = parseDuration(opt)
		if result.Date.IsZero() {
			return Option{}, false
		}
	}
	return result, true
}

func parseTwoOpts(opts []string) (Option, bool) {
	result := Option{}

	result.Limit = parseInt(opts[0])
	if result.Limit == 0 {
		result.Limit = parseInt(opts[1])
		if result.Limit == 0 {
			return Option{}, false
		}
		result.Date = parseDuration(opts[0])
		if result.Date.IsZero() {
			return Option{}, false
		}
	} else {
		result.Date = parseDuration(opts[1])
		if result.Date.IsZero() {
			return Option{}, false
		}
	}

	return result, true
}

func parseInt(str string) int {
	result, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return int(result)
}

func findIndexName(opts []string) int {
	for i, opt := range opts {
		if parseDuration(opt).IsZero() && parseInt(opt) == 0 {
			return i
		}
	}
	return -1
}

func parseDuration(str string) time.Time {
	var (
		result time.Time
		period time.Duration
		err    error
	)

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

	if period < 0 {
		return time.Time{}
	}

	result = time.Now().Add(-1 * period).Truncate(time.Hour * 24)
	return result
}
