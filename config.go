package main

import (
	"os"
	"strings"
)

type Config struct {
	Depends []string `yaml:"depends"`
}



func RemoveProtoFile(path string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}

	if !f.IsDir() {
		if strings.HasSuffix(path, ".proto") {
			os.RemoveAll(path)
		}
	}
	return nil
}
