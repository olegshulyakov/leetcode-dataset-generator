package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	repoPath     = flag.String("repo", ".", "Path to leetcode repository")
	outputFormat = flag.String("convert", PARQUET, "Output format: parquet, csv, or json")
	outputName   = flag.String("output", "leetcode-solutions", "Base output filename")
)

func main() {
	flag.Parse()
	if err := validateFlags(); err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}

	f, err := outputFile()
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	writer, err := NewDataWriter(*outputFormat, f)
	if err != nil {
		log.Printf("Failed to create writer: %v", err)
		return
	}
	defer (*writer).Stop()

	processor := &Processor{
		root:   filepath.Join(*repoPath, "solution"),
		writer: writer,
	}
	err = processor.Process()
	if err != nil {
		log.Printf("Error walking directory: %v", err)
	}
}

func validateFlags() error {
	if *repoPath == "" {
		return errors.New("repository path cannot be empty")
	}

	if *outputName == "" {
		return errors.New("output name cannot be empty")
	}

	validFormats := map[string]bool{
		PARQUET: true,
		CSV:     true,
		JSON:    true,
	}

	if !validFormats[strings.ToLower(*outputFormat)] {
		return fmt.Errorf("unsupported format: %s", *outputFormat)
	}

	return nil
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
