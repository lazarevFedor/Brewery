package entities

// ParameterValue содержит проверенные ссылки и значение параметра класса.
type ParameterValue struct {
	FieldName        string
	RangeParameterID *int64
	NumericValue     *float64
	EnumParameterID  *int64
	EnumValueID      *int64
}

// ParameterChanges описывает частичное обновление значений параметров класса.
type ParameterChanges struct {
	Fields             []string
	Values             []ParameterValue
	Replace            bool
	ExpectedClassID    int64
	ExpectedParameters map[string]any
}

// ProductUpdate передает репозиторию поля изделия и проверенные изменения параметров.
type ProductUpdate struct {
	Fields             map[string]any
	ParameterFields    []string
	ParameterValues    []ParameterValue
	ReplaceParameters  bool
	ExpectedClassID    int64
	ExpectedBaseID     int64
	ExpectedParameters map[string]any
}
