package main

import (
	"bufio"
	"fmt"
	"os"

	dk "./DockerRun"
)

func main() {
	ans, err := dk.NewMockContainer(`docker run --rm -v $PWD:/workspace --network=test zhenshaw/1000010020_angular ng build`)
	if err != nil {
		panic(err)
	}
	ans.Printf()
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Docker command: # ")
		cmd, _ := reader.ReadString('\n')
		ctr, err := dk.NewMockContainer(cmd)
		if err != nil {
			fmt.Println("Error: ", err)
		} else {
			res := dk.Judge(&ctr, &ans)
			fmt.Println(res)
		}
	}
}
