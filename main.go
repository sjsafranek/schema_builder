package schema_builder

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
	flag.IntVar(&SelectorUniqueValueThreshold, "s", DEFAULT_SELECTOR_UNIQUE_VALUE_THRESHOLD, "Selector unique value threshold")
	flag.IntVar(&VarcharPadding, "vp", DEFAULt_VARCHAR_PADDING, "Varchar padding")
	flag.IntVar(&NumericPadding, "np", DEFAULT_NUMERIC_PADDING, "Numeric padding")
	flag.IntVar(&PrecisionPadding, "pp", DEFAULT_PRECISION_PADDING, "Precision padding")

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
