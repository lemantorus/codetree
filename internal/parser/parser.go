package parser

import "github.com/lemantorus/codetree/internal/model"

type Parser interface {
	Language() string
	Extensions() []string
	LibDirs() []string
	Parse(content []byte) ([]model.CodeEntity, error)
}
