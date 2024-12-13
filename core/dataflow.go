package core

import (
	"fmt"
	"log"
	"os"
	"variable-dataflow-tracer/crawler"
	"variable-dataflow-tracer/logger"
	"variable-dataflow-tracer/models"
	"variable-dataflow-tracer/services/dataFlowService"
	"variable-dataflow-tracer/services/languageService"
	"variable-dataflow-tracer/services/nodeService"
)

// -----------------------------------------------------------------------------
// RunDataflowAnalysis - Runs data flow analysis based on the provided configuration
// -----------------------------------------------------------------------------
//
// Parameters:
//   - config (models.Config): Configuration settings for the data flow analysis.
//
// Returns:
//   - ([]models.DataFlow): A slice of DataFlow models representing the result of the analysis.
//   - (error): An error object if an error occurred during the analysis.
//
// -----------------------------------------------------------------------------
func RunDataflowAnalysis(config models.Config) ([]models.DataFlow, error) {
	// Logger configuration
	logger.Setup(
		func(format string, v ...interface{}) {
			if config.Verbose {
				log.Printf("[INFO] "+format+"\n", v...)
			}
		},
		func(format string, v ...interface{}) {
			if config.Verbose {
				log.Printf("[WARNING] "+format+"\n", v...)
			}
		},
		func(format string, v ...interface{}) {
			log.Printf("[ERROR] "+format+"\n", v...)
		},
		func(format string, v ...interface{}) {
			if config.Debug {
				log.Printf("[DEBUG] "+format+"\n", v...)
			}
		},
	)

	dataflowInitial := dataFlowService.CreateDataflowInitial(config)

	// Check if the variable is defined
	if config.Variable == "" {
		logger.PrintError("Variable is not defined in config")
		return dataflowInitial, nil
	}

	// Set the global language
	models.GlobalLanguage = config.Language
	logger.PrintDebug("Global language set to: %s", config.Language)

	// Read the file content
	content, err := os.ReadFile(config.FilePath)
	if err != nil {
		logger.PrintError("error reading file: %v", err)
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse the content into a syntax tree
	tree := languageService.ParseContent(content, config.Language)
	if tree == nil {
		logger.PrintError("failed to parse the file into a syntax tree")
		return nil, fmt.Errorf("failed to parse the file into a syntax tree")
	}

	logger.PrintDebug("variable selected: %s", config.Variable)

	// Variables to track
	variablesToTrack := map[string]bool{config.Variable: true}

	// Analyze the variable in the function
	root := tree.RootNode()
	startingFunction := nodeService.FindFunctionByLine(root, uint32(config.StartLine), config.Language)
	if startingFunction == nil {
		return dataflowInitial, nil
	}

	visitedLines := make(map[uint32]bool)
	visitedFunctions := make(map[string]*models.VisitInfo)

	// Start data flow analysis
	result := crawler.CrawlFromLine(root, startingFunction, content, variablesToTrack, uint32(config.StartLine), true, visitedLines, visitedFunctions)

	// delete duplicate steps
	result = dataFlowService.RemoveDuplicateDataFlowStep(result, uint32(config.StartLine), config.Variable)

	// Add a verification step for the global variable
	result = nodeService.AddGlobalVariableSteps(result, root, content, uint32(config.StartLine))

	// Print the data flow
	if config.Verbose {
		models.PrintDataFlow(result)
	}

	// Create the data flow model
	dataflow := dataFlowService.CreateDataflow(result, content, config.StartLine, config.Language, config.FilePath, config.Variable)

	return dataflow, nil
}
