package main

import (
	"flag"
	"log"
	"os"
)

var (
	repoPath     = flag.String("repo", ".", "Path to leetcode repository")
	outputFormat = flag.String("convert", "parquet", "Output format: parquet, csv, or json")
	outputName   = flag.String("output", "leetcode-solutions", "Base output filename")
)

func main() {
	flag.Parse()

	f, err := outputFile()
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	writer, err := getDataWriter(*outputFormat, f)
	defer (*writer).Stop()
	if err != nil {
		log.Printf("Failed to create writer: %v", err)
		return
	}

	processor := &Processor{
		writer: writer,
	}
	err = processor.Process()
	if err != nil {
		log.Printf("Error walking directory: %v", err)
	}
}

func outputFile() (*os.File, error) {
	extension := "parquet"
	switch *outputFormat {
	case "csv":
		extension = "csv"
	case "json":
		extension = "json"
	case "parquet":
		extension = "parquet"
	default:
		log.Fatalf("Unsupported format: %s", *outputFormat)
	}

	outputFile := *outputName + "." + extension
	return os.Create(outputFile)
}
