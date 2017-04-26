package schema_builder

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Run starts worker
func (self Worker) Run() {
	self.Column.ColumnId = self.column_id
	self.startTime = time.Now()
	go self.processQueue()
}

// Worker thread to classify column of csv file
func (self Worker) processQueue() {

	isString := false
	isInt := false
	isFloat := false

	// item count
	count := 0

	// unique values
	unique_values := make(map[string]int)

	// report starting job
	if Verbose && "" != self.Column.ColumnId {
		message := fmt.Sprintf(`{"column_id":"%v","status":"classifying column"}`, self.Column.ColumnId)
		log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", message)
	}

	// if reserved column
	if StringInSlice(self.Column.ColumnId, ReservedColumns) {
		if Verbose {
			message := fmt.Sprintf(`{"column_id":"%v","status":"ets reserved column"}`, self.Column.ColumnId)
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
		if "" == self.Column.ColumnId {

			// set column_id
			self.Column.ColumnId = item

			if Verbose {
				message := fmt.Sprintf(`{"column_id":"%v","status":"classifying column"}`, self.Column.ColumnId)
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

				// check self.Column.Attributes.Length of string
				// ** used for varchar columns
				if unique_values[item] > self.Column.Attributes.Length {
					self.Column.Attributes.Length = unique_values[item]
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
						if 0 == self.Column.Attributes.MinValue && 0 == self.Column.Attributes.MaxValue {
							self.Column.Attributes.MinValue = n
							self.Column.Attributes.MaxValue = n
						} else {
							if self.Column.Attributes.MinValue > n {
								self.Column.Attributes.MinValue = n
							} else if self.Column.Attributes.MaxValue < n {
								self.Column.Attributes.MaxValue = n
							}
						}

						// update fix_point precision
						index := strings.Index(item, ".")
						if -1 != index {
							if len(item)-index > self.Column.Attributes.Precision {
								self.Column.Attributes.Precision = len(item) - index
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
						if 0 == self.Column.Attributes.MinValue && 0 == self.Column.Attributes.MaxValue {
							self.Column.Attributes.MinValue = n
							self.Column.Attributes.MaxValue = n
						} else {
							if self.Column.Attributes.MinValue > n {
								self.Column.Attributes.MinValue = n
							} else if self.Column.Attributes.MaxValue < n {
								self.Column.Attributes.MaxValue = n
							}
						}

					} else {

						self.Column.Classification.StringReason = fmt.Sprintf("row%v\nitem %v", count, item)

						// classify column as string
						isString = true
						isInt = false
						isFloat = false

					}
				}

			}
		}

	}

	// get unique values for varchar columns and job metadata
	values := []string{}
	for i := range unique_values {
		values = append(values, i)
	}

	// metadata on classification
	self.Column.Classification.UniqueValues = len(values)
	self.Column.Classification.TotalValues = count

	// Determine data type of column
	if isFloat {

		// classify as geographic_point or fixed_point column
		if "latitude" == self.Column.ColumnId || "longitude" == self.Column.ColumnId {

			// classify as geographic_point column
			self.Column.Type = "geographic_point"
			self.Column.ColumnId = "location"
			self.Column.Attributes.MinValue = 0
			self.Column.Attributes.MaxValue = 0
			self.Column.Attributes.Precision = 0
			self.Column.Attributes.Length = 0

		} else {

			// classify as fixed_point column
			self.Column.Type = "fixed_point"
			self.Column.Attributes.MinValue -= NumericPadding
			self.Column.Attributes.MaxValue += NumericPadding
			self.Column.Attributes.Precision += PrecisionPadding
			self.Column.Attributes.Length = 0
		}

	} else if isInt {

		// classify as integer column
		self.Column.Type = "integer"
		self.Column.Attributes.MinValue -= NumericPadding
		self.Column.Attributes.MaxValue += NumericPadding
		self.Column.Attributes.Precision = 0
		self.Column.Attributes.Length = 0

	} else if isString {

		// classify as varchar or selector column
		if len(values) > count/3 || len(values) > SelectorUniqueValueThreshold {

			// classify as varchar
			self.Column.Type = "varchar"
			self.Column.Attributes.MinValue = 0
			self.Column.Attributes.MaxValue = 0
			self.Column.Attributes.Precision = 0
			self.Column.Attributes.Length += VarcharPadding

		} else {

			// classify as selector
			// sort and store values
			sort.Strings(values)
			self.Column.Type = "selector"
			self.Column.Attributes.MinValue = 0
			self.Column.Attributes.MaxValue = 0
			self.Column.Attributes.Precision = 0
			self.Column.Attributes.Length = 0
			self.Column.Attributes.Values = values

		}

	} else {

		// not enough information to classify
		self.Column.Type = "unknown"

	}

	// check if reserved column
	if Verbose {
		classification := fmt.Sprintf(`{"column_id":"%v","size":%v,"type":"%v","status":"classified column"}`, self.Column.ColumnId, len(values), self.Column.Type)
		log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", classification)
	}

	self.Column.Classification.RunTime = fmt.Sprintf("%v", time.Since(self.startTime).Seconds())

	// add column schema to columns
	guard.Lock()
	// if "location" == column_schema.ColumnId {
	if "location" == self.Column.ColumnId {
		found := false
		for i := range Jobs[self.job_id].Columns {
			if "location" == Jobs[self.job_id].Columns[i].ColumnId {
				found = true
			}
		}
		if !found {
			Jobs[self.job_id].Columns = append(Jobs[self.job_id].Columns, self.Column)
		}
	} else {
		Jobs[self.job_id].Columns = append(Jobs[self.job_id].Columns, self.Column)
	}
	guard.Unlock()

	if Verbose {
		classification := fmt.Sprintf(`{"column_id":"%v","run_time":%v,"status":"complete"}`, self.Column.ColumnId, time.Since(self.startTime).Seconds())
		log.Println("[Worker-"+self.id+"] ["+self.job_id+"]", classification)
	}

	// signal second work group as completed
	self.workwg.Done()

}
