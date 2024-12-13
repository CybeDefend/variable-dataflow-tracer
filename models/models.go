package models

import (
	"fmt"
	logger "variable-dataflow-tracer/logger"

	sitter "github.com/smacker/go-tree-sitter"
)

var (
	GlobalLanguage string
)

// DataFlowStep représente une étape dans le flux de données d'une variable.
type DataFlowStep struct {
	Line     uint32
	Type     string
	Method   string
	Function string
	Value    string
	Variable string
}

type CodeLine struct {
	Line    int    `json:"line"`
	Content string `json:"content"`
}

type DataFlow struct {
	NameHighlight string     `json:"nameHighlight"`
	Line          int        `json:"line"`
	Code          []CodeLine `json:"code"`
	Language      string     `json:"language"`
	Path          string     `json:"path"`
	Type          string     `json:"type"`
	Order         int        `json:"order"`
}

type VisitInfo struct {
	VisitedDef   bool
	VisitedCalls map[int]bool
	VisitCount   int
}

type FunctionCallSite struct {
	Line     uint32
	CallNode *sitter.Node
}

type Config struct {
	FilePath  string
	StartLine int
	Language  string
	Verbose   bool
	Debug     bool
	Variable  string
}

type AIRequestBody struct {
	Model       string      `json:"model"`
	Messages    []AIMessage `json:"messages"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float64     `json:"temperature"`
}

type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type IdentifyVariableRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type IdentifyVariableResponse struct {
	VariableOrValue string `json:"variableOrValue"`
	IsVariable      bool   `json:"isVariable"`
}

/**
 * -----------------------------------------------------------------------------
 * PrintDataFlow - Prints the data flow steps for a given variable.
 * -----------------------------------------------------------------------------
 *
 * Parameters:
 *   - dataFlow ([]DataFlowStep): List of data flow steps.
 *
 * Returns:
 *   - (void): This function does not return a value.
 *
 * -----------------------------------------------------------------------------
 */
func PrintDataFlow(dataFlow []DataFlowStep) {
	if len(dataFlow) == 0 {
		logger.PrintWarning("No data flow steps found.")
		return
	}

	fmt.Printf("Flux de données pour la variable '%s':\n", dataFlow[0].Variable)
	fmt.Println("----------------------------------------")
	for i, step := range dataFlow {
		fmt.Printf("Étape %d:\n", i+1)
		fmt.Printf(" Ligne: %d\n", step.Line)
		fmt.Printf(" Type: %s\n", step.Type)
		if step.Method != "" {
			fmt.Printf(" Méthode: %s\n", step.Method)
		}
		fmt.Printf(" Fonction: %s\n", step.Function)
		fmt.Printf(" Valeur: %s\n", step.Value)
		fmt.Printf(" Variable: %s\n", step.Variable)
		fmt.Println()
	}
}
