package tree

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lemantorus/codetree/internal/model"
	"github.com/lemantorus/codetree/internal/parser"
)

type EntityType = model.EntityType

type Builder struct {
	languages         map[string]bool
	entityTypes       map[EntityType]bool
	maxDepth          int
	showDocstr        bool
	extensions        map[string]bool
	includeLibs       bool
	excludePattern    *regexp.Regexp
	excludeIgnoreCase bool
}

type BuilderOption func(*Builder)

func WithLanguages(langs []string) BuilderOption {
	return func(b *Builder) {
		for _, lang := range langs {
			b.languages[lang] = true
		}
	}
}

func WithEntityTypes(types []EntityType) BuilderOption {
	return func(b *Builder) {
		for _, t := range types {
			b.entityTypes[t] = true
		}
	}
}

func WithMaxDepth(depth int) BuilderOption {
	return func(b *Builder) {
		b.maxDepth = depth
	}
}

func WithDocstrings(show bool) BuilderOption {
	return func(b *Builder) {
		b.showDocstr = show
	}
}

func WithExtensions(exts []string) BuilderOption {
	return func(b *Builder) {
		for _, ext := range exts {
			b.extensions[ext] = true
		}
	}
}

func WithIncludeLibs(include bool) BuilderOption {
	return func(b *Builder) {
		b.includeLibs = include
	}
}

func WithExcludePattern(pattern string, ignoreCase bool) BuilderOption {
	return func(b *Builder) {
		if pattern == "" {
			return
		}
		b.excludeIgnoreCase = ignoreCase
		if ignoreCase {
			pattern = "(?i)" + pattern
		}
		if re, err := regexp.Compile(pattern); err == nil {
			b.excludePattern = re
		}
	}
}

func NewBuilder(opts ...BuilderOption) *Builder {
	b := &Builder{
		languages:   make(map[string]bool),
		entityTypes: make(map[EntityType]bool),
		maxDepth:    -1,
		showDocstr:  true,
		extensions:  make(map[string]bool),
	}
	for _, opt := range opts {
		opt(b)
	}

	if len(b.languages) > 0 && len(b.extensions) == 0 {
		for lang := range b.languages {
			if p := parser.Get(lang); p != nil {
				for _, ext := range p.Extensions() {
					b.extensions[ext] = true
				}
			}
		}
	}

	if len(b.extensions) == 0 {
		for _, ext := range parser.AllExtensions() {
			b.extensions[ext] = true
		}
	}

	return b
}

func (b *Builder) Build(rootPath string) (*model.DirNode, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	root := &model.DirNode{
		Name:  filepath.Base(absPath),
		Path:  absPath,
		IsDir: true,
	}

	err = b.walk(absPath, root, 0)
	return root, err
}

func (b *Builder) shouldSkipDir(name string) bool {
	for _, skip := range parser.GlobalSkipDirs {
		if name == skip {
			return true
		}
	}

	if b.excludePattern != nil && b.excludePattern.MatchString(name) {
		return true
	}

	if !b.includeLibs {
		for _, libDir := range parser.AllLibDirs() {
			if strings.Contains(libDir, "*") {
				matched, _ := filepath.Match(libDir, name)
				if matched {
					return true
				}
			} else if name == libDir {
				return true
			}
		}
	}

	return false
}

func (b *Builder) walk(path string, parent *model.DirNode, depth int) error {
	if b.maxDepth >= 0 && depth > b.maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() && b.shouldSkipDir(name) {
			continue
		}

		fullPath := filepath.Join(path, name)

		if entry.IsDir() {
			dirNode := &model.DirNode{
				Name:  name,
				Path:  fullPath,
				IsDir: true,
			}
			parent.Children = append(parent.Children, dirNode)
			b.walk(fullPath, dirNode, depth+1)
		} else {
			ext := filepath.Ext(name)
			if !b.extensions[ext] {
				continue
			}

			p := parser.GetByExtension(ext)
			if p == nil {
				continue
			}

			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			entities, err := p.Parse(content)
			if err != nil {
				continue
			}

			entities = b.filterEntities(entities)

			fileNode := &model.DirNode{
				Name:     name,
				Path:     fullPath,
				IsDir:    false,
				Entities: entities,
			}
			parent.Children = append(parent.Children, fileNode)
		}
	}

	return nil
}

func (b *Builder) filterEntities(entities []model.CodeEntity) []model.CodeEntity {
	if len(b.entityTypes) == 0 {
		return entities
	}

	var filtered []model.CodeEntity
	for _, e := range entities {
		if b.entityTypes[e.Type] {
			filteredEntity := e
			filteredEntity.Children = b.filterEntities(e.Children)
			filtered = append(filtered, filteredEntity)
		}
	}
	return filtered
}

func (b *Builder) ShouldShowDocstrings() bool {
	return b.showDocstr
}
