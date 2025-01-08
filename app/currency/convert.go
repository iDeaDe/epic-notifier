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
