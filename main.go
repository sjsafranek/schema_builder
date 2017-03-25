package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	csv_file string
	verbose  bool = false
)

func main() {
	Jobs = make(map[string]*Job)

	flag.StringVar(&csv_file, "f", "", "csv file")
	flag.BoolVar(&verbose, "v", false, "verbose")
	//flag.IntVar(&MAX_LINES, "m", MAX_LINES, "max lines to process")
	flag.Parse()

	if "" == csv_file {
		fmt.Println("Incorrect usage!")
		os.Exit(1)
	}

	results, err := ProcessCsvFile(csv_file)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Println(string(results))

}
