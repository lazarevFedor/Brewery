package mocks

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/usecase.BeerService -o . -s _mock.go
//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/usecase.EnumService -o . -s _mock.go
