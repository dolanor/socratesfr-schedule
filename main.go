package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dolanor/caldav-go/caldav"
	"github.com/dolanor/caldav-go/icalendar/components"
	"github.com/dolanor/caldav-go/icalendar/values"
	"github.com/fatih/color"
	"github.com/rs/xid"
)

func mustParseURL(urlstr string) *url.URL {
	u, err := url.Parse(urlstr)
	if err != nil {
		panic(err)
	}
	return u
}

var (
	locURL = map[string][]*url.URL{
		"Bastide 1":                       {mustParseURL("https://www.openstreetmap.org/way/736667138"), mustParseURL("geo:44.23264,4.79316")},
		"Bastide 2":                       {mustParseURL("https://www.openstreetmap.org/way/736667140"), mustParseURL("geo:44.23256,4.79299")},
		"Bastide Mob":                     {mustParseURL("https://www.openstreetmap.org/way/736667139"), mustParseURL("geo:44.23272,4.79321")},
		"Bibliothèque (across yard left)": {mustParseURL("https://www.openstreetmap.org/way/736667136"), mustParseURL("geo:44.23284,4.7928")},
		"Cheminée (across yard right)":    {mustParseURL("https://www.openstreetmap.org/way/736667137"), mustParseURL("geo:44.23279,4.79278")},
		"Marquis (right of yard)":         {mustParseURL("https://www.openstreetmap.org/way/736667135"), mustParseURL("geo:44.23281,4.79307")},
		"Somewhere else":                  {mustParseURL("https://www.openstreetmap.org/changeset/75916932")},
	}
)

func main() {
	r := bufio.NewScanner(os.Stdin)
	r.Split(bufio.ScanLines)
	var lineNb int
	var rooms []string

	host, login, password, calendar, basePath := loadEnvVar()
	calClient, err := newClient(host, login, password, calendar)
	if err != nil {
		panic(err)
	}
	for r.Scan() {
		lineNb++
		fmt.Printf("line: %d : %s\n", lineNb, r.Text())
		// Let's ignore the header separator line
		//max := 12
		if lineNb == 2 /* the marketplace line*/ {
			println("bail out, wrong line")
			continue
		}
		tokens := strings.Split(r.Text(), "|")
		// Let's save the room
		if lineNb == 1 {
			for _, v := range tokens[2:] {
				rooms = append(rooms, strings.TrimSpace(v))
			}
			continue
		}
		//if lineNb > max {
		//	break
		//}
		evts, err := parseEvents(tokens, time.Now(), rooms)
		if err != nil {
			log.Printf("could not parse line %d: %v", lineNb+1, err)
			return
		}

		for _, evt := range evts {
			uid := xid.New()
			calfile := fmt.Sprintf("%s/%s/%s.ics", basePath, calendar, uid.String())
			//err = calClient.PutEvents(calfile, evt)
			_, _, _ = calClient, evt, calfile
			if err != nil {
				panic(err)
			}
		}
	}
}

func parseEvents(tokens []string, day time.Time, rooms []string) ([]*components.Event, error) {
	if len(tokens) < 10 {
		return nil, errors.New("wrong number of columns")
	}

	// let's ignore the session number column
	tokens = tokens[1:]

	timesStrs := strings.Split(tokens[0], " - ")

	hm := strings.Split(timesStrs[0], "h")

	var h, m int
	var err error
	h, err = strconv.Atoi(strings.TrimSpace(hm[0]))
	if err != nil {
		panic(err)
	}

	if len(hm) > 1 {
		mstr := strings.TrimSpace(hm[1])
		if mstr != "" {
			m, err = strconv.Atoi(mstr)
			if err != nil {
				panic(err)
			}
		}
	} else {
		m = 0
	}

	tz, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		panic(err)
	}

	start := time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, tz)
	if len(timesStrs) == 2 {
		hm = strings.Split(timesStrs[1], "h")
		h, err = strconv.Atoi(hm[0])
		if err != nil {
			panic(err)
		}

		if len(hm) > 1 {
			mstr := strings.TrimSpace(hm[1])
			if mstr != "" {
				m, err = strconv.Atoi(mstr)
				if err != nil {
					panic(err)
				}
			}
		} else {
			m = 0
		}
	}
	end := time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, tz)

	var events []*components.Event
	// We ignore the first columns which is time and has been parsed already
	for col, session := range tokens[1:] {
		uid := xid.New()
		if strings.TrimSpace(session) == "" {
			// No session here
			continue
		}
		facilitatorTitle := strings.Split(session, " - ")
		if len(facilitatorTitle) < 2 {
			fmt.Println(facilitatorTitle)
			return nil, errors.New("wrong session format (is it Author Name - Text for so-called session?)")
		}
		facilitator := strings.TrimSpace(facilitatorTitle[1])
		// We re-join in case some " - " string has been put in the session description
		title := strings.TrimSpace(facilitatorTitle[0])

		event := components.NewEventWithEnd(uid.String(), start, end)
		event.Summary = title
		event.Organizer = &values.OrganizerContact{Entry: mail.Address{Name: facilitator, Address: fmt.Sprintf("%s@socratesfr2019.fr", strings.ReplaceAll(facilitator, " ", "."))}}

		locurls := locURL[rooms[col]]
		_ = locurls

		event.Location = values.NewLocation(rooms[col], locurls...)
		color.Red("%v\n", *event.Location)

		event.Attendees = []*values.AttendeeContact{&values.AttendeeContact{Entry: event.Organizer.Entry}}

		events = append(events, event)
		fmt.Println(*event)
	}
	return events, nil
}

func loadEnvVar() (host, login, password, calendar, basePath string) {
	host = os.Getenv("SCHED_HOST")
	login = os.Getenv("SCHED_LOGIN")
	password = os.Getenv("SCHED_PASSWORD")
	calendar = os.Getenv("SCHED_CALENDAR")
	basePath = os.Getenv("SCHED_BASEPATH")
	if host == "" || login == "" || password == "" || calendar == "" {
		panic("wrong env var setup")
	}
	return host, login, password, calendar, basePath
}

func newClient(host, login, password, calendar string) (*caldav.Client, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	ui := url.UserPassword(login, password)

	caldavServer, err := caldav.NewServer(fmt.Sprintf("%s://%s@%s/", u.Scheme, ui, u.Host))
	if err != nil {
		return nil, err
	}
	client := caldav.NewDefaultClient(caldavServer)
	return client, nil
}
