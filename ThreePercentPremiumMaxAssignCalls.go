package main

import (
	"sort"
)

/*
	ThreePercentPremiumMaxAssignCalls strategy finds all options contracts that are out-of-the-money (OTM) CALLs
	with a premium ((OptionIdea.Bid / OptionIdea.Strike) * 100) of 3 percent or more.  The Calculated Value provided
	is ((((OptionIdea.Bid + OptionIdea.Strike) / UnderlyingPrice) - 1) * 100)
*/

type ThreePercentPremiumMaxAssignCalls struct{}

func (s ThreePercentPremiumMaxAssignCalls) Name() string {
	return "ThreePercentPremiumMaxAssignCalls"
}

func (s *ThreePercentPremiumMaxAssignCalls) Run(ideas []OptionIdea, underlyingPrice float64) StrategyResponse {
	fIdeas := Filter(ideas, OutOfTheMoney, IsCall, func(idea OptionIdea) bool {
		premPercent := (idea.Bid / idea.Strike) * 100
		return premPercent > 3
	})

	details := []ContractDetail{}
	for _, i := range fIdeas {
		det := ContractDetail{
			FinalCalculatedVal: (((i.Bid + i.Strike) / underlyingPrice) - 1) * 100,
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
	Strategies = append(Strategies, &ThreePercentPremiumMaxAssignCalls{})
}
