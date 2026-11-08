package handlers

// Handlers объединяет все HTTP-обработчики для регистрации в роутере.
type Handlers struct {
	ClassHandler      ClassHandlers
	ProductHandler    ProductHandlers
	MaterialHandler   MaterialHandlers
	HistoryHandler    HistoryHandlers
	EnumHandler       EnumHandlers
	ParametersHandler ParametersHandlers
	AggregatesHandler AggregateHandlers
	AuthHandler       AuthHandlers
}
