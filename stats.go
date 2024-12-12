package main

import (
	"fmt"
	"sort"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const OutOfRange = 99999
const DaysInLastSixMonths = 183
const WeeksInLastSixMonths = 26

type column []int

func stats(email string) {
	commits := ProcessRepositories(email)
	PrintCommitStats(commits)
}

func GetBeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return startOfDay
}

func CountDaysSinceDate(date time.Time) int {
	days := 0
	now := GetBeginningOfDay(time.Now())
	for date.Before(now) {
		date = date.Add(time.Hour * 24)
		days++
		if days > DaysInLastSixMonths {
			return OutOfRange
		}
	}
	return days
}

func FillCommits(email string, path string, commits map[int]int) map[int]int {
	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}

	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}

	iterator, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		panic(err)
	}

	offset := calcOffset()
	err = iterator.ForEach(func(c *object.Commit) error {
		daysAgo := CountDaysSinceDate(c.Author.When) + offset
		if c.Author.Email != email {
			return nil
		}

		if daysAgo != OutOfRange {
			commits[daysAgo]++
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
	return commits
}

func ProcessRepositories(email string) map[int]int {
	FilePath := GetDotFilePath()
	repos := ParseFileLinesToSlice(FilePath)
	DaysInMap := DaysInLastSixMonths

	commits := make(map[int]int, DaysInMap)
	for i := DaysInMap; i > 0; i-- {
		commits[i] = 0
	}

	for _, path := range repos {
		commits = FillCommits(email, path, commits)
	}

	return commits
}

func calcOffset() int {
	var offset int
	weekday := time.Now().Weekday()

	switch weekday {
	case time.Sunday:
		offset = 7
	case time.Monday:
		offset = 6
	case time.Tuesday:
		offset = 5
	case time.Wednesday:
		offset = 4
	case time.Thursday:
		offset = 3
	case time.Friday:
		offset = 2
	case time.Saturday:
		offset = 1
	}
	return offset
}

func PrintCell(val int, today bool) {
	escape := "\033[0;37;30m"

	switch {
	case val > 0 && val < 5:
		escape = "\033[0;30;47m"
	case val > 5 && val < 10:
		escape = "\033[0;30;43m"
	case val >= 10:
		escape = "\033[0;30;42m"
	}

	if today {
		escape = "\033[1;37;45m"
	}

	if val == 0 {
		fmt.Printf(escape + "  - " + "\033[0m")
		return
	}

	str := "  %d "
	switch {
	case val >= 10:
		str = " %d "
	case val >= 100:
		str = "%d "
	}

	fmt.Printf(escape+str+"\033[0m", val)
}

func PrintCommitStats(commits map[int]int) {
	keys := SortMapIntoSlice(commits)
	cols := BuildCols(keys, commits)
	PrintCells(cols)
}

func SortMapIntoSlice(m map[int]int) []int {
	var keys []int
	for k := range m {
		keys = append(keys, k)

	}

	sort.Ints(keys)

	return keys
}

func BuildCols(keys []int, commits map[int]int) map[int]column {
	cols := make(map[int]column)
	col := column{}

	for _, k := range keys {
		week := int(k / 7)
		dayinweek := k % 7

		if dayinweek == 0 {
			col = column{}
		}

		col = append(col, commits[k])

		if dayinweek == 6 {
			cols[week] = col
		}
	}
	return cols
}

func PrintCells(cols map[int]column) {
	PrintMonths()

	for j := 6; j >= 0; j-- {
		for i := WeeksInLastSixMonths + 1; i >= 0; i-- {
			if i == WeeksInLastSixMonths+1 {
				PrintDayCol(j)
			}
			if col, ok := cols[i]; ok {
				if i == 0 && j == calcOffset()-1 {
					PrintCell(col[j], true)
					continue
				} else {
					if len(col) > j {
						PrintCell(col[j], false)
						continue
					}
				}
			}
			PrintCell(0, false)

		}
		fmt.Printf("\n")
	}
}

func PrintMonths() {
	week := GetBeginningOfDay(time.Now()).Add(-(DaysInLastSixMonths * time.Hour * 24))
	month := week.Month()
	fmt.Printf("	")
	for {
		if week.Month() != month {
			fmt.Printf("%s ", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("	")
		}

		week = week.Add(7 * time.Hour * 24)

		if week.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

func PrintDayCol(day int) {
	out := "     "
	switch day {
	case 1:
		out = " Mon "
	case 3:
		out = " Wed "
	case 5:
		out = " Fri "
	}

	fmt.Printf(out)
}
