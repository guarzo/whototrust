package xlog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func getFilePathWithDir(file string) string {
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return file
	}
	relativePath, err := filepath.Rel(projectRoot, file)
	if err != nil {
		fmt.Println("Error getting relative path:", err)
		return file
	}
	return relativePath
}

// logMessage is a helper function that logs a message with its location
func logMessage(message string, file string, line int, ok bool) {
	if ok {
		fmt.Printf("%s:%d - %s\n", getFilePathWithDir(file), line, message)
	} else {
		fmt.Printf("%s\n", message)
	}
}

func Log(message string) {
	_, file, line, ok := runtime.Caller(1)
	logMessage(message, file, line, ok)
}

func Logf(format string, a ...interface{}) {
	_, file, line, ok := runtime.Caller(1)
	logMessage(fmt.Sprintf(format, a...), file, line, ok)
}

func LogIndirect(message string) {
	_, file, line, ok := runtime.Caller(2)
	logMessage(message, file, line, ok)
}

func Logt(functionName string, start time.Time) {
	_, file, line, ok := runtime.Caller(1)
	logMessage(fmt.Sprintf("%s - %vms", functionName, time.Since(start).Milliseconds()), file, line, ok)
}
