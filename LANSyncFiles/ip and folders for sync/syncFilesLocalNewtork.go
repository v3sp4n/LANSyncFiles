package main

import (
	// "net/values"
	"io/ioutil"
	// "net/http"
	"fmt"
	"os"
)

func main() {
	// var ips map[string][]string
	if _,err := os.Stat("ip and folders for sync/"); err != nil {
		os.Mkdir("ip and folders for sync/", os.ModeDir)
	}
	files,_ := ioutil.ReadDir("ip and folders for sync/")
	for _,v := range files {
		fmt.Println(v.Name())
	}
}