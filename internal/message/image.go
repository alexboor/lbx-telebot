package message

import (
	"fmt"
	"github.com/alexboor/lbx-telebot/internal/model"
	"github.com/wcharczuk/go-chart/v2"
	"os"
)

func GenerateProfileRatingImage(profile model.Profile, opt model.Option) (string, error) {
	title := getProfileTitle(profile, opt)
	avg := profile.Count.GetAvgStatistic()
	bar := chart.BarChart{
		Title: title,
		DPI:   100,
		Background: chart.Style{
			Padding: chart.Box{
				Top: 60,
			},
		},
		Height:   710,
		BarWidth: 100,
		YAxis:    chart.YAxis{Style: chart.Style{Hidden: true}, Range: nil},
		Bars: []chart.Value{
			{Value: float64(profile.Count.Word), Label: fmt.Sprintf("%v\nWords", profile.Count.Word)},
			{Value: float64(profile.Count.Message), Label: fmt.Sprintf("%v\nMessages", profile.Count.Message)},
			{Value: float64(profile.Count.Reply), Label: fmt.Sprintf("%v\nReplies", profile.Count.Reply)},
			{Value: float64(profile.Count.Forward), Label: fmt.Sprintf("%v\nForwards", profile.Count.Forward)},
			{Value: float64(profile.Count.Sticker), Label: fmt.Sprintf("%v\nStickers", profile.Count.Sticker)},
			{Value: float64(profile.Count.Media), Label: fmt.Sprintf("%v\nMedia", profile.Count.Media)},
			{Value: avg, Label: fmt.Sprintf("%v\ntotal/messages", avg)},
		},
	}

	filename := fmt.Sprintf("%v.png", profile.Id)
	f, _ := os.Create(filename)
	defer func() { _ = f.Close() }()
	if err := bar.Render(chart.PNG, f); err != nil {
		return "", err
	}
	return filename, nil
}
