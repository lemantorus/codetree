package parser

import "codetree/internal/model"

type Parser interface {
	Language() string
	Extensions() []string
	LibDirs() []string
	Parse(content []byte) ([]model.CodeEntity, error)
}
