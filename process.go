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

type Metadata struct {
	Difficulty string   `yaml:"difficulty"`
	Tags       []string `yaml:"tags"`
}

type Processor struct {
	root   string
	writer *DataWriter
}

func (pr *Processor) Process() error {
	return filepath.WalkDir(pr.root, pr.walkDir)
}

func (pr *Processor) walkDir(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if !d.IsDir() && path != filepath.Join(pr.root, "README_EN.md") && filepath.Base(path) == "README_EN.md" {
		pr.processDir(path)
	}

	return nil
}

func (pr *Processor) processDir(path string) {
	dir := filepath.Dir(path)

	id, title, err := pr.parseDir(dir)
	if err != nil {
		log.Printf("Fail to parse id and title: %s", path)
		return
	}

	metadata, description, err := pr.parseReadme(path)
	if err != nil {
		log.Printf("Error reading README %s: %v", path, err)
		return
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Printf("Error reading directory %s: %v", dir, err)
		return
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

		if err = (*pr.writer).WriteRecord(record); err != nil {
			log.Printf("Error writing record: %v", err)
		}
	}
}

func (pr *Processor) parseDir(path string) (id int64, title string, err error) {
	base := filepath.Base(path)

	titleRegex := regexp.MustCompile(`^(\d+)\.(.+)$`)
	if matches := titleRegex.FindStringSubmatch(base); matches != nil {
		title = matches[2]
		idStr := matches[1]
		id, err = strconv.ParseInt(idStr, 10, 0)
	} else {
		err = errors.New("title does not match")
	}
	return id, title, err
}

func (pr *Processor) parseReadme(path string) (metadata Metadata, description string, err error) {
	readme, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading README %s: %v", path, err)
		return metadata, description, err
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

	// Extract description between <!-- description:start --> and <!-- description:end -->
	const descStart = "<!-- description:start -->"
	const descEnd = "<!-- description:end -->"

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
