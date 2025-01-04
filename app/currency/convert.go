package currency

type Pair struct {
	From string
	To   string
}

func NewPair(from string, to string) *Pair {
	return &Pair{from, to}
}

func (pair *Pair) String() string {
	return pair.From + "-" + pair.To
}

func Convert(sum float64, from, to string) float64 {
	if from == to {
		return sum
	}

	pair := NewPair(from, to)

	rate, ok := exchangeRates[*pair]
	if !ok {
		return sum
	}

	return sum * rate
}
