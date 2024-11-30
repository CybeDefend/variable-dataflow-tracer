// Fonctions liées à la gestion du flux de données, en particulier celles qui manipulent et analysent les variables au sein du code.

package dataFlowService

import (
	"dataflow/logger"
	"dataflow/models"
	"os"
	"strings"
)

// Fonction auxiliaire pour calculer le maximum
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// Fonction auxiliaire pour calculer le minimumContainsString
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func CreateDataflow(dataflow []models.DataFlowStep, content []byte, startLine int, language, filePath, variable string) []models.DataFlow {
	var Dataflows []models.DataFlow

	lines := strings.Split(string(content), "\n")

	for i, step := range dataflow {
		// Créer l'objet Dataflow
		dto := models.DataFlow{
			NameHighlight: dataflow[i].Variable,
			Line:          int(step.Line),
			Language:      language,
			Path:          filePath,
			Type:          dataflow[i].Type,
			Order:         i + 1,
		}

		// Ajouter les lignes de code autour de la ligne concernée
		start := max(int(step.Line)-8, 0)
		end := min(int(step.Line)+7, len(lines)-1)

		for j := start; j <= end; j++ {
			code := models.CodeLine{
				Line:    j + 1,
				Content: lines[j],
			}
			dto.Code = append(dto.Code, code)
		}

		Dataflows = append(Dataflows, dto)
	}

	return Dataflows
}

// RemoveDuplicateDataFlowStep removes duplicates in the data flow while preserving the order
// and limits entries per line to a maximum of two (in case there are two variables on the same line).
func RemoveDuplicateDataFlowStep(elements []models.DataFlowStep, startLine uint32, variable string) []models.DataFlowStep {
	// Map to keep track of entries per variable per line
	varLineMap := make(map[string]map[uint32]models.DataFlowStep)
	// Map to keep count of entries per line
	lineCountMap := make(map[uint32]int)
	// Result slice to store the filtered data flow steps
	var result []models.DataFlowStep

	// Flag to track if a step exists on the start line for the specified variable
	stepExistsOnStartLine := false

	for _, element := range elements {
		line := element.Line
		variableName := element.Variable

		// Initialize the map for this variable if it doesn't exist
		if varLineMap[variableName] == nil {
			varLineMap[variableName] = make(map[uint32]models.DataFlowStep)
		}

		// Check if an entry for this variable on this line already exists
		existingElement, exists := varLineMap[variableName][line]
		if exists {
			// Compare priorities to decide whether to replace the existing element
			priorityExisting := getTypePriority(existingElement.Type)
			priorityNew := getTypePriority(element.Type)

			if priorityNew > priorityExisting {
				// Replace the existing element with the new one
				varLineMap[variableName][line] = element
				// Update the element in the result slice
				for i := range result {
					if result[i].Line == line && result[i].Variable == variableName {
						result[i] = element
						break
					}
				}
			}
			// Else, keep the existing element (no action needed)
		} else {
			// Check if we've already added two entries for this line
			if lineCountMap[line] >= 2 {
				continue // Skip adding more entries for this line
			}
			// Add the new element
			varLineMap[variableName][line] = element
			lineCountMap[line]++
			result = append(result, element)
		}
	}

	for _, element := range result {
		line := element.Line
		variableName := element.Variable

		if line == startLine && variableName == variable {
			stepExistsOnStartLine = true
		}
	}

	logger.PrintDebug("Data flow steps after removing duplicates: %v", stepExistsOnStartLine)

	// If no step exists on the start line for the variable, add it
	if !stepExistsOnStartLine {
		logger.PrintDebug("Adding missing step for variable '%s' on start line %d.", variable, startLine)
		newStep := models.DataFlowStep{
			Line:     startLine,
			Type:     "Use of variable",
			Method:   "", // No method
			Function: "", // No function
			Value:    variable,
			Variable: variable,
		}
		result = append(result, newStep)
	}

	if len(result) == 0 {
		return nil
	}
	for result[0].Line != startLine {
		result = result[1:]
		if len(result) == 0 {
			return nil
		}
	}
	// The result slice preserves the original order
	return result
}

// CreateDataflowWithValue crée une étape de dataflow basée sur une valeur identifiée.
func CreateDataflowInitial(config models.Config) []models.DataFlow {
	// Lire le contenu du fichier et diviser en lignes
	content, err := os.ReadFile(config.FilePath)
	if err != nil {
		logger.PrintError("Failed to read file: %v", err)
		return nil
	}
	lines := strings.Split(string(content), "\n")

	// Identifier les lignes de code autour de la ligne concernée
	start := max(config.StartLine-8, 0)
	end := min(config.StartLine+7, len(lines)-1)

	// Créer les lignes de code pour le champ "Code"
	var codeLines []models.CodeLine
	for i := start; i <= end; i++ {
		codeLines = append(codeLines, models.CodeLine{
			Line:    i + 1,
			Content: lines[i],
		})
	}

	// Créer le Dataflow
	dataflow := models.DataFlow{
		NameHighlight: config.Variable,
		Line:          config.StartLine,
		Code:          codeLines,
		Language:      config.Language,
		Path:          config.FilePath,
		Type:          "Vulnerability Value Usage",
		Order:         1,
	}

	return []models.DataFlow{dataflow}
}

// getTypePriority assigns a priority to each type of data flow step.
// Higher numbers indicate higher priority.
func getTypePriority(stepType string) int {
	switch stepType {
	case "Assignment of value":
		return 5
	case "Function parameters":
		return 4
	case "Variable used in return statement":
		return 3
	case "Variable used in 'if' condition":
		return 2
	case "Variable used in function call":
		return 2
	case "Variable used in assignment":
		return 1
	case "Use of variable":
		return 0
	default:
		return 0
	}
}