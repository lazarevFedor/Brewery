package mapper

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@latest gen .

// goverter:output:file ./mapper_impl.go
// goverter:output:package Brewery/internal/repository/mapper
// goverter:converter
// goverter:ignoreMissing
// goverter:useZeroValueOnPointerInconsistency
type ClassMapper interface {
	FromInsert(classRow queries.InsertClassRow) entities.Class

	// goverter:map NumericParameterIds NumParamIDs
	// goverter:map EnumParameterIds EnumParamIDs
	FromDBClass(dbClass queries.Class) entities.Class

	FromDBClasses(dbClasses []queries.Class) []entities.Class
}

// goverter:output:file ./mapper_impl.go
// goverter:output:package Brewery/internal/repository/mapper
// goverter:converter
// goverter:ignoreMissing
// goverter:useZeroValueOnPointerInconsistency
type ProductMapper interface {
	FromInsert(classRow queries.InsertProductRow) entities.Product
	FromUpdate(classRow queries.UpdateProductRow) entities.Product

	// goverter:map . ProdClass | BuildClass
	FromDBFullProduct(dbClass queries.GetProductByIDRow) entities.Product
	FromDBProduct(dbClass queries.Product) entities.Product
	FromDBProducts(dbClass []queries.Product) []entities.Product
}

func BuildClass(row queries.GetProductByIDRow) entities.Class {
	if row.ClassID == 0 {
		return entities.Class{}
	}

	return entities.Class{
		ID:            row.ClassID,
		Name:          row.ClassName,
		ParentID:      *row.ClassParentID,
		RangeParamIDs: row.ClassNumParams,
		EnumParamIDs:  row.ClassEnumParams,
	}
}
