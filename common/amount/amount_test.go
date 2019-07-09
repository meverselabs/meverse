package amount

import (
	"log"
	"testing"
)

func Test_Amount(t *testing.T) {
	a := COIN.DivC(1000)
	b := COIN.MulC(10000)
	log.Println(a.String())
	log.Println(b.String())
	log.Println(a.Add(b).String())
	log.Println(a.Sub(b).String())
	log.Println(a.DivC(10000).String())
	log.Println(a.MulC(90000).String())
	c, _ := ParseAmount("10000.00121454")
	log.Println(c.String())
}
