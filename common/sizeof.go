package common

import (
	"reflect"
)

// reflect.TypeOf(packetHeader)
func Sizeof(t reflect.Type) int {
	switch t.Kind() {
	case reflect.Array:
		//fmt.Println("reflect.Array")
		if s := Sizeof(t.Elem()); s >= 0 {
			return s * t.Len()
		}
	case reflect.Map:
		//fmt.Println("reflect.Array")
		if s := Sizeof(t.Key()); s >= 0 {
			if s2 := Sizeof(t.Elem()); s2 >= 0 {
				return (s * t.Len()) + (s2 * t.Len())
			}
		}

	case reflect.Struct:
		//fmt.Println("reflect.Struct")
		sum := 0
		for i, n := 0, t.NumField(); i < n; i++ {
			s := Sizeof(t.Field(i).Type)
			if s < 0 {
				return -1
			}
			sum += s
		}
		return sum

	case reflect.Uint8, reflect.Bool, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		//fmt.Println("reflect.int")
		return int(t.Size())
	case reflect.Slice:
		//fmt.Println("reflect.Slice:", sizeof(t.Elem()))
		return 0
	}

	return 0

}

type PacketHeader struct {
	N1 int32 // sizeof에서 사용할 때 int로 하면 안 됨
	N2 int16
	N3 int64
}
