package parser

import (
	"path/filepath"
	"sync"
)

var GlobalSkipDirs = []string{
	".git", ".svn", ".hg", ".idea", ".vscode", ".vs",
	".DS_Store", "Thumbs.db",
}

var (
	registry = make(map[string]Parser)
	extMap   = make(map[string]string)
	mu       sync.RWMutex
)

func Register(p Parser) {
	mu.Lock()
	defer mu.Unlock()
	registry[p.Language()] = p
	for _, ext := range p.Extensions() {
		extMap[ext] = p.Language()
	}
}

func Get(language string) Parser {
	mu.RLock()
	defer mu.RUnlock()
	return registry[language]
}

func GetByExtension(ext string) Parser {
	mu.RLock()
	defer mu.RUnlock()
	if lang, ok := extMap[ext]; ok {
		return registry[lang]
	}
	return nil
}

func GetParserForFile(path string) Parser {
	ext := filepath.Ext(path)
	return GetByExtension(ext)
}

func AvailableLanguages() []string {
	mu.RLock()
	defer mu.RUnlock()
	langs := make([]string, 0, len(registry))
	for lang := range registry {
		langs = append(langs, lang)
	}
	return langs
}

func AllExtensions() []string {
	mu.RLock()
	defer mu.RUnlock()
	exts := make([]string, 0, len(extMap))
	for ext := range extMap {
		exts = append(exts, ext)
	}
	return exts
}

func AllLibDirs() []string {
	mu.RLock()
	defer mu.RUnlock()
	var dirs []string
	seen := make(map[string]bool)
	for _, p := range registry {
		for _, dir := range p.LibDirs() {
			if !seen[dir] {
				seen[dir] = true
				dirs = append(dirs, dir)
			}
		}
	}
	return dirs
}
