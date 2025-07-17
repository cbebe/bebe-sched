package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/arran4/golang-ical"
	"github.com/bitfield/script"

	_ "embed"
)

//go:embed ratings.csv
var ratings string

//go:embed scrape.js
var scrapeScript string

var scheduleURL string
var exportURL string

func getRatings() (map[string]string, error) {
	records, err := csv.NewReader(bytes.NewBuffer([]byte(ratings))).ReadAll()
	if err != nil {
		return nil, err

	}
	ratings := make(map[string]string, len(records))
	for _, v := range records {
		ratings[v[0]] = v[1]
	}
	return ratings, nil
}

var SYMBOLS = map[string]string{
	"D": "Days",
	"N": "Nights",
	"E": "Evenings",
	"A": "Days 12 (A)",
	"B": "Nights 12 (B)",
}

type Shift struct {
	Symbol string
	Start  time.Time
	End    time.Time
	Status string
	Unit   string
}

type JSONShift struct {
	Date      string `json:"date"`      // Day Mon Date
	StartTime string `json:"starttime"` // HH:mm
	EndTime   string `json:"endtime"`   // HH:mm
	PaidHours string `json:"paidhours"`
	PayReason string `json:"payreason"`
	Occ       string `json:"occ"`
	Status    string `json:"status"` // Booked Off / Relieved
	Symbol    string `json:"symbol"`
	Unit      string `json:"unit"`
}

var jsonFileMatch = regexp.MustCompilePOSIX("jb.*\\.json$")

// Gets the path latest of the latest JSON file from Downloads
func getJSONFile() (string, error) {
	u, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Assuming this is the right path
	downloads := path.Join(u, "Downloads")
	s, err := script.ListFiles(downloads).MatchRegexp(jsonFileMatch).String()
	if err != nil {
		return "", err
	}
	files := strings.Split(strings.TrimSpace(s), "\n")
	switch len(files) {
	case 0:
		return "", fmt.Errorf("No files found")
	default:
		files = files[0 : len(files)-1]
		slices.Sort(files)
		slices.Reverse(files)
		fallthrough
	case 1:
		return files[0], nil
	}
}

func parseTime(dateStr, timeStr string) (time.Time, error) {
	// Assume date time is local timezone
	now := time.Now()
	_, zoneSecs := now.Zone()
	offset := -(time.Duration(zoneSecs) * time.Second)

	dateTimeStr := fmt.Sprintf("%s %d %s", dateStr, now.Year(), timeStr)
	t, err := time.Parse("Mon Jan 02 2006 15:04", dateTimeStr)
	if err != nil {
		return time.Time{}, err
	}
	return t.Add(offset).Local(), nil
}

func shiftsFromJSON(jsonShifts []JSONShift) ([]Shift, error) {
	shifts := make([]Shift, 0, len(jsonShifts))

	for _, s := range jsonShifts {
		start, err := parseTime(s.Date, s.StartTime)
		if err != nil {
			return nil, err
		}
		end, err := parseTime(s.Date, s.EndTime)
		if err != nil {
			return nil, err
		}

		if end.Hour() < start.Hour() {
			// This is assuming you don't work 24 hours
			end = end.Add(time.Hour * 24)
		}

		if end.Sub(time.Now()) > 0 && strings.Contains(s.Status, "Relieved") {
			shifts = append(shifts, Shift{
				Symbol: strings.TrimSpace(s.Symbol),
				Start:  start,
				End:    end,
				Status: s.Status,
				Unit:   strings.Split(s.Unit, "-")[1],
			})
		}
	}

	return shifts, nil
}

// Pipe to convert a JSON array from the given Reader into an iCal file
// written to the given Writer.
func jsonToICal(r io.Reader, w io.Writer) error {
	var jsonShifts []JSONShift
	if err := json.NewDecoder(r).Decode(&jsonShifts); err != nil {
		return err
	}

	shifts, err := shiftsFromJSON(jsonShifts)
	if err != nil {
		return err
	}

	ratings, err := getRatings()
	if err != nil {
		log.Printf("ratings: %v\n", err)
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodAdd)
	for _, s := range shifts {
		// Use ID tied to date and unit so we can keep importing
		id := fmt.Sprintf("%s-%s", s.Unit, s.Start.Format(time.RFC3339))
		event := cal.AddEvent(id)

		now := time.Now()
		event.SetCreatedTime(now)
		event.SetDtStampTime(now)
		event.SetModifiedAt(now)

		event.SetStartAt(s.Start)
		event.SetEndAt(s.End)

		event.SetSummary(fmt.Sprintf("%s at %s", SYMBOLS[s.Symbol], s.Unit))
		event.SetLocation("Hospital")

		if ratings == nil {
			continue
		}

		if rating, ok := ratings[s.Unit]; ok {
			event.SetDescription(fmt.Sprintf("Bebe's rating: %s", rating))
		} else {
			log.Printf("warning: no rating for unit %s\n", s.Unit)
		}
	}

	return cal.SerializeTo(w)
}

func makeICal() {
	jsonFile, err := getJSONFile()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Reading JSON from", jsonFile)
	_, err = script.File(jsonFile).Filter(jsonToICal).WriteFile("bebe.ical")
	if err != nil {
		log.Fatal(err)
	}
	if exportURL != "" {
		log.Println("Import the calendar in", exportURL)
	}
}

func main() {
	skip := flag.Bool("skip", false, "Whether to skip opening the browser")
	flag.Parse()

	if scheduleURL == "" {
		log.Println("ScheduleURL is not defined")
		makeICal()
		return
	}

	if !*skip {
		fmt.Println("Opening in browser...")
		script.Exec(fmt.Sprintf("open %s", scheduleURL))
		fmt.Println("Log in to the website and paste the JavaScript code into the console.")

		script.Echo(scrapeScript).Exec("pbcopy").Wait()

		fmt.Print("Press enter to continue. ")
		fmt.Scanln()
	}

	makeICal()
}
