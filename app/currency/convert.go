package currency

type Updater interface {
	SetErrorHandler(func(error))
	AddPair(Pair)
	Convert(float64, Pair) float64
	Update()
}

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
