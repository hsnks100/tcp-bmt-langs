package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// var clients uint64 = 0
func worker(id int, jobs <-chan int, results chan<- uint64) {
	for j := range jobs {
		_ = j

		conn, err := net.Dial("tcp", "127.0.0.1:"+os.Args[1])
		if nil != err {
			log.Println(err)
		}
		send := []byte{0x11, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x5, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		send = append(send, []byte{0x99, 0x99, 0x99, 0x99, 0x99}...)
		conn.Write(send)
		go func() {
			// atomic.AddUint64(&ops, uint64(1))
			data := make([]byte, 4096)
			byteSum := 0

			for {
				n, err := conn.Read(data)
				byteSum += n
				if err != nil {
					log.Println(err)
					return
				}
				results <- uint64(n)
				if n == 5 {
				}
			}
		}()
	}
}
func main() {

	nn := make(chan uint64, 10000)
	sockets := uint64(2000)

	jobs := make(chan int, 1000)
	for w := 1; w <= 5; w++ {
		go worker(w, jobs, nn)
	}
	start := time.Now()
	go func() {
		sum := uint64(0)
		for {
			select {
			case i := <-nn:
				sum += i
				// if sum >= (sockets*(sockets+1)/2)*5 {
				if sum >= sockets*5 {
					duration := time.Since(start)
					fmt.Println(sum, duration)
				}
				fmt.Println(sum)
			}
		}

	}()
	for i := uint64(0); i < sockets; i++ {
		jobs <- int(i)
		// time.Sleep(10 * time.Millisecond)
	}
	fmt.Scanln()
	fmt.Println("done")
}
