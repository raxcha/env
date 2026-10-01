package main

import (
	"os"
	"strings"
	"env/master"
)


func main() {

	master.CreateMaster(strings.Join(os.Args[1:], " "))
	select {}
}