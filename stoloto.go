package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocolly/colly"
)

type circulation struct {
	id             int
	digit1         int
	digit2         int
	digit3         int
	digit4         int
	digit5         int
	digitExtra     int
	countOfTickets int
	date           string
	time           string
}

var (
	numbersFrequency = map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0, 7: 0, 8: 0, 9: 0, 10: 0, 11: 0, 12: 0, 13: 0, 14: 0, 15: 0, 16: 0,
		17: 0, 18: 0, 19: 0, 20: 0, 21: 0, 22: 0, 23: 0, 24: 0, 25: 0, 26: 0, 27: 0, 28: 0, 29: 0, 30: 0, 31: 0, 32: 0, 33: 0, 34: 0,
		35: 0, 36: 0}
	extraDigitFrequency = map[int]int{1: 0, 2: 0, 3: 0, 4: 0}
)

func main() {
	var cmdLineArgs = os.Args
	//circulationLimit must be lower than circulationCounter
	circulationCounter := convertToInteger(cmdLineArgs[1])
	circulationLimit := convertToInteger(cmdLineArgs[2])

	db, err := sql.Open("mysql", "mysql:321asd680@tcp(127.0.0.1:3306)/stoloto")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	c := colly.NewCollector()

	c.OnHTML("#content", func(e *colly.HTMLElement) {
		circ := new(circulation)
		circ.id = circulationCounter

		b := e.DOM.Find("div.winning_numbers.cleared > ul > li:nth-child(1) > p").Text()
		circ.digit1 = convertToInteger(b)

		b = e.DOM.Find("div.winning_numbers.cleared > ul > li:nth-child(2) > p").Text()
		circ.digit2 = convertToInteger(b)

		b = e.DOM.Find("div.winning_numbers.cleared > ul > li:nth-child(3) > p").Text()
		circ.digit3 = convertToInteger(b)

		b = e.DOM.Find("div.winning_numbers.cleared > ul > li:nth-child(4) > p").Text()
		circ.digit4 = convertToInteger(b)

		b = e.DOM.Find("div.winning_numbers.cleared > ul > li:nth-child(5) > p").Text()
		circ.digit5 = convertToInteger(b)

		b = e.DOM.Find("div.winning_numbers.cleared > ul > li:nth-child(6) > div > p").Text()
		circ.digitExtra = convertToInteger(b)

		b = e.DOM.Find("div.col.drawing_details > div > div > table > tbody > tr:nth-child(1) > td.numeric").Text()
		circ.countOfTickets = convertToInteger(b)

		b = e.DOM.Find("p").Text()
		//getting date and time out of text with regex
		re, _ := regexp.Compile(`(\d{2}\.){2}20\d{2}`)
		circ.date = re.FindString(b)

		re, _ = regexp.Compile(`\d{2}:\d{2}`)
		circ.time = re.FindString(b)

		circ.digitCounter()
		circ.dbWriting(db)

		a := fmt.Sprintf("Числа тиража %d: %d, %d, %d, %d, %d + extra: %d. Кол-во билетов: %d. Дата: %s, время: %s",
			circ.id, circ.digit1, circ.digit2, circ.digit3, circ.digit4, circ.digit5, circ.digitExtra, circ.countOfTickets, circ.date, circ.time)
		fmt.Println(a)
	})

	for ; circulationCounter >= circulationLimit; circulationCounter-- {
		link := fmt.Sprintf("https://www.stoloto.ru/5x36plus/archive/%d", circulationCounter)
		c.Visit(link)
	}

	fmt.Println(numbersFrequency)
	fmt.Println(extraDigitFrequency)
}

func convertToInteger(word string) int {
	a, err := strconv.Atoi(word)
	if err != nil {
		fmt.Println(err)
	}

	return a
}

func (circ *circulation) digitCounter() {
	numbersFrequency[circ.digit1]++
	numbersFrequency[circ.digit2]++
	numbersFrequency[circ.digit3]++
	numbersFrequency[circ.digit4]++
	numbersFrequency[circ.digit5]++

	extraDigitFrequency[circ.digitExtra]++
}

func (circ *circulation) dbWriting(db *sql.DB) {
	//making array of numbers to then iterate of it
	numberArr := make([]int, 0, 5)
	numberArr = append(numberArr, circ.digit1, circ.digit2, circ.digit3, circ.digit4, circ.digit5)

	k, err := time.Parse("02.01.2006", circ.date)
	if err != nil {
		fmt.Println(err)
	}
	date := k.Format("2006.01.02")

	//I made 1 db row containing only 1 circulation number to make counting query much easier
	for _, val := range numberArr {
		query := fmt.Sprintf("INSERT INTO `digits` (`circulation_number`, `digit`, `extra_digit`, `date`, `time`) VALUES('%d', '%d',"+
			" '%d', '%s', '%s')", circ.id, val, circ.digitExtra, date, circ.time)
		insert, err := db.Query(query)
		if err != nil {
			fmt.Println(err)
		}
		insert.Close()
	}
}
