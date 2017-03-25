package schema_builder

const (
	DEFAULT_SELECTOR_UNIQUE_VALUE_THRESHOLD int = 200
	DEFAULt_VARCHAR_PADDING                 int = 1
	DEFAULT_NUMERIC_PADDING                 int = 1
	DEFAULT_PRECISION_PADDING               int = 0
)

var (
	SelectorUniqueValueThreshold int = DEFAULT_SELECTOR_UNIQUE_VALUE_THRESHOLD
	VarcharPadding               int = DEFAULt_VARCHAR_PADDING
	NumericPadding               int = DEFAULT_NUMERIC_PADDING
	PrecisionPadding             int = DEFAULT_PRECISION_PADDING
	ReservedColumns              []string
)

func init() {
	ReservedColumns = append(ReservedColumns, "event_timestamp")
	ReservedColumns = append(ReservedColumns, "event_duration")
}
