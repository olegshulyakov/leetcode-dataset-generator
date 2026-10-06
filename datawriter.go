package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
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

	defaultParallelNumber = 4
)

type Record struct {
	ID          int64  `parquet:"name=id, type=INT64"                                   json:"id"`
	Title       string `parquet:"name=title, type=BYTE_ARRAY, convertedtype=UTF8"       json:"title"`
	Difficulty  string `parquet:"name=difficulty, type=BYTE_ARRAY, convertedtype=UTF8"  json:"difficulty"`
	Description string `parquet:"name=description, type=BYTE_ARRAY, convertedtype=UTF8" json:"description"`
	Tags        string `parquet:"name=tags, type=BYTE_ARRAY, convertedtype=UTF8"        json:"tags"`
	Language    string `parquet:"name=language, type=BYTE_ARRAY, convertedtype=UTF8"    json:"language"`
	Approach    int64  `parquet:"name=approach, type=INT64"                             json:"approach"`
	Solution    string `parquet:"name=solution, type=BYTE_ARRAY, convertedtype=UTF8"    json:"solution"`
}

type DataWriter interface {
	WriteRecord(Record) error
	Stop() error
}

type ParquetWriter struct {
	pw *writer.ParquetWriter
}

func (w *ParquetWriter) WriteRecord(r Record) error {
	return w.pw.Write(r)
}

func (w *ParquetWriter) Stop() error {
	return w.pw.WriteStop()
}

type CSVWriter struct {
	cw *csv.Writer
}

func (w *CSVWriter) WriteRecord(r Record) error {
	return w.cw.Write([]string{
		strconv.FormatInt(r.ID, 10),
		r.Title,
		r.Difficulty,
		r.Description,
		r.Tags,
		r.Language,
		strconv.FormatInt(r.Approach, 10),
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

func (w *JSONWriter) WriteRecord(r Record) error {
	return w.encoder.Encode(r)
}

func (w *JSONWriter) Stop() error {
	return nil
}

func NewDataWriter(format string, f *os.File) (*DataWriter, error) {
	var out DataWriter
	switch strings.ToLower(format) {
	case PARQUET:
		pw, err := writer.NewParquetWriterFromWriter(f, new(Record), defaultParallelNumber)
		if err != nil {
			log.Printf("Failed to create parquet writer: %v", err)
			return nil, err
		}
		out = &ParquetWriter{pw: pw}
	case CSV:
		cw := csv.NewWriter(f)
		if err := cw.Write([]string{"id", "title", "difficulty", "description", "tags", "language", "approach", "solution"}); err != nil {
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
