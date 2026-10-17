package main

import (
	"errors"
	stdhtml "html"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const (
	TEMPLATE_FILENAME  = "template.html"
	MARKDOWN_EXTENSION = ".md"
	HTML_EXTENSION     = ".html"
	TITLE_SLOT         = "{{title}}"
	CONTENT_SLOT       = "{{content}}"
	SOURCE_DIR_NAME    = "pages"
	TARGET_DIR_NAME    = "target"
)

var rootSourceDir, targetDir string

func main() {
	rootSourceDir, targetDir = resolveDirectoriesFromArguments()
	if err := os.RemoveAll(targetDir); err != nil {
		log.Fatalf("Cannot clear target directory: %v", err)
	}
	buildDirectory(rootSourceDir, []string{})
}

func resolveDirectoriesFromArguments() (sourceDir string, outputDir string) {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <absolute path to the directory containing %s>", os.Args[0], SOURCE_DIR_NAME)
	}
	root := os.Args[1]
	if !filepath.IsAbs(root) {
		log.Fatal("Source directory path must be absolute")
	}

	sourceDir = filepath.Join(root, SOURCE_DIR_NAME)
	outputDir = filepath.Join(root, TARGET_DIR_NAME)
	return
}

func buildDirectory(sourceDir string, templates []string) {
	template := readDirectoryTemplate(sourceDir)
	if template != nil {
		templates = append(templates, string(template))
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatalf("Cannot list source directory: %v", err)
	}
	for _, entry := range entries {
		entryName := entry.Name()
		entryPath := filepath.Join(sourceDir, entryName)
		if strings.HasPrefix(entryName, ".") || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		if entry.IsDir() {
			buildDirectory(entryPath, templates)
		} else if hasExtension(entryName, MARKDOWN_EXTENSION) {
			page := renderMarkdownFile(entryPath, entryName, templates)
			writeRenderedPage(page, entryPath)
		} else if entryName != TEMPLATE_FILENAME {
			copyStaticFile(entryPath)
		}
	}
}

func readDirectoryTemplate(dirPath string) []byte {
	filename := filepath.Join(dirPath, TEMPLATE_FILENAME)

	content, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		log.Fatalf("Cannot read template: %v", err)
	}

	if !strings.Contains(string(content), CONTENT_SLOT) {
		log.Fatalf("%s has no %s slot", filename, CONTENT_SLOT)
	}
	return content
}

func hasExtension(filename, extension string) bool {
	return strings.ToLower(filepath.Ext(filename)) == extension
}

func renderMarkdownFile(path, name string, templates []string) string {
	markdownContent, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read markdown page: %v", err)
	}
	title := strings.TrimSuffix(name, filepath.Ext(name))
	return renderPage(title, markdownContent, templates)
}

func renderPage(title string, markdownContent []byte, templates []string) string {
	markdownParser := parser.NewWithExtensions(parser.CommonExtensions)
	document := markdownParser.Parse(markdownContent)
	renderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags})
	body := markdown.Render(document, renderer)

	layout := composeLayout(templates)
	layout = strings.ReplaceAll(layout, TITLE_SLOT, stdhtml.EscapeString(title))
	return strings.ReplaceAll(layout, CONTENT_SLOT, string(body))
}

func composeLayout(templates []string) string {
	count := len(templates)
	if count == 0 {
		return CONTENT_SLOT
	}
	if count == 1 {
		return templates[0]
	}
	outer := composeLayout(templates[:count-1])
	return strings.ReplaceAll(outer, CONTENT_SLOT, templates[count-1])
}

func writeRenderedPage(page, path string) {
	targetPath := toTargetPath(path)
	targetPath = strings.TrimSuffix(targetPath, filepath.Ext(targetPath)) + HTML_EXTENSION
	ensureParentDirectoryExists(targetPath)
	if err := os.WriteFile(targetPath, []byte(page), 0644); err != nil {
		log.Fatalf("Cannot write rendered page: %v", err)
	}
}

func toTargetPath(path string) string {
	return targetDir + strings.TrimPrefix(path, rootSourceDir)
}

func copyStaticFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read static file: %v", err)
	}

	targetPath := toTargetPath(path)
	ensureParentDirectoryExists(targetPath)
	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		log.Fatalf("Cannot copy static file: %v", err)
	}
}

func ensureParentDirectoryExists(path string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Fatalf("Cannot create target directory: %v", err)
	}
}
