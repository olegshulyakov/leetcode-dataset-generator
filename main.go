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
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	if err = validateFlags(); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}

	f, err := outputFile()
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close output file: %w", closeErr)
		}
	}()

	writer, err := NewDataWriter(*outputFormat, f)
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}

	processor := &Processor{
		root:   filepath.Join(*repoPath, "solution"),
		writer: writer,
	}
	processErr := processor.Process()

	if err = (*writer).Stop(); err != nil {
		return fmt.Errorf("failed to finalize output: %w", err)
	}

	return processErr
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
