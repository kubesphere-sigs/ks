package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GetNightlyTag parse the nightly tag
func GetNightlyTag(date string) (dateStr, tag string) {
	if strings.HasPrefix(date, "latest") {
		increment := strings.ReplaceAll(date, "latest", "")
		incrementDays := 0
		if increment != "" {
			if increment == "-" {
				increment = "-1"
			}
			incrementDays, _ = strconv.Atoi(increment)
		}
		incrementDays--

		dateStr = time.Now().AddDate(0, 0, incrementDays).Format("20060102")
		tag = fmt.Sprintf("nightly-%s", dateStr)
	} else if date != "" {
		var targetDate time.Time
		var err error
		// try to parse the date from different layouts
		if targetDate, err = time.Parse("2006-01-02", date); err != nil {
			targetDate, err = time.Parse("20060102", date)
		}

		if err == nil {
			dateStr = targetDate.Format("20060102")
			tag = fmt.Sprintf("nightly-%s", dateStr)
		}
	}
	return
}
