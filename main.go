package main

import (
	"flag"
	"os"
)

var (
	prod string 	// 产品线
	srv  string // 模块名
	idlConf string // idl 配置文件
)



func main() {
	if len(prod) == 0 || len(srv) == 0 || len(idlConf) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	srvDir, _ = os.Getwd()

	idl := NewIDLFolder(prod, srv)

	idl.LoadConfig(idlConf)

	idl.PrepareEnv()
	idl.Extract()
	idl.Cleanup()
	idl.Transfer()
}

func init() {
	flag.StringVar(&prod, "p", "", "product name")
	flag.StringVar(&srv, "m", "", "service name")
	flag.StringVar(&idlConf, "f", "", "idl config file path")
	flag.Parse()
}
