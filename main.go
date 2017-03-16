package main

import (
	"flag"
	"fmt"
	"log"
	//"os"
	"time"
)

var (
	start_time time.Time
	csv_file   string
	verbose    bool = false
)

func main() {
	Jobs = make(map[string]*Job)

	start_time = time.Now()

	flag.StringVar(&csv_file, "f", "", "csv file")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.StringVar(&port, "p", PORT, "server port")
	//flag.IntVar(&MAX_LINES, "m", MAX_LINES, "max lines to process")
	flag.Parse()

	if "" == csv_file {
		//fmt.Println("Incorrect usage!")
		//fmt.Println("Please provide csv file to process.")
		//flag.Usage()
		//os.Exit(1)
		Server()
	} else {
		results, err := processCsvFile(csv_file)
		if nil != err {
			log.Fatal(err)
		}
		fmt.Println(string(results))
	}

}
