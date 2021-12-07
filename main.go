package main

import (
	"fmt"

	"gitlab.qredo.com/qredo-server/core-client/service"
)

func main() {
	if err := service.New().Start(); err != nil {
		fmt.Println(err)
	}
}
