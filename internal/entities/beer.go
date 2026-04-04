// Package entities contains app's main objects' models
package entities

// Beer представляет полное описание пива со всеми связанными данными.
type Beer struct {
	ID          uint            `json:"id" info:"ID записи в бд"`
	Name        string          `json:"name" info:"Название напитка"`
	Rating      float32         `json:"rating" info:"Средний рейтинг"`
	Description string          `json:"description,omitempty" info:"Описание продукта"`
	ABV         float32         `json:"abv" info:"Крепость в %"`
	IBU         uint8           `json:"ibu" info:"Индекс горечи"`
	City        string          `json:"city" info:"Город производства"`
	Country     string          `json:"country" info:"Страна производства"`
	Type        string          `json:"type" info:"Тип напитка"`
	Category    ProductCategory `json:"category" info:"Информация о категории товара"`
	Features    []string        `json:"features,omitempty" info:"Список особенностей"`
}

// Beers представляет собой срез изделий Beer.
//
//easyjson:json
type Beers []Beer
