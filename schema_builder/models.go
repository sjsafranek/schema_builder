package schema_builder

import (
	"sync"
	"time"
)

type ColumnSchema struct {
	Type       string           `json:"type"`
	ColumnId   string           `json:"column_id"`
	Attributes AttributesSchema `json:"attributes"`
}

type AttributesSchema struct {
	Length    int      `json:"length,omitempty"`
	MinValue  int      `json:"min_value,omitempty"`
	MaxValue  int      `json:"max_value,omitempty"`
	Precision int      `json:"precision,omitempty"`
	Values    []string `json:"values,omitempty"`
}

type Job struct {
	Id       string         `json:"id"`
	FileName string         `json:"filename"`
	Columns  []ColumnSchema `json:"columns"`
	RunTime  float64        `json:"run_time"`
	Rows     int            `json:"rows"`
	Cols     int            `json:"cols"`
}

type JobOptions struct {
	SelectorUniqueValueThreshold int
	VarcharPadding               int
	NumericPadding               int
	PrecisionPadding             int
}

type Worker struct {
	Queue     chan string
	job_id    string
	id        string
	column_id string
	workwg    *sync.WaitGroup
	startTime time.Time
	//Column ColumnMetadata
	Column ColumnSchema
}

// type ColumnMetadata struct {
// 	isString      bool
// 	isInt         bool
// 	isFloat       bool
// 	minValue      int
// 	maxValue      int
// 	count         int
// 	length        int
// 	precision     int
// 	unique_values map[string]int
// 	column_id string
// }
