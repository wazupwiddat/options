package main

import (
	"sort"
)

/*
	AnnualizeOutOfTheMoneyPuts strategy finds all options contracts that are out-of-the-money (OTM) PUTs
	and annualizes (if we could make this trade every OptionIdea.DaysToExpiration) the premium of the option against the OptionIdea.DaysToExpiration.
	The Calculated Value provided
	is ((((OptionIdea.Strike / (OptionIdea.Strike - OptionIdea.Bid)) - 1) * 100) / OptionIdea.DaysToExpiration * 365).
*/

type AnnualizeOutOfTheMoneyPuts struct{}

func (s AnnualizeOutOfTheMoneyPuts) Name() string {
	return "AnnualizeOutOfTheMoneyPuts"
}

func (s *AnnualizeOutOfTheMoneyPuts) Run(ideas []OptionIdea, underlyingPrice float64) StrategyResponse {
	fIdeas := Filter(ideas, OutOfTheMoney, IsPut)

	details := []ContractDetail{}
	for _, i := range fIdeas {
		det := ContractDetail{
			FinalCalculatedVal: (((i.Strike / (i.Strike - i.Bid)) - 1) * 100) / float64(i.DaysToExpiration) * 365,
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
	Strategies = append(Strategies, &AnnualizeOutOfTheMoneyPuts{})
}
