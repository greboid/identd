package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/kouhin/envflag"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	response = flag.String("response", "ident", "Response to ident request")
	port     = flag.Int("port", 8000, "Port to listen on")
)

func main() {
	if err := envflag.Parse(); err != nil {
		fmt.Printf("Unable to load config: %s\r\n", err.Error())
		return
	}
	sigWait := make(chan os.Signal, 1)
	signal.Notify(sigWait, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	done := make(chan bool, 1)
	go func() {
		<-sigWait
		done <- true
	}()

	fmt.Println("Starting Identd")
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("Unable to listen: %s\r\n", err.Error())
		return
	}
	go handleConnections(listener)
	<-done
	_ = listener.Close()
	fmt.Println("Exiting.")
}

func handleConnections(listener net.Listener) {
	for {
		c, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	line, _, _ := bufio.NewReader(c).ReadLine()
	data := strings.Split(strings.ReplaceAll(string(line), " ", ""), ",")
	if len(data) != 2 {
		_ = c.Close()
		return
	}
	_, _ = c.Write([]byte(fmt.Sprintf("%s, %s : USERID : UNIX : %s\r\n", strings.TrimSpace(data[0]), strings.TrimSpace(data[1]), *response)))
	_ = c.Close()
}
