package main

import (
  "fmt"
  "sync"

  // "mime/multipart"

  "net"
  // "os"
  "runtime/debug"
  "time"

  "github.com/jinzhu/configor"
  "github.com/nareix/joy4/format"
  log "github.com/sirupsen/logrus"
)

/*
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
  start := time.Now()

  go func() {
      for {
          select {
          case <-jobChan:
              {
                  if userCnt % 1000 == 0 {
                      fmt.Println("start measure")
                      start = time.Now()
                  }
                  userCnt += 1
                  if userCnt % 1000 == 0 {
                      duration := time.Since(start)
                      fmt.Println("end measure: ", userCnt, duration)
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
