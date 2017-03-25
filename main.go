package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

import "./schema_builder"

var (
	csv_file string
)

func main() {
	// Jobs = make(map[string]*Job)

	flag.StringVar(&csv_file, "f", "", "csv file")
	flag.BoolVar(&schema_builder.Verbose, "v", false, "verbose")
	//flag.IntVar(&MAX_LINES, "m", MAX_LINES, "max lines to process")
	flag.IntVar(&schema_builder.SelectorUniqueValueThreshold, "s", schema_builder.DEFAULT_SELECTOR_UNIQUE_VALUE_THRESHOLD, "Selector unique value threshold")
	flag.IntVar(&schema_builder.VarcharPadding, "vp", schema_builder.DEFAULt_VARCHAR_PADDING, "Varchar padding")
	flag.IntVar(&schema_builder.NumericPadding, "np", schema_builder.DEFAULT_NUMERIC_PADDING, "Numeric padding")
	flag.IntVar(&schema_builder.PrecisionPadding, "pp", schema_builder.DEFAULT_PRECISION_PADDING, "Precision padding")

	flag.Parse()

	if "" == csv_file {
		fmt.Println("Incorrect usage!")
		os.Exit(1)
	}

	results, err := schema_builder.ProcessCsvFile(csv_file)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Println(string(results))

}
