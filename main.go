package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

var (
	repoPath     = flag.String("repo", ".", "Path to leetcode repository")
	outputFormat = flag.String("convert", PARQUET, "Output format: parquet, csv, or json")
	outputName   = flag.String("output", "leetcode-solutions", "Base output filename")
)

func main() {
	flag.Parse()

	f, err := outputFile()
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	writer, err := NewDataWriter(*outputFormat, f)
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
	extension := PARQUET
	switch strings.ToLower(*outputFormat) {
	case CSV:
		extension = CSV
	case JSON:
		extension = JSON
	case PARQUET:
		extension = PARQUET
	default:
		log.Fatalf("Unsupported format: %s", *outputFormat)
	}

	outputFile := *outputName + "." + extension
	return os.Create(outputFile)
}
