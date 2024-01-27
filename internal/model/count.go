package model

import "math"

type Count struct {
	Word    int
	Reply   int
	Forward int
	Media   int
	Sticker int
	Message int
	Total   int
}

func (c Count) GetAvgStatistic() float64 {
	if c.Message == 0 {
		c.Message = 1
	}
	avg := float64(c.Word+c.Media+c.Reply+c.Sticker+c.Forward) / float64(c.Message)
	return math.Round(avg*100) / 100
}
