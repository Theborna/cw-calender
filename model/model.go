package model

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/apognu/gocal"
	tele "gopkg.in/telebot.v3"
)

type Event struct {
	Name        string    `json:"name"`
	Prof        string    `json:"prof"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
}

func (e *Event) ToString() string {
	return fmt.Sprintf(`
Name: %v
Prof: %v
Summary: %v
Description: %v
Deadline: %v
Time left: %v
`, e.Name, e.Prof, e.Summary, e.Description, e.Deadline.Format(time.ANSIC), time.Until(e.Deadline).Truncate(time.Minute))
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
		name, prof := nameOfCourse(e.Categories[len(e.Categories)-1])
		events = append(events, Event{
			Name:        name,
			Prof:        prof,
			Summary:     e.Summary,
			Description: e.Description,
			Deadline:    *e.End,
		})
	}
	c.Events = events
}
func nameOfCourse(info string) (string, string) {
	f, err := os.Open("courses.csv")
	if err != nil {
		log.Fatal(err)
		return "", ""
	}
	defer f.Close()
	r := csv.NewReader(f)
	ids := strings.Split(info, "-")
	for record, err := r.Read(); err != io.EOF; record, err = r.Read() {
		if err != nil {
			log.Fatal(err)
		}
		id, group, title, prof := record[0], record[1], record[3], record[4]
		for _, value := range ids {
			if value == id || value == strings.Join([]string{id, group}, "") {
				return title, prof
			}
		}
	}
	return "", ""
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

func (u *User) AddCall(cal *Calendar) bool {
	for _, c := range u.Schedule {
		if c.Link() == cal.Link() {
			return false
		}
	}
	u.Schedule = append(u.Schedule, *cal)
	return true
}
