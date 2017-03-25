package schema_builder

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (self Worker) Run() {
	self.Column.unique_values = make(map[string]int)
	self.Column.precision = 1

	self.startTime = time.Now()
	go self.processQueue()
}

// Worker thread to classify column of csv file
func (self Worker) processQueue() {


	// hold unique values
	values := []string{}

	// item count
	count := 0

	// max length of string for varchar columns
	length := 0

	// decimal precision
	precision := 1

	// unique values
	unique_values := make(map[string]int)

	// report starting job
	if Verbose && "" != self.Column.column_id {
		message := fmt.Sprintf(`{"column_id":"%v","status":"classifying column"}`, self.Column.column_id)
		log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", message)
	}

	// if reserved column
	if StringInSlice(self.Column.column_id, ReservedColumns) {
		if Verbose {
			message := fmt.Sprintf(`{"column_id":"%v","status":"ets reserved column"}`, self.Column.column_id)
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
		if "" == self.Column.column_id {

			// set column_id
			self.Column.column_id = item

			if Verbose {
				message := fmt.Sprintf(`{"column_id":"%v","status":"classifying column"}`, self.Column.column_id)
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
				if !self.Column.isString {

					if strIsFloat(item) {

						// classify column as float
						self.Column.isFloat = true
						self.Column.isInt = false
						self.Column.isString = false

						// update min and max
						n, _ := strconv.Atoi(item)
						if 0 == self.Column.minValue && 0 == self.Column.maxValue {
							self.Column.minValue = n
							self.Column.maxValue = n
						} else {
							if self.Column.minValue > n {
								self.Column.minValue = n
							} else if self.Column.maxValue < n {
								self.Column.maxValue = n
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
						if !self.Column.isFloat {
							self.Column.isInt = true
							self.Column.isFloat = false
							self.Column.isString = false
						}

						// update min and max
						n, _ := strconv.Atoi(item)
						if 0 == self.Column.minValue && 0 == self.Column.maxValue {
							self.Column.minValue = n
							self.Column.maxValue = n
						} else {
							if self.Column.minValue > n {
								self.Column.minValue = n
							} else if self.Column.maxValue < n {
								self.Column.maxValue = n
							}
						}

					} else {

						// classify column as string
						self.Column.isString = true
						self.Column.isInt = false
						self.Column.isFloat = false

					}
				}

			}
		}

		// signal first work group as completed
		//workwg.Done()

	}

	// create column schema object
	column_schema := ColumnSchema{ColumnId: self.Column.column_id}

	// get unique values for varchar columns and job metadata
	for i := range unique_values {
		values = append(values, i)
	}

	// Determine data type of column
	if self.Column.isFloat {

		// classify as geographic_point or fixed_point column
		if "latitude" == self.Column.column_id || "longitude" == self.Column.column_id {

			// classify as geographic_point column
			column_schema.Type = "geographic_point"
			column_schema.ColumnId = "location"

		} else {

			// classify as fixed_point column
			column_schema.Type = "fixed_point"
			column_schema.Attributes.MinValue = self.Column.minValue - NumericPadding
			column_schema.Attributes.MaxValue = self.Column.maxValue + NumericPadding
			column_schema.Attributes.Precision = precision + PrecisionPadding

		}

	} else if self.Column.isInt {

		// classify as integer column
		column_schema.Type = "integer"
		column_schema.Attributes.MinValue = self.Column.minValue - NumericPadding
		column_schema.Attributes.MaxValue = self.Column.maxValue + NumericPadding

	} else if self.Column.isString {

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
	if Verbose {
		classification := fmt.Sprintf(`{"column_id":"%v","size":%v,"type":"%v","status":"classified column"}`, self.Column.column_id, len(values), column_schema.Type)
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

	if Verbose {
		classification := fmt.Sprintf(`{"column_id":"%v","run_time":%v,"status":"complete"}`, self.Column.column_id, time.Since(self.startTime).Seconds())
		log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", classification)
	}

	// signal second work group as completed
	self.workwg.Done()

}
