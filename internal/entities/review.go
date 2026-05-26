// Package entities содержит слой сущностей
package entities

// Review представляет полное описание отзыва о пиве.
type Review struct {
	ID     uint   `json:"id" info:"ID записи в бд"`
	Author string `json:"author" info:"Имя автора отзыва (пусто = аноним)"`
	Body   string `json:"body" info:"Комментарий отзыва"`
	BeerID uint   `json:"beer_id" info:"ID напитка, к которому прикреплен отзыв"`
	Rating uint   `json:"rating" info:"Оценка напитка"`
}

// Reviews представляет собой массив отзывов о пиве.
//
//easyjson:json
type Reviews []Review
