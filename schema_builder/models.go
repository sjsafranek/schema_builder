package schema_builder

import (
	"sync"
	"time"
)

type Classification struct {
	TotalValues  int    `json:"total"`
	UniqueValues int    `json:"unique"`
	StringReason string `json:"string_classification,omitempty"`
	// RunTime
}

// ColumnSchema struct to store column information.
type ColumnSchema struct {
	Type           string           `json:"type"`
	ColumnId       string           `json:"column_id"`
	Attributes     AttributesSchema `json:"attributes"`
	Classification Classification   `json:"classification"`
}

// AttributesSchema struct to store column metadata for classification.
type AttributesSchema struct {
	Length    int      `json:"length,omitempty"`
	MinValue  int      `json:"min_value,omitempty"`
	MaxValue  int      `json:"max_value,omitempty"`
	Precision int      `json:"precision,omitempty"`
	Values    []string `json:"values,omitempty"`
}

// Job struct for storing results and metadata.
type Job struct {
	Id       string         `json:"id"`
	FileName string         `json:"filename"`
	Columns  []ColumnSchema `json:"columns"`
	RunTime  float64        `json:"run_time"`
	Rows     int            `json:"rows"`
	Cols     int            `json:"cols"`
}

// Worker struct for classifying data types from a string channel.
type Worker struct {
	Queue     chan string
	job_id    string
	id        string
	column_id string
	workwg    *sync.WaitGroup
	startTime time.Time
	Column    ColumnSchema
}
