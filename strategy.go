package main

type StrategyResponse struct {
	StrategyName    string
	UnderlyingPrice float64
	Contracts       []ContractDetail
}

type ContractDetail struct {
	ContractName       string
	BidAmount          float64
	DaysRemaining      int
	StrikePrice        float64
	FinalCalculatedVal float64
}

type ByContractDetail []ContractDetail

func (a ByContractDetail) Len() int { return len(a) }
func (a ByContractDetail) Less(i, j int) bool {
	return a[i].FinalCalculatedVal > a[j].FinalCalculatedVal
}
func (a ByContractDetail) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type Strategy interface {
	Name() string
	Run(ideas []OptionIdea, underlyingPrice float64) StrategyResponse
}

var Strategies []Strategy
