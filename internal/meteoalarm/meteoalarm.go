package meteoalarm

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const meteoalarmURL = "https://www.meteo.co.me/page.php?id=10"

type Alert struct {
	Region string
	Level  string
	Text   string
}

func Extract() ([]Alert, []Alert, error) {
	page, err := getRawHTML(meteoalarmURL)
	if err != nil {
		fmt.Errorf("error getting content: %s\n", err)
		return nil, nil, err
	}

	scripts, err := extractScriptContent(page)
	if err != nil {
		fmt.Errorf("error extract script content: %s\n", err)
		return nil, nil, err
	}

	var today []Alert
	var tomorrow []Alert

	for _, s := range scripts {
		norm := normScriptContent(s)
		data := removeUselessLines(norm)

		t := isToday(s)
		for _, l := range strings.Split(data, "\n") {
			alert := extractAlert(l)
			if len(alert.Region) == 3 {
				if t {
					today = append(today, alert)
				} else {
					tomorrow = append(tomorrow, alert)
				}
			}

		}
	}

	return today, tomorrow, nil
}

func getRawHTML(url string) (string, error) {
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: customTransport,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	return string(body), nil
}

func extractScriptContent(htmlContent string) ([]string, error) {
	// Define a regular expression to find content between <script> tags
	re := regexp.MustCompile(`(?s)<script.*?>(.*?)</script>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var scriptContents []string
	for _, match := range matches {
		if len(match) > 1 {
			scriptContents = append(scriptContents, match[1])
		}
	}

	return scriptContents, nil
}

// extractIDs extracts IDs from script content
func extractIDs(script string) []string {
	re := regexp.MustCompile(`document\.getElementById\("([^"]+)"\)`)
	matches := re.FindAllStringSubmatch(script, -1)
	var ids []string
	for _, match := range matches {
		if len(match) > 1 {
			ids = append(ids, match[1])
		}
	}
	return ids
}

func isToday(s string) bool {
	ids := extractIDs(s)
	for _, id := range ids {
		if strings.Contains(id, "D1") {
			return true
		}
	}
	return false
}

// normScriptContent normalise given script content by removing whitespaces and adding newlines
func normScriptContent(s string) string {
	res := strings.ReplaceAll(s, "\n", "")
	res = strings.ReplaceAll(res, "\t", "")

	re := regexp.MustCompile(`;\s*`)
	res = re.ReplaceAllString(res, ";")

	res = strings.ReplaceAll(res, ";", ";\n")
	res = strings.ReplaceAll(res, ";\n\\", ";\\")

	return res
}

// removeUselessLines removes useless lines from the given string
//
//	for the extraction only lines containing innerHTML are useful
func removeUselessLines(s string) string {
	var res []string
	list := strings.Split(s, "\n")

	for _, l := range list {
		if strings.Contains(l, "innerHTML") {
			res = append(res, l)
		}
	}

	return strings.Join(res, "\n") + "\n"
}

// extractAlert extracts alert from the given normalized string
//
//	see tests for the real example, copied from the website
func extractAlert(s string) Alert {
	var alert Alert

	re := regexp.MustCompile(`getElementById\(.+-(\d\d\d)"\).innerHTML.+color:\s(.+);\\"><\/i>(.+)<\/li>";`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 3 {
		alert = Alert{
			Region: matches[1],
			Level:  matches[2],
			Text:   strings.Trim(matches[3], " "),
		}
		return alert
	}

	if strings.Contains(s, "Nema upozorenja") {
		re = regexp.MustCompile(`getElementById\(.+-(\d\d\d)"\).innerHTML=.*Nema upozorenja.*$`)
		matches = re.FindStringSubmatch(s)

		alert = Alert{
			Region: matches[1],
			Level:  "green",
			Text:   "No alert",
		}
		return alert
	}

	return alert
}
