package p2p

import (
	"io"
	"reflect"

	"github.com/fletaio/fleta_v2/common/bin"
	"github.com/fletaio/fleta_v2/common/hash"
	"github.com/pkg/errors"
)

type Serializable interface {
	TypeID() uint32
	io.WriterTo
	io.ReaderFrom
}

func ToTypeID(name string) uint32 {
	h := hash.Hash([]byte(name))
	return bin.Uint32(h[len(h)-4:])
}

var gSerializableTypeMap = map[uint32]reflect.Type{}
var gSerializableNameMap = map[uint32]string{}

func RegisterSerializableType(s Serializable) uint32 {
	rt := reflect.TypeOf(s)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.Name()
	if pkgPath := rt.PkgPath(); len(pkgPath) > 0 {
		name = pkgPath + "." + name
	}
	h := hash.Hash([]byte(name))
	TypeID := bin.Uint32(h[len(h)-8:])
	gSerializableNameMap[TypeID] = name

	if _, has := gSerializableTypeMap[TypeID]; has {
		panic(ErrExistSerializableType)
	}
	gSerializableTypeMap[TypeID] = rt
	return TypeID
}

func CreateSerializable(TypeID uint32) (Serializable, error) {
	rt, has := gSerializableTypeMap[TypeID]
	if !has {
		return nil, errors.WithStack(ErrInvalidSerializableTypeID)
	}
	return reflect.New(rt).Interface().(Serializable), nil
}

func SerializableName(TypeID uint32) string {
	return gSerializableNameMap[TypeID]
}
