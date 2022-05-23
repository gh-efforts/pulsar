package main

import (
	"fmt"

	"github.com/bitrainforest/pulsar/chain/agen/generator"
)

func main() {
	if err := generator.Gen(); err != nil {
		fmt.Println(err)
	}
}
