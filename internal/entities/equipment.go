package entities

// Equipment представляет структуру оборудования / прибора.
type Equipment struct {
	Name       string `json:"name" info:"Название оборудования"`
	WorkCenter string `json:"work_center" info:"Название цеха"`
}
