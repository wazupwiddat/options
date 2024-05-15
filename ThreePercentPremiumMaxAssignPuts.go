package main

import (
	"sort"
)

/*
	ThreePercentPremiumMaxAssignPuts strategy finds all options contracts that are out-of-the-money (OTM) PUTs
	with a premium ((OptionIdea.Bid / OptionIdea.Strike) * 100) of 3 percent or more.  The Calculated Value provided
	is (((UnderlyingPrice / (OptionIdea.Strike - OptionIdea.Bid)) - 1) * 100)
*/

type ThreePercentPremiumMaxAssignPuts struct{}

func (s ThreePercentPremiumMaxAssignPuts) Name() string {
	return "ThreePercentPremiumMaxAssignPuts"
}

func (s *ThreePercentPremiumMaxAssignPuts) Run(ideas []OptionIdea, underlyingPrice float64) StrategyResponse {
	fIdeas := Filter(ideas, OutOfTheMoney, IsPut, func(idea OptionIdea) bool {
		premPercent := (idea.Bid / idea.Strike) * 100
		return premPercent > 3
	})

	details := []ContractDetail{}
	for _, i := range fIdeas {
		det := ContractDetail{
			FinalCalculatedVal: ((underlyingPrice / (i.Strike - i.Bid)) - 1) * 100,
			ContractName:       i.Contract,
			BidAmount:          i.Bid,
			DaysRemaining:      i.DaysToExpiration,
			StrikePrice:        i.Strike,
		}
		details = append(details, det)
	}

	sort.Sort(ByContractDetail(details))
	return StrategyResponse{
		StrategyName:    s.Name(),
		UnderlyingPrice: underlyingPrice,
		Contracts:       details,
	}
}

func init() {
	Strategies = append(Strategies, &ThreePercentPremiumMaxAssignPuts{})
}
