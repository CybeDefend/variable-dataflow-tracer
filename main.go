package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/CybeDefend/variable-dataflow-tracer/core"
	"github.com/CybeDefend/variable-dataflow-tracer/logger"
	"github.com/CybeDefend/variable-dataflow-tracer/models"
)

func main() {
	logger.Setup(
		func(format string, v ...interface{}) { fmt.Printf("[INFO] "+format, v...) },
		func(format string, v ...interface{}) { fmt.Printf("[WARNING] "+format, v...) },
		func(format string, v ...interface{}) { fmt.Printf("[ERROR] "+format, v...) },
		func(format string, v ...interface{}) { fmt.Printf("[DEBUG] "+format, v...) },
	)

	start := time.Now()

	filePath := flag.String("f", "", "Path to the code file to analyze")
	startLine := flag.Int("l", 0, "Line number to start the dataflow analysis")
	language := flag.String("lang", "go", "Programming language of the file (e.g., go, python, java)")
	variable := flag.String("var", "", "Variable to analyze")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	debug := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	if *filePath == "" || *startLine == 0 || *language == "" || *variable == "" {
		logger.PrintError("Usage: go run main.go -f <file_path> -l <line_number> -lang <language> -var <variable> [-verbose] [-debug]")
		return
	}

	config := models.Config{
		FilePath:  *filePath,
		StartLine: *startLine,
		Language:  strings.ToLower(*language),
		Verbose:   *verbose,
		Debug:     *debug,
		Variable:  *variable,
	}

	result, err := core.RunDataflowAnalysis(config)
	if err != nil {
		logger.PrintError("Error: %v\n", err)
		return
	}
	_ = result

	elapsed := time.Since(start)
	logger.PrintInfo("Total execution time: %s\n", elapsed)
}
