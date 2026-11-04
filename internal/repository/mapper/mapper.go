package mapper

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
)

func ToDomainClass(dbClass queries.Class) entities.Class{
	return entities.Class{
		ID: dbClass.ID,
		Name: dbClass.Name,
		ParentID: *dbClass.ParentID,
		NumParamIDs: dbClass.NumericParameterIds,
		EnumParamIDs: dbClass.EnumParameterIds,
	}
}

func BaseToDomainClass(dbClass queries.InsertClassRow) entities.Class{
	return entities.Class{
		ID: dbClass.ID,
		Name: dbClass.Name,
		ParentID: *dbClass.ParentID,
	}
}

func ToDomainClasses(dbClasses []queries.Class) []entities.Class{
	classes := make([]entities.Class, len(dbClasses))
	for i, cl := range dbClasses{
		classes[i] = ToDomainClass(cl)
	}
	return classes
}

func Row2DomainProduct(product queries.InsertProductRow) entities.Product{
	return entities.Product{
		ID: product.ID,
		Name: product.Name,
		ProdClass: entities.Class{ID: product.ID},
    	BaseID: *product.BaseID,
    	Unit: *product.Units,
		Amount: *product.Amount,
	}
}

func ToDomainProduct(product queries.Product) entities.Product{
	return entities.Product{
		ID: product.ID,
		Name: product.Name,
		ProdClass: entities.Class{ID: product.ID},
    	BaseID: *product.BaseID,
    	Unit: *product.Units,
		Amount: *product.Amount,
	}
}

func ToDomainProducts(dbProducts []queries.Product) []entities.Product{
	products := make([]entities.Product, len(dbProducts))
	for i, prod := range dbProducts{
		products[i] = ToDomainProduct(prod)
	}
	return products
}