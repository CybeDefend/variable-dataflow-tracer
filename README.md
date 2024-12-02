# Variable Dataflow Tracer

Variable Dataflow Tracer is an open-source tool developed by Cybedefend that performs variable-specific data flow analysis across multiple programming languages. It traces the usage and origin of a specified variable within codebases to create comprehensive dataflow graphs. This aids developers and security engineers in understanding how data moves through their applications, particularly focusing on individual variables. The project is licensed under the Apache 2 and welcomes contributions from the community.

---

# Variable Dataflow Tracer

![License](https://img.shields.io/badge/license-apache--2.0-blue)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.16-blue)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen)

## Table of Contents

1. [Introduction](#introduction)
2. [Key Features](#key-features)
3. [Supported Languages](#supported-languages)
4. [Prerequisites](#prerequisites)
5. [Installation](#installation)
6. [Usage](#usage)
   - [As a Command-Line Tool](#as-a-command-line-tool)
   - [As a Library](#as-a-library)
   - [Arguments](#arguments)
   - [Examples](#examples)
7. [Testing](#testing)
8. [Code Structure](#code-structure)
9. [Limitations](#limitations)
10. [Recommendations for Use](#recommendations-for-use)
11. [Contributing](#contributing)
12. [License](#license)
13. [Contact](#contact)

## Introduction

**Variable Dataflow Tracer** is a versatile tool designed to analyze the data flow of specific variables in programs written in various programming languages. Leveraging [Tree-sitter](https://tree-sitter.github.io/tree-sitter/) for syntax parsing, it provides insights into how a particular variable traverses through a codebase, tracing it back to its origin and tracking its usage throughout the application. This tool is particularly useful for developers and security engineers aiming to understand the lifecycle of variables, assisting in debugging, vulnerability assessments, and code reviews.

## Key Features

- **Variable-Specific Analysis**: Focuses on tracing a specified variable within the code.
- **Multilingual Support**: Analyze code across multiple programming languages.
- **Recursive Data Flow Analysis**: Traces variables back to their sources by recursively traversing functions and function calls.
- **Automatic Variable Detection**: Optionally detects variables at a specified line if not provided.
- **Global Variable Tracking**: Tracks global variables and their values throughout the codebase.
- **Hybrid Usage**: Utilize as both a command-line tool and a library in other Go projects.
- **Detailed Logging**: Provides verbose and debug logging options to help diagnose issues and understand the analysis process.

## Supported Languages

The following languages are currently supported by **Variable Dataflow Tracer**:

- Go
- Python
- Java
- JavaScript
- C
- C++
- C#
- PHP
- Ruby
- Rust

*Note: The project is under active development. Additional languages and features will be added in future releases.*

## Prerequisites

- [Go](https://golang.org/doc/install) installed on your machine (version **1.16** or later recommended).

## Installation

### Prerequisites

Ensure you have **Docker** and **Visual Studio Code** installed, along with the **Dev Containers** extension enabled. This setup provides a consistent development environment using a container.

### Installation Steps

1. **Clone the Repository**:

   Clone the project locally to start development.

   ```bash
   git clone https://github.com/your-username/variable-dataflow-tracer.git
   cd variable-dataflow-tracer
   ```

2. **Open the Project in a Dev Container**:

   Open the project folder in Visual Studio Code, and let the Dev Containers extension automatically set up the environment.

   1. Open the command palette in VS Code (`Ctrl + Shift + P`).
   2. Search for and select: **Dev Containers: Rebuild and Reopen in Container**.
   3. Wait for the container to be built and ready for use.

   Once the Dev Container is active, you will have a fully configured environment with all necessary dependencies, including Go, GCC, and Clang.

3. **Install Project Dependencies**:

   Inside the Dev Container, run the following command to install the Go dependencies:

   ```bash
   go mod tidy
   ```

   This ensures that all required dependencies for the project are correctly installed.

---

### Summary

- **Why use a Dev Container?**: It simplifies dependency management (Go, GCC, Clang) and ensures a consistent development environment for all contributors.
- **Included Dependencies**: Go `1.22.7`, GCC, Clang, musl-dev.

## Usage

### As a Command-Line Tool

You can run **Variable Dataflow Tracer** directly from the command line.

```bash
go run main.go -f <file_path> -l <line_number> -lang <language> -var <variable_name> [--verbose] [--debug]
```

#### Arguments

- `-f <file_path>`: Path to the code file to be analyzed.
- `-l <line_number>`: Line number to start the dataflow analysis.
- `-lang <language>`: Programming language of the file (e.g., `go`, `python`, `java`, `javascript`, etc.).
- `-var <variable_name>`: Name of the variable to analyze in the data flow.
- `--verbose`: Enable verbose output for detailed information.
- `--debug`: Enable debug output for even more detailed information.

#### Examples

**Example 1: Specifying the Variable**

```bash
go run main.go -f ./tests/py/example1.py -l 20 -lang python -var myVariable --verbose --debug
```

This command analyzes the data flow of `myVariable` starting from line 20 in `example1.py`.

### As a Library

To use **Variable Dataflow Tracer** in your Go project:

1. **Import the Package**:

   ```go
   import (
       "github.com/your-username/variable-dataflow-tracer/core"
       "github.com/your-username/variable-dataflow-tracer/models"
   )
   ```

2. **Use the `RunDataflowAnalysis` Function**:

   ```go
   package main

   import (
       "fmt"
       "github.com/your-username/variable-dataflow-tracer/core"
       "github.com/your-username/variable-dataflow-tracer/models"
   )

   func main() {
       config := models.Config{
           FilePath:  "path/to/file",
           StartLine: 20,
           Language:  "go",
           Verbose:   true,
           Debug:     false,
           Variable:  "myVariable",
       }

       result, err := core.RunDataflowAnalysis(config)
       if err != nil {
           // Handle error
           fmt.Println("Error:", err)
           return
       }
       // Process result
       fmt.Println("Dataflow Result:", result)
   }
   ```

   In this example, the `Variable` field specifies the variable to trace.

## Testing

We have provided an initial test suite to help you get started. To run the tests:

1. Navigate to the `tests` directory:

   ```bash
   cd tests
   ```

2. Run the test script:

   ```bash
   go run test_all_languages.go
   ```

   *Note: You may need to configure the test files, line numbers, and variable names within `test_all_languages.go` according to your needs. An initial set of test cases is provided for example purposes.*

## Code Structure

The project is organized into several key components:

- **`core`**: Contains the main function `RunDataflowAnalysis`, which is the primary entry point for data flow analysis. This function configures the analysis based on provided parameters and invokes the crawler.
- **`crawler`**: Implements the core logic to traverse the code and build the dataflow graph. It recursively explores the code to trace the specified variable.
- **`logger`**: Provides logging functionalities at different levels (info, warning, error, debug), enabling users to monitor the analysis process.
- **`models`**: Defines the data structures used throughout the analysis process, such as `Config`, `Dataflow`, and other entities.
- **`services`**:
  - **`dataFlowService`**: Handles the creation and management of dataflow structures, assembling the traced paths.
  - **`languageService`**: Manages language-specific processing and supports multiple programming languages by abstracting language details.
  - **`nodeService`**: Deals with node recognition and processing based on the language syntax. This is the most extensive part of the code and may require significant improvements for enhanced performance and language support.
  - **`utilityService`**: Contains helper functions used across the application, such as file handling and string manipulation.
- **`tests`**: Contains test scripts and example code files for different languages to validate the tool's functionality.
- **`main.go`**: The entry point of the command-line interface (CLI), which parses arguments and initiates the analysis.

## Limitations

- **Incomplete Testing**: The project is in active development and is not fully tested. Users may encounter bugs or incomplete features.
- **Language Support**: While multiple languages are supported, the tool may not fully handle all language-specific constructs or complex code patterns.
- **Node Recognition**: The `nodeService` may require further refinement to improve accuracy and performance.
- **Complex Patterns**: The tool may struggle with highly complex codebases or unconventional coding patterns.

## Recommendations for Use

- **Contributions Welcome**: As the project is under development, we encourage contributions to improve language support, performance, and features.
- **Error Handling**: When integrating **Variable Dataflow Tracer** as a library, implement additional error handling and result processing to suit your specific use case.
- **Testing**: Before using the tool on critical codebases, run it against the provided tests and your own examples to understand its behavior.
- **Performance Considerations**: For large codebases, consider running the tool with optimized settings and be aware that analysis may take longer.

## Contributing

We appreciate contributions from the community to enhance **Variable Dataflow Tracer**. To contribute:

1. **Fork the Repository**:

   Click on the "Fork" button at the top right of the repository page.

2. **Create a Feature Branch**:

   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Commit Your Changes**:

   ```bash
   git commit -am "Add new feature"
   ```

4. **Push to Your Fork**:

   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request**:

   Submit a pull request to the `main` branch of the original repository.

### Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project, you agree to abide by its terms.

## License

This project is licensed under the terms of the Apache 2 license. See the [LICENSE](LICENSE) file for details.

## Contact

For any questions or suggestions, feel free to reach out to us:

- **Company**: Cybedefend
- **Email**: [contact@cybedefend.com](mailto:contact@cybedefend.com)
- **Website**: [www.cybedefend.com](https://www.cybedefend.com)
