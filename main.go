package main // import "github.com/fletaio/fleta"
import (
	"log"

	"github.com/fletaio/fleta/common"
)

func main() {
	log.Println(common.NewAddress(0, 37, 0).String())
}
