package main

import (
	"bytes"
	"encoding/binary"
	_ "fmt"
	"net"

	log "github.com/sirupsen/logrus"
    "sync"
)

import "C"

/*
#cgo CFLAGS: -I./
#define CGO
   #include <memory.h>
   #include <stdio.h>
   #include <stdlib.h>
#include "cheader.h"
*/
import "C"

type Session struct {
	socket net.Conn
}

var channels map[int]chan []byte = map[int]chan[]byte{}
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
				// log.Info("woww")
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
		// fmt.Println("받은 데이터:\n", hex.Dump(data[:n]))
		netBuffer.Write(data[:n])
		// PKT_HDR 프로토콜 처리하는 부분
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
				// log.Infof("len(bb): %d", len(bb))
                sendChannel <- bb
                // lock.Lock()
                // sendChannel <- bb
                // // for _, j := range channels {
                // //     j <- bb
                // // }
                // lock.Unlock()
			}
		}
	}
}
