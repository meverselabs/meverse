package types

import (
	"reflect"
	"strings"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

var gContractTypeMap = map[uint64]reflect.Type{}
var gContractNameMap = map[uint64]string{}

// IMPORTANT: RegisterContractType must be called only at initialization time
// and never have to called concurrently with CreateContract, IsValidClassID, ContractName functions
func RegisterContractType(cont Contract) (uint64, error) {
	name, rt, ClassID := _getContractClassID(cont)

	if v, has := gContractNameMap[ClassID]; has {
		if name != v {
			return 0, errors.WithStack(ErrExistContractType)
		} else {
			return ClassID, nil
		}
	}
	gContractNameMap[ClassID] = name
	gContractTypeMap[ClassID] = rt
	return ClassID, nil
}

func _getContractClassID(cont Contract) (string, reflect.Type, uint64) {
	rt := reflect.TypeOf(cont)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.Name()
	if pkgPath := rt.PkgPath(); len(pkgPath) > 0 {
		pkgPath = strings.Replace(pkgPath, "meverselabs/meverse", "fletaio/fleta_v2", -1)
		name = pkgPath + "." + name
	}
	h := hash.Hash([]byte(name))
	ClassID := bin.Uint64(h[len(h)-8:])
	return name, rt, ClassID
}

func GetContractClassID(cont Contract) (string, uint64) {
	name, _, ClassID := _getContractClassID(cont)
	return name, ClassID
}

func CreateContract(cd *ContractDefine) (Contract, error) {
	rt, has := gContractTypeMap[cd.ClassID]
	if !has {
		return nil, errors.WithStack(ErrInvalidClassID)
	}
	cont := reflect.New(rt).Interface().(Contract)
	cont.Init(cd.Address, cd.Owner)
	return cont, nil
}

func IsValidClassID(ClassID uint64) bool {
	_, has := gContractTypeMap[ClassID]
	return has
}

func ContractName(ClassID uint64) string {
	return gContractNameMap[ClassID]
}

func ContractType(ClassID uint64) string {
	return gContractTypeMap[ClassID].String()
}
