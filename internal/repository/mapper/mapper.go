package mapper

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
)

func ToDomainClass(dbClass queries.Class) entities.Class{
	var parentID *int64
	if dbClass.ParentID.Valid {
		parentID = &dbClass.ParentID.Int64
	}
	return entities.Class{
		ID: dbClass.ID,
		Name: dbClass.Name,
		ParentID: parentID,
		NumParamIDs: dbClass.NumericParameterIds,
		EnumParamIDs: dbClass.EnumParameterIds,
	}
}

func ToDomainClasses(dbClasses []queries.Class) []entities.Class{
	classes := make([]entities.Class, len(dbClasses))
	for i, cl := range dbClasses{
		classes[i] = *ToDomainClass(cl)
	}
	return classes
}