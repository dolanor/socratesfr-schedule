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

var (
	mdSchedule = `What ?         | Timeslot    | Bastide 1                              | Bastide 2                                                   | Bastide Mob                                               | Bibliothèque (across yard left)                              | Cheminée (across yard right)                              | Marquis (right of yard)                             | Somewhere else                   | Somewhere else 
--             | --          | --                                     | --                                                          | --                                                        | --                                                           | --                                                        | --                                                  | --                               | --         
Marketplace    | 9h  - 10h   | Marketplace - Romeu et al.| | | | | | |
Session 1      | 10h - 11h   | Discover event storming 1/2 -  Nelson  |                                                             | Learn me some vi tricks - Thierry                         | Storybook for my #TDD - Xavier                               |                                                           | BDD + TDD kata / Beginners friendly - Adrien        |                                  |                     
Session 2      | 11h - 12h   | Discover event storming 2/2 -  Nelson  |                                                             | Chaos Monkey with Spring Boot - John                      | Co-creating kata - Romeu                                     | How to kickstart a continuous delivery program - Thierry  | Gambling-driven developemnt - ToF / Nico / Xavier D |                                  | Kotlin anti-patterns - Germain - Garden
Session 3      | 12h - 13h   | Fuzzzing -  Tanguy                     | Always green / Teaching tech of continuous delivery - Yohan | RPN calc open/closed principle - Jose                     | Pattern matching / Event sourcing / Mob / Haskell - Benoit   |                                                           |                                                     | Rugby RWC 1/2 - Who ? - Lounge   |                     
Lunch          | 13h - 15h   |                                        |                                                             | Feedback on mentoring - Yann                              |                                                              |                                                           |                                                     | Rugby RWC 2/2 - Who ? - Lounge   | Kata - Romeu - Lunch room
Session 4      | 15h - 16h   | Let's code meetup / DDD style - Nelson | Trunk-based development - Pascal                            | Java Quickperf - Jean                                     | Gossip kata / DSL / DDD - Mohammed                           | How to involve people to learn to evolve ? - Laurent      | Formal methods intro / Alloy - Alex V.              |                                  | Is feedback when we use information ? - ToF
Session 5      | 16h - 17h   | RPG web game (MtG) - Tanguy            |                                                             | Phone numbers kata - José                                 | Your code as a crime scene - Nicolas & Yohan                 | 50 nuances of (grey) testing - ToF                        | PureScript - Jerome A                               |                                  |                     
Session 6      | 17h - 18h   |                                        | Learning's pedagogy - Rémy & Houssam                        | Unmaintainable code that passes Sonar checks - Nicolas M. | Meta open-space ideas / Blog post co-creation - Romeu        | Contribute to an open-source app / Code - Adrien          | Practicing elm - Pascal                             |                                  |                     
Retro          | 18h - 18h30 | Retro - Romeu et al.                   |                                                             |                                                           |                                                              |                                                           |                                                     |                                  | 
Slack time     | 18h30 - 20h |                                        | Lightning talks - Niklaas                                   | Mob gaming - Arthur                                       | Factorio / TOC / Video gaming - Chris                        | OCaml discovery - NSA                                     | Take decisions as a group - Jhalil & Vincent        | Wine tasting (19h15) - Jerome    | Error-driven development brainstorm - ToF 
Dinner         | 20h - 22h   |                                        | Let's define lazy developer - Nicolas                       |                                                           |                                                              |                                                           |                                                     |                                  | Teach TDD to people - Romeu - Lunch
Night sessions | 22h - 23h59 |                                        |                                                             |                                                           | Overcooked 2 - Juke                                          |                                                           |                                                     | Whiskey strikes back - Christian | Board games & RPG - Arthuer - Lobby`
)

func main() {
	sr := strings.NewReader(mdSchedule)
	r := bufio.NewScanner(sr)
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
			rooms = tokens[2:]
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

	if len(hm) > 2 {
		m, err = strconv.Atoi(strings.TrimSpace(hm[1]))
		if err != nil {
			panic(err)
		}
	} else {
		m = 0
	}

	start := time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, time.Local)
	if len(timesStrs) == 2 {
		hm = strings.Split(timesStrs[1], "h")
		h, err = strconv.Atoi(hm[0])
		if err != nil {
			panic(err)
		}

		if len(hm) > 2 {
			m, err = strconv.Atoi(hm[1])
			if err != nil {
				panic(err)
			}
		} else {
			m = 0
		}
	}
	end := time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, time.Local)

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
		facilitator := facilitatorTitle[1]
		// We re-join in case some " - " string has been put in the session description
		title := facilitatorTitle[0]

		event := components.NewEventWithEnd(uid.String(), start.In(time.UTC), end.In(time.UTC))
		event.Summary = title
		event.Organizer = &values.OrganizerContact{Entry: mail.Address{Name: facilitator, Address: fmt.Sprintf("%s@socratesfr2019.fr", strings.ReplaceAll(facilitator, " ", "."))}}

		locurls := []*url.URL{}
		_ = locurls

		event.Location = values.NewLocation(rooms[col])
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
