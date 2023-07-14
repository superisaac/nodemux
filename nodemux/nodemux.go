package main

import (
	//"fmt"
	"github.com/superisaac/nodemux/server"
	//"net/http"
	//_ "net/http/pprof"
)

func main() {
	// go func() {
	// 	fmt.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	server.CommandStartServer()

}
