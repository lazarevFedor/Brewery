package mapper

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
	"time"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen .

// goverter:output:file ./mapper_impl.go
// goverter:output:package Brewery/internal/repository/mapper
// goverter:converter
// goverter:ignoreMissing
// goverter:useZeroValueOnPointerInconsistency
type ClassMapper interface {
	FromInsert(queries.InsertClassRow) entities.Class

	// goverter:map NumericParameterIds RangeParamIDs
	// goverter:map EnumParameterIds EnumParamIDs
	FromDBClass(queries.Class) entities.Class

	FromDBClasses([]queries.Class) []entities.Class
}

// goverter:output:file ./mapper_impl.go
// goverter:output:package Brewery/internal/repository/mapper
// goverter:converter
// goverter:ignoreMissing
// goverter:useZeroValueOnPointerInconsistency
type ProductMapper interface {
	// goverter:map ClassID ProdClass | ClassFromID
	FromInsert(queries.InsertProductRow) entities.Product
	// goverter:map ClassID ProdClass | ClassFromID
	FromUpdate(queries.Product) entities.Product

	// goverter:map . ProdClass | BuildClass
	FromDBFullProduct(queries.GetProductByIDRow) entities.Product
	// goverter:map ClassID ProdClass | ClassFromID
	FromDBProduct(queries.Product) entities.Product
	FromDBProducts([]queries.Product) []entities.Product
}

func ClassFromID(id int64) entities.Class {
	return entities.Class{ID: id}
}

func BuildClass(row queries.GetProductByIDRow) entities.Class {
	if row.ClassID == 0 {
		return entities.Class{}
	}

	return entities.Class{
		ID:            row.ClassID,
		Name:          row.ClassName,
		ParentID:      row.ClassParentID,
		RangeParamIDs: row.ClassNumParams,
		EnumParamIDs:  row.ClassEnumParams,
	}
}

// goverter:output:file ./mapper_impl.go
// goverter:output:package Brewery/internal/repository/mapper
// goverter:converter
// goverter:useZeroValueOnPointerInconsistency
type MaterialMapper interface {
	FromGetByProduct([]queries.GetMaterialsByProductIDRow) []entities.MaterialProduct
	FromGetByMaterial([]queries.GetProductsByMaterialRow) []entities.MaterialProduct
	FromDBMaterial(queries.Material) entities.Material

	// goverter:map . Product
	// goverter:map . Material
	// goverter:ignore Materials
	FromProductRow(queries.GetMaterialsByProductIDRow) entities.MaterialProduct

	// goverter:map . Product
	// goverter:map . Material
	// goverter:ignore Materials
	FromMaterialRow(queries.GetProductsByMaterialRow) entities.MaterialProduct

	// goverter:map ClassID ProdClass | ClassFromID
	// goverter:map Units Unit
	// goverter:ignore Amount Parameters
	ProductFromProductRow(queries.GetMaterialsByProductIDRow) entities.Product

	// goverter:map . ProdClass | ClassFromMaterialUsage
	// goverter:map Units Unit
	// goverter:ignore Amount Parameters
	ProductFromMaterialRow(queries.GetProductsByMaterialRow) entities.Product

	// goverter:map ID MaterialID
	// goverter:ignore ProductID
	MaterialFromProductRow(queries.GetMaterialsByProductIDRow) entities.Material

	// goverter:map ID ProductID
	// goverter:ignore MaterialID
	MaterialFromMaterialRow(queries.GetProductsByMaterialRow) entities.Material
}

// goverter:output:file ./mapper_impl.go
// goverter:output:package Brewery/internal/repository/mapper
// goverter:converter
// goverter:ignoreMissing
// goverter:useZeroValueOnPointerInconsistency
type HistoryMapper interface {
	// goverter:map CreatedAt CreatedAt | TimeToTime
	FromDBChangeRecord(queries.ChangeHistory) entities.HistoryRecord
	FromDBChangeHistory([]queries.ChangeHistory) []entities.HistoryRecord
}

func TimeToTime(t time.Time) time.Time {
	return t
}

func ClassFromMaterialUsage(row queries.GetProductsByMaterialRow) entities.Class {
	return entities.Class{ID: row.ClassID, Name: row.ClassName}
}
