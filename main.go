package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dolanor/caldav-go/icalendar/components"
)

var (
	mediaWikiSchedule = `! Timeslot !! Bastide 1 !! Bastide 2 !! Bastide Mob !! Library (across yard left) !! Cheminée (across yard right)  !! Marquis (right of yard) !! Somewhere else !! Somewhere else 
|-
| #1 -10h30    || Arnaud - How to convince team in pair programming || Thierry -  Workshop on continuous delivery pipeline || TOF  - Meeting driven development || Bastien- Speedrun kata || Xavier - How to learn and deploy TDD || Benoit - Install Haskell on your computer || Tanguy - Introduction to Go || Romeu - Can open spaces be co-creation ?
|-
| #2 - 11h30    || Yann - How to organize a code review || Romeu - Mob facilitation of the open space || TOF - Smalltalk kata || Arthur - Show me some production Haskell || Johann - Split my monolith || Nicolas S - Discover and install Elixir and prepare for kata
|-
| 12h30 - 14h30 (lunch) || Nico S - Nix OS sharing || x ||  x || x || Nelson - Domain Driven Design || Benoit - Teaching programming languages || Jerome Avoustin - Power naps (Where ?) || Romeu - Katas ! (Restaurant)
|-
| #3 - 14h30 - 15h30 || Jerome Avoustin - Test Driven Development vs Type Driven Development || Nelis - Design a team ||  Nicolas M - Tests that exercise code by running .. the compiler ! || Geoffrey - Aggregate refactoring || Rob - KOBOLO enoeavour / One page RPG ! || Alex V. - Software skills timeline || Rémy - Independant release - Decoupled architecture 
|-
| #4 - 15h30 - 16h30 || Xavier A. - Digital sobriety || Romeu - Evolvin open space || Benoit - Declarative programing with Prolog in one kata || Andrei - Error handling in Kotlin || Thierry - Introduction to Theory of Constraints || Pascal - How ELM made my developer experience delightly fun || Pierre - Books I liked: Georges Orwell & Structure of scientific revolutions (Garden if not raining - Near fountain if raining) || Tanguy - How to deal with estimates, how to split tasks into subtasks ? 
|-
| #5 - 16h30 - 17h30 || Christian + Clément + Benoit - HELP ! / How to teach/coach testing, TDD, slicing, pair prog,  crafmanship, pair programming / WE HAVE COOKIES ! || Bastien - Automatic refactoring kata || Nelis - Kata on transformation priority premise || Andrei - Error handling in Kotlin || Thierry - Introduction to Theory of Constraints || Pascal - How ELM made my developer experience delightly fun || Pierre - Books I liked: Georges Orwell & Structure of scientific revolutions (Garden if not raining - Near fountain if raining) || Tanguy - How to deal with estimates, how to split tasks into subtasks ? 
|-
| 17h30 - 18h30 - Retro  || x || || x || x || x || x || x || x || x 
|-
| 18h30 - 20h - Slack time  || x || Arnaud - Quantic computers || x || Andrei - Kotlin tapas || Arthur - "Le banquet" / Board game (10 people MAX)  || Juke - Let's play overcooked || x || Jerome Avoustin / Guillaume G. -  Wine tasting - 19h15 Bar
|-
| 20h - 22h - Diner || x || ? - Running a meetup and building community  || x || x || x || x || x   
|-
| 22h onwards - Night session || x || x  || Romeu - K*T*S ! || ? - Board Games || ? - Team game (4 people MAX - french speaking) + Rémy - How to create and keep great team  || Bastien - Mob gaming || Christian - Whisky leaks  
|-`
	mdSchedule = `Timeslot | Bastide 1 | Bastide 2 | Bastide Mob | Library (accross yard middle) | Cheminée (accross yard left) | Marquis (right of yard) | Somewhere else | Somewhere else
--|--|--|--|--|--|--|--|--
10h30 - 11h30|Arnaud - How to convince team in pair programming | Thierry - Workshop on continuous delivery pipeline | TOF - Meeting driven development | Bastien Speedrun kata | Xavier - How to learn and deploy TDD | Benoit - Install Haskell on your computer | Tanguy - Introduction to Go | Romeu - Can open spaces be co-creation?
`
)

func main() {
	sr := strings.NewReader(mdSchedule)
	r := bufio.NewScanner(sr)
	r.Split(bufio.ScanLines)
	var lineNb int
	var rooms []string
	for r.Scan() {
		lineNb++
		// Let's ignore the header separator line
		if lineNb == 2 {
			continue
		}
		tokens := strings.Split(r.Text(), "|")
		// Let's save the room
		if lineNb == 1 {
			rooms = tokens
			_ = rooms
			continue
		}
		evt, err := parseEvent(tokens)
		if err != nil {
			log.Printf("could not parse line %d: %v", lineNb+1, err)
		}
		_ = evt
	}
}

func parseEvent(tokens []string) (*components.Event, error) {
	if len(tokens) < 9 {
		return nil, errors.New("wrong number of columns")
	}

	timesStrs := strings.Split(tokens[0], " - ")
	if len(timesStrs) < 2 {
		return nil, errors.New("wrong number of times (is it hh:mm - hh:mm?)")
	}
	start, err := time.Parse("15h04", timesStrs[0])
	if err != nil {
		return nil, errors.New("wrong time format (is it hh:mm)?")
	}
	end, err := time.Parse("15h04", timesStrs[1])
	if err != nil {
		return nil, errors.New("wrong time format (is it hh:mm)?")
	}
	fmt.Println(start.Format("15h04"), end.Format("15h04"))
	return nil, nil
}
