package types

// Process defines process functions
type Process interface {
	ID() uint8
	Version() string
	Name() string
}
