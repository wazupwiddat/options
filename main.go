package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/piquette/finance-go"
	"github.com/piquette/finance-go/datetime"
	"github.com/piquette/finance-go/options"
	"github.com/piquette/finance-go/quote"
)

type OptionResponse struct {
	Ideas []OptionIdea
	Bid   float64
}

func main() {
	http.HandleFunc("/options", handleOptionsRequest)                                  // Handle API requests
	http.Handle("/spec/", http.StripPrefix("/spec/", http.FileServer(http.Dir("./")))) // Correctly serve files under /spec

	http.HandleFunc("/", rootHandler) // Root handler that checks the path

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r) // Respond with 404 for any requests not exactly matching the root
		return
	}
	healthCheck(w) // Call healthCheck only if the path is exactly "/"
}

func healthCheck(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleOptionsRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ticker := query.Get("ticker")
	if ticker == "" {
		ticker = "AAPL" // default value if not specified
	}
	weeksOutStr := query.Get("weeksout")
	weeksOut, err := strconv.Atoi(weeksOutStr)
	if err != nil || weeksOut <= 0 {
		weeksOut = 5 // default value if not specified or invalid
	}

	// current quote
	q, err := quote.Get(ticker)
	if err != nil {
		http.Error(w, "Unable to retrieve quote for "+ticker, http.StatusInternalServerError)
		return
	}

	// collect Fridays
	fridays := nextSoFridays(weeksOut)

	// collect Ideas from Straddles
	ideas := []OptionIdea{}
	for _, friday := range fridays {
		dt := datetime.New(&friday)
		formattedDate := fmt.Sprintf("%02d-%02d-%d", friday.Month(), friday.Day(), friday.Year())
		log.Println("Collecting Straddles for", ticker, formattedDate)
		iter := options.GetStraddleP(&options.Params{
			UnderlyingSymbol: strings.ToUpper(ticker),
			Expiration:       dt,
		})
		if iter.Count() == 0 {
			log.Println("No options available for", formattedDate)
			thursday := friday.AddDate(0, 0, -1)
			dt = datetime.New(&thursday)
			formattedDate = fmt.Sprintf("%02d-%02d-%d", thursday.Month(), thursday.Day(), thursday.Year())
			log.Println("There must be a holiday that day, trying Thursday", formattedDate)
			iter = options.GetStraddleP(&options.Params{
				UnderlyingSymbol: strings.ToUpper(ticker),
				Expiration:       dt,
			})
			if iter.Count() == 0 {
				log.Println("No options available for", thursday)
				log.Println("Must be something wrong??? SKIPPING")
				continue
			}
		}
		ideas = append(ideas, collectIdeas(iter, q.Bid)...)
	}

	res := OptionResponse{
		Ideas: ideas,
		Bid:   q.Bid,
	}

	// Write the ideas back to the client
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, "Error encoding response object", http.StatusInternalServerError)
	}
}

func collectIdeas(iter *options.StraddleIter, bid float64) []OptionIdea {
	ideas := []OptionIdea{}
	for iter.Next() {
		straddle := *iter.Straddle()
		if straddle.Call == nil || straddle.Put == nil {
			continue
		}
		days := daysToExpiration(straddle)

		callIdea := OptionIdea{
			Call:             true,
			Strike:           straddle.Strike,
			Contract:         straddle.Call.Symbol,
			Bid:              straddle.Call.Bid,
			InTheMoney:       straddle.Call.InTheMoney,
			DaysToExpiration: days,
			ReturnIfFlat:     straddle.Call.Bid / straddle.Strike, // Premium
			ReturnIfAssigned: ((straddle.Strike + straddle.Call.Bid) / bid) - 1,
		}
		ideas = append(ideas, callIdea)
		putIdea := OptionIdea{
			Call:             false,
			Strike:           straddle.Strike,
			Contract:         straddle.Put.Symbol,
			Bid:              straddle.Put.Bid,
			InTheMoney:       straddle.Put.InTheMoney,
			DaysToExpiration: days,
			ReturnIfFlat:     straddle.Put.Bid / straddle.Strike, // Premium
			ReturnIfAssigned: ((straddle.Strike - straddle.Put.Bid) / bid) - 1,
		}
		ideas = append(ideas, putIdea)
	}
	if err := iter.Err(); err != nil {
		fmt.Println(err)
	}
	return ideas
}

func daysToExpiration(straddle finance.Straddle) int {
	today := zeroOutTime(time.Now().UTC())
	exp := time.Unix(int64(straddle.Call.Expiration), 0).UTC()
	diff := exp.Sub(today)
	return int(diff.Hours() / 24)
}

func nextSoFridays(weeks int) []time.Time {
	var fridays []time.Time

	// Get today's date
	today := zeroOutTime(time.Now().UTC())

	// Find the first Friday
	for i := 0; i < 7; i++ {
		if today.Weekday() == time.Friday {
			break
		}
		today = today.AddDate(0, 0, 1)
	}

	// Collect the next so many Fridays
	thisYear := time.Now().Year()
	for i := 0; i < weeks; i++ {
		if today.Year() == thisYear || today.Year() > thisYear { // Ensure it's within the current year
			fridays = append(fridays, today)
		}
		today = today.AddDate(0, 0, 7) // Move to the next week
	}

	return fridays
}

func zeroOutTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
