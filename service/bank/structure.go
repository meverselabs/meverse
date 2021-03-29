package bank

type Transaction struct {
	Type   uint16
	Data   []byte
	Result uint8
}
