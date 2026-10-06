package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var extensionToLanguage = map[string]string{
	".c":     "C",
	".cj":    "Cangjie",
	".cpp":   "C++",
	".cs":    "C#",
	".dart":  "Dart",
	".go":    "Go",
	".java":  "Java",
	".js":    "JavaScript",
	".kt":    "Kotlin",
	".nim":   "Nim",
	".php":   "PHP",
	".py":    "Python",
	".rb":    "Ruby",
	".rs":    "Rust",
	".scala": "Scala",
	".sh":    "Bash",
	".sql":   "SQL",
	".swift": "Swift",
	".ts":    "TypeScript",
}

const (
	metadataFile = "README_EN.md"
	zhReadmeFile = "README.md"
	descStart    = "<!-- description:start -->"
	descEnd      = "<!-- description:end -->"
	tabsStart    = "<!-- tabs:start -->"
)

var (
	solutionBlockRegex = regexp.MustCompile(`(?s)<!-- solution:start -->(.*?)<!-- solution:end -->`)
	headingRegex       = regexp.MustCompile(`(?m)^###\s+(.*)$`)
	thinkingRegex      = regexp.MustCompile(`(?s)<!-- thinking:start -->(.*?)<!-- thinking:end -->`)
	problemURLRegex    = regexp.MustCompile(`(?m)^# \[[^\]]*\]\((https://leetcode\.com/problems/([^/)]+)/?)\)`)
)

// Approach is the README explanation of one solution approach.
type Approach struct {
	Name        string
	Thinking    string
	Explanation string
}

type Metadata struct {
	Difficulty string   `yaml:"difficulty"`
	Rating     *int64   `yaml:"rating"`
	Source     string   `yaml:"source"`
	Tags       []string `yaml:"tags"`
	URL        string   `yaml:"-"`
	Slug       string   `yaml:"-"`
}

var solutionFileRegex = regexp.MustCompile(`^Solution(\d*)\.\w+$`)

var errNoSolutions = errors.New("no solution files found")

type Processor struct {
	root        string
	layout      string
	includeZh   bool
	writer      *DataWriter
	processed   int
	failed      int
	skipped     int
	badFiles    int
	writeErrors int
	records     int
}

func (proc *Processor) Process() error {
	err := filepath.WalkDir(proc.root, proc.walkDir)
	log.Printf("Processing complete. Processed: %d, Failed: %d, Skipped (no solutions): %d, Bad files: %d, Write errors: %d, Records: %d",
		proc.processed, proc.failed, proc.skipped, proc.badFiles, proc.writeErrors, proc.records)
	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}
	if proc.writeErrors > 0 {
		return fmt.Errorf("failed to write %d records", proc.writeErrors)
	}
	return nil
}

// Validate checks the run against quality thresholds; a negative maxFailures disables that check.
func (proc *Processor) Validate(minRecords, maxFailures int) error {
	if proc.records < minRecords {
		return fmt.Errorf("too few records: %d < %d", proc.records, minRecords)
	}
	if failures := proc.failed + proc.badFiles; maxFailures >= 0 && failures > maxFailures {
		return fmt.Errorf("too many failures: %d failed problems + %d bad solution files > %d",
			proc.failed, proc.badFiles, maxFailures)
	}
	return nil
}

func (proc *Processor) walkDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() && path != filepath.Join(proc.root, metadataFile) && filepath.Base(path) == metadataFile {
		dir := filepath.Dir(path)
		err = proc.processDir(dir)
		if errors.Is(err, errNoSolutions) {
			proc.skipped++
		} else if err != nil {
			proc.failed++
			log.Printf("Error processing %s: %v", filepath.Base(dir), err)
		}

		proc.processed++
		if proc.processed%100 == 0 {
			log.Printf("Processed %d directories...", proc.processed)
		}
	}

	return nil
}

func (proc *Processor) processDir(dir string) (err error) {
	dirTitle := filepath.Base(dir)
	id, title, err := proc.parseDir(dirTitle)
	if err != nil {
		return err
	}

	metadata, description, approaches, err := proc.parseMetadata(dir)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	tags := metadata.Tags
	if tags == nil {
		tags = []string{}
	}

	problem := Record{
		ID:          id,
		Title:       title,
		Slug:        metadata.Slug,
		URL:         metadata.URL,
		Difficulty:  metadata.Difficulty,
		Rating:      metadata.Rating,
		Source:      metadata.Source,
		Description: description,
		Tags:        tags,
	}

	if proc.includeZh {
		problem.DescriptionZh = parseDescriptionZh(dir)
	}

	solutionsFound := 0
	var records []Record
	for _, file := range files {
		matches := solutionFileRegex.FindStringSubmatch(file.Name())
		if file.IsDir() || matches == nil {
			continue
		}

		solutionsFound++
		if record, ok := proc.readSolution(dir, file.Name(), matches[1], problem, approaches); ok {
			records = append(records, record)
		}
	}

	if solutionsFound == 0 {
		return errNoSolutions
	}

	if proc.layout == ProblemsLayout {
		if len(records) > 0 {
			proc.write(toProblemRecord(problem, records), dirTitle)
		}
		return nil
	}

	for _, record := range records {
		proc.write(record, dirTitle+"/"+record.Language)
	}
	return nil
}

func (proc *Processor) write(record any, label string) {
	if err := (*proc.writer).Write(record); err != nil {
		proc.writeErrors++
		log.Printf("Error writing record %s: %v", label, err)
		return
	}
	proc.records++
}

// toProblemRecord groups the solution records of one problem; problem holds the problem-level fields.
func toProblemRecord(problem Record, records []Record) ProblemRecord {
	solutions := make([]SolutionRecord, 0, len(records))
	for _, r := range records {
		solutions = append(solutions, SolutionRecord{
			Language:    r.Language,
			Approach:    r.Approach,
			Name:        r.Name,
			Thinking:    r.Thinking,
			Explanation: r.Explanation,
			Solution:    r.Solution,
		})
	}
	return ProblemRecord{
		ID:            problem.ID,
		Title:         problem.Title,
		Slug:          problem.Slug,
		URL:           problem.URL,
		Difficulty:    problem.Difficulty,
		Rating:        problem.Rating,
		Source:        problem.Source,
		Description:   problem.Description,
		DescriptionZh: problem.DescriptionZh,
		Tags:          problem.Tags,
		Solutions:     solutions,
	}
}

// readSolution reads one solution file as a record; problem holds the problem-level fields.
func (proc *Processor) readSolution(
	dir, fileName, approachNumber string,
	problem Record,
	approaches []Approach,
) (Record, bool) {
	dirTitle := filepath.Base(dir)

	approach := int64(1)
	if approachNumber != "" {
		var err error
		approach, err = strconv.ParseInt(approachNumber, 10, 0)
		if err != nil {
			proc.badFiles++
			log.Printf("Invalid approach number in %s/%s: %v", dirTitle, fileName, err)
			return Record{}, false
		}
	}

	ext := filepath.Ext(fileName)
	lang, ok := extensionToLanguage[ext]
	if !ok {
		proc.badFiles++
		log.Printf("Unknown language for solution file %s/%s: %s", dirTitle, fileName, ext)
		return Record{}, false
	}

	content, err := os.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		proc.badFiles++
		log.Printf("Error reading solution file %s/%s: %v", dirTitle, fileName, err)
		return Record{}, false
	}

	var info Approach
	if approach <= int64(len(approaches)) {
		info = approaches[approach-1]
	}

	record := problem
	record.Language = lang
	record.Approach = approach
	record.Name = info.Name
	record.Thinking = info.Thinking
	record.Explanation = info.Explanation
	record.Solution = string(content)
	return record, true
}

func (proc *Processor) parseDir(dirTitle string) (id int64, title string, err error) {
	titleRegex := regexp.MustCompile(`^(\d+)\.(.+)$`)
	if matches := titleRegex.FindStringSubmatch(dirTitle); matches != nil {
		title = matches[2]
		idStr := matches[1]
		id, err = strconv.ParseInt(idStr, 10, 0)
	} else {
		err = fmt.Errorf("title does not match: %s", dirTitle)
	}
	return id, title, err
}

func (proc *Processor) parseMetadata(dir string) (metadata Metadata, description string, approaches []Approach, err error) {
	readme, err := os.ReadFile(filepath.Join(dir, metadataFile))
	if err != nil {
		return metadata, description, nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	content := string(readme)

	// Split the markdown into lines
	lines := strings.Split(content, "\n")

	// Find the end of the YAML frontmatter
	var yamlLines []string
	inYaml := false
	started := false
	yamlEndIndex := 0

	for i, line := range lines {
		if line == "---" {
			if !started {
				inYaml = true
				started = true
				continue
			}

			yamlEndIndex = i
			break
		}
		if inYaml {
			yamlLines = append(yamlLines, line)
		}
	}

	// Parse the metadata
	yamlContent := strings.Join(yamlLines, "\n")
	err = yaml.Unmarshal([]byte(yamlContent), &metadata)
	if err != nil {
		return metadata, "", nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	if matches := problemURLRegex.FindStringSubmatch(content); matches != nil {
		metadata.URL = matches[1]
		metadata.Slug = matches[2]
	}

	description, found := extractDescription(lines[yamlEndIndex+1:])
	if !found {
		return metadata, "", nil, errors.New("description markers not found")
	}

	return metadata, description, parseApproaches(content), nil
}

// extractDescription returns the text between the description markers.
func extractDescription(lines []string) (string, bool) {
	startIndex := -1
	for i, line := range lines {
		if strings.Contains(line, descStart) {
			startIndex = i
		}
		if strings.Contains(line, descEnd) {
			if startIndex == -1 {
				return "", false
			}
			return strings.TrimSpace(strings.Join(lines[startIndex+1:i], "\n")), true
		}
	}
	return "", false
}

// parseDescriptionZh returns the Chinese description from README.md, or "" when it is unavailable.
func parseDescriptionZh(dir string) string {
	readme, err := os.ReadFile(filepath.Join(dir, zhReadmeFile))
	if err != nil {
		return ""
	}
	description, _ := extractDescription(strings.Split(string(readme), "\n"))
	return description
}

// parseApproaches extracts the approach sections in order; section N describes Solution{N}.* files.
func parseApproaches(content string) []Approach {
	var approaches []Approach
	for _, block := range solutionBlockRegex.FindAllStringSubmatch(content, -1) {
		body := block[1]
		if i := strings.Index(body, tabsStart); i >= 0 {
			body = body[:i]
		}

		var approach Approach
		if loc := headingRegex.FindStringSubmatchIndex(body); loc != nil {
			approach.Name = parseApproachName(body[loc[2]:loc[3]])
			body = body[loc[1]:]
		}
		if loc := thinkingRegex.FindStringSubmatchIndex(body); loc != nil {
			approach.Thinking = cleanThinking(body[loc[2]:loc[3]])
			body = body[:loc[0]] + body[loc[1]:]
		}
		approach.Explanation = strings.TrimSpace(body)

		approaches = append(approaches, approach)
	}
	return approaches
}

// parseApproachName turns "Solution 1: Sliding Window" into "Sliding Window"; "Solution 1" yields "".
func parseApproachName(heading string) string {
	heading = strings.TrimSpace(heading)
	if _, name, found := strings.Cut(heading, ":"); found {
		return strings.TrimSpace(name)
	}
	if strings.HasPrefix(heading, "Solution") {
		return ""
	}
	return heading
}

// cleanThinking strips the blockquote markup and the "**Thinking**" title.
func cleanThinking(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i, line := range lines {
		line = strings.TrimPrefix(line, ">")
		lines[i] = strings.TrimPrefix(line, " ")
	}
	result := strings.TrimSpace(strings.Join(lines, "\n"))
	result = strings.TrimPrefix(result, "**Thinking**")
	return strings.TrimSpace(result)
}
