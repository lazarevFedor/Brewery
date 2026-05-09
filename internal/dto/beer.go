package dto

type BeerBase struct{
	Name string `json:"name"`
	Rating float32 `json:"rating"`
	Description string `json:"description"`
}

type Beer struct {
	BeerBase
	
	ID          uint            `json:"id" info:"ID записи в бд"`
	ABV         float32         `json:"abv" info:"Крепость в %"`
	IBU         uint8           `json:"ibu" info:"Индекс горечи"`
	Amount      uint            `json:"amount" info:"Кол-во в наличие"`
	Unit        string          `json:"units" info:"Единица измерений"`
	City        string          `json:"city" info:"Город производства"`
	Country     string          `json:"country" info:"Страна производства"`
	Category    CategoryBase `json:"category" info:"Информация о категории товара"`
	Features    []string        `json:"features,omitempty" info:"Список особенностей"`
}

type BeerCreate struct {
	BeerBase

	ABV         float32         `json:"abv" info:"Крепость в %"`
	IBU         uint8           `json:"ibu" info:"Индекс горечи"`
	Amount      uint            `json:"amount" info:"Кол-во в наличие"`
	Unit        string          `json:"units" info:"Единица измерений"`
	City        string          `json:"city" info:"Город производства"`
	Country     string          `json:"country" info:"Страна производства"`
	Category    CategoryBase `json:"category" info:"Информация о категории товара"`
	Features    []string        `json:"features,omitempty" info:"Список особенностей"`
}

type BeerUpdate struct{
	Beer
}

type PaginatedBeer struct {
	Offset uint `json:"offset"`
	Limit uint `json:"limit"`
	Total uint `json:"total"`
	TotalPages uint `json:"total_pages"`
	HasNext bool `json:"has_next"`
	HasPrev bool `json:"has_prev"`
}