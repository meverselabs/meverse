package bank

import (
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	lediscfg "github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"

	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/service/p2p"

	"github.com/fletaio/fleta/common/amount"

	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/vault"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/core/backend"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/service/apiserver"
)

type Bank struct {
	sync.Mutex
	keyStore  backend.StoreBackend
	st        *chain.Store
	cn        types.Provider
	nd        *p2p.Node
	db        *ledis.DB
	seqMap    map[common.Address]uint64
	waitTxMap map[hash.Hash256]*chan string
}

// NewBank returns a Bank
func NewBank(keyStore backend.StoreBackend, dbpath string) *Bank {
	cfg := lediscfg.NewConfigDefault()
	cfg.DataDir = dbpath
	l, err := ledis.Open(cfg)
	if err != nil {
		panic(err)
	}
	db, err := l.Select(0)
	if err != nil {
		panic(err)
	}

	s := &Bank{
		keyStore:  keyStore,
		db:        db,
		seqMap:    map[common.Address]uint64{},
		waitTxMap: map[hash.Hash256]*chan string{},
	}
	return s
}

// Name returns the name of the service
func (s *Bank) Name() string {
	return "fleta.bank"
}

// InitFromStore initializes account datas from the store
func (s *Bank) InitFromStore(st *chain.Store) error {
	s.st = st

	if err := s.updateAccountData(); err != nil {
		return err
	}
	return nil
}

func (s *Bank) SetNode(nd *p2p.Node) {
	s.nd = nd
}

// Init called when initialize service
func (s *Bank) Init(pm types.ProcessManager, cn types.Provider) error {
	s.cn = cn

	if vs, err := pm.ServiceByName("fleta.apiserver"); err != nil {
		//ignore when not loaded
	} else if v, is := vs.(*apiserver.APIServer); !is {
		//ignore when not loaded
	} else {
		as, err := v.JRPC("bank")
		if err != nil {
			return err
		}
		as.Set("keyNames", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			names, err := s.KeyNames()
			if err != nil {
				return nil, err
			}
			return names, nil
		})
		as.Set("accounts", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 1 {
				return nil, apiserver.ErrInvalidArgument
			}
			name, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			addrs, err := s.Accounts(name)
			if err != nil {
				return nil, err
			}
			return addrs, nil
		})
		as.Set("createKey", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 2 {
				return nil, apiserver.ErrInvalidArgument
			}
			name, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			Password, err := arg.String(1)
			if err != nil {
				return nil, err
			}
			if err := s.CreateKey(name, Password); err != nil {
				return nil, err
			}
			return nil, nil
		})
		as.Set("importKey", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 3 {
				return nil, apiserver.ErrInvalidArgument
			}
			name, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			hexed, err := arg.String(1)
			if err != nil {
				return nil, err
			}
			bs, err := hex.DecodeString(hexed)
			if err != nil {
				return nil, err
			}
			Password, err := arg.String(2)
			if err != nil {
				return nil, err
			}
			if err := s.ImportKey(name, bs, Password); err != nil {
				return nil, err
			}
			return nil, nil
		})
		as.Set("checkPassword", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 2 {
				return nil, apiserver.ErrInvalidArgument
			}
			name, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			Password, err := arg.String(1)
			if err != nil {
				return nil, err
			}
			if err := s.CheckPassword(name, Password); err != nil {
				return nil, err
			}
			return nil, nil
		})
		as.Set("changePassword", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 3 {
				return nil, apiserver.ErrInvalidArgument
			}
			name, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			oldPassword, err := arg.String(1)
			if err != nil {
				return nil, err
			}
			Password, err := arg.String(2)
			if err != nil {
				return nil, err
			}
			if err := s.ChangePassword(name, oldPassword, Password); err != nil {
				return nil, err
			}
			return nil, nil
		})
		as.Set("deleteKey", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 2 {
				return nil, apiserver.ErrInvalidArgument
			}
			name, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			Password, err := arg.String(1)
			if err != nil {
				return nil, err
			}
			if err := s.DeleteKey(name, Password); err != nil {
				return nil, err
			}
			return nil, nil
		})
		as.Set("accountDetail", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 1 {
				return nil, apiserver.ErrInvalidArgument
			}
			addrStr, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			addr, err := common.ParseAddress(addrStr)
			if err != nil {
				return nil, err
			}
			acc, err := s.cn.NewLoaderWrapper(1).Account(addr)
			if err != nil {
				return nil, err
			}
			return acc, nil
		})
		as.Set("send", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 4 {
				return nil, apiserver.ErrInvalidArgument
			}
			fromStr, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			from, err := common.ParseAddress(fromStr)
			if err != nil {
				return nil, err
			}
			toStr, err := arg.String(1)
			if err != nil {
				return nil, err
			}
			to, err := common.ParseAddress(toStr)
			if err != nil {
				return nil, err
			}
			amStr, err := arg.String(2)
			if err != nil {
				return nil, err
			}
			am, err := amount.ParseAmount(amStr)
			if err != nil {
				return nil, err
			}
			Password, err := arg.String(3)
			if err != nil {
				return nil, err
			}

			name, err := s.NameByAddress(from)
			if err != nil {
				return nil, err
			}

			s.Lock()
			Seq, has := s.seqMap[from]
			ChainSeq := s.cn.Seq(from)
			if !has || Seq < ChainSeq {
				Seq = ChainSeq
			}
			Seq++
			s.seqMap[from] = Seq
			s.Unlock()

			tx := &vault.Transfer{
				Timestamp_: uint64(time.Now().UnixNano()),
				Seq_:       Seq,
				From_:      from,
				To:         to,
				Amount:     am,
			}
			TxHash := chain.HashTransaction(s.cn.ChainID(), tx)
			sig, err := s.Sign(name, Password, TxHash)
			if err != nil {
				return nil, err
			}
			if err := s.nd.AddTx(tx, []common.Signature{sig}); err != nil {
				return nil, err
			}
			if err := s.addPending(tx); err != nil {
				return nil, err
			}
			return TxHash, nil
		})
		as.Set("transaction", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 1 {
				return nil, apiserver.ErrInvalidArgument
			}
			txid, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			height, index, err := types.ParseTransactionID(txid)
			if err != nil {
				return nil, err
			}
			b, err := s.cn.Block(height)
			if err != nil {
				return nil, err
			}
			if len(b.Transactions) <= int(index) {
				return nil, ErrInvalidTXID
			}
			t := b.TransactionTypes[index]
			tx := b.Transactions[index]
			result := b.TransactionResults[index]

			fc := encoding.Factory("transaction")
			bs, err := tx.MarshalJSON()
			if err != nil {
				return nil, err
			}
			mp := map[string]interface{}{}
			if err := json.Unmarshal(bs, &mp); err != nil {
				return nil, err
			}
			name, err := fc.TypeName(t)
			if err != nil {
				return nil, err
			}
			mp["type"] = name
			mp["txid"] = txid
			mp["result"] = result
			return mp, nil
		})
		as.Set("transactions", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 3 {
				return nil, apiserver.ErrInvalidArgument
			}
			addrStr, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			addr, err := common.ParseAddress(addrStr)
			if err != nil {
				return nil, err
			}
			offset, err := arg.Int(1)
			if err != nil {
				return nil, err
			}
			count, err := arg.Int(2)
			if err != nil {
				return nil, err
			}
			txids, txs, results, err := s.getTransactionFromTail(addr, int32(offset), int32(count))
			if err != nil {
				return nil, err
			}
			fc := encoding.Factory("transaction")
			txmps := []map[string]interface{}{}
			for i, tx := range txs {
				bs, err := tx.MarshalJSON()
				if err != nil {
					return nil, err
				}
				mp := map[string]interface{}{}
				if err := json.Unmarshal(bs, &mp); err != nil {
					return nil, err
				}
				t, err := fc.TypeOf(tx)
				if err != nil {
					return nil, err
				}
				name, err := fc.TypeName(t)
				if err != nil {
					return nil, err
				}
				mp["type"] = name
				mp["txid"] = txids[i]
				mp["result"] = results[i]
				txmps = append(txmps, mp)
			}
			return txmps, nil
		})
		as.Set("transferSends", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 3 {
				return nil, apiserver.ErrInvalidArgument
			}
			addrStr, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			addr, err := common.ParseAddress(addrStr)
			if err != nil {
				return nil, err
			}
			offset, err := arg.Int(1)
			if err != nil {
				return nil, err
			}
			count, err := arg.Int(2)
			if err != nil {
				return nil, err
			}
			txids, txs, results, err := s.getTransferSendFromTail(addr, int32(offset), int32(count))
			if err != nil {
				return nil, err
			}
			fc := encoding.Factory("transaction")
			txmps := []map[string]interface{}{}
			for i, tx := range txs {
				bs, err := tx.MarshalJSON()
				if err != nil {
					return nil, err
				}
				mp := map[string]interface{}{}
				if err := json.Unmarshal(bs, &mp); err != nil {
					return nil, err
				}
				t, err := fc.TypeOf(tx)
				if err != nil {
					return nil, err
				}
				name, err := fc.TypeName(t)
				if err != nil {
					return nil, err
				}
				mp["type"] = name
				mp["txid"] = txids[i]
				mp["result"] = results[i]
				txmps = append(txmps, mp)
			}
			return txmps, nil
		})
		as.Set("transferRecvs", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 3 {
				return nil, apiserver.ErrInvalidArgument
			}
			addrStr, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			addr, err := common.ParseAddress(addrStr)
			if err != nil {
				return nil, err
			}
			offset, err := arg.Int(1)
			if err != nil {
				return nil, err
			}
			count, err := arg.Int(2)
			if err != nil {
				return nil, err
			}
			txids, txs, results, err := s.getTransferRecvFromTail(addr, int32(offset), int32(count))
			if err != nil {
				return nil, err
			}
			fc := encoding.Factory("transaction")
			txmps := []map[string]interface{}{}
			for i, tx := range txs {
				bs, err := tx.MarshalJSON()
				if err != nil {
					return nil, err
				}
				mp := map[string]interface{}{}
				if err := json.Unmarshal(bs, &mp); err != nil {
					return nil, err
				}
				t, err := fc.TypeOf(tx)
				if err != nil {
					return nil, err
				}
				name, err := fc.TypeName(t)
				if err != nil {
					return nil, err
				}
				mp["type"] = name
				mp["txid"] = txids[i]
				mp["result"] = results[i]
				txmps = append(txmps, mp)
			}
			return txmps, nil
		})
		as.Set("pendings", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
			if arg.Len() != 1 {
				return nil, apiserver.ErrInvalidArgument
			}
			addrStr, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			addr, err := common.ParseAddress(addrStr)
			if err != nil {
				return nil, err
			}
			txs, err := s.getPendingsByAddress(addr)
			if err != nil {
				return nil, err
			}
			fc := encoding.Factory("transaction")
			txmps := []map[string]interface{}{}
			for _, tx := range txs {
				bs, err := tx.MarshalJSON()
				if err != nil {
					return nil, err
				}
				mp := map[string]interface{}{}
				if err := json.Unmarshal(bs, &mp); err != nil {
					return nil, err
				}
				t, err := fc.TypeOf(tx)
				if err != nil {
					return nil, err
				}
				name, err := fc.TypeName(t)
				if err != nil {
					return nil, err
				}
				mp["tx_hash"] = chain.HashTransactionByType(s.cn.ChainID(), t, tx).String()
				mp["type"] = name
				txmps = append(txmps, mp)
			}
			return txmps, nil
		})
	}
	return nil
}

// OnLoadChain called when the chain loaded
func (s *Bank) OnLoadChain(loader types.Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (s *Bank) OnBlockConnected(b *types.Block, events []types.Event, loader types.Loader) {
	s.keyStore.View(func(txn backend.StoreReader) error {
		for i, t := range b.Transactions {
			TXID := types.TransactionID(b.Header.Height, uint16(i))
			res := b.TransactionResults[i]
			if at, is := t.(chain.AccountTransaction); is {
				s.removePending(at)

				if tx, is := t.(*vault.Transfer); is {
					_, err := txn.Get(toAddressNameKey(tx.From()))
					if err != nil {
						_, err := txn.Get(toAddressNameKey(tx.To))
						if err != nil {
							continue
						}
					}
				} else {
					_, err := txn.Get(toAddressNameKey(at.From()))
					if err != nil {
						continue
					}
				}

				s.addTransaction(TXID, b.TransactionTypes[i], at, res)
				CreatedAddr := common.NewAddress(b.Header.Height, uint16(i), 0)
				if res == 1 {
					switch tx := t.(type) {
					case *vault.CreateAccount:
						s.addAccount(CreatedAddr, tx.KeyHash)
					case *vault.IssueAccount:
						s.addAccount(CreatedAddr, tx.KeyHash)
					case *formulator.ChangeOwner:
						s.removeAccount(tx.From())
						s.addAccount(tx.From(), tx.KeyHash)
					case *formulator.CreateAlpha:
						s.addAccount(CreatedAddr, tx.KeyHash)
					case *formulator.Transmute:
						s.addAccount(CreatedAddr, tx.KeyHash)
					case *formulator.CreateSigma:
						for i, addr := range tx.AlphaFormulators {
							if i > 0 {
								s.removeAccount(addr)
							}
						}
					case *formulator.CreateOmega:
						for i, addr := range tx.SigmaFormulators {
							if i > 0 {
								s.removeAccount(addr)
							}
						}
					case *formulator.CreateHyper:
						s.addAccount(CreatedAddr, tx.KeyHash)
					case *formulator.Unstaking:
						s.addUnstaking(tx.HyperFormulator, tx.From(), b.Header.Height+2592000, tx.Amount)
					case *formulator.RevertUnstaking:
						s.removeUnstaking(tx.HyperFormulator, tx.From(), tx.UnstakedHeight, tx.Amount)
					case *vault.Transfer:
						s.addTransfer(TXID, tx)
					}
				}
			}
		}
		return nil
	})
}

// KeyNames returns names of keys from the wallet
func (s *Bank) KeyNames() ([]string, error) {
	names := []string{}
	if err := s.keyStore.View(func(txn backend.StoreReader) error {
		txn.Iterate(tagSecret, func(key []byte, value []byte) error {
			name, err := fromSecretKey(key)
			if err != nil {
				return err
			}
			names = append(names, name)
			return nil
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return names, nil
}

// Accounts returns accounts of the name from the wallet
func (s *Bank) Accounts(name string) ([]common.Address, error) {
	addrs := []common.Address{}
	if err := s.keyStore.View(func(txn backend.StoreReader) error {
		txn.Iterate(toNameAddressPrefix(name), func(key []byte, value []byte) error {
			addr, err := fromNameAddress(key, name)
			if err != nil {
				return err
			}
			addrs = append(addrs, addr)
			return nil
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return addrs, nil
}

// NameByAddress returns the name of the address from the wallet
func (s *Bank) NameByAddress(addr common.Address) (string, error) {
	var name string
	if err := s.keyStore.View(func(txn backend.StoreReader) error {
		bs, err := txn.Get(toAddressNameKey(addr))
		if err != nil {
			return err
		}
		name = string(bs)
		return nil
	}); err != nil {
		return "", err
	}
	return name, nil
}

// CreateKey creates the private key with password to the wallet
func (s *Bank) CreateKey(name string, Password string) error {
	if err := s.keyStore.Update(func(txn backend.StoreWriter) error {
		k, err := key.NewMemoryKey()
		if err != nil {
			return nil
		}
		pubhash := common.NewPublicHash(k.PublicKey())
		ens, err := Cipher(k.Bytes(), Password)
		k.Clear()
		if err != nil {
			return err
		}
		if _, err := txn.Get(toSecretKey(name)); err == nil {
			return ErrExistKeyName
		}
		if err := txn.Set(toSecretKey(name), ens); err != nil {
			return err
		}
		if err := txn.Set(toPublicHashKey(pubhash), []byte(name)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// ImportKey adds the private key with password to the wallet
func (s *Bank) ImportKey(name string, bs []byte, Password string) error {
	k, err := key.NewMemoryKeyFromBytes(bs)
	if err != nil {
		return nil
	}
	pubhash := common.NewPublicHash(k.PublicKey())
	k.Clear()

	if err := s.keyStore.Update(func(txn backend.StoreWriter) error {
		ens, err := Cipher(bs, Password)
		copy(bs, make([]byte, len(bs)))
		if err != nil {
			return err
		}
		if _, err := txn.Get(toSecretKey(name)); err == nil {
			return ErrExistKeyName
		}
		if err := txn.Set(toSecretKey(name), ens); err != nil {
			return err
		}
		if err := txn.Set(toPublicHashKey(pubhash), []byte(name)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	s.updateAccountData()
	return nil
}

// CheckPassword checks the password of the key
func (s *Bank) CheckPassword(name string, Password string) error {
	if err := s.keyStore.View(func(txn backend.StoreReader) error {
		bs, err := txn.Get(toSecretKey(name))
		if err != nil {
			return err
		}
		des, err := Decipher(bs, Password)
		if err != nil {
			return err
		}
		copy(des, make([]byte, len(des)))
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// ChangePassword changes the password of the key
func (s *Bank) ChangePassword(name string, oldPassword string, Password string) error {
	if err := s.keyStore.Update(func(txn backend.StoreWriter) error {
		bs, err := txn.Get(toSecretKey(name))
		if err != nil {
			return err
		}
		des, err := Decipher(bs, oldPassword)
		if err != nil {
			return err
		}
		ens, err := Cipher(des, Password)
		copy(des, make([]byte, len(des)))
		if err != nil {
			return err
		}
		if err := txn.Set(toSecretKey(name), ens); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// DeleteKey removes the private key from the wallet
func (s *Bank) DeleteKey(name string, Password string) error {
	if err := s.keyStore.Update(func(txn backend.StoreWriter) error {
		bs, err := txn.Get(toSecretKey(name))
		if err != nil {
			return err
		}
		des, err := Decipher(bs, Password)
		if err != nil {
			return err
		}
		copy(des, make([]byte, len(des)))
		if err := txn.Delete(toSecretKey(name)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// Sign makes a signature of the message using the private key of the name
func (s *Bank) Sign(name string, Password string, MessageHash hash.Hash256) (common.Signature, error) {
	var Key key.Key
	if err := s.keyStore.View(func(txn backend.StoreReader) error {
		bs, err := txn.Get(toSecretKey(name))
		if err != nil {
			return err
		}
		des, err := Decipher(bs, Password)
		if err != nil {
			return err
		}
		v, err := key.NewMemoryKeyFromBytes(des)
		copy(des, make([]byte, len(des)))
		if err != nil {
			return err
		}
		Key = v
		return nil
	}); err != nil {
		return common.Signature{}, err
	}
	defer Key.Clear()

	sig, err := Key.Sign(MessageHash)
	if err != nil {
		return common.Signature{}, err
	}
	return sig, nil
}

func (s *Bank) updateAccountData() error {
	accs, err := s.st.Accounts()
	if err != nil {
		return err
	}
	if err := s.keyStore.Update(func(txn backend.StoreWriter) error {
		for _, a := range accs {
			switch acc := a.(type) {
			case *vault.SingleAccount:
				bsName, err := txn.Get(toPublicHashKey(acc.KeyHash))
				if err != nil {
					continue
				}
				if err := txn.Set(toNameAddressKey(string(bsName), acc.Address()), []byte("vault.SingleAccount")); err != nil {
					return err
				}
				if err := txn.Set(toAddressNameKey(acc.Address()), bsName); err != nil {
					return err
				}
			case *formulator.FormulatorAccount:
				bsName, err := txn.Get(toPublicHashKey(acc.KeyHash))
				if err != nil {
					continue
				}
				if err := txn.Set(toNameAddressKey(string(bsName), acc.Address()), []byte("formulator.FormulatorAccount")); err != nil {
					return err
				}
				if err := txn.Set(toAddressNameKey(acc.Address()), bsName); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
