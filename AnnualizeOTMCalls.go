package main

import (
	"sort"
)

/*
AnnualizeOutOfTheMoneyCalls strategy finds all options contracts that are out-of-the-money (OTM) CALLs
and annualizes (if we could make this trade every OptionIdea.DaysToExpiration) the premium of the option against the OptionIdea.DaysToExpiration.
The Calculated Value provided
is (((((OptionIdea.Bid + OptionIdea.Strike) / (OptionIdea.Strike)) - 1) * 100) / OptionIdea.DaysToExpiration * 365).
*/

type AnnualizeOutOfTheMoneyCalls struct{}

func (s AnnualizeOutOfTheMoneyCalls) Name() string {
	return "AnnualizeOutOfTheMoneyCalls"
}

func (s *AnnualizeOutOfTheMoneyCalls) Run(ideas []OptionIdea, underlyingPrice float64) StrategyResponse {
	fIdeas := Filter(ideas, OutOfTheMoney, IsCall)

	details := []ContractDetail{}
	for _, i := range fIdeas {
		det := ContractDetail{
			FinalCalculatedVal: ((((i.Bid + i.Strike) / (i.Strike)) - 1) * 100) / float64(i.DaysToExpiration) * 365,
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
	Strategies = append(Strategies, &AnnualizeOutOfTheMoneyCalls{})
}
