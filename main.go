package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	prod string 	// 产品线
	srv  string // 模块名
	idlConf string // idl 配置文件
	update bool
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
	flag.StringVar(&idlConf, "f", "idl.yaml", "idl config file path")
	flag.BoolVar(&update, "u", false, "update by git or not")
	flag.Parse()

	if len(prod) == 0 || len(srv) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		wdAbs, _ := filepath.Abs(wd)
		ss := strings.Split(filepath.ToSlash(wdAbs), "/")
		//fmt.Println("wdAbs:", wdAbs, " ss:", ss)
		if len(ss) < 2 {
			fmt.Println("product or service is missing")
			os.Exit(1)
		}

		prod = ss[len(ss)-2]
		srv = ss[len(ss)-1]
	}
}
