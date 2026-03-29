package entities

// Review представляет полное описание отзыва о пиве
type Review struct {
	ID     uint     `json:"id" info:"ID записи в бд"`
	Body   string  `json:"body" info:"Комментарий отзыва"`
	BeerID uint    `json:"beer_id" info:"ID напитка, к которому прикреплен отзыв"`
	Rating float32 `json:"rating" info:"Оценка напитка"`
}

//easyjson:json
type Reviews []Review
