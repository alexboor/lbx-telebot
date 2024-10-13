package internal

import "time"

const (
	VersionFile = "app.version"

	Timeout     = 10 * time.Second
	RatingLimit = 5
	MarkdownOpt = "Markdown" // TODO change to v2

	ShareBtn = "share"

	HelpCmd    = "/help"
	HCmd       = "/h"
	StartCmd   = "/start"
	VerCmd     = "/ver"
	VCmd       = "/v"
	TopCmd     = "/top"
	BottomCmd  = "/bottom"
	ProfileCmd = "/profile"
	TopicCmd   = "/topic"
	EventCmd   = "/event"
	TodayCmd   = "/today"
	MeteoAlarm = "/meteoalarm"
)

const (
	MemkeyMeteoalarmToday    = "meteoalarm_today"
	MemkeyMeteoalarmTomorrow = "meteoalarm_tomorrow"
)
