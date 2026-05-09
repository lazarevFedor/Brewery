package dto

type FeatureBase struct {
	Name string `json:"name"`
}

type Feature struct {
	FeatureBase
	
	ID uint `json:"id"`
}
