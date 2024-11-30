package main

import (
	"dataflow/core"
	"dataflow/logger"
	"dataflow/models"
	"flag"
	"fmt"
	"strings"
	"time"
)

func main() {
	// Initialisation du logger
	logger.Setup(
		func(format string, v ...interface{}) { fmt.Printf("[INFO] "+format, v...) },
		func(format string, v ...interface{}) { fmt.Printf("[WARNING] "+format, v...) },
		func(format string, v ...interface{}) { fmt.Printf("[ERROR] "+format, v...) },
		func(format string, v ...interface{}) { fmt.Printf("[DEBUG] "+format, v...) },
	)

	start := time.Now()

	// Définir les flags CLI
	filePath := flag.String("f", "", "Path to the code file to analyze")
	startLine := flag.Int("l", 0, "Line number to start the dataflow analysis")
	language := flag.String("lang", "go", "Programming language of the file (e.g., go, python, java)")
	variable := flag.String("var", "", "Variable to analyze")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	debug := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	// Vérification des arguments
	if *filePath == "" || *startLine == 0 || *language == "" || *variable == "" {
		logger.PrintError("Usage: go run main.go -f <file_path> -l <line_number> -lang <language> -var <variable> [-verbose] [-debug]")
		return
	}

	// Construire la configuration
	config := models.Config{
		FilePath:  *filePath,
		StartLine: *startLine,
		Language:  strings.ToLower(*language),
		Verbose:   *verbose,
		Debug:     *debug,
		Variable:  *variable,
	}

	// Exécuter l'analyse du flux de données
	result, err := core.RunDataflowAnalysis(config)
	if err != nil {
		logger.PrintError("Error: %v\n", err)
		return
	}
	_ = result

	elapsed := time.Since(start)
	logger.PrintInfo("Total execution time: %s\n", elapsed)
}
