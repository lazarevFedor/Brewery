// Package entities contains app's main objects' models
package entities

type Beer struct {
	ID          uint
	Name        string
	Rating      float32
	Description string
	ABV         float32
	IBU         uint8
	City        string
	BeerType    string
	Category    string
	Features    []string
}
