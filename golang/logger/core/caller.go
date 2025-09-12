// logger/core/caller.go - Caller information and stack trace utilities
package core

import (
	"runtime"
	"strings"
)

// Enhanced caller information gathering
func getEnhancedCaller(skip int) *CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}
	
	// Get function information
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return nil
	}
	
	fullFuncName := fn.Name()
	
	// Extract package and function name
	var pkg, funcName string
	if lastSlash := strings.LastIndex(fullFuncName, "/"); lastSlash >= 0 {
		pkg = fullFuncName[:lastSlash]
		funcName = fullFuncName[lastSlash+1:]
	} else {
		funcName = fullFuncName
	}
	
	// Clean up the function name (remove package prefix if present)
	if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
		if len(funcName) > lastDot+1 {
			// Keep the package part for context
			pkg = funcName[:lastDot]
			funcName = funcName[lastDot+1:]
		}
	}
	
	// Clean up file path to show only relevant part
	if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
		file = file[lastSlash+1:]
	}
	
	return &CallerInfo{
		Function: funcName,
		File:     file,
		Line:     line,
		Package:  pkg,
	}
}

// Capture stack trace for better error debugging
func captureStackTrace(skip, depth int) []CallerInfo {
	var stack []CallerInfo
	
	for i := skip; i < skip+depth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		fullFuncName := fn.Name()
		
		// Extract package and function name
		var pkg, funcName string
		if lastSlash := strings.LastIndex(fullFuncName, "/"); lastSlash >= 0 {
			pkg = fullFuncName[:lastSlash]
			funcName = fullFuncName[lastSlash+1:]
		} else {
			funcName = fullFuncName
		}
		
		// Clean up the function name
		if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
			if len(funcName) > lastDot+1 {
				pkg = funcName[:lastDot]
				funcName = funcName[lastDot+1:]
			}
		}
		
		// Clean up file path
		if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
			file = file[lastSlash+1:]
		}
		
		stack = append(stack, CallerInfo{
			Function: funcName,
			File:     file,
			Line:     line,
			Package:  pkg,
		})
	}
	
	return stack
}