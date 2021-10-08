package main

import (
	"fmt"

	// "mime/multipart"

	// "os"
	"runtime/debug"
	"time"

	"bytes"
	"encoding/binary"
	_ "fmt"
	"net"

	"github.com/jinzhu/configor"
	"github.com/nareix/joy4/format"

	"sync"

	log "github.com/sirupsen/logrus"
)

/*
#cgo CFLAGS: -I./
#define CGO
   #include <memory.h>
   #include <stdio.h>
   #include <stdlib.h>
#include "cheader.h"
*/
import "C"

var Config = struct {
	RedisUrl string `required:"true"`
	IpifyUrl string `required:"true"`
	Port     uint   `required:"true"`
	Message  string `required:"true"`
}{}

var ip string
var adminPort string

var wait sync.WaitGroup

func sendKeepAlive(ticker *time.Ticker) {
	defer wait.Done()
	for {
		select {
		case <-ticker.C:
			{
			}
			// do stuff
		}
	}
}

func init() {
	format.RegisterAll()
}
func main() {
	wait.Add(2)
	configor.Load(&Config, "config.yml")
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)

	// file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err == nil {
	//   log.SetOutput(file)
	// } else {
	//   log.Info("Failed to log to file, using default stderr")
	// }

	defer func() {
		if r := recover(); r != nil {
			log.Infoln(string(debug.Stack()))
		}
	}()

	log.Infoln("START: ", Config.Message)

	// getRedis()
	ticker := time.NewTicker(5 * time.Second)
	go sendKeepAlive(ticker)

	log.Println("======================= START!!")
	go startServer()

	//goroutine 종료를 기다림
	wait.Wait()

	fmt.Println("exit!!")
}

var userCnt int = 0
var jobChan chan int = make(chan int, 1000)

func startServer() {
	defer wait.Done()
	ln, err := net.Listen("tcp", ":"+"3334") // TCP 프로토콜에 8000 포트로 연결을 받음
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close() // main 함수가 끝나기 직전에 연결 대기를 닫음

	go func() {
		start := time.Now()

		for {
			select {
			case <-jobChan:
				{
					userCnt += 1
					elapsed := time.Since(start)
					if elapsed.Milliseconds() >= 1000 {
						fmt.Printf("send: %d/s\n", int64(userCnt*1000)/elapsed.Milliseconds())
						userCnt = 0
						start = time.Now()
					}

				}

			}
		}
	}()

	for {
		conn, err := ln.Accept() // 클라이언트가 연결되면 TCP 연결을 리턴
		if err != nil {
			fmt.Println(err)
			continue
		}
		_ = conn
		// defer conn.Close() // main 함수가 끝나기 직전에 TCP 연결을 닫음

		s := Session{conn}
		// log.Info("======================= Accept!!!!!")
		// defer conn.Close() // main 함수가 끝나기 직전에 TCP 연결을 닫음
		sendChannel := make(chan []byte, 150)
		lock.Lock()
		channelIndex += 1
		insertIndex := channelIndex
		channels[channelIndex] = sendChannel
		// fmt.Println("len: ", len(channels))
		lock.Unlock()
		go s.Handler(conn, sendChannel, insertIndex) // 패킷을 처리할 함수를 고루틴으로 실행
	}
}

type Session struct {
	socket net.Conn
}

var channels map[int]chan []byte = map[int]chan []byte{}
var channelIndex = 0
var lock = sync.RWMutex{}

func (thiz *Session) Handler(splitterSocket net.Conn, sendChannel chan []byte, insertIndex int) {
	thiz.socket = splitterSocket
	data := make([]byte, 4096)
	headerPacket := C.PKT_HDR{}
	headerSize := 4 + 4 + 4 + 4 // C.sizeof_PKT_HDR
	needBytes := headerSize
	step := 1
	var netBuffer bytes.Buffer
	doClose := make(chan bool)
	go func() {
		for {
			select {
			case msg := <-sendChannel:
				jobChan <- 1
				if _, err := splitterSocket.Write(msg); err != nil {
					// log.Error("[writer] ", err, len(sendChannel))
					// return
				} else {
					// log.Infof("================ CHANNEL WRITE OK[%d] ===================", len(msg))
				}
			case <-doClose:
				log.Info("================ CHANNEL WRITER END ===================")
				return
			}
		}
	}()
	for {
		n, err := thiz.socket.Read(data) // 서버에서 받은 데이터를 읽음
		if err != nil {
			// log.Error("================== Socket Error ==================")
			// log.Error(err)
			close(doClose)
			lock.Lock()
			delete(channels, insertIndex)
			lock.Unlock()
			splitterSocket.Close()
			return
		}
		netBuffer.Write(data[:n])
		for netBuffer.Len() >= needBytes {
			if step == 1 {
				bb := make([]byte, needBytes)
				_, err := netBuffer.Read(bb)
				if err != nil {
				}
				_ = binary.Read(bytes.NewBuffer(bb), binary.LittleEndian, &headerPacket)
				step = 2
				needBytes = int(headerPacket.DataLen) // 다음 받을 바이트 수
			} else if step == 2 {
				bb := make([]byte, needBytes)
				_, err := netBuffer.Read(bb)
				if err != nil {
				}

				step = 1
				needBytes = headerSize
				sendChannel <- bb
			}
		}
	}
}
