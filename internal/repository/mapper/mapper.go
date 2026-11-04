package mapper

import (
	"Brewery/internal/entities"
	"Brewery/internal/repository/queries"
)

func ToDomainClass(dbClass queries.Class) entities.Class{
	return entities.Class{
		ID: dbClass.ID,
		Name: &dbClass.Name,
		ParentID: dbClass.ParentID,
		NumParamIDs: dbClass.NumericParameterIds,
		EnumParamIDs: dbClass.EnumParameterIds,
	}
}

func BaseToDomainClass(dbClass queries.InsertClassRow) entities.Class{
	return entities.Class{
		ID: dbClass.ID,
		Name: &dbClass.Name,
		ParentID: dbClass.ParentID,
	}
}

func ToDomainClasses(dbClasses []queries.Class) []entities.Class{
	classes := make([]entities.Class, len(dbClasses))
	for i, cl := range dbClasses{
		classes[i] = ToDomainClass(cl)
	}
	return classes
}