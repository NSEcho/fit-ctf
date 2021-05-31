package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	listenAddr = ":3000"
	socketType = "tcp"

	correctUsername = "second"
	correctChars    = "0BrBXDmHto"
	secondPassword  = "+zV_H:ERDkBWjR4$"
)

func main() {
	l, err := net.Listen(socketType, listenAddr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			os.Exit(1)
		}

		go handleReq(conn)
	}
}

func handleReq(conn net.Conn) {
	fmt.Println("Connection from", conn.RemoteAddr())
	defer conn.Close()
	conn.Write([]byte("Username: "))
	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		switch err {
		case io.EOF:
		default:
			fmt.Println("Error occurred", err)
		}
		return
	}
	username := strings.TrimSpace(string(data))
	if username != correctUsername {
		conn.Write([]byte("Wrong username!!!"))
		return
	}
	conn.Write([]byte("Ten consecutives characters from the private key: "))
	data, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error ocurred", err)
		return
	}
	chars := strings.TrimSpace(string(data))
	if chars != correctChars {
		conn.Write([]byte("Wrong!!! Did you count correctly???"))
		return
	}
	message := fmt.Sprintf("Password for the provided username: %s", secondPassword)
	conn.Write([]byte(message))
}
