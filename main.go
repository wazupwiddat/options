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
	http.HandleFunc("/options", handleOptionsRequest)
	http.HandleFunc("/options/strategies", handleStrategiesRequest)
	http.Handle("/spec/", http.StripPrefix("/spec/", http.FileServer(http.Dir("./"))))
	http.HandleFunc("/", rootHandler)

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	healthCheck(w)
}

func healthCheck(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleOptionsRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ticker := query.Get("ticker")
	if ticker == "" {
		http.Error(w, "Ticker is required", http.StatusBadRequest)
		return
	}
	weeksOutStr := query.Get("weeksout")
	weeksOut, err := strconv.Atoi(weeksOutStr)
	if err != nil || weeksOut <= 0 {
		weeksOut = 5
	}

	q, err := quote.Get(ticker)
	if err != nil {
		http.Error(w, "Unable to retrieve quote for "+ticker, http.StatusInternalServerError)
		return
	}

	fridays := nextSoFridays(weeksOut)

	ideas := []OptionIdea{}
	for _, friday := range fridays {
		dt := datetime.New(&friday)
		iter := options.GetStraddleP(&options.Params{
			UnderlyingSymbol: strings.ToUpper(ticker),
			Expiration:       dt,
		})
		if iter.Count() == 0 {
			thursday := friday.AddDate(0, 0, -1)
			dt = datetime.New(&thursday)
			iter = options.GetStraddleP(&options.Params{
				UnderlyingSymbol: strings.ToUpper(ticker),
				Expiration:       dt,
			})
			if iter.Count() == 0 {
				continue
			}
		}
		ideas = append(ideas, collectIdeas(iter, q.Bid)...)
	}

	res := OptionResponse{
		Ideas: ideas,
		Bid:   q.Bid,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, "Error encoding response object", http.StatusInternalServerError)
	}
}

func handleStrategiesRequest(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	ticker := query.Get("ticker")
	if ticker == "" {
		http.Error(w, "Ticker is required", http.StatusBadRequest)
		return
	}
	weeksOutStr := query.Get("weeksout")
	weeksOut, err := strconv.Atoi(weeksOutStr)
	if err != nil || weeksOut <= 0 {
		weeksOut = 5
	}

	q, err := quote.Get(ticker)
	if err != nil {
		http.Error(w, "Unable to retrieve quote for "+ticker, http.StatusInternalServerError)
		return
	}

	fridays := nextSoFridays(weeksOut)

	ideas := []OptionIdea{}
	for _, friday := range fridays {
		dt := datetime.New(&friday)
		iter := options.GetStraddleP(&options.Params{
			UnderlyingSymbol: strings.ToUpper(ticker),
			Expiration:       dt,
		})
		if iter.Count() == 0 {
			thursday := friday.AddDate(0, 0, -1)
			dt = datetime.New(&thursday)
			iter = options.GetStraddleP(&options.Params{
				UnderlyingSymbol: strings.ToUpper(ticker),
				Expiration:       dt,
			})
			if iter.Count() == 0 {
				continue
			}
		}
		ideas = append(ideas, collectIdeas(iter, q.Bid)...)
	}

	strategyName := query.Get("strategy")
	strategyResponse := buildStrategyResponse(strategyName, q.Bid, ideas)

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(strategyResponse)
	if err != nil {
		http.Error(w, "Error encoding response object", http.StatusInternalServerError)
	}
}

func buildStrategyResponse(strategyName string, underlyingPrice float64, ideas []OptionIdea) []StrategyResponse {
	response := []StrategyResponse{}
	for _, strat := range Strategies {
		if strategyName == "" || strategyName == strat.Name() {
			response = append(response, strat.Run(ideas, underlyingPrice))
		}
	}
	return response
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
			ReturnIfFlat:     straddle.Call.Bid / straddle.Strike,
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
			ReturnIfFlat:     straddle.Put.Bid / straddle.Strike,
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

	today := zeroOutTime(time.Now().UTC())

	for i := 0; i < 7; i++ {
		if today.Weekday() == time.Friday {
			break
		}
		today = today.AddDate(0, 0, 1)
	}

	thisYear := time.Now().Year()
	for i := 0; i < weeks; i++ {
		if today.Year() == thisYear || today.Year() > thisYear {
			fridays = append(fridays, today)
		}
		today = today.AddDate(0, 0, 7)
	}

	return fridays
}

func zeroOutTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
