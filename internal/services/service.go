package services

import (
	"errors"
	"final-project/internal/models"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("repeat is empty")
	}

	startDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}

	switch {
	case strings.HasPrefix(repeat, "d "):
		return dailyRule(now, startDate, repeat)
	case strings.HasPrefix(repeat, "y"):
		return yearlyRule(now, startDate)
	case strings.HasPrefix(repeat, "w "):
		return weeklyRule(now, repeat)
	case strings.HasPrefix(repeat, "m "):
		return monthlyRule(now, startDate, repeat)
	default:
		return "", errors.New("invalid repeat")
	}
}

func dailyRule(now, startDate time.Time, repeat string) (string, error) {
	days, err := parseDays(repeat)
	if err != nil {
		return "", err
	}

	for {
		startDate = startDate.AddDate(0, 0, days)
		if startDate.After(now) {
			return startDate.Format("20060102"), nil
		}
	}
}

func parseDays(repeat string) (int, error) {
	parts := strings.Split(repeat, " ")
	if len(parts) != 2 {
		return 0, errors.New("invalid repeat format for d")
	}

	days, err := strconv.Atoi(parts[1])
	if err != nil || days < 0 || days > 400 {
		return 0, errors.New("invalid days: " + parts[1])
	}

	return days, nil
}

func yearlyRule(now, startDate time.Time) (string, error) {
	for {
		startDate = startDate.AddDate(1, 0, 0)
		if startDate.After(now) {
			return startDate.Format("20060102"), nil
		}
	}
}

func weeklyRule(now time.Time, repeat string) (string, error) {
	weekdays, err := parseWeekdays(repeat)
	if err != nil {
		return "", err
	}

	for {
		now = now.AddDate(0, 0, 1)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}

		for _, day := range weekdays {
			if weekday == day {
				return now.Format("20060102"), nil
			}
		}
	}
}

func parseWeekdays(repeat string) ([]int, error) {
	parts := strings.Split(repeat, " ")
	if len(parts) != 2 {
		return nil, errors.New("invalid repeat format for w")
	}

	partsWeek := strings.Split(parts[1], ",")
	var weekdays []int
	for _, part := range partsWeek {
		day, err := strconv.Atoi(part)
		if err != nil || day < 1 || day > 7 {
			return nil, errors.New("invalid weekday: " + part)
		}
		weekdays = append(weekdays, day)
	}

	return weekdays, nil
}

func monthlyRule(now, startDate time.Time, repeat string) (string, error) {
	days, months, err := parseMonthsDays(repeat)
	if err != nil {
		return "", err
	}

	for {
		year, month, day := startDate.Date()

		if len(months) > 0 && !months[int(month)] {
			startDate = startDate.AddDate(0, 1, -startDate.Day()+1)
			continue
		}

		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, startDate.Location()).Day()
		for d := range days {
			if (d > 0 && d == day) || (d == -1 && day == lastDay) || (d == -2 && day == lastDay-1) {
				if startDate.After(now) {
					return startDate.Format("20060102"), nil
				}
			}
		}

		startDate = startDate.AddDate(0, 0, 1)
	}
}

func parseMonthsDays(repeat string) (map[int]bool, map[int]bool, error) {
	parts := strings.Split(repeat, " ")
	if len(parts) < 2 || len(parts) > 3 {
		return nil, nil, errors.New("invalid repeat format for m")
	}

	partsDays := strings.Split(parts[1], ",")
	days := make(map[int]bool)
	for _, part := range partsDays {
		day, err := strconv.Atoi(part)
		if err != nil || day == 0 || day < -2 || day > 31 {
			return nil, nil, errors.New("invalid day: " + part)
		}
		days[day] = true
	}

	months := make(map[int]bool)
	if len(parts) == 3 {
		partsMonth := strings.Split(parts[2], ",")
		for _, part := range partsMonth {
			month, err := strconv.Atoi(part)
			if err != nil || month < 1 || month > 12 {
				return nil, nil, errors.New("invalid month: " + part)
			}
			months[month] = true
		}
	}

	return days, months, nil
}

func ProcessTask(task *models.Task) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if task.Title == "" {
		return errors.New("task title is required")
	}

	if task.Date == "" {
		task.Date = today.Format("20060102")
		return nil
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		return errors.New("invalid date format, expected YYYYMMDD")
	}

	if parsedDate.Before(today) {
		if task.Repeat == "" {
			task.Date = today.Format("20060102")
			return nil
		}

		nextDate, err := NextDate(today, task.Date, task.Repeat)
		if err != nil {
			return err
		}
		task.Date = nextDate
		return nil
	}

	if task.Repeat != "" {
		_, err = NextDate(today, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	return nil
}
