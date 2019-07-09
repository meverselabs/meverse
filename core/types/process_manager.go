package types

// ProcessManager managers processes in the chain
type ProcessManager interface {
	Processes() []Process
	Process(id uint8) (Process, error)
	ProcessByName(name string) (Process, error)
}
