package main

import (
	"bytes"
	"encoding/binary"
	_ "fmt"
	"net"

	log "github.com/sirupsen/logrus"
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

func (thiz *Session) Handler(splitterSocket net.Conn) {
	thiz.socket = splitterSocket
	data := make([]byte, 4096)
	headerPacket := C.PKT_HDR{}
	headerSize := 4 + 4 + 4 + 4 // C.sizeof_PKT_HDR
	needBytes := headerSize
	step := 1
	var netBuffer bytes.Buffer

	sendChannel := make(chan []byte, 1500)
	doClose := make(chan bool)
	go func(conn net.Conn) {
		for {
			select {
			case msg := <-sendChannel:
				if _, err := conn.Write(msg); err != nil {
					log.Error("[writer] ", err, len(sendChannel))
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
	}(splitterSocket)
	for {
		n, err := thiz.socket.Read(data) // 서버에서 받은 데이터를 읽음
		if err != nil {
			log.Error("================== Socket Error ==================")
			log.Error(err)
			close(doClose)
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
			}
		}
	}
}