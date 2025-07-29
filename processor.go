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
	descStart    = "<!-- description:start -->"
	descEnd      = "<!-- description:end -->"
)

type Metadata struct {
	Difficulty string   `yaml:"difficulty"`
	Tags       []string `yaml:"tags"`
}

type Processor struct {
	root     string
	writer   *DataWriter
	dirCount int
}

func (proc *Processor) Process() error {
	defer func() {
		log.Printf("Processed %d directories...", proc.dirCount)
	}()
	return filepath.WalkDir(proc.root, proc.walkDir)
}

func (proc *Processor) walkDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() && path != filepath.Join(proc.root, metadataFile) && filepath.Base(path) == metadataFile {
		err = proc.processDir(path)
		if err != nil {
			log.Printf("Error processing %s: %v", path, err)
		}

		proc.dirCount++
		if proc.dirCount%100 == 0 {
			log.Printf("Processed %d directories...", proc.dirCount)
		}
	}

	return nil
}

func (proc *Processor) processDir(path string) (err error) {
	dir := filepath.Dir(path)

	id, title, err := proc.parseDir(dir)
	if err != nil {
		return err
	}

	metadata, description, err := proc.parseReadme(path)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %w", dir, err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasPrefix(file.Name(), "Solution.") {
			continue
		}

		fileName := file.Name()
		ext := filepath.Ext(fileName)
		lang, ok := extensionToLanguage[ext]
		if !ok {
			log.Printf("Unknown language for solution file %s/%s: %s", dir, fileName, ext)
			continue
		}

		var content []byte
		content, err = os.ReadFile(filepath.Join(dir, fileName))
		if err != nil {
			log.Printf("Error reading solution file %s: %v", fileName, err)
			continue
		}

		record := Record{
			ID:          id,
			Title:       title,
			Difficulty:  metadata.Difficulty,
			Description: description,
			Tags:        strings.Join(metadata.Tags, "; "),
			Language:    lang,
			Solution:    string(content),
		}

		if err = (*proc.writer).WriteRecord(record); err != nil {
			log.Printf("Error writing record: %v", err)
		}
	}
	return nil
}

func (proc *Processor) parseDir(path string) (id int64, title string, err error) {
	base := filepath.Base(path)

	titleRegex := regexp.MustCompile(`^(\d+)\.(.+)$`)
	if matches := titleRegex.FindStringSubmatch(base); matches != nil {
		title = matches[2]
		idStr := matches[1]
		id, err = strconv.ParseInt(idStr, 10, 0)
	} else {
		err = fmt.Errorf("title does not match: %s", base)
	}
	return id, title, err
}

func (proc *Processor) parseReadme(path string) (metadata Metadata, description string, err error) {
	readme, err := os.ReadFile(path)
	if err != nil {
		return metadata, description, fmt.Errorf("failed to read README %s: %w", path, err)
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
		return metadata, "", fmt.Errorf("failed to parse metadata: %w", err)
	}

	descStartIndex := -1
	descEndIndex := -1

	for i, line := range lines[yamlEndIndex+1:] {
		if strings.Contains(line, descStart) {
			descStartIndex = yamlEndIndex + 1 + i
		}
		if strings.Contains(line, descEnd) {
			descEndIndex = yamlEndIndex + 1 + i
			break
		}
	}

	if descStartIndex == -1 || descEndIndex == -1 {
		return metadata, "", errors.New("description markers not found")
	}

	descriptionLines := lines[descStartIndex+1 : descEndIndex]
	description = strings.Join(descriptionLines, "\n")
	description = strings.TrimSpace(description)

	return metadata, description, nil
}
