package entities

// Material - строка таблицы materials: материал MaterialID входит в изделие ProductID в количестве MaterialAmount.
type Material struct {
	MaterialID     int64
	ProductID      int64
	MaterialAmount float32
}

// MaterialProduct - материал вместе с данными изделия (материал - это тоже изделие).
// Одна структура на все запросы по составу:
//   - состав изделия (FIND_BOM): Product - материал, MaterialAmount - сколько его на единицу родителя,
//     Materials - из чего состоит сам материал (пусто, если материал не составной);
//   - нормы расхода: Product - материал, MaterialAmount - сколько всего его нужно на изделие, Materials пустой;
//   - где используется материал: Product - изделие, MaterialAmount - сколько материала в него входит.
type MaterialProduct struct {
	Material
	Product

	Materials []MaterialProduct
}
