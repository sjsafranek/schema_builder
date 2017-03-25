package schema_builder

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

var (
	guard sync.RWMutex
	//MAX_LINES int = 10000
)

// ProcessCsvFile reads in a csv and attempts to classify the data type
// of each column.
// Supported data types: integer, fixed_point, varchar and selctor.
func ProcessCsvFile(inCsvFile string) ([]byte, error) {

	t1 := time.Now()

	job_id := newJobId(8)
	for {
		if _, ok := Jobs[job_id]; ok {
			job_id = newJobId(8)
		} else {
			break
		}
	}

	if Verbose {
		log.Println(`[Job] [`+job_id+`]`, `{"file":"`+inCsvFile+`","status":"start"}`)
	}

	// create work group for workers
	var workwg sync.WaitGroup

	// check if file exists
	if _, err := os.Stat(inCsvFile); os.IsNotExist(err) {
		return []byte(`{}`), err
	}

	// track number of lines in csv file
	num_lines := 0
	num_columns := 0

	// create job
	job := Job{Id: job_id, FileName: inCsvFile}
	Jobs[job_id] = &job

	// Load a TXT file.
	f, _ := os.Open(inCsvFile)
	// Create a new reader.
	r := csv.NewReader(bufio.NewReader(f))

	if Verbose {
		log.Println(`[Job] [`+job_id+`]`, `{"file":"`+inCsvFile+`","status":"reading file"}`)
	}

	// Read first line
	record, err := r.Read()

	var workers []Worker

	// loop through headers in csv
	for value := range record {
		// create worker and chan for each column
		workwg.Add(1)
		id := ""
		if 10 > value {
			id = fmt.Sprintf("0%v", value)
		} else {
			id = fmt.Sprintf("%v", value)
		}
		// create new worker
		worker := Worker{id: id, job_id: job_id, column_id: record[value], workwg: &workwg, Queue: make(chan string, 1000)}
		//worker.Queue <- record[value]
		workers = append(workers, worker)
		// track number of columns
		num_columns++
	}

	// if file is empty
	if err == io.EOF {
		for i := range workers {
			close(workers[i].Queue)
		}
		log.Println(job_id, "No lines in file")
		return []byte(`{}`), err
	}

	// run workers
	for i := range workers {
		workers[i].Run()
	}

	num_lines += 1

	if Verbose {
		message := fmt.Sprintf(`{"file":"`+inCsvFile+`","status":"classifying columns","details":{"workers":%v}}`, num_columns)
		log.Println(`[Job] [`+job_id+`]`, message)
	}

	// no columns in file
	if 0 == num_columns {
		return []byte(`{}`), fmt.Errorf("Couldn't find headers")
	}

	// read lines from csv file
	for {
		// Read line
		record, err := r.Read()

		num_lines += 1
		// Stop at EOF. - OR - once max number of lines is reached
		//if err == io.EOF || num_lines > MAX_LINES {
		if err == io.EOF {
			for i := range workers {
				close(workers[i].Queue)
			}
			break
		}

		for value := range record {
			// remove empty values
			if "" != record[value] {
				workers[value].Queue <- record[value]
			}
		}
	}

	if Verbose {
		log.Println(`[Job] [`+job_id+`]`, `{"file":"`+inCsvFile+`","status":"processing data"}`)
	}

	// wait for work groups to complete
	workwg.Wait()

	if Verbose {
		log.Println(`[Job] [`+job_id+`]`, `{"file":"`+inCsvFile+`","status":"processing complete"}`)
	}

	// collect job metadata
	runTime := time.Since(t1).Seconds()
	Jobs[job_id].RunTime = runTime
	Jobs[job_id].Rows = num_lines
	Jobs[job_id].Cols = num_columns
	results, err := json.Marshal(Jobs[job_id])
	if err != nil {
		log.Fatal(err)
	}

	// delete job from jobs
	guard.Lock()
	delete(Jobs, job_id)
	guard.Unlock()

	if Verbose {
		message := fmt.Sprintf(`{"file":"`+inCsvFile+`","status":"complete","details":{"run_time":%v,"rows":%v,"cols":%v}}`, runTime, num_lines, num_columns)
		log.Println(`[Job] [`+job_id+`]`, message)
	}

	return results, nil

}
