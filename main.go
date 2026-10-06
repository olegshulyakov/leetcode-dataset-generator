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
	layout       = flag.String("layout", SolutionsLayout, "Row layout: solutions (one row per solution) or problems (one row per problem)")
	minRecords   = flag.Int("min-records", 0, "Fail if fewer records are written")
	split        = flag.String("split", AllSplit, "Problems to include: all, train, or test")
	testPercent  = flag.Int("test-percent", 5, "Percentage of problems assigned to the test split")
	descFormat   = flag.String("description-format", HTMLFormat, "Description format: html or markdown")
	includeZh    = flag.Bool("include-zh", false, "Fill description_zh with the Chinese problem description")
	maxFailures  = flag.Int("max-failures", -1, "Fail if more problems or solution files fail to parse (-1 disables)")
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

	writer, err := NewDataWriter(*outputFormat, *layout, f)
	if err != nil {
		return fmt.Errorf("failed to create writer: %w", err)
	}

	processor := &Processor{
		root:        filepath.Join(*repoPath, "solution"),
		layout:      *layout,
		split:       *split,
		testPercent: *testPercent,
		includeZh:   *includeZh,
		descFormat:  *descFormat,
		writer:      writer,
	}
	processErr := processor.Process()

	if err = (*writer).Stop(); err != nil {
		return fmt.Errorf("failed to finalize output: %w", err)
	}

	if processErr != nil {
		return processErr
	}

	return processor.Validate(*minRecords, *maxFailures)
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

	if *layout != SolutionsLayout && *layout != ProblemsLayout {
		return fmt.Errorf("unsupported layout: %s", *layout)
	}

	if *split != AllSplit && *split != TrainSplit && *split != TestSplit {
		return fmt.Errorf("unsupported split: %s", *split)
	}

	if *descFormat != HTMLFormat && *descFormat != MarkdownFormat {
		return fmt.Errorf("unsupported description format: %s", *descFormat)
	}

	if *testPercent < 0 || *testPercent > percentBase {
		return fmt.Errorf("test percent must be between 0 and 100: %d", *testPercent)
	}

	if *layout == ProblemsLayout && strings.EqualFold(*outputFormat, CSV) {
		return fmt.Errorf("the %s layout is not supported for CSV", ProblemsLayout)
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
