package chain

// Application is a interface of the chain application
type Application interface {
	Process
	InitGenesis(ctp *ContextProcess) error
}
