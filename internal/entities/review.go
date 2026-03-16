package entities

type review struct {
	ID     int
	Body   string
	BeerID uint
	Rating float32
}
