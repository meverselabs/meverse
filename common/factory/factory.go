package factory

import (
	"reflect"

	"github.com/fletaio/fleta/common/hash"
)

// Factory provide account's handlers of the target chain
type Factory struct {
	nameTypeName   map[hash.Hash256]uint16
	typeReflectMap map[uint16]reflect.Type
}

// NewFactory returns a Factory
func NewFactory() *Factory {
	fc := &Factory{
		nameTypeName:   map[hash.Hash256]uint16{},
		typeReflectMap: map[uint16]reflect.Type{},
	}
	return fc
}

// Register add the account type with handler loaded by the name from the global account registry
func (fc *Factory) Register(t uint16, v interface{}) error {
	rt := reflect.TypeOf(v)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := TypeHashOf(rt)
	if _, has := fc.typeReflectMap[t]; has {
		return ErrExistType
	}
	if _, has := fc.nameTypeName[name]; has {
		return ErrExistTypeName
	}
	fc.nameTypeName[name] = t
	fc.typeReflectMap[t] = rt
	return nil
}

// Create creates the instance of the type
func (fc *Factory) Create(t uint16) (interface{}, error) {
	rt, has := fc.typeReflectMap[t]
	if !has {
		return nil, ErrUnknownType
	}
	return reflect.New(rt).Interface(), nil
}

// TypeOf returns the type of the account
func (fc *Factory) TypeOf(v interface{}) (uint16, error) {
	name := TypeHashOf(reflect.TypeOf(v))

	t, has := fc.nameTypeName[name]
	if !has {
		return 0, ErrUnknownType
	}
	return t, nil
}

// TypeHashOf returns the hash of the type
func TypeHashOf(rt reflect.Type) hash.Hash256 {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.Name()
	if pkgPath := rt.PkgPath(); len(pkgPath) > 0 {
		name = pkgPath + "." + name
	}
	return hash.Hash([]byte(name))
}
