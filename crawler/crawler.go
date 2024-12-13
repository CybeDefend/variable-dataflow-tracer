package crawler

import (
	logger "variable-dataflow-tracer/logger"
	models "variable-dataflow-tracer/models"
	"variable-dataflow-tracer/services/nodeService"
	"variable-dataflow-tracer/services/utilityService"

	sitter "github.com/smacker/go-tree-sitter"
)

var visitedFunctionStack []string

// -----------------------------------------------------------------------------
// CrawlFromLine - Performs data flow analysis starting from a specific line.
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - node (*sitter.Node): The current node in the syntax tree.
//   - content ([]byte): The content of the file being analyzed.
//   - variablesToTrack (map[string]bool): A map of variables to track during the analysis.
//   - startLine (uint32): The line number to start the analysis from.
//   - startFromEnd (bool): Flag indicating whether to start the analysis from the end of the function.
//   - visitedLines (map[uint32]bool): A map of lines that have already been visited.
//   - visitedFunctions (map[string]*models.VisitInfo): A map of functions that have already been visited.
//
// Returns:
//   - ([]models.DataFlowStep): A slice of data flow steps identified during the analysis.
//
// -----------------------------------------------------------------------------
func CrawlFromLine(
	root,
	node *sitter.Node,
	content []byte,
	variablesToTrack map[string]bool,
	startLine uint32,
	startFromEnd bool,
	visitedLines map[uint32]bool,
	visitedFunctions map[string]*models.VisitInfo,
) []models.DataFlowStep {
	var dataFlow []models.DataFlowStep

	functionStart, functionEnd := nodeService.FindFunctionBounds(root, node, startLine)
	var line uint32
	if startFromEnd {
		line = startLine
		logger.PrintInfo("Starting analysis from line %d.", line)
	}

	step := int32(-1) // Backward analysis by default

	for (startFromEnd && line >= functionStart) || (!startFromEnd && line <= functionEnd) { //
		if visitedLines[line] {
			logger.PrintDebug("Line %d already analyzed. Skipping to the next line.", line)
			line = uint32(int32(line) + step)
			continue
		}

		logger.PrintDebug("Analyzing line %d.", line)
		currentNode := nodeService.FindNodeAtLine(root, line)
		if currentNode != nil {
			logger.PrintDebug("Node of type '%s' found at line %d.", currentNode.Type(), line)
			for variable := range variablesToTrack {
				logger.PrintDebug("Analyzing node for variable '%s' at line %d.", variable, line)
				dataFlow = append(dataFlow, analyzeNode(root, currentNode, content, variable, visitedLines, visitedFunctions, variablesToTrack, startLine)...)
			}
		} else {
			logger.PrintWarning("No node found at line %d.", line)
		}

		visitedLines[line] = true
		line = uint32(int32(line) + step) // Move to next/previous line based on analysis direction
	}

	var filteredDataFlow []models.DataFlowStep
	for _, step := range dataFlow {
		if step.Value != "" {
			filteredDataFlow = append(filteredDataFlow, step)
		} else {
			logger.PrintInfo("Skipping data flow step with empty value at line %d.", step.Line)
		}
	}

	logger.PrintInfo("Data flow analysis complete for variable '%s'.", variablesToTrack)
	return filteredDataFlow
}

// analyzeNode - Analyzes a node in the syntax tree to trace variable data flow
// -----------------------------------------------------------------------------
//
// Parameters:
//   - root (*sitter.Node): The root node of the syntax tree.
//   - node (*sitter.Node): The current node being analyzed.
//   - content ([]byte): The content of the source code.
//   - variable (string): The variable to track in the data flow analysis.
//   - visitedLines (map[uint32]bool): A map to keep track of visited lines to avoid duplicate analysis.
//   - visitedFunctions (map[string]*models.VisitInfo): A map to keep track of visited functions and their visit information.
//   - variablesToTrack (map[string]bool): A map of variables to track during the analysis.
//   - startLine (uint32): The starting line number for the analysis.
//
// Returns:
//   - ([]models.DataFlowStep): A slice of DataFlowStep representing the steps in the variable's data flow.
//
// -----------------------------------------------------------------------------
func analyzeNode(
	root, node *sitter.Node,
	content []byte,
	variable string,
	visitedLines map[uint32]bool,
	visitedFunctions map[string]*models.VisitInfo,
	variablesToTrack map[string]bool,
	startLine uint32,
) []models.DataFlowStep {
	var dataFlow []models.DataFlowStep
	line := node.StartPoint().Row + 1

	// Skip analyzing 'block' and 'body_statement' nodes directly to avoid duplicate analysis for langages like python and ruby
	if node.Type() == "block" || node.Type() == "body_statement" || node.Type() == "compound_statement" {
		logger.PrintDebug("Skipping 'block' node at line %d; processing its children instead.", line)
		return append(dataFlow, analyzeNode(
			root, node.Child(0), content, variable,
			visitedLines, visitedFunctions, variablesToTrack, startLine)...)
	}

	// 1. Check if the node is an assignment
	assignment, newVariable := nodeService.IsAssignment(node, content, variable)
	if assignment && line != startLine {
		if !visitedLines[line] {
			logger.PrintInfo("Assignment found for variable '%s' at line %d", variable, line)
			rightNode := node.ChildByFieldName("right")
			value := nodeService.SafeContent(rightNode, content)

			// Check if the right side is a function call
			if rightNode != nil {
				callExprNode := nodeService.FindCallExpression(rightNode)
				if callExprNode != nil {
					functionIdentifier := callExprNode.ChildByFieldName("function")
					if functionIdentifier == nil {
						functionIdentifier = callExprNode.Child(0)
					}
					functionName := nodeService.SafeContent(functionIdentifier, content)
					logger.PrintInfo("Function call to '%s' detected at line %d", functionName, line)

					// Check if the variable is passed as an argument
					variablePassedAsArgument := false
					argumentsNode := callExprNode.ChildByFieldName("arguments")
					if argumentsNode != nil {
						for i := 0; i < int(argumentsNode.NamedChildCount()); i++ {
							argNode := argumentsNode.NamedChild(i)
							argContent := nodeService.SafeContent(argNode, content)
							if argContent == variable {
								variablePassedAsArgument = true
								break
							}
						}
					}

					// Find the function declaration
					funcDeclNode := nodeService.FindFunctionDeclaration(root, functionName, content)

					if funcDeclNode != nil {
						funcLine := funcDeclNode.StartPoint().Row + 1

						if variablePassedAsArgument {
							// Add "Function Declaration" step only if the variable is passed as an argument
							dataFlow = append(dataFlow, models.DataFlowStep{
								Line:     funcLine,
								Type:     "Function Declaration",
								Function: functionName,
								Value:    functionName,
								Variable: variable,
							})
							logger.PrintInfo("Dataflow step added for function '%s' at line %d", functionName, funcLine)
						} else {
							logger.PrintInfo("Variable '%s' is not passed as an argument to function '%s', skipping 'Function Declaration' data flow step", variable, functionName)
						}

						logger.PrintInfo("Function declaration for '%s' found at line %d", functionName, funcLine)

						if variablePassedAsArgument {
							// Proceed to analyze the function only if the variable is passed as an argument
							if visitedFunctions[functionName] == nil {
								visitedFunctions[functionName] = &models.VisitInfo{
									VisitedCalls: make(map[int]bool),
								}
								logger.PrintDebug("Initialized VisitInfo for new function: %s", functionName)
							}

							// Log already visited functions
							logger.PrintInfo("visitedFunctionStack before joining = %v", visitedFunctionStack)

							// Check if function is already in visitedFunctions
							visitInfo, exists := visitedFunctions[functionName]
							if !exists {
								// Initialize if it doesn't exist
								visitedFunctions[functionName] = &models.VisitInfo{VisitedDef: false}
								visitInfo = visitedFunctions[functionName]
							}

							logger.PrintDebug("VisitInfo for function '%s': %+v", functionName, visitInfo)

							if !visitInfo.VisitedDef {
								// Mark the definition as visited
								visitInfo.VisitedDef = true
								visitedFunctionStack = append(visitedFunctionStack, functionName)
								logger.PrintInfo("Entering function '%s' to analyze variable '%s'", functionName, variable)

								// Get the corresponding parameter name
								paramVariable := nodeService.GetParameterName(funcDeclNode, content, variable, node)
								logger.PrintInfo("Tracking variable '%s' as '%s' inside function '%s'", variable, paramVariable, functionName)

								// Create a new variablesToTrack map only for relevant variables
								newVariablesToTrack := make(map[string]bool)
								for varName := range variablesToTrack {
									if varName == variable && paramVariable != "" {
										newVariablesToTrack[paramVariable] = true
										logger.PrintDebug("Mapped variable '%s' to '%s' in function '%s'", varName, paramVariable, functionName)
									} else if nodeService.IsVariableInScope(varName, funcDeclNode, content) {
										// Track only variables that are in scope
										newVariablesToTrack[varName] = true
									}
								}

								// Continue with data flow analysis inside the called function if a variable is mapped
								if len(newVariablesToTrack) > 0 {
									dataFlow = append(dataFlow, CrawlFromLine(
										root, funcDeclNode, content, newVariablesToTrack, funcDeclNode.EndPoint().Row+1, true, visitedLines, visitedFunctions)...)
								} else {
									logger.PrintInfo("No relevant variables to track in function '%s'. Skipping analysis.", functionName)
								}

								// Remove the function from the stack after analysis
								visitedFunctionStack = visitedFunctionStack[:len(visitedFunctionStack)-1]
							} else {
								logger.PrintInfo("Function '%s' has already been visited. Skipping to prevent infinite recursion.", functionName)
							}
						}
					} else {
						logger.PrintInfo("Function declaration for '%s' not found", functionName)
					}
				}
			}

			// Add the assignment step
			dataFlow = append(dataFlow, models.DataFlowStep{
				Line:     line,
				Type:     "Assignment of value",
				Function: nodeService.FindParentFunction(node, content),
				Value:    value,
				Variable: variable,
			})
			visitedLines[line] = true

			// Handle new variables
			if newVariable != "" && !nodeService.IsLiteral(rightNode) {
				// Check if newVariable is an identifier
				if nodeService.IsVariableUsedInExpression(root, newVariable, content) && nodeService.IsValidVariableToTrack(root, newVariable, content) {
					variablesToTrack[newVariable] = true
					logger.PrintInfo("New variable '%s' found in assignment at line %d", newVariable, line)
					dataFlow = append(dataFlow, models.DataFlowStep{
						Line:     line,
						Type:     "Assignment of value",
						Function: nodeService.FindParentFunction(node, content),
						Value:    value,
						Variable: newVariable,
					})
				} else {
					logger.PrintInfo("New variable '%s' is not a valid variable to track.", newVariable)
				}
			} else {
				logger.PrintInfo("Right-hand side is a literal or newVariable is empty; no new variable tracked.")
			}

		} else {
			logger.PrintInfo("Line %d already visited for assignment.", line)
		}
	}

	// 2. Check if the node is a function call
	functionCall, newVariableFromCall := nodeService.IsFunctionCall(node, content, variable)
	if functionCall {
		line := node.StartPoint().Row + 1

		// Extract the method name from the function call node
		methodName := nodeService.ExtractFunctionNameFromCall(node, content)

		logger.PrintInfo("Function call found for variable '%s' at line %d, calling method '%s'", variable, line, methodName)

		if visitedFunctions[methodName] == nil {
			visitedFunctions[methodName] = &models.VisitInfo{
				VisitedCalls: make(map[int]bool),
			}
		}

		methodNameInt := int(line)
		if visitedFunctions[methodName].VisitedCalls[methodNameInt] {
			logger.PrintInfo("Skipping previously visited function '%s' for variable '%s'", methodName, variable)
			return dataFlow
		}
		visitedFunctions[methodName].VisitedCalls[methodNameInt] = true

		if methodName == nodeService.FindParentFunction(node, content) {
			logger.PrintInfo("Skipping recursive call to function '%s' at line %d", methodName, line)
			return dataFlow
		}

		dataFlow = append(dataFlow, models.DataFlowStep{
			Line:     line,
			Type:     "Function parameters",
			Method:   methodName,
			Function: nodeService.FindParentFunction(node, content),
			Value:    variable,
			Variable: variable,
		})
		visitedLines[line] = true

		newFunction := nodeService.FindFunctionByName(root, methodName, content)
		if newFunction != nil {
			// Check if the function has already been visited
			logger.PrintInfo("visitedFunctionStack = %v", visitedFunctionStack)
			if utilityService.ContainsString(visitedFunctionStack, methodName) {
				logger.PrintInfo("Function '%s' has already been visited. Skipping to prevent infinite recursion.", methodName)
				return dataFlow
			}

			// Add the function to the visited functions stack
			visitedFunctionStack = append(visitedFunctionStack, methodName)
			logger.PrintInfo("Entering function '%s' to analyze variable '%s'", methodName, variable)

			// Get the corresponding parameter name
			paramVariable := nodeService.GetParameterName(newFunction, content, variable, node)
			logger.PrintInfo("Tracking variable '%s' as '%s' inside function '%s'", variable, paramVariable, methodName)

			// Create a new variablesToTrack map only for relevant variables
			newVariablesToTrack := make(map[string]bool)
			if paramVariable != "" {
				for varName := range variablesToTrack {
					if varName == variable {
						newVariablesToTrack[paramVariable] = true
						logger.PrintDebug("Mapped variable '%s' to '%s' in function '%s'", varName, paramVariable, methodName)
					}
				}
			}

			// Continue with data flow analysis inside the called function if a variable is mapped
			if len(newVariablesToTrack) > 0 {
				functionStart, functionEnd := nodeService.FindFunctionBounds(root, newFunction, newFunction.StartPoint().Row+1)
				logger.PrintInfo("Function '%s' bounds: %d - %d", methodName, functionStart, functionEnd)
				dataFlow = append(dataFlow, CrawlFromLine(root, newFunction, content, newVariablesToTrack, functionEnd-1, true, visitedLines, visitedFunctions)...)
			} else {
				logger.PrintInfo("No relevant variables to track for function '%s'. Skipping analysis.", methodName)
			}

			// Remove the function from the stack after analysis
			visitedFunctionStack = visitedFunctionStack[:len(visitedFunctionStack)-1]
		} else {
			logger.PrintInfo("Function '%s' not found, treating it as an assignment", methodName)
			dataFlow = append(dataFlow, models.DataFlowStep{
				Line:     line,
				Type:     "Assignment of value",
				Function: nodeService.FindParentFunction(node, content),
				Value:    nodeService.SafeContent(node.ChildByFieldName("right"), content),
				Variable: variable,
			})
			visitedLines[line] = true
		}

		if newVariableFromCall != "" {
			logger.PrintInfo("New variable '%s' assigned from function call at line %d", newVariableFromCall, line)
			if nodeService.IsValidVariableToTrack(root, variable, content) {
				variablesToTrack[variable] = true
				logger.PrintInfo("New variable '%s' found in assignment at line %d", newVariableFromCall, line)
				dataFlow = append(dataFlow, models.DataFlowStep{
					Line:     line,
					Type:     "Assignment of value",
					Function: nodeService.FindParentFunction(node, content),
					Value:    newVariableFromCall,
					Variable: newVariableFromCall,
				})
			} else {
				delete(variablesToTrack, variable)
			}

			dataFlow = append(dataFlow, analyzeNode(root, node, content, newVariableFromCall, visitedLines, visitedFunctions, variablesToTrack, startLine)...)
		}
	}

	// 3. Check if the node is a function declaration
	funcName := nodeService.IsFunctionDeclaration(root, node, content)
	if funcName != "" {
		// Check if the function has already been visited
		logger.PrintInfo("visitedFunctionStack = %v", visitedFunctionStack)
		if utilityService.ContainsString(visitedFunctionStack, funcName) {
			logger.PrintInfo("Function '%s' has already been visited. Skipping to prevent infinite recursion.", funcName)
			return dataFlow
		}

		// Add the function to the visited functions stack
		visitedFunctionStack = append(visitedFunctionStack, funcName)

		if visitedFunctions[funcName] == nil {
			visitedFunctions[funcName] = &models.VisitInfo{
				VisitedCalls: make(map[int]bool),
			}
			logger.PrintDebug("Initialized VisitInfo for new function: %s", funcName)
		}

		// Put the function declaration in the visited functions map
		visitInfo := visitedFunctions[funcName]
		if !visitInfo.VisitedDef {
			visitInfo.VisitedDef = true
			logger.PrintInfo("Function declaration detected and marked: %s", funcName)
		} else {
			logger.PrintInfo("Skipping revisited function declaration: %s", funcName)
			// Remove the function from the stack after analysis
			visitedFunctionStack = visitedFunctionStack[:len(visitedFunctionStack)-1]
			return dataFlow
		}

		// From the function declaration to the call sites
		callSites := nodeService.FindFunctionCallSites(root, funcName, content)
		logger.PrintDebug("Call sites for function '%s'", funcName)
		for _, callSite := range callSites {
			callSiteInt := int(callSite.Line)
			if !visitedFunctions[funcName].VisitedCalls[callSiteInt] {
				visitedFunctions[funcName].VisitedCalls[callSiteInt] = true

				// Map the variables to the function parameters
				newVariablesToTrack := make(map[string]bool)
				variableMapped := false

				logger.PrintInfo("Starting variable mapping at call site line %d", callSite.Line)
				for varName := range variablesToTrack {
					logger.PrintDebug("Analyzing variable '%s' at call site line %d", varName, callSite.Line)

					argVariable := nodeService.GetArgumentVariable(callSite.CallNode, varName, node, content)

					logger.PrintInfo("argVariable '%s' et varName '%s'", argVariable, varName)

					if argVariable != "" {
						if argVariable != varName {
							logger.PrintInfo("Tracking variable '%s' as '%s' at call site line %d", varName, argVariable, callSite.Line)
							newVariablesToTrack[argVariable] = true
							logger.PrintDebug("Mapped variable '%s' to '%s' at call site line %d", varName, argVariable, callSite.Line)
							variableMapped = true
						} else {
							logger.PrintDebug("No mapping needed for variable '%s' at call site line %d", varName, callSite.Line)
							newVariablesToTrack[varName] = true
							variableMapped = true
						}
					} else {
						logger.PrintInfo("Variable '%s' not found in arguments at call site line %d", varName, callSite.Line)
					}
				}

				// Delete variables that were not mapped
				for varName := range variablesToTrack {
					if !newVariablesToTrack[varName] {
						logger.PrintInfo("Removing variable '%s' from tracking as it was not found in the function parameters.", varName)
					}
				}

				// If no variables are mapped, skip the analysis
				if !variableMapped {
					logger.PrintInfo("No relevant variables found in call site at line %d. Skipping analysis for this call.", callSite.Line)
					continue
				}

				// Continue with data flow analysis inside the called function if a variable is mapped
				dataFlow = append(dataFlow, CrawlFromLine(root, node, content, newVariablesToTrack, callSite.Line, true, visitedLines, visitedFunctions)...)
			}
		}

		// Remove the function from the stack after analysis
		visitedFunctionStack = visitedFunctionStack[:len(visitedFunctionStack)-1]
		return dataFlow
	}

	// 4. Check if the node is a control structure
	controlType := nodeService.GetControlType(node.Type())
	if controlType != "" && nodeService.IsVariableUsedInExpression(node, variable, content) {
		line := node.StartPoint().Row + 1
		functionName := nodeService.FindParentFunction(node, content)
		dataFlow = append(dataFlow, models.DataFlowStep{
			Line:     line,
			Type:     controlType,
			Function: functionName,
			Value:    variable,
			Variable: variable,
		})
		logger.PrintInfo("Variable '%s' used in %s at line %d within function '%s'", variable, controlType, line, functionName)
	}

	// 5. Check if the node is a return statement
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		dataFlow = append(dataFlow, analyzeNode(
			root, child, content, variable,
			visitedLines, visitedFunctions, variablesToTrack, startLine)...)
	}

	return dataFlow
}
