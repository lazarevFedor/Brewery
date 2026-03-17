package entities

type review struct {
	ID     int `json:"id" info:"ID записи в бд"`
	Body   string `json:"body" info:"Комментарий отзыва"`
	BeerID uint `json:"beer_id" info:"ID напитка, к которому прикреплен отзыв"`
	Rating float32 `json:"rating" info:"Оценка напитка"`
}
