// Package entities содержит слой сущностей
package entities

// Beer представляет полное описание пива со всеми связанными данными.
type Beer struct {
	ID          uint            `json:"id" info:"ID записи в бд"`
	Name        string          `json:"name" info:"Название напитка"`
	Rating      float32         `json:"rating" info:"Средний рейтинг"`
	Description string          `json:"description,omitempty" info:"Описание продукта"`
	ABV         float32         `json:"abv" info:"Крепость в %"`
	IBU         uint8           `json:"ibu" info:"Индекс горечи"`
	Amount      uint            `json:"amount" info:"Кол-во в наличие"`
	Unit        string          `json:"units" info:"Единица измерений"`
	City        string          `json:"city" info:"Город производства"`
	Country     string          `json:"country" info:"Страна производства"`
	Category    ProductCategory `json:"category" info:"Информация о категории товара"`
	Features    []string        `json:"features,omitempty" info:"Список особенностей"`
}

// Beers представляет собой срез изделий Beer.
//
//easyjson:json
type Beers []Beer
