package jobs

func ScrapeBufferedPages(a int, b int) int {
	return a + b
}

type AA struct {
	a int
	b int
}

func (aa *AA) DOASTUFF() int {
	return aa.a + aa.b
}

type BB struct {
	a int
	b int
}

func (bb *BB) DOBSTUFF() int {
	return bb.a + bb.b
}

type MM interface {
	DOASTUFF() int
	DOBSTUFF() int
}

type DD struct {
	AA
	BB
}

func Doer() MM {

	return &DD{
		AA: AA{
			a: 1,
			b: 2,
		},
		BB: BB{
			a: 3,
			b: 4,
		},
	}
}
