package model

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/apognu/gocal"
	tele "gopkg.in/telebot.v3"
)

type Calendar struct {
	link string
	Data string
}

// const URL = `https:\\/\\/cw\\.sharif\\.edu\\/calendar\\/export_execute\\.php\\?userid=\\d+&authtoken=\\S+`
const URL = `.+`

func NewCalendar(link string) (*Calendar, error) {
	link = strings.TrimSpace(link)
	if len(link) == 0 {
		return nil, fmt.Errorf("must provide a calendar link")
	}
	// matched, _ := regexp.MatchString(link, URL)
	// if !matched {
	// 	return nil, fmt.Errorf("invalid calendar url")
	// }
	data, _ := getLinkData(link)
	start, end := time.Now(), time.Now().Add(12*30*24*time.Hour)

	events := []string{}

	c := gocal.NewParser(bytes.NewReader(data))
	c.Start, c.End = &start, &end
	c.Parse()

	for _, e := range c.Events {
		event := fmt.Sprintf("Summary: %s\nDescription: %s\ndeadline: %s", e.Summary, e.Description, e.End)
		events = append(events, event)
	}
	return &Calendar{
		link: link,
		Data: strings.Join(events, `

		=================================

		`),
	}, nil
}

func (c *Calendar) Link() string {
	return c.link
}

func getLinkData(link string) ([]byte, error) {
	resp, err := http.Get(link)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	return body, nil
}

type User struct {
	*tele.User

	Calendar Calendar
}
