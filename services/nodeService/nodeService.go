// All functions that manipulate syntax nodes (Tree-sitter), such as finding a node, a function, or exploring the tree.

package nodeService

import (
	"strings"

	logger "github.com/CybeDefend/variable-dataflow-tracer/logger"
	models "github.com/CybeDefend/variable-dataflow-tracer/models"
	"github.com/CybeDefend/variable-dataflow-tracer/services/utilityService"

	sitter "github.com/smacker/go-tree-sitter"
)

/**** Functions Declaration Functions ****/

// -----------------------------------------------------------------------------
// FindParentFunction - Finds the name of the parent function for a given node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node for which to find the parent function.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The name of the parent function if found, otherwise an empty string.
//
// -----------------------------------------------------------------------------
func FindParentFunction(node *sitter.Node, content []byte) string {
	for node != nil {
		switch node.Type() {
		case "function_declaration", "method_declaration", "function_definition", "function_item", "method":
			// Attempt to get the function name from the "name" field
			funcName := node.ChildByFieldName("name")
			if funcName != nil {
				return SafeContent(funcName, content)
			}

			// Handle special cases where the function name is nested differently
			if node.Type() == "function_definition" {
				// For C and C++, the function name is under the "declarator"
				declarator := node.ChildByFieldName("declarator")
				if declarator != nil {
					funcNameNode := findIdentifierInDeclarator(declarator)
					if funcNameNode != nil {
						return SafeContent(funcNameNode, content)
					}
				}
			}
		}
		node = node.Parent()
	}
	return ""
}

// -----------------------------------------------------------------------------
// IsFunctionDeclaration - Determines if a node represents a function declaration.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - node (*sitter.Node): The node to check if it is a function declaration.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The name of the function if the node is a function declaration, otherwise an empty string.
//
// -----------------------------------------------------------------------------
func IsFunctionDeclaration(root, node *sitter.Node, content []byte) string {
	switch node.Type() {
	case "function_declaration", "method_declaration", "function_definition", "function_item", "method":
		// Attempt to get the function name from the "name" field
		funcNameNode := node.ChildByFieldName("name")
		if funcNameNode != nil {
			return SafeContent(funcNameNode, content)
		}

		// Handle special cases where the function name is nested differently
		if node.Type() == "function_definition" {
			// For C and C++, the function name is under the "declarator"
			declarator := node.ChildByFieldName("declarator")
			if declarator != nil {
				funcNameNode := findIdentifierInDeclarator(declarator)
				if funcNameNode != nil {
					return SafeContent(funcNameNode, content)
				}
			}
		}
	}
	return ""
}

// -----------------------------------------------------------------------------
// IsFunctionDeclared - Checks if a function with the given name is declared in the syntax tree.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - functionName (string): The name of the function to search for.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (bool): True if the function is declared, false otherwise.
//
// -----------------------------------------------------------------------------
func IsFunctionDeclared(root *sitter.Node, functionName string, content []byte) bool {
	var found bool

	var walkFunc func(node *sitter.Node)
	walkFunc = func(node *sitter.Node) {
		if node == nil || found {
			return
		}

		switch node.Type() {
		case "function_declaration", "method_declaration", "function_definition", "function_item", "method":
			// Attempt to get the function name from the "name" field
			funcNameNode := node.ChildByFieldName("name")
			if funcNameNode != nil {
				funcName := SafeContent(funcNameNode, content)
				if funcName == functionName {
					found = true
					return
				}
			} else if node.Type() == "function_definition" {
				// For C, C++, and PHP, the function name may be nested in the declarator
				declarator := node.ChildByFieldName("declarator")
				if declarator != nil {
					funcNameNode := findIdentifierInDeclarator(declarator)
					if funcNameNode != nil {
						funcName := SafeContent(funcNameNode, content)
						if funcName == functionName {
							found = true
							return
						}
					}
				}
			}
		}

		// Recursively check child nodes
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			walkFunc(child)
			if found {
				return
			}
		}
	}

	walkFunc(root)
	return found
}

// -----------------------------------------------------------------------------
// FindFunctionDeclaration - Finds the declaration of a function by its name in the syntax tree.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - functionName (string): The name of the function to find.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (*sitter.Node): The node representing the function declaration if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func FindFunctionDeclaration(root *sitter.Node, functionName string, content []byte) *sitter.Node {
	var result *sitter.Node
	var traverse func(node *sitter.Node)
	traverse = func(node *sitter.Node) {
		if node == nil || result != nil {
			return
		}
		switch node.Type() {
		case "function_declaration", "method_declaration", "function_definition", "function_item", "method":
			// Attempt to get the function name from the "name" field
			funcNameNode := node.ChildByFieldName("name")
			if funcNameNode != nil {
				name := SafeContent(funcNameNode, content)
				if name == functionName {
					result = node
					return
				}
			} else if node.Type() == "function_definition" {
				// For C, C++, PHP, and Python, the function name is under the "declarator"
				declarator := node.ChildByFieldName("declarator")
				if declarator != nil {
					funcNameNode := findIdentifierInDeclarator(declarator)
					if funcNameNode != nil {
						name := SafeContent(funcNameNode, content)
						if name == functionName {
							result = node
							return
						}
					}
				}
			} else if node.Type() == "function_item" {
				// For Rust, the function name is under the "name" field
				// Already handled by funcNameNode
			} else if node.Type() == "method" {
				// For Ruby, the function name is under the "name" field
				// Already handled by funcNameNode
			}
		}
		// Recursively traverse child nodes
		for i := 0; i < int(node.NamedChildCount()); i++ {
			traverse(node.NamedChild(i))
		}
	}
	traverse(root)
	return result
}

// -----------------------------------------------------------------------------
// FindFunctionDeclaration - Finds the declaration of a function by its name in the syntax tree.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - functionName (string): The name of the function to find.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (*sitter.Node): The node representing the function declaration if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func FindFunctionByName(root *sitter.Node, name string, content []byte) *sitter.Node {
	logger.PrintDebug("Searching for function '%s'...", name)

	var result *sitter.Node
	var traverse func(node *sitter.Node)
	traverse = func(node *sitter.Node) {
		if node == nil || result != nil {
			return
		}
		switch node.Type() {
		case "function_declaration", "method_declaration", "function_definition", "function_item", "method":
			// Attempt to get the function name from the "name" field
			funcNameNode := node.ChildByFieldName("name")
			if funcNameNode != nil {
				funcName := SafeContent(funcNameNode, content)
				logger.PrintDebug("Comparing function name '%s' with '%s'", funcName, name)
				if funcName == name {
					logger.PrintInfo("Function '%s' found at line %d", funcName, funcNameNode.StartPoint().Row+1)
					result = node
					return
				}
			} else if node.Type() == "function_definition" {
				// For C, C++, PHP, and Python, the function name may be nested under the "declarator"
				declarator := node.ChildByFieldName("declarator")
				if declarator != nil {
					funcNameNode := findIdentifierInDeclarator(declarator)
					if funcNameNode != nil {
						funcName := SafeContent(funcNameNode, content)
						logger.PrintDebug("Comparing function name '%s' with '%s'", funcName, name)
						if funcName == name {
							logger.PrintInfo("Function '%s' found at line %d", funcName, funcNameNode.StartPoint().Row+1)
							result = node
							return
						}
					}
				}
			}
		}
		// Recursively traverse child nodes
		for i := 0; i < int(node.NamedChildCount()); i++ {
			traverse(node.NamedChild(i))
		}
	}
	traverse(root)
	if result == nil {
		logger.PrintDebug("Function '%s' not found", name)
	}
	return result
}

// -----------------------------------------------------------------------------
// FindFunctionByLine - Finds the function node that contains the specified line.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - line (uint32): The line number to search for.
//   - lang (string): The programming language of the source code.
//
// Returns:
//   - (*sitter.Node): The node representing the function if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func FindFunctionByLine(root *sitter.Node, line uint32, lang string) *sitter.Node {
	var result *sitter.Node

	var traverse func(node *sitter.Node)
	traverse = func(node *sitter.Node) {
		if node == nil || result != nil {
			return
		}

		// Check if the node is a function node
		switch node.Type() {
		case "function_declaration", "method_declaration", "function_definition", "function_item", "method":
			startLine := node.StartPoint().Row + 1
			endLine := node.EndPoint().Row + 1
			if startLine <= line && line <= endLine {
				result = node
				return
			}
		}

		// Recursively traverse child nodes
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			traverse(child)
			if result != nil {
				return
			}
		}
	}

	traverse(root)

	return result
}

// -----------------------------------------------------------------------------
// FindFunctionBounds - Finds the start and end lines of a function containing the specified line.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - node (*sitter.Node): The node to check if it is a function declaration.
//   - startLine (uint32): The line number to search for.
//
// Returns:
//   - (functionStart, functionEnd uint32): The start and end lines of the function.
//
// -----------------------------------------------------------------------------
func FindFunctionBounds(root *sitter.Node, node *sitter.Node, startLine uint32) (functionStart, functionEnd uint32) {
	var currentNode *sitter.Node

	// Find the node at the specified line
	currentNode = FindNodeAtLine(root, startLine)

	// Traverse up the tree to find the function node
	for currentNode != nil {
		switch currentNode.Type() {
		case "function_declaration", "method_declaration", "function_definition", "function_item", "method":
			functionStart = currentNode.StartPoint().Row + 1
			functionEnd = currentNode.EndPoint().Row + 1
			return functionStart, functionEnd
		}
		currentNode = currentNode.Parent()
	}

	// If no function node is found, default to startLine
	functionStart = startLine
	functionEnd = startLine

	return functionStart, functionEnd
}

/**** Utils Functions ****/

// -----------------------------------------------------------------------------
// SafeContent - Securely extracts the content of a node from the content of the file.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node from which to extract content.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The extracted content as a string.
//
// -----------------------------------------------------------------------------
func SafeContent(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}

	start := int(node.StartByte())
	end := int(node.EndByte())

	// Validation des bornes de début et de fin
	if start < 0 || end > len(content) || start >= end {
		return ""
	}

	return string(content[start:end])
}

// -----------------------------------------------------------------------------
// IsValidVariableToTrack - Verifies if the variable name corresponds to a function declaration.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - variable (string): The variable name to check.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (bool): True if the variable name does not correspond to a function declaration, false otherwise.
//
// -----------------------------------------------------------------------------
func IsValidVariableToTrack(root *sitter.Node, variable string, content []byte) bool {
	// Vérifier si le nom de la variable correspond à une déclaration de fonction
	return !IsFunctionDeclared(root, variable, content)
}

// -----------------------------------------------------------------------------
// FindNodeAtLine - Finds the node at the specified line in the syntax tree.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The root node of the syntax tree.
//   - targetLine (uint32): The line number to search for.
//
// Returns:
//   - (*sitter.Node): The node at the specified line if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func FindNodeAtLine(node *sitter.Node, targetLine uint32) *sitter.Node {
	if node.StartPoint().Row+1 == targetLine {
		return node
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if result := FindNodeAtLine(child, targetLine); result != nil {
			return result
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// GetControlType - Returns a descriptive string for control flow statements across multiple languages.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - nodeType (string): The type of the node.
//
// Returns:
//   - (string): A descriptive string for the control flow statement.
//
// -----------------------------------------------------------------------------
func GetControlType(nodeType string) string {
	switch nodeType {
	// Conditional statements
	case "if_statement", "if_expression", // Go, C, C++, Java, JavaScript, C#, PHP, Python, Rust
		"if",                     // Ruby
		"conditional_expression": // JavaScript ternary operator
		return "Variable used in 'if' condition"

	// Loop statements
	case "for_statement", "enhanced_for_statement", "foreach_statement", "for_in_statement", "for_of_statement", // Go, C, C++, Java, JavaScript, C#, PHP
		"for_expression",                                   // Rust
		"for",                                              // Ruby
		"do_statement", "do_expression", "loop_expression": // Go, C, C++, Java, JavaScript, C#, PHP, Rust
		return "Variable used in loop condition"

	// While loops
	case "while_statement", "while_expression", // C, C++, Java, JavaScript, C#, PHP, Python, Rust
		"while": // Ruby
		return "Variable used in 'while' loop condition"

	// Switch statements
	case "switch_statement", "switch_expression", "match_expression": // Go, C, C++, Java, JavaScript, C#, PHP, Rust
		return "Variable used in 'switch' or 'match' statement"

	// Return statements
	case "return_statement", "return_expression": // Go, C, C++, Java, JavaScript, C#, PHP, Python, Rust
		return "Variable used in return statement"

	// Function calls
	case "call_expression", "function_call_expression", "method_invocation", "invocation_expression", "call": // Go, C, C++, Java, JavaScript, C#, PHP, Python, Ruby
		return "Variable used in function call"

	// Goroutines (Go-specific)
	case "go_statement": // Go
		return "Variable used in goroutine launch"

	// Identifier usage
	case "identifier", "variable_name", "variable": // General identifier usage across languages
		return "Use of variable"

	// Exception handling
	case "try_statement", "catch_clause", "except_clause", "finally_clause": // Java, C#, JavaScript, Python, PHP
		return "Variable used in exception handling"

	// Others
	case "assignment_expression", "assignment_statement", "short_var_declaration": // Assignments
		return "Variable used in assignment"

	default:
		return ""
	}
}

// -----------------------------------------------------------------------------
// IsLiteral - Checks if the node represents a literal value in the syntax tree.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to check.
//
// Returns:
//   - (bool): True if the node represents a literal value, false otherwise.
//
// -----------------------------------------------------------------------------
func IsLiteral(node *sitter.Node) bool {
	if node == nil {
		return false
	}

	// Node types representing literals in different languages
	literalNodeTypes := []string{
		// Go
		"interpreted_string_literal", "raw_string_literal", "int_literal", "float_literal", "rune_literal",
		// C/C++
		"string_literal", "char_literal", "number_literal",
		// C#
		"string_literal", "character_literal", "integer_literal", "real_literal",
		// Java
		"string_literal", "character_literal", "decimal_integer_literal", "hex_integer_literal", "octal_integer_literal", "binary_integer_literal", "floating_point_literal",
		// JavaScript
		"string", "number",
		// PHP
		"string", "encapsed_string", "integer", "floating_point",
		// Python
		"string", "integer", "float",
		// Ruby
		"string", "integer", "float",
		// Rust
		"string_literal", "char_literal", "integer_literal", "float_literal",
	}

	if utilityService.ContainString(literalNodeTypes, node.Type()) {
		return true
	}

	// Check for nested literal content (e.g., string content)
	nestedLiteralTypes := []string{"string_content", "string_fragment", "string_literal_content"}
	if utilityService.ContainString(nestedLiteralTypes, node.Type()) {
		return true
	}

	// Recursively check child nodes
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if IsLiteral(child) {
			return true
		}
	}

	return false
}

// -----------------------------------------------------------------------------
// findIdentifierInDeclarator - Recursively searches for an identifier in a declarator node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The declarator node to search.
//
// Returns:
//   - (*sitter.Node): The identifier node if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func findIdentifierInDeclarator(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}
	if node.Type() == "identifier" || node.Type() == "variable_name" || node.Type() == "name" {
		return node
	}
	// Recurse into all children, not just named ones
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if result := findIdentifierInDeclarator(child); result != nil {
			return result
		}
	}
	return nil
}

/**** Call Functions ****/

// -----------------------------------------------------------------------------
// FindCallExpression - Finds the call expression node in a given node for multiple languages.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to search for a call expression.
//
// Returns:
//   - (*sitter.Node): The call expression node if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func FindCallExpression(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}

	switch node.Type() {
	case "call_expression", "invocation_expression", "method_invocation", "function_call_expression", "call", "macro_invocation":
		return node
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		result := FindCallExpression(child)
		if result != nil {
			return result
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// ExtractFunctionNameFromCall - Extracts the function name from a function call node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The function call node.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The name of the function being called.
//
// -----------------------------------------------------------------------------
func ExtractFunctionNameFromCall(node *sitter.Node, content []byte) string {
	if node == nil {
		return ""
	}

	// Try to extract function name from different possible fields
	possibleFields := []string{"function", "call", "method", "name", "function_name"}
	for _, field := range possibleFields {
		funcNode := node.ChildByFieldName(field)
		if funcNode != nil {
			return SafeContent(funcNode, content)
		}
	}

	// If no function node is found, attempt to find an identifier or constant among the children
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		if child.Type() == "identifier" || child.Type() == "constant" || child.Type() == "variable_name" || child.Type() == "name" {
			return SafeContent(child, content)
		}
	}

	return ""
}

// -----------------------------------------------------------------------------
// IsFunctionCall - Checks if the node represents a function call involving the given variable.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to check.
//   - content ([]byte): The content of the source code.
//   - variable (string): The variable to check for.
//
// Returns:
//   - (bool, string): True if the node represents a function call involving the variable, and the variable name if assigned.
//
// -----------------------------------------------------------------------------
func IsFunctionCall(node *sitter.Node, content []byte, variable string) (bool, string) {
	if node == nil {
		return false, ""
	}

	// Check if the node is a function call in any of the supported languages
	switch node.Type() {
	case "call_expression", // Go, C, C++, JavaScript, Rust
		"invocation_expression",    // C#
		"method_invocation",        // Java
		"function_call_expression", // PHP
		"call",                     // Python, Ruby
		"macro_invocation":         // Rust (for macros like println!)
		// Proceed to extract arguments
	default:
		return false, ""
	}

	// Extract arguments from the function call
	args := node.ChildByFieldName("arguments")
	if args == nil && node.Type() == "call" {
		// For Ruby, arguments might be direct children under "argument_list"
		args = node.ChildByFieldName("argument_list")
	}
	if args != nil {
		for i := 0; i < int(args.NamedChildCount()); i++ {
			arg := args.NamedChild(i)
			if SafeContent(arg, content) == variable {
				// Check for assignment to a new variable on the left
				parent := node.Parent()
				if parent != nil {
					variableName := getAssignedVariableName(parent, content)
					if variableName != "" {
						return true, variableName
					}
				}
				return true, ""
			}
		}
	}

	return false, ""
}

// -----------------------------------------------------------------------------
// getAssignedVariableName - Checks if the parent node represents an assignment to a variable and returns the variable name.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - parent (*sitter.Node): The parent node to check.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The name of the assigned variable if found, otherwise an empty string.
//
// -----------------------------------------------------------------------------
func getAssignedVariableName(parent *sitter.Node, content []byte) string {
	if parent == nil {
		return ""
	}

	var leftSide *sitter.Node

	switch parent.Type() {
	case "short_var_declaration", "assignment_statement", "assignment_expression", "assignment":
		// Common assignment node types
		leftSide = parent.ChildByFieldName("left")
	case "local_variable_declaration", "local_declaration_statement", "lexical_declaration":
		// Declarations with assignments in Java, C#, JavaScript
		for i := 0; i < int(parent.NamedChildCount()); i++ {
			child := parent.NamedChild(i)
			if child.Type() == "variable_declarator" {
				leftSide = child.ChildByFieldName("name")
				break
			}
		}
	case "declaration":
		// For C and C++ declarations
		declarator := parent.ChildByFieldName("declarator")
		if declarator != nil {
			leftSide = findIdentifierInDeclarator(declarator)
		}
	case "let_declaration":
		// Rust let declaration
		leftSide = parent.ChildByFieldName("pattern")
	case "variable_declarator", "init_declarator":
		// Variable declarators in various languages
		leftSide = parent.ChildByFieldName("name")
		if leftSide == nil {
			// For C++, the name might be nested in the declarator
			declarator := parent.ChildByFieldName("declarator")
			if declarator != nil {
				leftSide = findIdentifierInDeclarator(declarator)
			}
		}
	}

	if leftSide != nil {
		return SafeContent(leftSide, content)
	}

	return ""
}

// -----------------------------------------------------------------------------
// FindFunctionCallSites - Searches all lines where a specific function is called.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - functionName (string): The name of the function to search for.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - ([]models.FunctionCallSite): A slice of function call sites.
//
// -----------------------------------------------------------------------------
func FindFunctionCallSites(root *sitter.Node, functionName string, content []byte) []models.FunctionCallSite {
	var callSites []models.FunctionCallSite

	var exploreNode func(node *sitter.Node)
	exploreNode = func(node *sitter.Node) {
		if node == nil {
			return
		}

		// Check if the node represents a function call in any of the specified languages
		switch node.Type() {
		case "call_expression", "invocation_expression", "method_invocation", "function_call_expression", "call", "macro_invocation":
			// Extract the function being called
			funcNode := getFunctionNode(node)
			if funcNode != nil {
				funcName := extractFunctionName(funcNode, content)
				if funcName == functionName {
					// Add the call site to the list
					callSites = append(callSites, models.FunctionCallSite{
						Line:     node.StartPoint().Row + 1,
						CallNode: node,
					})
				}
			}
		}

		// Recursively explore all child nodes
		for i := 0; i < int(node.ChildCount()); i++ {
			exploreNode(node.Child(i))
		}
	}

	// Start exploring from the root
	exploreNode(root)

	return callSites
}

// -----------------------------------------------------------------------------
// GetArgumentVariable - Finds the argument variable corresponding to a parameter variable in a function call.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - callNode (*sitter.Node): The function call node.
//   - parameterVariable (string): The parameter variable to search for.
//   - functionNode (*sitter.Node): The function node containing the parameter.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The argument variable corresponding to the parameter variable.
//
// -----------------------------------------------------------------------------
func GetArgumentVariable(callNode *sitter.Node, parameterVariable string, functionNode *sitter.Node, content []byte) string {
	// Get the list of parameter names from the function's parameter list
	parameters := getParametersNode(functionNode)
	logger.PrintDebug("Parameters: %v", parameters)
	if parameters != nil {
		var parameterNames []string
		for i := 0; i < int(parameters.NamedChildCount()); i++ {
			param := parameters.NamedChild(i)
			paramNames := extractParameterNames(param, content)
			parameterNames = append(parameterNames, paramNames...)

		}

		logger.PrintDebug("Parameter names: %v", parameterNames)
		// Find the index of the parameter variable
		var paramIndex = -1
		for idx, paramName := range parameterNames {
			if paramName == parameterVariable {
				paramIndex = idx
				break
			}
		}
		if paramIndex != -1 {
			// Get the argument at the same index in the call node
			arguments := getArgumentsNode(callNode)
			if arguments != nil && paramIndex < int(arguments.NamedChildCount()) {
				arg := arguments.NamedChild(paramIndex)
				argName := extractArgumentName(arg, content)
				if argName != "" {
					return argName
				} else {
					// If argName is empty, try to extract identifiers from expressions
					identifiers := extractIdentifiers(arg, content)
					if len(identifiers) > 0 {
						return identifiers[0] // Return the first identifier found
					}
				}
			}
		}
	}
	return ""
}

// -----------------------------------------------------------------------------
// extractArgumentName - Extracts the variable name from an argument node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - argNode (*sitter.Node): The argument node.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The name of the argument variable.
//
// -----------------------------------------------------------------------------
func extractArgumentName(argNode *sitter.Node, content []byte) string {
	if argNode == nil {
		return ""
	}
	switch argNode.Type() {
	case "identifier", "variable_name", "name":
		return SafeContent(argNode, content)
	default:
		// Attempt to find an identifier within the argument node
		identifiers := extractIdentifiers(argNode, content)
		if len(identifiers) > 0 {
			return identifiers[0] // Return the first identifier found
		}
	}
	return ""
}

// -----------------------------------------------------------------------------
// getFunctionNode - Returns the node representing the function being called.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to search for the function.
//
// Returns:
//   - (*sitter.Node): The function node if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func getFunctionNode(node *sitter.Node) *sitter.Node {
	// Depending on the node type, the function node may be in different child fields
	switch node.Type() {
	case "call_expression", "invocation_expression", "method_invocation", "function_call_expression", "macro_invocation":
		// For these node types, the function is in the "function" or "function_name" field
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			funcNode = node.ChildByFieldName("function_name")
		}
		if funcNode == nil {
			// For some languages, the function might be in a different field
			funcNode = node.ChildByFieldName("name")
		}
		return funcNode
	case "call":
		// For Python and Ruby, the function being called may be in the "function" field or "method" field
		funcNode := node.ChildByFieldName("function")
		if funcNode == nil {
			funcNode = node.ChildByFieldName("method")
		}
		return funcNode
	default:
		return nil
	}
}

// -----------------------------------------------------------------------------
// extractFunctionName - Extracts the function name from the function node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - funcNode (*sitter.Node): The function node.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The name of the function.
//
// -----------------------------------------------------------------------------
func extractFunctionName(funcNode *sitter.Node, content []byte) string {
	if funcNode == nil {
		return ""
	}

	// In some languages, the function node may be an identifier or a more complex expression
	switch funcNode.Type() {
	case "identifier", "variable_name", "name", "constant":
		return SafeContent(funcNode, content)
	case "member_expression", "field_access", "scoped_identifier", "member_access_expression":
		// The function is a method of an object (e.g., object.method)
		propertyNode := funcNode.ChildByFieldName("property")
		if propertyNode == nil {
			// For some languages, the property might be under "field" or "name"
			propertyNode = funcNode.ChildByFieldName("field")
		}
		if propertyNode != nil {
			return SafeContent(propertyNode, content)
		}
		// If we can't find the property, attempt to extract from the function node
		return SafeContent(funcNode, content)
	case "call":
		// In nested calls, recursively extract the function name
		return extractFunctionName(funcNode.ChildByFieldName("function"), content)
	default:
		// Attempt to find an identifier among the children
		for i := 0; i < int(funcNode.NamedChildCount()); i++ {
			child := funcNode.NamedChild(i)
			if child.Type() == "identifier" || child.Type() == "variable_name" || child.Type() == "name" || child.Type() == "constant" {
				return SafeContent(child, content)
			}
		}
	}

	return ""
}

// Functions to get parameter name

// -----------------------------------------------------------------------------
// GetParameterName - Finds the name of the parameter corresponding to a variable in a function call.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - functionNode (*sitter.Node): The function node containing the parameter.
//   - content ([]byte): The content of the source code.
//   - originalVariable (string): The original variable name.
//   - callSite (*sitter.Node): The function call node.
//
// Returns:
//   - (string): The name of the parameter corresponding to the variable.
//
// -----------------------------------------------------------------------------
func GetParameterName(functionNode *sitter.Node, content []byte, originalVariable string, callSite *sitter.Node) string {
	// Find the position of the variable in the function call arguments
	arguments := getArgumentsNode(callSite)
	if arguments != nil {
		for i := 0; i < int(arguments.NamedChildCount()); i++ {
			arg := arguments.NamedChild(i)
			argName := SafeContent(arg, content)
			if argName == originalVariable {
				// Now, i is the index of the parameter in the call
				parameters := getParametersNode(functionNode)
				if parameters != nil && i < int(parameters.NamedChildCount()) {
					param := parameters.NamedChild(i)
					paramName := extractParameterName(param, content)
					if paramName != "" {
						return paramName
					}
				}
			}
		}
	}
	return originalVariable // Return the original variable if no parameter name is found
}

// -----------------------------------------------------------------------------
// getParametersNode - Returns the node containing the parameters of a function definition.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - functionNode (*sitter.Node): The function node.
//
// Returns:
//   - (*sitter.Node): The parameters node if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func getParametersNode(functionNode *sitter.Node) *sitter.Node {
	switch functionNode.Type() {
	case "function_declaration", "method_declaration", "function_definition", "function_item", "method":
		// Common field names for parameters
		parameters := functionNode.ChildByFieldName("parameters")
		if parameters != nil {
			return parameters
		}
		// For C and C++, parameters may be under nested 'declarator' fields
		if functionNode.Type() == "function_definition" {
			declarator := functionNode.ChildByFieldName("declarator")
			parameters := findParametersInDeclarator(declarator)
			if parameters != nil {
				return parameters
			}
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// getArgumentsNode - Returns the node containing the arguments of a function call.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - callSite (*sitter.Node): The function call node.
//
// Returns:
//   - (*sitter.Node): The arguments node if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func getArgumentsNode(callSite *sitter.Node) *sitter.Node {
	switch callSite.Type() {
	case "call_expression", "method_invocation", "invocation_expression", "function_call_expression", "call", "macro_invocation":
		// Common field names for arguments
		arguments := callSite.ChildByFieldName("arguments")
		if arguments != nil {
			return arguments
		}
		// Some languages may use "argument_list"
		arguments = callSite.ChildByFieldName("argument_list")
		if arguments != nil {
			return arguments
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// findParametersInDeclarator - Recursively searches for 'parameters' in declarator nodes.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The declarator node to search.
//
// Returns:
//   - (*sitter.Node): The parameters node if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func findParametersInDeclarator(node *sitter.Node) *sitter.Node {
	if node == nil {
		return nil
	}
	parameters := node.ChildByFieldName("parameters")
	if parameters != nil {
		return parameters
	}
	// Recurse into the 'declarator' child
	declarator := node.ChildByFieldName("declarator")
	if declarator != nil {
		return findParametersInDeclarator(declarator)
	}
	return nil
}

// -----------------------------------------------------------------------------
// extractParameterName - Extracts the name of a parameter node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - param (*sitter.Node): The parameter node.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (string): The name of the parameter.
//
// -----------------------------------------------------------------------------
func extractParameterName(param *sitter.Node, content []byte) string {
	if param == nil {
		return ""
	}
	switch param.Type() {
	case "parameter_declaration", "parameter", "simple_parameter", "formal_parameter":
		// Try to get the name field
		paramNameNode := param.ChildByFieldName("name")
		if paramNameNode != nil {
			return SafeContent(paramNameNode, content)
		}
		// For C and C++, parameter name may be under 'declarator'
		declaratorNode := param.ChildByFieldName("declarator")
		if declaratorNode != nil {
			identifierNode := findIdentifierInDeclarator(declaratorNode)
			if identifierNode != nil {
				paramName := SafeContent(identifierNode, content)
				logger.PrintDebug("Extracted parameter name: %s", paramName)
				return paramName
			}
		}
		// For languages like Rust, the parameter pattern contains the name
		patternNode := param.ChildByFieldName("pattern")
		if patternNode != nil {
			return SafeContent(patternNode, content)
		}
	case "identifier":
		// For languages where parameters are identifiers directly
		return SafeContent(param, content)
	default:
		// Try to find an identifier among the children
		for i := 0; i < int(param.NamedChildCount()); i++ {
			child := param.NamedChild(i)
			if child.Type() == "identifier" || child.Type() == "variable_name" || child.Type() == "name" {
				return SafeContent(child, content)
			}
		}
	}
	return ""
}

// -----------------------------------------------------------------------------
// extractParameterNames - Extracts the names of parameters from a parameter_declaration node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - param (*sitter.Node): The parameter node.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - ([]string): A slice of parameter names.
//
// -----------------------------------------------------------------------------
func extractParameterNames(param *sitter.Node, content []byte) []string {
	var names []string
	seen := make(map[string]bool) // Map pour éviter les doublons

	if param == nil {
		return names
	}

	switch param.Type() {
	case "parameter_declaration", "parameter", "simple_parameter", "formal_parameter":
		for i := 0; i < int(param.NamedChildCount()); i++ {
			child := param.NamedChild(i)
			if child.Type() == "identifier" {
				paramName := SafeContent(child, content)
				if !seen[paramName] { // Vérifier si le nom est déjà ajouté
					names = append(names, paramName)
					seen[paramName] = true
				}
			}
		}

		nameNode := param.ChildByFieldName("name")
		if nameNode != nil {
			paramName := SafeContent(nameNode, content)
			if !seen[paramName] {
				names = append(names, paramName)
				seen[paramName] = true
			}
		}

		declaratorNode := param.ChildByFieldName("declarator")
		if declaratorNode != nil {
			identifierNode := findIdentifierInDeclarator(declaratorNode)
			if identifierNode != nil {
				paramName := SafeContent(identifierNode, content)
				if !seen[paramName] {
					names = append(names, paramName)
					seen[paramName] = true
				}
			}
		}

		patternNode := param.ChildByFieldName("pattern")
		if patternNode != nil {
			paramName := SafeContent(patternNode, content)
			if !seen[paramName] {
				names = append(names, paramName)
				seen[paramName] = true
			}
		}
	case "identifier":
		paramName := SafeContent(param, content)
		if !seen[paramName] {
			names = append(names, paramName)
			seen[paramName] = true
		}
	default:
		for i := 0; i < int(param.NamedChildCount()); i++ {
			child := param.NamedChild(i)
			if child.Type() == "identifier" || child.Type() == "variable_name" || child.Type() == "name" {
				paramName := SafeContent(child, content)
				if !seen[paramName] {
					names = append(names, paramName)
					seen[paramName] = true
				}
			}
		}
	}

	return names
}

/**** Variable Functions ****/

// -----------------------------------------------------------------------------
// IsAssignment - Checks if a variable is used in an assignment or declaration across multiple languages.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to check.
//   - content ([]byte): The content of the source code.
//   - variable (string): The variable to check for.
//
// Returns:
//   - (bool, string): True if the variable is found in the assignment, and the value assigned to it.
//
// -----------------------------------------------------------------------------
func IsAssignment(node *sitter.Node, content []byte, variable string) (bool, string) {
	// Handle expression_statement nodes that may contain assignments
	if node.Type() == "expression_statement" && node.NamedChildCount() > 0 {
		return IsAssignment(node.NamedChild(0), content, variable)
	}

	// Node types representing assignments and declarations in various languages
	assignmentNodeTypes := map[string]bool{
		// Assignments
		"short_var_declaration":           true, // Go
		"assignment_statement":            true, // Go
		"assignment_expression":           true, // C, C++, Java, JavaScript, C#
		"assignment":                      true, // Python, Ruby
		"reference_assignment_expression": true, // PHP
		// Declarations with initialization
		"var_declaration":             true, // Go
		"const_declaration":           true, // Go
		"let_declaration":             true, // Rust
		"static_item":                 true, // Rust
		"declaration":                 true, // C, C++
		"local_declaration_statement": true, // C#
		"local_variable_declaration":  true, // Java
		"lexical_declaration":         true, // JavaScript
		"variable_declaration":        true, // JavaScript
	}

	// Check if the node is an assignment or declaration
	if !assignmentNodeTypes[node.Type()] {
		return false, ""
	}

	// Variables to store left-hand side and right-hand side identifiers
	var lhsVars []string
	var rhsVars []string

	// Extract variables based on node type
	switch node.Type() {
	case "short_var_declaration", "assignment_statement", "assignment_expression", "assignment", "reference_assignment_expression":
		// Assignments with 'left' and 'right' fields
		leftSide := node.ChildByFieldName("left")
		rightSide := node.ChildByFieldName("right")

		// Extract variables from left side
		if leftSide != nil {
			lhsVars = extractIdentifiers(leftSide, content)
		}

		// Extract variables from right side
		if rightSide != nil {
			rhsVars = extractIdentifiers(rightSide, content)
		}
	case "var_declaration", "const_declaration":
		// Go var/const declarations
		for i := 0; i < int(node.NamedChildCount()); i++ {
			varSpec := node.NamedChild(i)
			if varSpec.Type() == "var_spec" || varSpec.Type() == "const_spec" {
				nameNode := varSpec.ChildByFieldName("name")
				valueNode := varSpec.ChildByFieldName("value")
				if nameNode != nil {
					lhsVars = append(lhsVars, extractIdentifiers(nameNode, content)...)
				}
				if valueNode != nil {
					rhsVars = append(rhsVars, extractIdentifiers(valueNode, content)...)
				}
			}
		}
	case "let_declaration":
		// Rust let declarations
		patternNode := node.ChildByFieldName("pattern")
		valueNode := node.ChildByFieldName("value")
		if patternNode != nil {
			lhsVars = append(lhsVars, extractIdentifiers(patternNode, content)...)
		}
		if valueNode != nil {
			rhsVars = append(rhsVars, extractIdentifiers(valueNode, content)...)
		}
	case "static_item":
		// Rust static declarations
		nameNode := node.ChildByFieldName("name")
		valueNode := node.ChildByFieldName("value")
		logger.PrintDebug("Static item name: %v", nameNode)
		logger.PrintDebug("Static item value: %v", valueNode)

		// Add the name to lhsVars (left-hand side)
		if nameNode != nil {
			lhsVars = append(lhsVars, extractIdentifiers(nameNode, content)...)
		}

		// Add identifiers from the value to rhsVars (right-hand side)
		if valueNode != nil {
			rhsVars = append(rhsVars, extractIdentifiers(valueNode, content)...)
		}
		logger.PrintDebug("Static item lhsVars: %v", lhsVars)
		logger.PrintDebug("Static item rhsVars: %v", rhsVars)

	case "declaration":
		// C/C++ declarations
		declarator := node.ChildByFieldName("declarator")
		if declarator != nil {
			var nameNode *sitter.Node
			if declarator.Type() == "init_declarator" {
				innerDeclarator := declarator.ChildByFieldName("declarator")
				nameNode = findIdentifierInDeclarator(innerDeclarator)
				valueNode := declarator.ChildByFieldName("value")
				if valueNode != nil {
					rhsVars = append(rhsVars, extractIdentifiers(valueNode, content)...)
				}
			} else {
				// Handle cases where declarator is directly an identifier
				nameNode = findIdentifierInDeclarator(declarator)
			}
			if nameNode != nil {
				lhsVars = append(lhsVars, SafeContent(nameNode, content))
			}
		}

	case "local_declaration_statement":
		// C# local declarations
		varDecl := node.NamedChild(0)
		logger.PrintDebug("Local declaration statement: %v", varDecl)
		if varDecl != nil && varDecl.Type() == "variable_declaration" {
			for i := 0; i < int(varDecl.NamedChildCount()); i++ {
				varDeclarator := varDecl.NamedChild(i)
				if varDeclarator.Type() == "variable_declarator" {
					nameNode := varDeclarator.ChildByFieldName("name")
					valueNode := varDeclarator.ChildByFieldName("value")
					// If valueNode is nil, check if there's an additional child
					if valueNode == nil && varDeclarator.NamedChildCount() > 1 {
						// Assume the second named child is the value node
						valueNode = varDeclarator.NamedChild(1)
					}
					logger.PrintDebug("Name node: %v", nameNode)
					logger.PrintDebug("Value node: %v", valueNode)
					if nameNode != nil {
						lhsVars = append(lhsVars, SafeContent(nameNode, content))
						logger.PrintDebug("Added variable: %s", SafeContent(nameNode, content))
					}
					if valueNode != nil {
						rhsVars = append(rhsVars, extractIdentifiers(valueNode, content)...)
						logger.PrintDebug("Added value: %s", SafeContent(valueNode, content))
					}
				}
			}
		}

	case "local_variable_declaration":
		// Java local variable declarations
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if child.Type() == "variable_declarator" {
				nameNode := child.ChildByFieldName("name")
				valueNode := child.ChildByFieldName("value")
				if nameNode != nil {
					lhsVars = append(lhsVars, SafeContent(nameNode, content))
				}
				if valueNode != nil {
					rhsVars = append(rhsVars, extractIdentifiers(valueNode, content)...)
				}
			}
		}
	case "lexical_declaration", "variable_declaration":
		// JavaScript variable declarations
		for i := 0; i < int(node.NamedChildCount()); i++ {
			varDeclarator := node.NamedChild(i)
			if varDeclarator.Type() == "variable_declarator" {
				nameNode := varDeclarator.ChildByFieldName("name")
				valueNode := varDeclarator.ChildByFieldName("value")
				if nameNode != nil {
					lhsVars = append(lhsVars, SafeContent(nameNode, content))
				}
				if valueNode != nil {
					rhsVars = append(rhsVars, extractIdentifiers(valueNode, content)...)
				}
			}
		}
	}

	// Check if the variable is on the left-hand side (assigned)
	for _, lhsVar := range lhsVars {
		if lhsVar == variable || (node.Type() == "static_item" && strings.EqualFold(lhsVar, variable)) {
			assignedValue := ""
			if len(rhsVars) > 0 {
				assignedValue = rhsVars[0]
			}
			return true, assignedValue
		}
	}

	// Check if the variable is on the right-hand side (used)
	for _, rhsVar := range rhsVars {
		if rhsVar == variable {
			assignedVariable := ""
			if len(lhsVars) > 0 {
				assignedVariable = lhsVars[0]
			}
			return true, assignedVariable
		}
	}

	return false, ""
}

// -----------------------------------------------------------------------------
// extractIdentifiers - Extracts identifier names from a node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to extract identifiers from.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - ([]string): A slice of identifier names.
//
// -----------------------------------------------------------------------------
func extractIdentifiers(node *sitter.Node, content []byte) []string {
	var identifiers []string

	if node == nil {
		return identifiers
	}

	nodeType := node.Type()

	if nodeType == "identifier" || nodeType == "variable_name" || nodeType == "name" || nodeType == "constant" {
		identifiers = append(identifiers, SafeContent(node, content))
		return identifiers
	}

	// Handle pointers, references, and declarators in C/C++
	if nodeType == "pointer_declarator" || nodeType == "reference_declarator" || nodeType == "array_declarator" {
		if node.ChildCount() > 0 {
			child := node.Child(0)
			identifiers = append(identifiers, extractIdentifiers(child, content)...)
		}
		return identifiers
	}

	// Handle Rust static items
	if nodeType == "static_item" {
		// Extract the name of the static variable
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			identifiers = append(identifiers, SafeContent(nameNode, content))
		}
		// Extract identifiers from the value if necessary (e.g., complex expressions)
		valueNode := node.ChildByFieldName("value")
		if valueNode != nil {
			identifiers = append(identifiers, extractIdentifiers(valueNode, content)...)
		}
		return identifiers
	}

	// Handle Rust patterns
	if nodeType == "pattern" || nodeType == "mutable_specifier" {
		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			identifiers = append(identifiers, extractIdentifiers(child, content)...)
		}
		return identifiers
	}

	// Recursively extract identifiers from children
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		identifiers = append(identifiers, extractIdentifiers(child, content)...)
	}

	return identifiers
}

// -----------------------------------------------------------------------------
// IsVariableUsedInExpression - Checks if a variable is used in an expression within a node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to check.
//   - variable (string): The variable to check for.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (bool): True if the variable is used in the expression, false otherwise.
//
// -----------------------------------------------------------------------------
func IsVariableUsedInExpression(node *sitter.Node, variable string, content []byte) bool {
	if node == nil {
		return false
	}

	switch node.Type() {
	case "identifier", "variable_name", "name":
		// Extraire le texte du nœud et comparer avec la variable
		identifier := string(content[node.StartByte():node.EndByte()])
		return identifier == variable
	default:
		// Parcourir récursivement tous les enfants du nœud
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if IsVariableUsedInExpression(child, variable, content) {
				return true
			}
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// IsVariableInScope - Checks if the given variable is in scope within the provided function node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - varName (string): The name of the variable to check.
//   - functionNode (*sitter.Node): The function node to check within.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (bool): True if the variable is in scope, false otherwise.
//
// -----------------------------------------------------------------------------
func IsVariableInScope(varName string, functionNode *sitter.Node, content []byte) bool {
	if functionNode == nil {
		return false
	}

	// Vérifier si la variable est définie dans les paramètres de la fonction
	parameters := functionNode.ChildByFieldName("parameters")
	if parameters != nil {
		for i := 0; i < int(parameters.NamedChildCount()); i++ {
			param := parameters.NamedChild(i)
			paramName := extractParameterName(param, content)
			if paramName == varName {
				return true // La variable est un paramètre de la fonction
			}
		}
	}

	// Vérifier si la variable est définie dans les déclarations locales
	var traverse func(*sitter.Node) bool
	traverse = func(n *sitter.Node) bool {
		if n == nil {
			return false
		}

		// Vérifier les déclarations locales (e.g. short_var_declaration, local_variable_declaration, etc.)
		switch n.Type() {
		case "short_var_declaration", "local_variable_declaration", "assignment_statement":
			leftSide := n.ChildByFieldName("left")
			if leftSide != nil {
				for i := 0; i < int(leftSide.NamedChildCount()); i++ {
					child := leftSide.NamedChild(i)
					if child.Type() == "identifier" || child.Type() == "variable_name" {
						varNameInNode := SafeContent(child, content)
						if varNameInNode == varName {
							return true
						}
					}
				}
			}
		}

		// Parcourir récursivement les enfants
		for i := 0; i < int(n.NamedChildCount()); i++ {
			if traverse(n.NamedChild(i)) {
				return true
			}
		}
		return false
	}

	// Vérifier dans le corps de la fonction
	body := functionNode.ChildByFieldName("body")
	if body != nil {
		return traverse(body)
	}

	return false
}

/**** Variable Global Functions ****/

// -----------------------------------------------------------------------------
// AddGlobalVariableSteps - Adds global variable declaration steps for all unique variables in the data flow steps.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - dataFlow ([]models.DataFlowStep): The data flow steps to process.
//   - root (*sitter.Node): The root node of the syntax tree.
//   - content ([]byte): The content of the source code.
//   - lineNumber (uint32): The line number to use for generic usage steps.
//
// Returns:
//   - ([]models.DataFlowStep): The updated data flow steps with global variable declaration steps.
//
// -----------------------------------------------------------------------------
func AddGlobalVariableSteps(dataFlow []models.DataFlowStep, root *sitter.Node, content []byte, lineNumber uint32) []models.DataFlowStep {
	logger.PrintDebug("Extracting unique variables from data flow steps.")
	processedVariables := make(map[string]string) // Tracks processed variables (lowercase -> original case)

	// Collect all unique variables from the data flow steps
	for _, step := range dataFlow {
		lowercaseVariable := strings.ToLower(step.Variable)
		if original, exists := processedVariables[lowercaseVariable]; !exists {
			processedVariables[lowercaseVariable] = step.Variable
			logger.PrintDebug("Added unique variable '%s' (original case: '%s') for processing.", lowercaseVariable, step.Variable)
		} else if original != step.Variable {
			logger.PrintWarning("Detected case-insensitive duplicate variables: '%s' and '%s'.", original, step.Variable)
		}
	}

	// Variables to add to the final dataFlow
	var finalSteps []models.DataFlowStep

	// Process each unique lowercase variable
	for lowercaseVariable, originalCaseVariable := range processedVariables {
		logger.PrintDebug("Checking if global variable step is needed for '%s' (original case: '%s').", lowercaseVariable, originalCaseVariable)
		isOriginalGlobal := IsVariableGlobal(root, originalCaseVariable, content)
		isLowercaseGlobal := IsVariableGlobal(root, lowercaseVariable, content)

		// Determine which variable to add based on global detection
		if isOriginalGlobal && isLowercaseGlobal {
			logger.PrintDebug("Both '%s' and '%s' are global. Keeping lowercase version: '%s'.", originalCaseVariable, lowercaseVariable, lowercaseVariable)
			originalCaseVariable = lowercaseVariable // Prefer lowercase if both are global
		} else if isLowercaseGlobal {
			logger.PrintDebug("Only lowercase '%s' is global. Adding this version.", lowercaseVariable)
			originalCaseVariable = lowercaseVariable
		} else if isOriginalGlobal {
			logger.PrintDebug("Only original case '%s' is global. Adding this version.", originalCaseVariable)
		} else {
			logger.PrintDebug("Neither '%s' nor '%s' is global. Skipping.", lowercaseVariable, originalCaseVariable)
			continue // Skip if neither is global
		}

		// Add the detected global variable to the data flow
		globalVariableNode := FindGlobalVariableDeclaration(root, originalCaseVariable, content)
		if globalVariableNode != nil {
			line := globalVariableNode.StartPoint().Row + 1
			logger.PrintInfo("Adding global variable declaration step for '%s' at line %d in node type '%s'.", originalCaseVariable, line, globalVariableNode.Type())
			finalSteps = append(finalSteps, models.DataFlowStep{
				Line:     line,
				Type:     "Global Variable Declaration",
				Function: "Global Scope",
				Value:    originalCaseVariable,
				Variable: originalCaseVariable,
			})
		} else {
			logger.PrintWarning("Global variable node not found for '%s'. Adding generic usage step at line %d.", originalCaseVariable, lineNumber)
			finalSteps = append(finalSteps, models.DataFlowStep{
				Line:     lineNumber,
				Type:     "Global Variable Usage",
				Function: "Global Scope",
				Value:    originalCaseVariable,
				Variable: originalCaseVariable,
			})
		}
	}

	// Return the original data flow combined with the new steps
	return append(dataFlow, finalSteps...)
}

// -----------------------------------------------------------------------------
// FindGlobalVariableDeclaration - Finds the declaration of a global variable in the syntax tree.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - variable (string): The name of the variable to search for.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (*sitter.Node): The node representing the global variable declaration if found, otherwise nil.
//
// -----------------------------------------------------------------------------
func FindGlobalVariableDeclaration(root *sitter.Node, variable string, content []byte) *sitter.Node {
	if root == nil {
		logger.PrintDebug("Root node is nil, skipping global variable search for '%s'.", variable)
		return nil
	}

	var result *sitter.Node
	var traverse func(node *sitter.Node, inFunctionScope bool, inClassScope bool)
	traverse = func(node *sitter.Node, inFunctionScope bool, inClassScope bool) {
		if node == nil || result != nil {
			return
		}

		nodeType := node.Type()

		// Determine if we're entering a new scope
		switch nodeType {
		case "function_definition", "method_declaration", "method_declaration_header",
			"function_declaration", "anonymous_function_creation_expression", "method", "function_item":
			inFunctionScope = true
			return
		case "class_declaration", "namespace_name", "interface_declaration", "trait_declaration":
			logger.PrintDebug("Entering class or namespace scope: '%s'.", nodeType)
			inClassScope = true
		}

		if !inFunctionScope && !inClassScope {
			// Handle PHP assignments
			if models.GlobalLanguage == "php" && nodeType == "assignment_expression" {
				leftNode := node.ChildByFieldName("left")
				if leftNode != nil && leftNode.Type() == "variable_name" {
					// Instead of using ChildByFieldName("name"), access the first named child
					varNameNode := leftNode.NamedChild(0)
					if varNameNode != nil && varNameNode.Type() == "name" {
						varName := SafeContent(varNameNode, content)
						logger.PrintDebug("Checking PHP assignment for global variable '%s'.", varName)
						if strings.EqualFold(varName, variable) {
							logger.PrintDebug("Global variable '%s' assigned at line %d in node type '%s'.",
								variable, node.StartPoint().Row+1, nodeType)
							result = node
							return
						}
					}
				}

			} else {
				// Use IsAssignment for other languages
				isAssign, _ := IsAssignment(node, content, variable)
				if isAssign {
					logger.PrintDebug("Global variable '%s' declared at line %d in node type '%s'.",
						variable, node.StartPoint().Row+1, nodeType)
					result = node
					return
				}
			}
		} else if inClassScope && !inFunctionScope {
			// For class field declarations
			if isClassStaticVariableDeclarationNode(node) {
				isAssign := IsClassFieldAssignment(node, content, variable)
				if isAssign {
					logger.PrintDebug("Global class variable '%s' declared at line %d in node type '%s'.",
						variable, node.StartPoint().Row+1, nodeType)
					result = node
					return
				}
			}
		}

		// Traverse children
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			traverse(child, inFunctionScope, inClassScope)
			if result != nil {
				return
			}
		}
	}

	logger.PrintDebug("Starting traversal to find global variable '%s'.", variable)
	traverse(root, false, false)
	if result == nil {
		logger.PrintDebug("Global variable '%s' not found at the global scope.", variable)
	}
	return result
}

// -----------------------------------------------------------------------------
// isClassStaticVariableDeclarationNode - Checks if a node represents a class field declaration.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to check.
//
// Returns:
//   - (bool): True if the node represents a class field declaration, false otherwise.
//
// -----------------------------------------------------------------------------
func isClassStaticVariableDeclarationNode(node *sitter.Node) bool {
	nodeType := node.Type()

	switch nodeType {
	case
		"field_declaration",    // Java, C#
		"field_declarator",     // Java, C#
		"constant_declaration", // C#
		"property_declaration": // C#
		return true
	default:
		return false
	}
}

// -----------------------------------------------------------------------------
// IsClassFieldAssignment - Checks if a node represents an assignment to a class field.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - node (*sitter.Node): The node to check.
//   - content ([]byte): The content of the source code.
//   - variable (string): The variable to check for.
//
// Returns:
//   - (bool): True if the node represents an assignment to the class field, false otherwise.
//
// -----------------------------------------------------------------------------
func IsClassFieldAssignment(node *sitter.Node, content []byte, variable string) bool {
	nodeType := node.Type()
	logger.PrintDebug("IsClassFieldAssignment: nodeType='%s'", nodeType)
	if nodeType == "field_declaration" || nodeType == "property_declaration" || nodeType == "constant_declaration" {
		var nameNode *sitter.Node

		// Handle Java syntax
		if models.GlobalLanguage == "java" {
			declaratorNode := node.ChildByFieldName("declarator")
			if declaratorNode != nil && declaratorNode.Type() == "variable_declarator" {
				nameNode = declaratorNode.ChildByFieldName("name")
			}
		}

		// Handle C# syntax
		if models.GlobalLanguage == "csharp" {
			// Navigate through the children to locate `variable_declaration`
			for i := 0; i < int(node.NamedChildCount()); i++ {
				child := node.NamedChild(i)
				if child.Type() == "variable_declaration" {
					for j := 0; j < int(child.NamedChildCount()); j++ {
						variableDeclarator := child.NamedChild(j)
						if variableDeclarator.Type() == "variable_declarator" {
							nameNode = variableDeclarator.ChildByFieldName("name")
							if nameNode != nil {
								varName := SafeContent(nameNode, content)
								logger.PrintDebug("Found class field variable '%s'.", varName)
								if strings.EqualFold(varName, variable) {
									return true
								}
							}
						}
					}
				}
			}
		}

		if nameNode != nil {
			varName := SafeContent(nameNode, content)
			logger.PrintDebug("Found class field variable '%s'.", varName)
			if strings.EqualFold(varName, variable) {
				return true
			}
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// IsVariableGlobal - Checks if a variable is global within the provided root node.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - variable (string): The name of the variable to check.
//   - content ([]byte): The content of the source code.
//
// Returns:
//   - (bool): True if the variable is global, false otherwise.
//
// -----------------------------------------------------------------------------
func IsVariableGlobal(root *sitter.Node, variable string, content []byte) bool {
	logger.PrintDebug("Checking if variable '%s' is global.", variable)
	globalDecl := FindGlobalVariableDeclaration(root, variable, content)
	if globalDecl != nil {
		logger.PrintInfo("Variable '%s' is global. Found at line %d in node type '%s'.", variable, globalDecl.StartPoint().Row+1, globalDecl.Type())
		return true
	}
	logger.PrintInfo("Variable '%s' is not global.", variable)
	return false
}
