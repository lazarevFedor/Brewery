package handlers

// Handlers объединяет все HTTP-обработчики для регистрации в роутере.
type Handlers struct {
	ClassHandler           ClassHandlers
	ProductHandler         ProductHandlers
	MaterialHandler        MaterialHandlers
	HistoryHandler         HistoryHandlers
	EnumParametersHandler  EnumParametersHandlers
	EnumValuesHandler      EnumValuesHandlers
	RangeParametersHandler RangeParametersHandlers
	AggregatesHandler      AggregateHandlers
	AuthHandler            AuthHandlers
	EquipmentHandler       EquipmentHandlers
	OperationHandler       OperationHandlers
}
