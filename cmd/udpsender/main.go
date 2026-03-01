package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	add, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {

		os.Exit(1)
	}
	con, err := net.DialUDP("udp", nil, add)

	if err != nil {

		return
	}
	defer con.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {

			os.Exit(1)
		}
		_, err = con.Write([]byte(line))
		if err != nil {
			os.Exit(1)
		}
		fmt.Printf("message: %s\n", line)
	}

}
