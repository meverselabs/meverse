package types

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
)

// ContextWrapper is an context for the process
type ContextWrapper struct {
	pid uint8
	ctx *Context
}

// NewContextWrapper returns a ContextWrapper
func NewContextWrapper(pid uint8, ctx *Context) *ContextWrapper {
	ctw := &ContextWrapper{
		pid: pid,
		ctx: ctx,
	}
	return ctw
}

// SwitchContextWrapper returns a SwitchContextWrapper of the pid
func SwitchContextWrapper(pid uint8, ctw *ContextWrapper) *ContextWrapper {
	cts := &ContextWrapper{
		pid: pid,
		ctx: ctw.ctx,
	}
	return cts
}

// Switch returns a SwitchContextWrapper of the pid
func (ctw *ContextWrapper) Switch(pid uint8) *ContextWrapper {
	if ctw.pid == pid {
		return ctw
	} else {
		return SwitchContextWrapper(pid, ctw)
	}
}

// Name returns the name of the chain
func (ctw *ContextWrapper) Name() string {
	return ctw.ctx.Name()
}

// Version returns the version of the chain
func (ctw *ContextWrapper) Version() uint16 {
	return ctw.ctx.Version()
}

// Hash returns the hash value of it
func (ctw *ContextWrapper) Hash() hash.Hash256 {
	return ctw.ctx.Hash()
}

// TargetHeight returns the recorded target height when ContextWrapper generation
func (ctw *ContextWrapper) TargetHeight() uint32 {
	return ctw.ctx.TargetHeight()
}

// LastHash returns the recorded prev hash when ContextWrapper generation
func (ctw *ContextWrapper) LastHash() hash.Hash256 {
	return ctw.ctx.LastHash()
}

// LastTimestamp returns the last timestamp of the chain
func (ctw *ContextWrapper) LastTimestamp() uint64 {
	return ctw.ctx.LastTimestamp()
}

// Top returns the top snapshot
func (ctw *ContextWrapper) Top() *ContextData {
	return ctw.ctx.Top()
}

// Seq returns the sequence of the target account
func (ctw *ContextWrapper) Seq(addr common.Address) uint64 {
	return ctw.ctx.Seq(addr)
}

// AddSeq update the sequence of the target account
func (ctw *ContextWrapper) AddSeq(addr common.Address) {
	ctw.ctx.AddSeq(addr)
}

// Account returns the account instance of the address
func (ctw *ContextWrapper) Account(addr common.Address) (Account, error) {
	return ctw.ctx.Account(addr)
}

// AddressByName returns the account address of the name
func (ctw *ContextWrapper) AddressByName(Name string) (common.Address, error) {
	return ctw.ctx.AddressByName(Name)
}

// HasAccount checks that the account of the address is exist or not
func (ctw *ContextWrapper) HasAccount(addr common.Address) (bool, error) {
	return ctw.ctx.HasAccount(addr)
}

// HasAccountName checks that the account of the name is exist or not
func (ctw *ContextWrapper) HasAccountName(Name string) (bool, error) {
	return ctw.ctx.HasAccountName(Name)
}

// CreateAccount inserts the account to the top snapshot
func (ctw *ContextWrapper) CreateAccount(acc Account) error {
	return ctw.ctx.CreateAccount(acc)
}

// DeleteAccount deletes the account from the top snapshot
func (ctw *ContextWrapper) DeleteAccount(acc Account) error {
	return ctw.ctx.DeleteAccount(acc)
}

// AccountDataKeys returns all data keys of the account in the context
func (ctw *ContextWrapper) AccountDataKeys(addr common.Address, Prefix []byte) ([][]byte, error) {
	return ctw.ctx.AccountDataKeys(addr, ctw.pid, Prefix)
}

// AccountData returns the account data from the top snapshot
func (ctw *ContextWrapper) AccountData(addr common.Address, name []byte) []byte {
	return ctw.ctx.AccountData(addr, ctw.pid, name)
}

// SetAccountData inserts the account data to the top snapshot
func (ctw *ContextWrapper) SetAccountData(addr common.Address, name []byte, value []byte) {
	ctw.ctx.SetAccountData(addr, ctw.pid, name, value)
}

// HasUTXO checks that the utxo of the id is exist or not
func (ctw *ContextWrapper) HasUTXO(id uint64) (bool, error) {
	return ctw.ctx.HasUTXO(id)
}

// UTXO returns the UTXO from the top snapshot
func (ctw *ContextWrapper) UTXO(id uint64) (*UTXO, error) {
	return ctw.ctx.UTXO(id)
}

// CreateUTXO inserts the UTXO to the top snapshot
func (ctw *ContextWrapper) CreateUTXO(id uint64, vout *TxOut) error {
	return ctw.ctx.CreateUTXO(id, vout)
}

// DeleteUTXO deletes the UTXO from the top snapshot
func (ctw *ContextWrapper) DeleteUTXO(utxo *UTXO) error {
	return ctw.ctx.DeleteUTXO(utxo)
}

// EmitEvent creates the event to the top snapshot
func (ctw *ContextWrapper) EmitEvent(e Event) error {
	return ctw.ctx.EmitEvent(e)
}

// Dump prints the top ContextWrapper data of the context
func (ctw *ContextWrapper) Dump() string {
	return ctw.ctx.Dump()
}

// Snapshot push a snapshot and returns the snapshot number of it
func (ctw *ContextWrapper) Snapshot() int {
	return ctw.ctx.Snapshot()
}

// Revert removes snapshots after the snapshot number
func (ctw *ContextWrapper) Revert(sn int) {
	ctw.ctx.Revert(sn)
}

// Commit apply snapshots to the top after the snapshot number
func (ctw *ContextWrapper) Commit(sn int) {
	ctw.ctx.Commit(sn)
}

// StackSize returns the size of the context data stack
func (ctw *ContextWrapper) StackSize() int {
	return ctw.ctx.StackSize()
}

// ProcessDataKeys returns all data keys of the process in the context
func (ctw *ContextWrapper) ProcessDataKeys(Prefix []byte) ([][]byte, error) {
	return ctw.ctx.ProcessDataKeys(ctw.pid, Prefix)
}

// ProcessData returns the process data from the top snapshot
func (ctw *ContextWrapper) ProcessData(name []byte) []byte {
	return ctw.ctx.ProcessData(ctw.pid, name)
}

// SetProcessData inserts the process data to the top snapshot
func (ctw *ContextWrapper) SetProcessData(name []byte, value []byte) {
	ctw.ctx.SetProcessData(ctw.pid, name, value)
}
