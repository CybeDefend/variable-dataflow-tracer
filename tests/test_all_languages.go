package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	core "github.com/CybeDefend/variable-dataflow-tracer/core"
	models "github.com/CybeDefend/variable-dataflow-tracer/models"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	reset     = "\033[0m"
	blue      = "\033[34m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	cyan      = "\033[36m"
	bold      = "\033[1m"
	underline = "\033[4m"
)

// -----------------------------------------------------------------------------
// clearConsole - Clears the console screen based on the operating system.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// -----------------------------------------------------------------------------
func clearConsole() {
	cmdName := "clear"
	if runtime.GOOS == "windows" {
		cmdName = "cls"
	}
	cmd := exec.Command(cmdName)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// -----------------------------------------------------------------------------
// displayMenu - Displays the main menu for selecting the test set.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - None
//
// Returns:
//   - None
//
// -----------------------------------------------------------------------------
func displayMenu() {
	clearConsole()
	fmt.Println(bold + cyan + "\n***************************************")
	fmt.Println("*                                     *")
	fmt.Println("*      Welcome to the test            *")
	fmt.Println("*                                     *")
	fmt.Println("***************************************\n" + reset)
	fmt.Println(underline + yellow + "Please choose the set of tests to run:" + reset)
	fmt.Println(blue + "  [1] Example 1: " + reset + "Tests with example1 files")
	fmt.Println(blue + "  [2] Example 2: " + reset + "Tests with example2 files")
	fmt.Println(blue + "  [3] Example 3: " + reset + "Tests with example3 files")
	fmt.Println(blue + "  [4] Example 4: " + reset + "Tests with example4 files")
	fmt.Println(blue + "  [5] Global Variables: " + reset + "Tests with exampleGlobal files")
	fmt.Print("\nEnter your choice (1, 2, 3, 4 or 5): ")
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	displayMenu()
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	// Définir les ensembles de tests
	tests1 := []struct {
		language  string
		filePath  string
		startLine int
		variable  string
	}{
		{"go", "tests/go/example1.go", 12, "filePath"},
		{"python", "tests/py/example1.py", 6, "newPath"},
		{"java", "tests/java/example1.java", 11, "newPath"},
		{"javascript", "tests/js/example1.js", 6, "newPath"},
		{"c", "tests/c/example1.c", 8, "newPath"},
		{"cpp", "tests/cpp/example1.cpp", 10, "newPath"},
		{"csharp", "tests/cs/example1.cs", 10, "newPath"},
		{"php", "tests/php/example1.php", 5, "newPath"},
		{"ruby", "tests/rb/example1.rb", 4, "newPath"},
		{"rust", "tests/rs/example1.rs", 7, "new_path"},
	}

	tests2 := []struct {
		language  string
		filePath  string
		startLine int
		variable  string
	}{
		{"go", "tests/go/example2.go", 11, "filePath"},
		{"python", "tests/py/example2.py", 3, "filePath"},
		{"java", "tests/java/example2.java", 5, "filePath"},
		{"javascript", "tests/js/example2.js", 3, "filePath"},
		{"c", "tests/c/example2.c", 7, "filePath"},
		{"cpp", "tests/cpp/example2.cpp", 9, "filePath"},
		{"csharp", "tests/cs/example2.cs", 8, "filePath"},
		{"php", "tests/php/example2.php", 5, "filePath"},
		{"ruby", "tests/rb/example2.rb", 3, "filePath"},
		{"rust", "tests/rs/example2.rs", 3, "filePath"},
	}

	tests3 := []struct {
		language  string
		filePath  string
		startLine int
		variable  string
	}{
		{"go", "tests/go/example3.go", 9, "filePath"},
		{"python", "tests/py/example3.py", 5, "filePath"},
		{"java", "tests/java/example3.java", 5, "filePath"},
		{"javascript", "tests/js/example3.js", 3, "filePath"},
		{"c", "tests/c/example3.c", 6, "filePath"},
		{"cpp", "tests/cpp/example3.cpp", 8, "filePath"},
		{"csharp", "tests/cs/example3.cs", 8, "filePath"},
		{"php", "tests/php/example3.php", 5, "filePath"},
		{"ruby", "tests/rb/example3.rb", 3, "filePath"},
		{"rust", "tests/rs/example3.rs", 5, "filePath"},
	}

	tests4 := []struct {
		language  string
		filePath  string
		startLine int
		variable  string
	}{
		{"go", "tests/go/example4.go", 17, "filePath"},
		{"python", "tests/py/example4.py", 12, "filePath"},
		{"java", "tests/java/example4.java", 15, "filePath"},
		{"javascript", "tests/js/example4.js", 12, "filePath"},
		{"c", "tests/c/example4.c", 22, "filePath"},
		{"cpp", "tests/cpp/example4.cpp", 23, "filePath"},
		{"csharp", "tests/cs/example4.cs", 19, "filePath"},
		{"php", "tests/php/example4.php", 15, "filePath"},
		{"ruby", "tests/rb/example4.rb", 13, "filePath"},
		{"rust", "tests/rs/example4.rs", 13, "filePath"},
	}

	testsGlobal := []struct {
		language  string
		filePath  string
		startLine int
		variable  string
	}{
		{"go", "tests/go/exampleGlobal.go", 29, "filePath"},
		{"python", "tests/py/exampleGlobal.py", 24, "filePath"},
		{"java", "tests/java/exampleGlobal.java", 25, "filePath"},
		{"javascript", "tests/js/exampleGlobal.js", 22, "filePath"},
		{"c", "tests/c/exampleGlobal.c", 16, "filePath"},
		{"cpp", "tests/cpp/exampleGlobal.cpp", 14, "filePath"},
		{"csharp", "tests/cs/exampleGlobal.cs", 34, "filePath"},
		{"php", "tests/php/exampleGlobal.php", 24, "filePath"},
		{"ruby", "tests/rb/exampleGlobal.rb", 11, "filePath"},
		{"rust", "tests/rs/exampleGlobal.rs", 22, "filePath"},
	}

	var tests []struct {
		language  string
		filePath  string
		startLine int
		variable  string
	}

	// Choix de l'ensemble de tests
	switch choice {
	case "1":
		tests = tests1
	case "2":
		tests = tests2
	case "3":
		tests = tests3
	case "4":
		tests = tests4
	case "5":
		tests = testsGlobal
	default:
		fmt.Println(red + "Invalid option. Please choose 1, 2, 3, 4 or 5" + reset)
		return
	}

	// Exécuter les tests choisis
	for _, test := range tests {
		executeTest(test, reader)
	}

	clearConsole()
	fmt.Println(bold + blue + "=== End of Dataflow Analysis Tests ===" + reset)
}

// executeTest - Executes a dataflow analysis test for a given programming language.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - test (struct): Contains the language, file path, start line, and variable to be analyzed.
//   - reader (*bufio.Reader): Reader to handle user input.
//
// Returns:
//   - (void): This function does not return a value.
//
// -----------------------------------------------------------------------------
func executeTest(test struct {
	language, filePath string
	startLine          int
	variable           string
}, reader *bufio.Reader) {
	clearConsole()

	// Construire la configuration
	config := models.Config{
		FilePath:  test.filePath,
		StartLine: test.startLine,
		Language:  test.language,
		Verbose:   false,
		Debug:     false,
	}

	// Exécuter l’analyse de dataflow
	dataflow, err := core.RunDataflowAnalysis(config)
	caser := cases.Title(language.Und) // Utiliser la bibliothèque pour gérer les majuscules

	if err != nil {
		fmt.Printf(red+"%s [Erreur]: %v\n"+reset, caser.String(test.language), err)
	} else if len(dataflow) == 0 {
		fmt.Printf(yellow+"%s : Aucun résultat de dataflow\n"+reset, caser.String(test.language))
	} else {
		// Affichage structuré pour un seul langage
		fmt.Println(bold + blue + "=== Dataflow Analysis for language: " + caser.String(test.language) + " ===\n" + reset)
		fmt.Printf(bold+"Files : %s\n"+reset, test.filePath)
		fmt.Printf(bold+"Start Line : %d\n\n"+reset, test.startLine)

		// Afficher les données dans l'ordre sans réorganiser
		for _, df := range dataflow {
			// Afficher directement les informations sans répéter la variable dans "Variable :"
			fmt.Printf("Ligne %d: %s -> %s\n", df.Line, df.Type, df.NameHighlight)
		}
	}

	fmt.Println(bold + cyan + "\n\nPress Enter to continue..." + reset)
	reader.ReadString('\n')
}

// extractCodeLine - Extracts a specific line of code from a file.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - filePath (string): The path to the file.
//   - lineNumber (int): The line number to extract.
//
// Returns:
//   - (string): The extracted line of code.
//   - (error): Error if the file cannot be opened or the line number is not found.
//
// -----------------------------------------------------------------------------
func extractCodeLine(filePath string, lineNumber int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == lineNumber {
			return strings.TrimSpace(scanner.Text()), nil
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return "", fmt.Errorf("line number %d not found in file", lineNumber)
}
