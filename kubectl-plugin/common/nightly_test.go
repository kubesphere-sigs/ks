package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetNightlyTag(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("20060102")

	table := []struct {
		date       string
		expectDate string
		expectTag  string
		message    string
	}{{
		date:       "",
		expectDate: "",
		expectTag:  "",
		message:    "should return an empty string if the date is empty",
	}, {
		date:       "invalid",
		expectDate: "",
		expectTag:  "",
		message:    "should return an empty string if the date is invalid",
	}, {
		date:       "20060102",
		expectDate: "20060102",
		expectTag:  "nightly-20060102",
		message:    "should return 20060102",
	}, {
		date:       "2020-01-01",
		expectDate: "20200101",
		expectTag:  "nightly-20200101",
		message:    "should return 20200101",
	}, {
		date:       "latest",
		expectDate: yesterday,
		expectTag:  fmt.Sprintf("nightly-%s", yesterday),
		message:    fmt.Sprintf("should return %s", yesterday),
	}, {
		date:       "latest-0",
		expectDate: yesterday,
		expectTag:  fmt.Sprintf("nightly-%s", yesterday),
		message:    fmt.Sprintf("should return %s", yesterday),
	}, {
		date:       "latest-1",
		expectDate: time.Now().AddDate(0, 0, -2).Format("20060102"),
		expectTag:  fmt.Sprintf("nightly-%s", time.Now().AddDate(0, 0, -2).Format("20060102")),
		message:    fmt.Sprintf("should return %s", time.Now().AddDate(0, 0, -2).Format("20060102")),
	}, {
		date:       "latest-2",
		expectDate: time.Now().AddDate(0, 0, -3).Format("20060102"),
		expectTag:  fmt.Sprintf("nightly-%s", time.Now().AddDate(0, 0, -3).Format("20060102")),
		message:    fmt.Sprintf("should return %s", time.Now().AddDate(0, 0, -3).Format("20060102")),
	}, {
		date:       "latest-",
		expectDate: time.Now().AddDate(0, 0, -2).Format("20060102"),
		expectTag:  fmt.Sprintf("nightly-%s", time.Now().AddDate(0, 0, -2).Format("20060102")),
		message:    fmt.Sprintf("should return %s", time.Now().AddDate(0, 0, -2).Format("20060102")),
	}}

	for _, item := range table {
		date, tag := GetNightlyTag(item.date)
		assert.Equal(t, date, item.expectDate, item.message)
		assert.Equal(t, tag, item.expectTag, item.message)
	}
}
