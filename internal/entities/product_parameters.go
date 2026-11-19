package entities

// ProductParameterValue содержит проверенные ссылки и значение для одной записи products_parameters.
type ProductParameterValue struct {
	FieldName        string
	RangeParameterID *int64
	NumericValue     *float64
	EnumParameterID  *int64
	EnumValueID      *int64
}

// ProductUpdate передает репозиторию поля изделия и проверенные изменения параметров.
type ProductUpdate struct {
	Fields             map[string]any
	ParameterFields    []string
	ParameterValues    []ProductParameterValue
	ReplaceParameters  bool
	ExpectedClassID    int64
	ExpectedParameters map[string]any
}
