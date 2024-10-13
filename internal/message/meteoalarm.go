package message

import (
	"fmt"
	"github.com/alexboor/lbx-telebot/internal/meteoalarm"
	"strings"
)

func GetMeteoAlarm(today []meteoalarm.Alert, tomorrow []meteoalarm.Alert) (string, bool, string, bool) {

	regionAlerts := make(map[string][]string)

	// Processing today (d0)
	for _, alert := range today {
		regionAlerts[alert.Region] = append(regionAlerts[alert.Region], fmt.Sprintf("%s %s", alertColor(alert.Level), alert.Text))
	}

	region001 := strings.Join(regionAlerts["001"], "\n")
	region002 := strings.Join(regionAlerts["002"], "\n")
	region003 := strings.Join(regionAlerts["003"], "\n")

	var d0 strings.Builder
	var d0alert bool

	d0.WriteString("TODAY\n\n")

	if !strings.Contains(region001, "No alert") {
		d0.WriteString("*Continental and Mountains*\n")
		d0.WriteString(region001)
		d0alert = true
	}
	if !strings.Contains(region002, "No alert") {
		d0.WriteString("\n\n*Central*\n")
		d0.WriteString(region002)
		d0alert = true
	}
	if !strings.Contains(region003, "No alert") {
		d0.WriteString("\n\n*Adriatic coast*\n")
		d0.WriteString(region003)
		d0alert = true
	}

	// Processing tomorrow (d1)
	regionAlerts = make(map[string][]string)

	for _, alert := range tomorrow {
		regionAlerts[alert.Region] = append(regionAlerts[alert.Region], fmt.Sprintf("%s %s", alertColor(alert.Level), alert.Text))
	}

	region001 = strings.Join(regionAlerts["001"], "\n")
	region002 = strings.Join(regionAlerts["002"], "\n")
	region003 = strings.Join(regionAlerts["003"], "\n")

	var d1 strings.Builder
	var d1alert bool

	d1.WriteString("TOMORROW\n\n")

	if !strings.Contains(region001, "No alert") {
		d1.WriteString("*Continental and Mountains*\n")
		d1.WriteString(region001)
		d1alert = true
	}
	if !strings.Contains(region002, "No alert") {
		d1.WriteString("\n\n*Central*\n")
		d1.WriteString(region002)
		d1alert = true
	}
	if !strings.Contains(region003, "No alert") {
		d1.WriteString("\n\n*Adriatic coast*\n")
		d1.WriteString(region003)
		d1alert = true
	}

	return d0.String(), d0alert, d1.String(), d1alert
}

func alertColor(level string) string {
	switch level {
	case "green":
		return "🟢"
	case "yellow":
		return "🟡"
	case "orange":
		return "🟠"
	case "red":
		return "🔴"
	default:
		return level
	}
}
