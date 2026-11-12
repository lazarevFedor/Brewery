package entities

// Equipment представляет структуру оборудования / прибора.
type Equipment struct {
	ID   int64  `json:"id" info:"ID оборудования"`
	Name string `json:"name" info:"Название оборудования"`
}
