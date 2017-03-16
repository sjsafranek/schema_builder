package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

type Worker struct {
	Queue     chan string
	job_id    string
	id        string
	column_id string
	workwg    *sync.WaitGroup
}

func (self Worker) Run() {
	go self.processQueue()
}

// Worker thread to classify column of csv file
func (self Worker) processQueue() {

	// record runtime for column classification
	var startTime time.Time

	// start time for classification
	startTime = time.Now()

	// column classification variables
	isString := false
	isInt := false
	isFloat := false

	// hold unique values
	values := []string{}

	// min and max values
	var minValue int
	var maxValue int

	// item count
	count := 0

	// max length of string for varchar columns
	length := 0

	// decimal precision
	precision := 1

	// unique values
	unique_values := make(map[string]int)

	// report starting job
	if verbose && "" != self.column_id {
		message := fmt.Sprintf(`{"column_id":"%v","status":"classifying column"}`, self.column_id)
		log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", message)
	}

	// if reserved column
	if StringInSlice(self.column_id, EtsReservedColumns) {
		if verbose {
			message := fmt.Sprintf(`{"column_id":"%v","status":"ets reserved column"}`, self.column_id)
			log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", message)
		}
		// drain chan contents
		for range self.Queue {
		}
		// signal second work group as completed
		self.workwg.Done()
		return
	}

	// read from channel
	for item := range self.Queue {

		// if column_id not defined
		if "" == self.column_id {

			// set column_id
			self.column_id = item

			if verbose {
				message := fmt.Sprintf(`{"column_id":"%v","status":"classifying column"}`, self.column_id)
				log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", message)
			}

		} else {

			// only classify column if item is not an empty string
			if "" != item {

				// count items
				// ** this will be used to determine selector vs varchar
				count++

				// store unique values
				// used to classify string column as varchar or selector
				unique_values[item] = len(item)

				// check length of string
				// ** used for varchar columns
				if unique_values[item] > length {
					length = unique_values[item]
				}

				// once column is classified as string stop checking
				if !isString {

					if strIsFloat(item) {

						// classify column as float
						isFloat = true
						isInt = false
						isString = false

						// update min and max
						n, _ := strconv.Atoi(item)
						if 0 == minValue && 0 == maxValue {
							minValue = n
							maxValue = n
						} else {
							if minValue > n {
								minValue = n
							} else if maxValue < n {
								maxValue = n
							}
						}

						// update fix_point precision
						index := strings.Index(item, ".")
						if -1 != index {
							if len(item)-index > precision {
								precision = len(item) - index
							}
						}

					} else if strIsInt(item) {

						// classify column as integer
						if !isFloat {
							isInt = true
							isFloat = false
							isString = false
						}

						// update min and max
						n, _ := strconv.Atoi(item)
						if 0 == minValue && 0 == maxValue {
							minValue = n
							maxValue = n
						} else {
							if minValue > n {
								minValue = n
							} else if maxValue < n {
								maxValue = n
							}
						}

					} else {

						// classify column as string
						isString = true
						isInt = false
						isFloat = false

					}
				}

			}
		}

		// signal first work group as completed
		//workwg.Done()

	}

	// create column schema object
	column_schema := ColumnSchema{ColumnId: self.column_id}

	// get unique values for varchar columns and job metadata
	for i := range unique_values {
		values = append(values, i)
	}

	// Determine data type of column
	if isFloat {

		// classify as geographic_point or fixed_point column
		if "latitude" == self.column_id || "longitude" == self.column_id {

			// classify as geographic_point column
			column_schema.Type = "geographic_point"
			column_schema.ColumnId = "location"

		} else {

			// classify as fixed_point column
			column_schema.Type = "fixed_point"
			column_schema.Attributes.MinValue = minValue - NumericPadding
			column_schema.Attributes.MaxValue = maxValue + NumericPadding
			column_schema.Attributes.Precision = precision + PrecisionPadding

		}

	} else if isInt {

		// classify as integer column
		column_schema.Type = "integer"
		column_schema.Attributes.MinValue = minValue - NumericPadding
		column_schema.Attributes.MaxValue = maxValue + NumericPadding

	} else if isString {

		// classify as varchar or selector column
		if len(values) > count/3 || len(values) > SelectorUniqueValueThreshold {

			// classify as varchar
			column_schema.Type = "varchar"
			column_schema.Attributes.Length = length + VarcharPadding

		} else {

			// classify as selector
			column_schema.Type = "selector"

			// sort and store values
			sort.Strings(values)
			column_schema.Attributes.Values = values

		}

	} else {

		// not enough information to classify
		column_schema.Type = "unknown"

	}

	// check if reserved column
	if verbose {
		classification := fmt.Sprintf(`{"column_id":"%v","size":%v,"type":"%v","status":"classified column"}`, self.column_id, len(values), column_schema.Type)
		log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", classification)
	}

	// add column schema to columns
	guard.Lock()
	if "location" == column_schema.ColumnId {
		found := false
		for i := range Jobs[self.job_id].Columns {
			if "location" == Jobs[self.job_id].Columns[i].ColumnId {
				found = true
			}
		}
		if !found {
			Jobs[self.job_id].Columns = append(Jobs[self.job_id].Columns, column_schema)
		}
	} else {
		Jobs[self.job_id].Columns = append(Jobs[self.job_id].Columns, column_schema)
	}
	guard.Unlock()

	if verbose {
		classification := fmt.Sprintf(`{"column_id":"%v","run_time":%v,"status":"complete"}`, self.column_id, time.Since(startTime).Seconds())
		log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", classification)
	}

	// signal second work group as completed
	self.workwg.Done()

}
