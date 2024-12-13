// All functions that handle languages and content parsing based on supported languages.

package languageService

import (
	logger "github.com/CybeDefend/variable-dataflow-tracer/logger"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
)

// -----------------------------------------------------------------------------
// GetLanguage - Returns the language configuration for the specified language.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - language (string): The name of the programming language.
//
// Returns:
//   - (*sitter.Language): The language configuration object or nil if unsupported.
//
// -----------------------------------------------------------------------------
func GetLanguage(language string) *sitter.Language {
	switch language {
	case "go":
		return golang.GetLanguage()
	case "python":
		return python.GetLanguage()
	case "java":
		return java.GetLanguage()
	case "javascript":
		return javascript.GetLanguage()
	case "c":
		return c.GetLanguage()
	case "cpp":
		return cpp.GetLanguage()
	case "csharp":
		return csharp.GetLanguage()
	case "php":
		return php.GetLanguage()
	case "ruby":
		return ruby.GetLanguage()
	case "rust":
		return rust.GetLanguage()
	default:
		logger.PrintError("Unsupported language: %s", language)
		return nil
	}
}

// -----------------------------------------------------------------------------
// ParseContent - Parses the given content based on the specified programming language.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - content ([]byte): The content to be parsed.
//   - language (string): The programming language of the content.
//
// Returns:
//   - (*sitter.Tree): The parsed syntax tree of the content.
//
// -----------------------------------------------------------------------------
func ParseContent(content []byte, language string) *sitter.Tree {
	logger.PrintDebug("Parsing content for language: %s", language)
	parser := sitter.NewParser()

	switch language {
	case "go":
		parser.SetLanguage(golang.GetLanguage())
	case "python":
		parser.SetLanguage(python.GetLanguage())
	case "java":
		parser.SetLanguage(java.GetLanguage())
	case "javascript":
		parser.SetLanguage(javascript.GetLanguage())
	case "c":
		parser.SetLanguage(c.GetLanguage())
	case "cpp":
		parser.SetLanguage(cpp.GetLanguage())
	case "csharp":
		parser.SetLanguage(csharp.GetLanguage())
	case "php":
		parser.SetLanguage(php.GetLanguage())
	case "ruby":
		parser.SetLanguage(ruby.GetLanguage())
	case "rust":
		parser.SetLanguage(rust.GetLanguage())
	default:
		logger.PrintWarning("Unsupported language: %s", language)
		return nil
	}

	tree := parser.Parse(nil, content)
	if tree == nil {
		logger.PrintError("Failed to parse the input file content for language: %s", language)
	}

	return tree
}
