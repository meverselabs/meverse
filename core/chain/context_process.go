package chain

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// ContextProcess is an context for the process
type ContextProcess struct {
	pid uint8
	ctx *types.Context
}

// NewContextProcess returns a ContextProcess
func NewContextProcess(pid uint8, ctx *types.Context) *ContextProcess {
	ctp := &ContextProcess{
		pid: pid,
		ctx: ctx,
	}
	return ctp
}

// Name returns the name of the chain
func (ctp *ContextProcess) Name() string {
	return ctp.ctx.Name()
}

// Version returns the version of the chain
func (ctp *ContextProcess) Version() uint16 {
	return ctp.ctx.Version()
}

// Hash returns the hash value of it
func (ctp *ContextProcess) Hash() hash.Hash256 {
	return ctp.ctx.Hash()
}

// TargetHeight returns the recorded target height when ContextProcess generation
func (ctp *ContextProcess) TargetHeight() uint32 {
	return ctp.ctx.TargetHeight()
}

// LastHash returns the recorded prev hash when ContextProcess generation
func (ctp *ContextProcess) LastHash() hash.Hash256 {
	return ctp.ctx.LastHash()
}

// LastTimestamp returns the last timestamp of the chain
func (ctp *ContextProcess) LastTimestamp() uint64 {
	return ctp.ctx.LastTimestamp()
}

// Top returns the top snapshot
func (ctp *ContextProcess) Top() *types.ContextData {
	return ctp.ctx.Top()
}

// Seq returns the sequence of the target account
func (ctp *ContextProcess) Seq(addr common.Address) uint64 {
	return ctp.ctx.Seq(addr)
}

// AddSeq update the sequence of the target account
func (ctp *ContextProcess) AddSeq(addr common.Address) {
	ctp.ctx.AddSeq(addr)
}

// Account returns the account instance of the address
func (ctp *ContextProcess) Account(addr common.Address) (types.Account, error) {
	return ctp.ctx.Account(addr)
}

// AddressByName returns the account address of the name
func (ctp *ContextProcess) AddressByName(Name string) (common.Address, error) {
	return ctp.ctx.AddressByName(Name)
}

// IsExistAccount checks that the account of the address is exist or not
func (ctp *ContextProcess) IsExistAccount(addr common.Address) (bool, error) {
	return ctp.ctx.IsExistAccount(addr)
}

// IsExistAccountName checks that the account of the name is exist or not
func (ctp *ContextProcess) IsExistAccountName(Name string) (bool, error) {
	return ctp.ctx.IsExistAccountName(Name)
}

// CreateAccount inserts the account to the top snapshot
func (ctp *ContextProcess) CreateAccount(acc types.Account) error {
	return ctp.ctx.CreateAccount(acc)
}

// DeleteAccount deletes the account from the top snapshot
func (ctp *ContextProcess) DeleteAccount(acc types.Account) error {
	return ctp.ctx.DeleteAccount(acc)
}

// AccountDataKeys returns all data keys of the account in the context
func (ctp *ContextProcess) AccountDataKeys(addr common.Address, Prefix []byte) ([][]byte, error) {
	return ctp.ctx.AccountDataKeys(addr, Prefix)
}

// AccountData returns the account data from the top snapshot
func (ctp *ContextProcess) AccountData(addr common.Address, name []byte) []byte {
	return ctp.ctx.AccountData(addr, name)
}

// SetAccountData inserts the account data to the top snapshot
func (ctp *ContextProcess) SetAccountData(addr common.Address, name []byte, value []byte) {
	ctp.ctx.SetAccountData(addr, name, value)
}

// IsExistUTXO checks that the utxo of the id is exist or not
func (ctp *ContextProcess) IsExistUTXO(id uint64) (bool, error) {
	return ctp.ctx.IsExistUTXO(id)
}

// UTXO returns the UTXO from the top snapshot
func (ctp *ContextProcess) UTXO(id uint64) (*types.UTXO, error) {
	return ctp.ctx.UTXO(id)
}

// CreateUTXO inserts the UTXO to the top snapshot
func (ctp *ContextProcess) CreateUTXO(id uint64, vout *types.TxOut) error {
	return ctp.ctx.CreateUTXO(id, vout)
}

// DeleteUTXO deletes the UTXO from the top snapshot
func (ctp *ContextProcess) DeleteUTXO(utxo *types.UTXO) error {
	return ctp.ctx.DeleteUTXO(utxo)
}

// EmitEvent creates the event to the top snapshot
func (ctp *ContextProcess) EmitEvent(e types.Event) error {
	return ctp.ctx.EmitEvent(e)
}

// Dump prints the top ContextProcess data of the context
func (ctp *ContextProcess) Dump() string {
	return ctp.ctx.Dump()
}

// Snapshot push a snapshot and returns the snapshot number of it
func (ctp *ContextProcess) Snapshot() int {
	return ctp.ctx.Snapshot()
}

// Revert removes snapshots after the snapshot number
func (ctp *ContextProcess) Revert(sn int) {
	ctp.ctx.Revert(sn)
}

// Commit apply snapshots to the top after the snapshot number
func (ctp *ContextProcess) Commit(sn int) {
	ctp.ctx.Commit(sn)
}

// StackSize returns the size of the context data stack
func (ctp *ContextProcess) StackSize() int {
	return ctp.ctx.StackSize()
}

// ProcessDataKeys returns all data keys of the process in the context
func (ctp *ContextProcess) ProcessDataKeys(Prefix []byte) ([][]byte, error) {
	return ctp.ctx.ProcessDataKeys(ctp.pid, Prefix)
}

// ProcessData returns the process data from the top snapshot
func (ctp *ContextProcess) ProcessData(name []byte) []byte {
	return ctp.ctx.ProcessData(ctp.pid, name)
}

// SetProcessData inserts the process data to the top snapshot
func (ctp *ContextProcess) SetProcessData(name []byte, value []byte) {
	ctp.ctx.SetProcessData(ctp.pid, name, value)
}
