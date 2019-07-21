package factory

import (
	"reflect"

	"github.com/fletaio/fleta/common/hash"
)

// Factory provides type's factory
type Factory struct {
	nameTypeMap     map[string]uint16
	nameHashTypeMap map[hash.Hash256]uint16
	typeNameMap     map[uint16]string
	typeReflectMap  map[uint16]reflect.Type
}

// NewFactory returns a Factory
func NewFactory() *Factory {
	fc := &Factory{
		nameTypeMap:     map[string]uint16{},
		nameHashTypeMap: map[hash.Hash256]uint16{},
		typeNameMap:     map[uint16]string{},
		typeReflectMap:  map[uint16]reflect.Type{},
	}
	return fc
}

// Register add the type
func (fc *Factory) Register(t uint16, v interface{}) error {
	rt := reflect.TypeOf(v)
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := typeNameOf(rt)
	h := hash.Hash([]byte(name))
	if _, has := fc.typeReflectMap[t]; has {
		return ErrExistType
	}
	if _, has := fc.nameHashTypeMap[h]; has {
		return ErrExistTypeName
	}
	fc.nameHashTypeMap[h] = t
	fc.nameTypeMap[name] = t
	fc.typeNameMap[t] = name
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

// TypeOf returns the type of the value
func (fc *Factory) TypeOf(v interface{}) (uint16, error) {
	name := typeNameOf(reflect.TypeOf(v))
	t, has := fc.nameTypeMap[name]
	if !has {
		return 0, ErrUnknownType
	}
	t, has = fc.nameHashTypeMap[hash.Hash([]byte(name))]
	if !has {
		return 0, ErrUnknownType
	}
	return t, nil
}

// TypeName returns the name of the type
func (fc *Factory) TypeName(t uint16) (string, error) {
	name, has := fc.typeNameMap[t]
	if !has {
		return "", ErrUnknownType
	}
	return name, nil
}

func typeNameOf(rt reflect.Type) string {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	name := rt.Name()
	if pkgPath := rt.PkgPath(); len(pkgPath) > 0 {
		name = pkgPath + "." + name
	}
	return name
}
