package entities

type Material struct {
	MaterialID     int64
	ProductID      int64
	MaterialAmount float32
}

type MaterialProduct struct{
	Material
	Product
}