package main

import (
	"context"
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

func main() {
	logger.Init()
	config.Init()
	for {
		run()
		time.Sleep(time.Second)
	}
}
func run() {
	var err error
	c, _, err := websocket.DefaultDialer.Dial(config.ServerWs, http.Header{"authorization": []string{config.Auth}})
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		c.Close()
		log.Println("disconnect with the server")
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, err := f.Read(buf)
				if err != nil {
					log.Println(err)
					cancel()
					return
				}
				if n > 0 {
					c.WriteMessage(websocket.BinaryMessage, buf[:n])
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				mt, message, err := c.ReadMessage()
				if err != nil {
					log.Println(err)
					cancel()
					return
				}
				if mt == websocket.BinaryMessage {
					_, err = f.Write(message)
					if err != nil {
						log.Println(err)
						cancel()
						return
					}
				}
			}
		}
	}()
	<-ctx.Done()
}
