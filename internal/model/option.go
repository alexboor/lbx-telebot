package model

import "time"

type Option struct {
	Date    time.Time
	Limit   int
	Profile string
}
