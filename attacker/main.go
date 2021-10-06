package main

import (
	"fmt"
	"log"
	"net"
	"time"
    "os"
)

func worker(id int, jobs <-chan int, results chan<- uint64) {
	for j := range jobs {
		_ = j
		conn, err := net.Dial("tcp", os.Args[1]) // "112.170.25.154:3333")
		if nil != err {
			log.Println(err)
		}

		go func() {
			data := make([]byte, 4096)

			for {
				n, err := conn.Read(data)
				// atomic.AddUint64(&ops, uint64(n))
				// atomic.AddUint64(&cnt, 1)

				if err != nil {
					log.Println(err)
					return
				}
				if n == 5 {
					results <- 1
					// conn.Close()
					// break
				}

				// log.Println("Server send : ", data[:n])
			}
		}()
		send := []byte{0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		send = append(send, []byte{0x99, 0x99, 0x99, 0x99, 0x99}...)
		conn.Write(send)
	}
}
func main() {

	nn := make(chan uint64, 10000)
	start := time.Now()
	sockets := uint64(3000)
	go func() {
		sum := uint64(0)
		for {
			select {
			case i := <-nn:
				sum += i
				if sum == sockets*1 {
					duration := time.Since(start)
					fmt.Println(sum, duration)
				}
			}
		}

	}()

	jobs := make(chan int, 10000)
	for w := 1; w <= 5; w++ {
		go worker(w, jobs, nn)
	}
	for i := uint64(0); i < sockets; i++ {
		jobs <- int(i)
	}
	for {
	}
}
