package schema_builder

const (
	// DEFAULT_SELECTOR_UNIQUE_VALUE_THRESHOLD default value for SelectorUniqueValueThreshold.
	DEFAULT_SELECTOR_UNIQUE_VALUE_THRESHOLD int = 200

	// DEFAULt_VARCHAR_PADDING default value for VarcharPadding.
	DEFAULt_VARCHAR_PADDING int = 1

	// DEFAULT_NUMERIC_PADDING default value for NumericPadding.
	DEFAULT_NUMERIC_PADDING int = 1

	// DEFAULT_PRECISION_PADDING default value for PrecisionPadding.
	DEFAULT_PRECISION_PADDING int = 0
)

var (
	// SelectorUniqueValueThreshold determines if a string column
	// is to be classified as a varchar or selector data type.
	SelectorUniqueValueThreshold int = DEFAULT_SELECTOR_UNIQUE_VALUE_THRESHOLD

	// VarcharPadding sets padding for max varchar length.
	VarcharPadding int = DEFAULt_VARCHAR_PADDING

	// NumericPadding sets padding for min and max values
	// in numeric columns.
	NumericPadding int = DEFAULT_NUMERIC_PADDING

	// PrecisionPadding sets padding for decimal precision
	// of fixed_point data types.
	PrecisionPadding int = DEFAULT_PRECISION_PADDING

	// ReservedColumns sets columns not to be classified.
	ReservedColumns []string

	// Verbose set verbose output during classification.
	Verbose bool = false

	// Jobs contains all active jobs.
	Jobs map[string]*Job
)

func init() {
	Jobs = make(map[string]*Job)
	ReservedColumns = append(ReservedColumns, "event_timestamp")
	ReservedColumns = append(ReservedColumns, "event_duration")
}
