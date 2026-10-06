package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/xitongsys/parquet-go/writer"
)

// File extentions.
const (
	PARQUET = "parquet"
	CSV     = "csv"
	JSON    = "json"

	SolutionsLayout = "solutions"
	ProblemsLayout  = "problems"

	AllSplit   = "all"
	TrainSplit = "train"
	TestSplit  = "test"

	defaultParallelNumber = 4
)

type Record struct {
	ID            int64    `parquet:"name=id, type=INT64"                                                                    json:"id"`
	Title         string   `parquet:"name=title, type=BYTE_ARRAY, convertedtype=UTF8"                                        json:"title"`
	Slug          string   `parquet:"name=slug, type=BYTE_ARRAY, convertedtype=UTF8"                                         json:"slug"`
	URL           string   `parquet:"name=url, type=BYTE_ARRAY, convertedtype=UTF8"                                          json:"url"`
	Difficulty    string   `parquet:"name=difficulty, type=BYTE_ARRAY, convertedtype=UTF8"                                   json:"difficulty"`
	Rating        *int64   `parquet:"name=rating, type=INT64, repetitiontype=OPTIONAL"                                       json:"rating"`
	Source        string   `parquet:"name=source, type=BYTE_ARRAY, convertedtype=UTF8"                                       json:"source"`
	Description   string   `parquet:"name=description, type=BYTE_ARRAY, convertedtype=UTF8"                                  json:"description"`
	DescriptionZh string   `parquet:"name=description_zh, type=BYTE_ARRAY, convertedtype=UTF8"                               json:"description_zh"`
	Tags          []string `parquet:"name=tags, type=MAP, convertedtype=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags"`
	Language      string   `parquet:"name=language, type=BYTE_ARRAY, convertedtype=UTF8"                                     json:"language"`
	Approach      int64    `parquet:"name=approach, type=INT64"                                                              json:"approach"`
	Name          string   `parquet:"name=approach_name, type=BYTE_ARRAY, convertedtype=UTF8"                                json:"approach_name"`
	Thinking      string   `parquet:"name=thinking, type=BYTE_ARRAY, convertedtype=UTF8"                                     json:"thinking"`
	Explanation   string   `parquet:"name=explanation, type=BYTE_ARRAY, convertedtype=UTF8"                                  json:"explanation"`
	Solution      string   `parquet:"name=solution, type=BYTE_ARRAY, convertedtype=UTF8"                                     json:"solution"`
}

// ProblemRecord is one problem with all its solutions, written in the problems layout.
type ProblemRecord struct {
	ID            int64            `parquet:"name=id, type=INT64"                                                                    json:"id"`
	Title         string           `parquet:"name=title, type=BYTE_ARRAY, convertedtype=UTF8"                                        json:"title"`
	Slug          string           `parquet:"name=slug, type=BYTE_ARRAY, convertedtype=UTF8"                                         json:"slug"`
	URL           string           `parquet:"name=url, type=BYTE_ARRAY, convertedtype=UTF8"                                          json:"url"`
	Difficulty    string           `parquet:"name=difficulty, type=BYTE_ARRAY, convertedtype=UTF8"                                   json:"difficulty"`
	Rating        *int64           `parquet:"name=rating, type=INT64, repetitiontype=OPTIONAL"                                       json:"rating"`
	Source        string           `parquet:"name=source, type=BYTE_ARRAY, convertedtype=UTF8"                                       json:"source"`
	Description   string           `parquet:"name=description, type=BYTE_ARRAY, convertedtype=UTF8"                                  json:"description"`
	DescriptionZh string           `parquet:"name=description_zh, type=BYTE_ARRAY, convertedtype=UTF8"                               json:"description_zh"`
	Tags          []string         `parquet:"name=tags, type=MAP, convertedtype=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags"`
	Solutions     []SolutionRecord `parquet:"name=solutions, type=LIST"                                                              json:"solutions"`
}

// SolutionRecord is one solution inside a ProblemRecord.
type SolutionRecord struct {
	Language    string `parquet:"name=language, type=BYTE_ARRAY, convertedtype=UTF8"      json:"language"`
	Approach    int64  `parquet:"name=approach, type=INT64"                               json:"approach"`
	Name        string `parquet:"name=approach_name, type=BYTE_ARRAY, convertedtype=UTF8" json:"approach_name"`
	Thinking    string `parquet:"name=thinking, type=BYTE_ARRAY, convertedtype=UTF8"      json:"thinking"`
	Explanation string `parquet:"name=explanation, type=BYTE_ARRAY, convertedtype=UTF8"   json:"explanation"`
	Solution    string `parquet:"name=solution, type=BYTE_ARRAY, convertedtype=UTF8"      json:"solution"`
}

// DataWriter writes Record values in the solutions layout or ProblemRecord values in the problems layout.
type DataWriter interface {
	Write(record any) error
	Stop() error
}

type ParquetWriter struct {
	pw *writer.ParquetWriter
}

func (w *ParquetWriter) Write(record any) error {
	return w.pw.Write(record)
}

func (w *ParquetWriter) Stop() error {
	return w.pw.WriteStop()
}

type CSVWriter struct {
	cw *csv.Writer
}

func (w *CSVWriter) Write(record any) error {
	r, ok := record.(Record)
	if !ok {
		return fmt.Errorf("CSV supports only the %s layout, got %T", SolutionsLayout, record)
	}

	rating := ""
	if r.Rating != nil {
		rating = strconv.FormatInt(*r.Rating, 10)
	}
	return w.cw.Write([]string{
		strconv.FormatInt(r.ID, 10),
		r.Title,
		r.Slug,
		r.URL,
		r.Difficulty,
		rating,
		r.Source,
		r.Description,
		r.DescriptionZh,
		strings.Join(r.Tags, "; "),
		r.Language,
		strconv.FormatInt(r.Approach, 10),
		r.Name,
		r.Thinking,
		r.Explanation,
		r.Solution,
	})
}

func (w *CSVWriter) Stop() error {
	w.cw.Flush()
	return w.cw.Error()
}

type JSONWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func (w *JSONWriter) Write(record any) error {
	return w.encoder.Encode(record)
}

func (w *JSONWriter) Stop() error {
	return nil
}

func NewDataWriter(format, layout string, f *os.File) (*DataWriter, error) {
	var schema any = new(Record)
	if layout == ProblemsLayout {
		schema = new(ProblemRecord)
	}

	var out DataWriter
	switch strings.ToLower(format) {
	case PARQUET:
		pw, err := writer.NewParquetWriterFromWriter(f, schema, defaultParallelNumber)
		if err != nil {
			log.Printf("Failed to create parquet writer: %v", err)
			return nil, err
		}
		out = &ParquetWriter{pw: pw}
	case CSV:
		if layout == ProblemsLayout {
			return nil, fmt.Errorf("CSV does not support the %s layout", ProblemsLayout)
		}
		cw := csv.NewWriter(f)
		if err := cw.Write(
			[]string{
				"id",
				"title",
				"slug",
				"url",
				"difficulty",
				"rating",
				"source",
				"description",
				"description_zh",
				"tags",
				"language",
				"approach",
				"approach_name",
				"thinking",
				"explanation",
				"solution",
			},
		); err != nil {
			log.Printf("Failed to write CSV header: %v", err)
			return nil, err
		}
		out = &CSVWriter{cw: cw}
	case JSON:
		out = &JSONWriter{
			file:    f,
			encoder: json.NewEncoder(f),
		}
	default:
		log.Printf("Unsupported format: %s", format)
		return nil, errors.New("unsupported format")
	}

	return &out, nil
}
