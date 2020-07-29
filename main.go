// main
package main

import (
	"pgp-sftp-proxy/lib"
	"flag"
	"fmt"
	"runtime"
)

func main() {
	conf_path := flag.String("c", "conf.json", "config json file")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	err, conf := lib.NewConfig(*conf_path)
	if err != nil {
		fmt.Println(err)
		return
	}

	service := lib.NewHTTP(conf)
	service.Start()
}
