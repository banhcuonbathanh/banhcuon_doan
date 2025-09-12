// logger/core/interfaces.go - Logger interfaces
package core

// Output interface for different output destinations
type Output interface {
	Write(entry *LogEntry) error
	Close() error
}

// OutputManager interface for managing multiple outputs
type OutputManager interface {
	WriteToAll(entry *LogEntry) error
	WriteToOutput(name string, entry *LogEntry) error
	AddOutput(name string, output Output) error
	RemoveOutput(name string) error
	Close() error
}