package types

import (
	"bytes"
	"encoding/hex"
	"strconv"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

// ContextData is a state data of the context
type ContextData struct {
	cache               *contextCache
	Parent              *ContextData
	mainToken           *common.Address
	AdminMap            map[common.Address]bool
	DeletedAdminMap     map[common.Address]bool
	AddrSeqMap          map[common.Address]uint64
	GeneratorMap        map[common.Address]bool
	DeletedGeneratorMap map[common.Address]bool
	ContractDefineMap   map[common.Address]*ContractDefine
	DataMap             map[string][]byte
	DeletedDataMap      map[string]bool
	TimeSlotMap         map[uint32]map[string]bool
	basicFee            *amount.Amount
	isTop               bool
	seq                 uint32
	size                uint64
}

// NewContextData returns a ContextData
func NewContextData(cache *contextCache, Parent *ContextData) *ContextData {
	ctd := &ContextData{
		cache:               cache,
		Parent:              Parent,
		AdminMap:            map[common.Address]bool{},
		DeletedAdminMap:     map[common.Address]bool{},
		AddrSeqMap:          map[common.Address]uint64{},
		GeneratorMap:        map[common.Address]bool{},
		DeletedGeneratorMap: map[common.Address]bool{},
		ContractDefineMap:   map[common.Address]*ContractDefine{},
		DataMap:             map[string][]byte{},
		DeletedDataMap:      map[string]bool{},
		TimeSlotMap:         map[uint32]map[string]bool{},
		isTop:               true,
	}
	return ctd
}

func (ctd *ContextData) GetPCSize() uint64 {
	return ctd.size
}

// IsAdmin returns the account is admin or not
func (ctd *ContextData) IsAdmin(addr common.Address) bool {
	if _, has := ctd.DeletedAdminMap[addr]; has {
		return false
	}
	if is, has := ctd.AdminMap[addr]; has {
		return is
	} else if ctd.Parent != nil {
		is := ctd.Parent.IsAdmin(addr)
		ctd.size += 21 //uint32(common.Sizeof(reflect.TypeOf(addr))) + uint32(common.Sizeof(reflect.TypeOf(is)))
		return is
	} else {
		is := ctd.cache.IsAdmin(addr)
		ctd.size += 21 //uint32(common.Sizeof(reflect.TypeOf(addr))) + uint32(common.Sizeof(reflect.TypeOf(is)))
		return is
	}
}

// SetAdmin adds the account as a admin or not
func (ctd *ContextData) SetAdmin(addr common.Address, is bool) error {
	if is {
		if ctd.IsAdmin(addr) {
			return errors.WithStack(ErrAlreadyAdmin)
		}
		delete(ctd.DeletedAdminMap, addr)
		ctd.AdminMap[addr] = true
		ctd.size += 42 // (uint32(common.Sizeof(reflect.TypeOf(addr))) + uint32(common.Sizeof(reflect.TypeOf(bool)))) * 2
	} else {
		if !ctd.IsAdmin(addr) {
			return errors.WithStack(ErrInvalidAdmin)
		}
		delete(ctd.AdminMap, addr)
		ctd.DeletedAdminMap[addr] = true
		ctd.size += 42 // (uint32(common.Sizeof(reflect.TypeOf(addr))) + uint32(common.Sizeof(reflect.TypeOf(bool)))) * 2
	}
	return nil
}

// IsGenerator returns the account is generator or not
func (ctd *ContextData) IsGenerator(addr common.Address) bool {
	if _, has := ctd.DeletedGeneratorMap[addr]; has {
		return false
	}
	ctd.size += 1 // uint32(common.Sizeof(reflect.TypeOf(bool)))
	if is, has := ctd.GeneratorMap[addr]; has {
		return is
	} else if ctd.Parent != nil {
		return ctd.Parent.IsGenerator(addr)
	} else {
		return ctd.cache.IsGenerator(addr)
	}
}

// SetGenerator adds the account as a generator or not
func (ctd *ContextData) SetGenerator(addr common.Address, is bool) error {
	if is {
		if ctd.IsGenerator(addr) {
			return errors.WithStack(ErrAlreadyGenerator)
		}
		delete(ctd.DeletedGeneratorMap, addr)
		ctd.GeneratorMap[addr] = true
		ctd.size += 42 // (uint32(common.Sizeof(reflect.TypeOf(addr))) + uint32(common.Sizeof(reflect.TypeOf(bool)))) * 2
	} else {
		if !ctd.IsGenerator(addr) {
			return errors.WithStack(ErrInvalidGenerator)
		}
		delete(ctd.GeneratorMap, addr)
		ctd.DeletedGeneratorMap[addr] = true
		ctd.size += 42 // (uint32(common.Sizeof(reflect.TypeOf(addr))) + uint32(common.Sizeof(reflect.TypeOf(bool)))) * 2
	}
	return nil
}

// UnsafeGetMainToken returns the MainToken or nil
func (ctd *ContextData) UnsafeGetMainToken() *common.Address {
	return ctd.mainToken
}

// MainToken returns the MainToken
func (ctd *ContextData) MainToken() *common.Address {
	if ctd.mainToken != nil {
		return ctd.mainToken
	}
	var mainToken *common.Address
	if ctd.Parent != nil {
		mainToken = ctd.Parent.MainToken()
	} else {
		mainToken = ctd.cache.MainToken()
	}
	if ctd.isTop {
		ctd.mainToken = mainToken
	}
	ctd.size += 20 // uint32(common.Sizeof(reflect.TypeOf(addr)))
	return mainToken
}

// SetMainToken is set the maintoken
func (ctd *ContextData) SetMainToken(addr common.Address) {
	ctd.mainToken = &addr
	ctd.size += 20 // uint32(common.Sizeof(reflect.TypeOf(addr)))
}

// IsContract returns is the contract
func (ctd *ContextData) IsContract(addr common.Address) bool {
	if _, has := ctd.ContractDefineMap[addr]; has {
		return true
	} else if ctd.Parent != nil {
		return ctd.Parent.IsContract(addr)
	} else {
		ctd.size += 20 // uint32(common.Sizeof(reflect.TypeOf(addr)))
		return ctd.cache.IsContract(addr)
	}

}

// Contract returns the contract
func (ctd *ContextData) Contract(addr common.Address) (Contract, error) {
	if cd, has := ctd.ContractDefineMap[addr]; has {
		return CreateContract(cd)
	} else if ctd.Parent != nil {
		ctd.size += 20 // uint32(common.Sizeof(reflect.TypeOf(addr)))
		return ctd.Parent.Contract(addr)
	} else {
		ctd.size += 20 // uint32(common.Sizeof(reflect.TypeOf(addr)))
		return ctd.cache.Contract(addr)
	}
}

// NextSeq returns the next squence number
func (ctd *ContextData) NextSeq() uint32 {
	ctd.seq++
	return ctd.seq
}

// DeployContract deploy contract to the chain
func (ctd *ContextData) DeployContract(sender common.Address, ClassID uint64, Args []byte) (Contract, error) {
	if !IsValidClassID(ClassID) {
		return nil, errors.WithStack(ErrInvalidClassID)
	}

	base := make([]byte, 1+common.AddressLength+8+4)
	base[0] = 0xff
	copy(base[1:], sender[:])
	copy(base[1+common.AddressLength:], bin.Uint64Bytes(ClassID))
	copy(base[1+common.AddressLength+8:], bin.Uint32Bytes(ctd.NextSeq()))
	height := ctd.cache.ctx.TargetHeight()
	if height > 0 {
		bs := bin.Uint32Bytes(height)
		base = append(base, bs...)
	}
	h := hash.Hash(base)
	addr := common.BytesToAddress(h[12:])
	cd := &ContractDefine{
		Address: addr,
		Owner:   sender,
		ClassID: ClassID,
	}
	cont, err := CreateContract(cd)
	if err != nil {
		return nil, err
	}
	ctd.ContractDefineMap[addr] = cd
	if err := cont.OnCreate(ctd.cache.ctx.ContractContext(cont, addr), Args); err != nil {
		return nil, err
	}
	ctd.size += 68 // uint32(common.Sizeof(reflect.TypeOf(addr))) + uint32(common.Sizeof(reflect.TypeOf(cd)))
	return cont, nil
}

// DeployContract deploy contract to the chain with address
func (ctd *ContextData) DeployContractWithAddress(sender common.Address, ClassID uint64, addr common.Address, Args []byte) (Contract, error) {
	cd := &ContractDefine{
		Address: addr,
		Owner:   sender,
		ClassID: ClassID,
	}
	cont, err := CreateContract(cd)
	if err != nil {
		return nil, err
	}
	ctd.ContractDefineMap[addr] = cd
	if err := cont.OnCreate(ctd.cache.ctx.ContractContext(cont, addr), Args); err != nil {
		return nil, err
	}
	ctd.size += 68 // uint32(common.Sizeof(reflect.TypeOf(addr))) + uint32(common.Sizeof(reflect.TypeOf(cd)))
	return cont, nil
}

// Data returns the data
func (ctd *ContextData) Data(cont common.Address, addr common.Address, name []byte) []byte {
	key := string(cont[:]) + string(addr[:]) + string(name)
	if _, has := ctd.DeletedDataMap[key]; has {
		return nil
	}
	if value, has := ctd.DataMap[key]; has {
		return value
	} else if ctd.Parent != nil {
		value := ctd.Parent.Data(cont, addr, name)
		ctd.size += uint64(len(name)) + uint64(len(value))
		if len(value) > 0 {
			if ctd.isTop {
				nvalue := make([]byte, len(value))
				copy(nvalue, value)
				return nvalue
			} else {
				return value
			}
		} else {
			return nil
		}
	} else {
		value := ctd.cache.Data(cont, addr, name)
		ctd.size += uint64(len(name)) + uint64(len(value))
		if len(value) > 0 {
			if ctd.isTop {
				nvalue := make([]byte, len(value))
				copy(nvalue, value)
				return nvalue
			} else {
				return value
			}
		} else {
			return nil
		}
	}
}

// SetData inserts the data
func (ctd *ContextData) SetData(cont common.Address, addr common.Address, name []byte, value []byte) {
	key := string(cont[:]) + string(addr[:]) + string(name)
	ctd.size += uint64(len(key))
	if len(value) == 0 {
		delete(ctd.DataMap, key)
		ctd.DeletedDataMap[key] = true
		ctd.size += 1 // bool
	} else {
		delete(ctd.DeletedDataMap, key)
		ctd.DataMap[key] = value
		ctd.size += uint64(len(value))
	}
}

// IsUsedTimeSlot returns timeslot is used or not
func (ctd *ContextData) IsUsedTimeSlot(slot uint32, key string) bool {
	if mp, has := ctd.TimeSlotMap[slot]; has {
		if _, has := mp[key]; has {
			return true
		}
	}
	if has := ctd.cache.IsUsedTimeSlot(slot, key); has {
		mp, has := ctd.TimeSlotMap[slot]
		if !has {
			mp = map[string]bool{}
			ctd.TimeSlotMap[slot] = mp
		}
		ctd.size += 5 // uint32 + bool
		mp[key] = true
		return true
	}
	return false
}

// UseTimeSlot consumes timeslot
func (ctd *ContextData) UseTimeSlot(slot uint32, key string) error {
	mp, has := ctd.TimeSlotMap[slot]
	if !has {
		mp = map[string]bool{}
		ctd.TimeSlotMap[slot] = mp
	} else if _, has := mp[key]; has {
		return errors.WithStack(ErrUsedTimeSlot)
	}
	if has := ctd.cache.IsUsedTimeSlot(slot, key); has {
		ctd.size += uint64(len(key) + 1) // string + bool
		mp[key] = true
		return errors.WithStack(ErrUsedTimeSlot)
	} else {
		ctd.size += uint64(len(key) + 1) // string + bool
		mp[key] = true
	}
	return nil
}

// Seq returns the number of txs using the UseSeq flag of the address.
func (ctd *ContextData) AddrSeq(addr common.Address) uint64 {
	var seq uint64
	var has bool
	if seq, has = ctd.AddrSeqMap[addr]; has {
		return seq
	}

	if ctd.Parent != nil {
		seq = ctd.Parent.AddrSeq(addr)
	} else {
		seq = ctd.cache.AddrSeq(addr)
	}
	if ctd.isTop {
		ctd.AddrSeqMap[addr] = seq
	}
	return seq
}

// AddSeq update the sequence of the target address
func (ctd *ContextData) AddAddrSeq(addr common.Address) {
	ctd.AddrSeqMap[addr] = ctd.AddrSeq(addr) + 1
	ctd.size += 28 // addr + uint64
}

// BasicFee returns the basic fee
func (ctd *ContextData) BasicFee() *amount.Amount {
	if ctd.basicFee != nil {
		return ctd.basicFee
	}

	var fee *amount.Amount
	if ctd.Parent != nil {
		fee = ctd.Parent.BasicFee()
	} else {
		fee = ctd.cache.BasicFee()
	}
	if ctd.isTop {
		ctd.basicFee = fee
	}

	return fee
}

// SetBasicFee update the basic fee
func (ctd *ContextData) SetBasicFee(fee *amount.Amount) {
	ctd.basicFee = fee
}

// Hash returns the hash value of it
func (ctd *ContextData) Hash() hash.Hash256 {
	var buffer bytes.Buffer
	buffer.WriteString("ChainID")
	buffer.Write(ctd.cache.ctx.ChainID().Bytes())
	buffer.WriteString("ChainVersion")
	buffer.Write(bin.Uint16Bytes(ctd.cache.ctx.Version()))
	buffer.WriteString("Height")
	buffer.Write(bin.Uint32Bytes(ctd.cache.ctx.TargetHeight()))
	buffer.WriteString("PrevHash")
	PrevHash := ctd.cache.ctx.PrevHash()
	buffer.Write(PrevHash[:])
	buffer.WriteString("AdminMap")
	EachAllAddressBool(ctd.AdminMap, func(key common.Address, value bool) error {
		buffer.Write(key[:])
		return nil
	})
	buffer.WriteString("DeletedAdminMap")
	EachAllAddressBool(ctd.DeletedAdminMap, func(key common.Address, value bool) error {
		buffer.Write(key[:])
		return nil
	})
	buffer.WriteString("AddrSeqMap")
	EachAllAddressUint64(ctd.AddrSeqMap, func(key common.Address, value uint64) error {
		buffer.Write(key[:])
		buffer.Write([]byte{0})
		buffer.Write(bin.Uint64Bytes(value))
		return nil
	})
	buffer.WriteString("GeneratorMap")
	EachAllAddressBool(ctd.GeneratorMap, func(key common.Address, value bool) error {
		buffer.Write(key[:])
		return nil
	})
	buffer.WriteString("MainToken")
	if ctd.mainToken != nil {
		buffer.Write((*ctd.mainToken)[:])
	}
	buffer.WriteString("DeletedGeneratorMap")
	EachAllAddressBool(ctd.DeletedGeneratorMap, func(key common.Address, value bool) error {
		buffer.Write(key[:])
		return nil
	})
	buffer.WriteString("DataMap")
	EachAllStringBytes(ctd.DataMap, func(key string, value []byte) error {
		buffer.Write([]byte(key))
		buffer.Write(value)
		return nil
	})
	buffer.WriteString("DeletedDataMap")
	EachAllStringBool(ctd.DeletedDataMap, func(key string, value bool) error {
		buffer.WriteString(key)
		return nil
	})
	buffer.WriteString("TimeSlotMap")
	EachAllTimeSlotMap(ctd.TimeSlotMap, func(key uint32, value map[string]bool) error {
		buffer.Write(bin.Uint32Bytes(key))
		EachAllStringBool(value, func(key string, value bool) error {
			buffer.WriteString(key)
			return nil
		})
		return nil
	})
	return hash.DoubleHash(buffer.Bytes())
}

// Dump prints the context data
func (ctd *ContextData) Dump() string {
	var buffer bytes.Buffer
	buffer.WriteString("ChainID\n")
	buffer.WriteString(ctd.cache.ctx.ChainID().String())
	buffer.WriteString("\n")
	buffer.WriteString("ChainVersion\n")
	buffer.WriteString(strconv.FormatUint(uint64(ctd.cache.ctx.Version()), 10))
	buffer.WriteString("\n")
	buffer.WriteString("Height\n")
	buffer.WriteString(strconv.FormatUint(uint64(ctd.cache.ctx.TargetHeight()), 10))
	buffer.WriteString("\n")
	buffer.WriteString("PrevHash\n")
	PrevHash := ctd.cache.ctx.PrevHash()
	buffer.WriteString(PrevHash.String())
	buffer.WriteString("\n")
	buffer.WriteString("AdminMap\n")
	EachAllAddressBool(ctd.AdminMap, func(key common.Address, value bool) error {
		buffer.WriteString(key.String())
		buffer.WriteString("\n")
		return nil
	})
	buffer.WriteString("DeletedAdminMap\n")
	EachAllAddressBool(ctd.DeletedAdminMap, func(key common.Address, value bool) error {
		buffer.WriteString(key.String())
		buffer.WriteString("\n")
		return nil
	})
	buffer.WriteString("AddrSeqMap\n")
	EachAllAddressUint64(ctd.AddrSeqMap, func(key common.Address, value uint64) error {
		buffer.WriteString(key.String())
		buffer.WriteString(":")
		buffer.WriteString(strconv.FormatUint(value, 10))
		buffer.WriteString("\n")
		return nil
	})
	buffer.WriteString("GeneratorMap\n")
	EachAllAddressBool(ctd.GeneratorMap, func(key common.Address, value bool) error {
		buffer.WriteString(key.String())
		buffer.WriteString("\n")
		return nil
	})
	buffer.WriteString("MainToken")
	if ctd.mainToken != nil {
		buffer.Write((*ctd.mainToken)[:])
	}
	buffer.WriteString("DeletedGeneratorMap\n")
	EachAllAddressBool(ctd.DeletedGeneratorMap, func(key common.Address, value bool) error {
		buffer.WriteString(key.String())
		buffer.WriteString("\n")
		return nil
	})
	buffer.WriteString("DataMap\n")
	EachAllStringBytes(ctd.DataMap, func(key string, value []byte) error {
		buffer.WriteString(hex.EncodeToString([]byte(key)))
		buffer.WriteString(":")
		buffer.WriteString(hash.Hash(value).String())
		buffer.WriteString("\n")
		return nil
	})
	buffer.WriteString("DeletedDataMap\n")
	EachAllStringBool(ctd.DeletedDataMap, func(key string, value bool) error {
		buffer.WriteString(hex.EncodeToString([]byte(key)))
		buffer.WriteString("\n")
		return nil
	})
	buffer.WriteString("TimeSlotMap")
	EachAllTimeSlotMap(ctd.TimeSlotMap, func(key uint32, value map[string]bool) error {
		buffer.WriteString(strconv.FormatUint(uint64(key), 10))
		buffer.WriteString(":\n")
		EachAllStringBool(value, func(key string, value bool) error {
			buffer.WriteString("    ")
			buffer.WriteString(hex.EncodeToString([]byte(key)))
			buffer.WriteString("\n")
			return nil
		})
		buffer.WriteString("\n")
		return nil
	})
	return buffer.String()
}
