package main

import (
	"log"
	"time"

	localnode "github.com/meverselabs/meverse/cmd/localnode"
	"github.com/meverselabs/meverse/cmd/localnode/sandbox"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	go func() {
		time.Sleep(time.Second * 2)
		localnode.Run()
	}()
	go sandbox.Run()
	time.Sleep(time.Second)
	select {}
}
