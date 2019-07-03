package types

// Application is a interface of the chain application
type Application interface {
	Process
	InitGenesis(ctw *ContextWrapper) error
}
