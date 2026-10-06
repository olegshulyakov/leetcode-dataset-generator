package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/reader"
)

const fixtureRoot = "testdata/solution"

func TestParseDir(t *testing.T) {
	tests := []struct {
		dir     string
		id      int64
		title   string
		wantErr bool
	}{
		{dir: "0001.Two Sum", id: 1, title: "Two Sum"},
		{dir: "3000.A. B", id: 3000, title: "A. B"},
		{dir: "Two Sum", wantErr: true},
		{dir: "0001.", wantErr: true},
	}
	for _, tt := range tests {
		id, title, err := (&Processor{}).parseDir(tt.dir)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseDir(%q) error = %v, wantErr %v", tt.dir, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && (id != tt.id || title != tt.title) {
			t.Errorf("parseDir(%q) = %d, %q; want %d, %q", tt.dir, id, title, tt.id, tt.title)
		}
	}
}

func TestParseMetadata(t *testing.T) {
	metadata, description, approaches, err := (&Processor{}).parseMetadata(filepath.Join(fixtureRoot, "0000-0099", "0001.Two Sum"))
	if err != nil {
		t.Fatal(err)
	}

	rating := int64(1234)
	wantMetadata := Metadata{
		Difficulty: "Easy",
		Rating:     &rating,
		Source:     "Weekly Contest 1 Q1",
		Tags:       []string{"Array", "Hash Table"},
		URL:        "https://leetcode.com/problems/two-sum",
		Slug:       "two-sum",
	}
	if !reflect.DeepEqual(metadata, wantMetadata) {
		t.Errorf("metadata = %+v, want %+v", metadata, wantMetadata)
	}

	if want := "<p>Find two numbers that add up to <code>target</code>.</p>"; description != want {
		t.Errorf("description = %q, want %q", description, want)
	}

	wantApproaches := []Approach{
		{
			Name:        "Hash Table",
			Thinking:    "Checking every pair is quadratic.\n\nA hash table finds the complement in constant time.",
			Explanation: "Store each number's index in a hash table.\n\nThe time complexity is $O(n)$.",
		},
		{Thinking: "Sorting also works."},
	}
	if !reflect.DeepEqual(approaches, wantApproaches) {
		t.Errorf("approaches = %+v, want %+v", approaches, wantApproaches)
	}
}

func TestParseMetadataErrors(t *testing.T) {
	_, _, _, err := (&Processor{}).parseMetadata(filepath.Join(fixtureRoot, "0000-0099", "0002.Broken Problem"))
	if err == nil {
		t.Error("expected error for missing description markers")
	}

	_, _, _, err = (&Processor{}).parseMetadata(filepath.Join(fixtureRoot, "0000-0099", "missing"))
	if err == nil {
		t.Error("expected error for missing README")
	}
}

func TestParseApproachName(t *testing.T) {
	tests := map[string]string{
		"Solution 1: Sliding Window":   "Sliding Window",
		"Solution 2":                   "",
		"Method 1: Enumerate Lines":    "Enumerate Lines",
		"Solution: Single Pass":        "Single Pass",
		"Binary Search":                "Binary Search",
		" Solution 3:  Two Pointers  ": "Two Pointers",
	}
	for heading, want := range tests {
		if got := parseApproachName(heading); got != want {
			t.Errorf("parseApproachName(%q) = %q, want %q", heading, got, want)
		}
	}
}

// runFixture processes the fixture repository into a file of the given format.
func runFixture(t *testing.T, format, layout string) (*Processor, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "out."+format)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer, err := NewDataWriter(format, layout, f)
	if err != nil {
		t.Fatal(err)
	}
	proc := &Processor{root: fixtureRoot, layout: layout, writer: writer}
	if err = proc.Process(); err != nil {
		t.Fatal(err)
	}
	if err = writer.Stop(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	return proc, path
}

func TestProcessJSON(t *testing.T) {
	proc, path := runFixture(t, JSON, SolutionsLayout)

	if proc.records != 4 || proc.processed != 2 || proc.failed != 1 || proc.skipped != 1 || proc.badFiles != 1 {
		t.Errorf("records=%d processed=%d failed=%d skipped=%d badFiles=%d; want 4, 2, 1, 1, 1",
			proc.records, proc.processed, proc.failed, proc.skipped, proc.badFiles)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	records := map[string]Record{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var r Record
		if err = json.Unmarshal(scanner.Bytes(), &r); err != nil {
			t.Fatal(err)
		}
		records[fmt.Sprintf("%s/%s/%d", r.Slug, r.Language, r.Approach)] = r
	}

	py2, ok := records["two-sum/Python/2"]
	if !ok {
		t.Fatalf("missing Solution2.py record; got %v", reflect.ValueOf(records).MapKeys())
	}
	if py2.Thinking != "Sorting also works." || py2.Solution != "class Solution:\n    sort = True\n" {
		t.Errorf("unexpected Solution2.py record: %+v", py2)
	}
	if py1 := records["two-sum/Python/1"]; py1.Name != "Hash Table" || *py1.Rating != 1234 {
		t.Errorf("unexpected Solution.py record: %+v", py1)
	}
	if _, found := records["two-sum/Go/1"]; !found {
		t.Error("missing Solution.go record")
	}
	unknown := records["unknown-language/Python/1"]
	if unknown.Rating != nil || unknown.Name != "" || unknown.Tags == nil {
		t.Errorf("unexpected record without approaches: %+v", unknown)
	}
}

func TestProcessCSV(t *testing.T) {
	_, path := runFixture(t, CSV, SolutionsLayout)

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 5 {
		t.Fatalf("got %d rows, want header + 4", len(rows))
	}
	if len(rows[0]) != reflect.TypeOf(Record{}).NumField() {
		t.Errorf("header has %d columns, Record has %d fields", len(rows[0]), reflect.TypeOf(Record{}).NumField())
	}
}

func TestProcessParquet(t *testing.T) {
	_, path := runFixture(t, PARQUET, SolutionsLayout)

	fr, err := local.NewLocalFileReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer fr.Close()

	pr, err := reader.NewParquetReader(fr, new(Record), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer pr.ReadStop()

	if n := pr.GetNumRows(); n != 4 {
		t.Fatalf("got %d rows, want 4", n)
	}
	rows := make([]Record, 4)
	if err = pr.Read(&rows); err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if r.Slug == "two-sum" && !reflect.DeepEqual(r.Tags, []string{"Array", "Hash Table"}) {
			t.Errorf("tags = %v, want [Array Hash Table]", r.Tags)
		}
	}
}

func TestValidate(t *testing.T) {
	proc := &Processor{records: 10, failed: 1, badFiles: 1}
	tests := []struct {
		minRecords, maxFailures int
		wantErr                 bool
	}{
		{0, -1, false},
		{10, 2, false},
		{11, -1, true},
		{0, 1, true},
	}
	for _, tt := range tests {
		if err := proc.Validate(tt.minRecords, tt.maxFailures); (err != nil) != tt.wantErr {
			t.Errorf("Validate(%d, %d) error = %v, wantErr %v", tt.minRecords, tt.maxFailures, err, tt.wantErr)
		}
	}
}

func TestParseDescriptionZh(t *testing.T) {
	got := parseDescriptionZh(filepath.Join(fixtureRoot, "0000-0099", "0001.Two Sum"))
	if want := "<p>找出和为目标值的两个整数。</p>"; got != want {
		t.Errorf("parseDescriptionZh() = %q, want %q", got, want)
	}

	if got = parseDescriptionZh(filepath.Join(fixtureRoot, "0000-0099", "0003.New Problem")); got != "" {
		t.Errorf("parseDescriptionZh() without README.md = %q, want empty", got)
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		lines     []string
		want      string
		wantFound bool
	}{
		{[]string{descStart, " text ", descEnd}, "text", true},
		{[]string{"intro", descStart, descEnd}, "", true},
		{[]string{descEnd, descStart}, "", false},
		{[]string{descStart, "text"}, "", false},
	}
	for _, tt := range tests {
		got, found := extractDescription(tt.lines)
		if got != tt.want || found != tt.wantFound {
			t.Errorf("extractDescription(%q) = %q, %v; want %q, %v", tt.lines, got, found, tt.want, tt.wantFound)
		}
	}
}

func TestProcessProblemsLayoutJSON(t *testing.T) {
	proc, path := runFixture(t, JSON, ProblemsLayout)
	if proc.records != 2 {
		t.Errorf("records = %d, want 2 problems", proc.records)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var problem ProblemRecord
	if err = json.NewDecoder(bytes.NewReader(data)).Decode(&problem); err != nil {
		t.Fatal(err)
	}
	if problem.Slug != "two-sum" || len(problem.Solutions) != 3 {
		t.Fatalf("got %s with %d solutions, want two-sum with 3", problem.Slug, len(problem.Solutions))
	}
	if s := problem.Solutions[2]; s.Language != "Python" || s.Approach != 2 || s.Thinking != "Sorting also works." {
		t.Errorf("unexpected third solution: %+v", s)
	}
}

func TestProcessProblemsLayoutParquet(t *testing.T) {
	_, path := runFixture(t, PARQUET, ProblemsLayout)

	fr, err := local.NewLocalFileReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer fr.Close()

	pr, err := reader.NewParquetReader(fr, new(ProblemRecord), 1)
	if err != nil {
		t.Fatal(err)
	}
	defer pr.ReadStop()

	rows := make([]ProblemRecord, pr.GetNumRows())
	if err = pr.Read(&rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || len(rows[0].Solutions) != 3 || rows[0].Solutions[0].Language != "Go" {
		t.Errorf("unexpected problems: %+v", rows)
	}
}

func TestCSVRejectsProblemsLayout(t *testing.T) {
	f, err := os.Create(filepath.Join(t.TempDir(), "out.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err = NewDataWriter(CSV, ProblemsLayout, f); err == nil {
		t.Error("expected error for CSV with problems layout")
	}
}

func TestInSplit(t *testing.T) {
	train := &Processor{split: TrainSplit, testPercent: 5}
	test := &Processor{split: TestSplit, testPercent: 5}
	all := &Processor{split: AllSplit, testPercent: 5}

	testCount := 0
	for id := int64(1); id <= 10000; id++ {
		inTrain, inTest := train.inSplit(id), test.inSplit(id)
		if inTrain == inTest {
			t.Fatalf("problem %d: train=%v test=%v; want exactly one", id, inTrain, inTest)
		}
		if !all.inSplit(id) {
			t.Fatalf("problem %d is missing from the all split", id)
		}
		if inTest {
			testCount++
		}
	}
	if testCount < 400 || testCount > 600 {
		t.Errorf("test split has %d of 10000 problems, want about 500", testCount)
	}
}

func TestHTMLToMarkdown(t *testing.T) {
	tests := map[string]string{
		"<p>Use <code>nums</code>.</p>":                            "Use `nums`.",
		"<code>1 &lt;= n &lt;= 10<sup>4</sup></code>":              "`1 <= n <= 10^4`",
		"<p>2<sup>n-1</sup> and x<sub>i</sub></p>":                 "2^(n-1) and x\\_i",
		"<p><strong>Example:</strong></p><pre>a\nb</pre>":          "**Example:**\n\n```\na\nb\n```",
		"<p><code>a</code>&nbsp;b</p><p>&nbsp;</p><p>c\u00a0d</p>": "`a` b\n\nc d",
	}
	for html, want := range tests {
		got, err := htmlToMarkdown(html)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("htmlToMarkdown(%q) = %q, want %q", html, got, want)
		}
	}
}
