// Package mocks содержит моки сервисов, которые используют HTTP-обработчики.
package mocks

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/http/handlers.ClassService -o . -s _mock.go
//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/http/handlers.ProductService -o . -s _mock.go
//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/http/handlers.DerivedProductService -o . -s _mock.go
//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/http/handlers.MaterialService -o . -s _mock.go
//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/http/handlers.HistoryService -o . -s _mock.go
//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/http/handlers.EnumService -o . -s _mock.go
//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/http/handlers.ParametersService -o . -s _mock.go
//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i Brewery/internal/http/handlers.AggregateService -o . -s _mock.go
