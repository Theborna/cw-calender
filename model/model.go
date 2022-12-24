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

type Event struct {
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
}

func (e *Event) ToString() string {
	return fmt.Sprintf("Summary: %v\nDescription: %v\nDeadline: %v", e.Summary, e.Description, e.Deadline)
}

type Calendar struct {
	link   string
	Events []Event
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
	cal := &Calendar{
		link:   link,
		Events: []Event{},
	}
	cal.UpdateEvents()
	return cal, nil
}

func (c *Calendar) UpdateEvents() {
	data, _ := getLinkData(c.link)
	start, end := time.Now(), time.Now().Add(12*30*24*time.Hour)

	cal := gocal.NewParser(bytes.NewReader(data))
	cal.Start, cal.End = &start, &end
	cal.Parse()

	events := []Event{}
	for _, e := range cal.Events {
		events = append(events, Event{
			Summary:     e.Summary,
			Description: e.Description,
			Deadline:    *e.End,
		})
	}
	c.Events = events
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

	Schedule []Calendar
}
