package types

// Application defines chain application functions
type Application interface {
	Process
	InitGenesis(ctw *ContextWrapper) error
}

// ApplicationBase is a base handler of the chain process
type ApplicationBase struct{}

// ID must returns 255
func (app *ApplicationBase) ID() uint8 {
	return 255
}

// InitGenesis initializes genesis data
func (app *ApplicationBase) InitGenesis(ctw *ContextWrapper) error {
	return nil
}

// OnLoadChain called when the chain loaded
func (app *ApplicationBase) OnLoadChain(loader LoaderWrapper) error {
	return nil
}

// BeforeExecuteTransactions called before processes transactions of the block
func (app *ApplicationBase) BeforeExecuteTransactions(ctw *ContextWrapper) error {
	return nil
}

// AfterExecuteTransactions called after processes transactions of the block
func (app *ApplicationBase) AfterExecuteTransactions(b *Block, ctw *ContextWrapper) error {
	return nil
}

// OnSaveData called when the context of the block saved
func (app *ApplicationBase) OnSaveData(b *Block, ctw *ContextWrapper) error {
	return nil
}
