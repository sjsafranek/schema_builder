package main

var (
	SelectorUniqueValueThreshold int = 200
	VarcharPadding               int = 1
	NumericPadding               int = 1
	PrecisionPadding             int = 0
	EtsReservedColumns           []string
)

func init() {
	EtsReservedColumns = append(EtsReservedColumns, "event_timestamp")
	EtsReservedColumns = append(EtsReservedColumns, "event_duration")
}
