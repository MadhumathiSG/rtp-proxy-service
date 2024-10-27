package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	bufferSize = 1500 // Standard UDP buffer size for RTP packets
	maxWorkers = 10
)

// ForwardRTP forwards RTP packets from the source to two destinations.
func ForwardRTP2Dest(srcAddr, destAddr1, destAddr2 string) error {

	srcConn, err := net.ListenPacket("udp", srcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", srcAddr, err)
	}
	defer srcConn.Close()

	destUDPAddr1, err := net.ResolveUDPAddr("udp", destAddr1)
	if err != nil {
		return fmt.Errorf("failed to resolve destination address %s: %w", destAddr1, err)
	}

	destUDPAddr2, err := net.ResolveUDPAddr("udp", destAddr2)
	if err != nil {
		return fmt.Errorf("failed to resolve destination address %s: %w", destAddr2, err)
	}

	destConn1, err := net.DialUDP("udp", nil, destUDPAddr1)
	if err != nil {
		return fmt.Errorf("failed to connect to destination %s: %w", destAddr1, err)
	}
	defer destConn1.Close()

	destConn2, err := net.DialUDP("udp", nil, destUDPAddr2)
	if err != nil {
		return fmt.Errorf("failed to connect to destination %s: %w", destAddr2, err)
	}
	defer destConn2.Close()

	fmt.Printf("Proxying RTP from %s to %s and %s\n", srcAddr, destAddr1, destAddr2)

	buffer := make([]byte, bufferSize)

	packetChan := make(chan []byte, maxWorkers) // Buffered channel to handle incoming packets

	for i := 0; i < maxWorkers; i++ {
		go func() {
			for pkt := range packetChan {
				_, err = destConn1.Write(pkt)
				if err != nil {
					log.Printf("failed to forward packet: %v\n", err)
				}
				_, err = destConn2.Write(pkt)
				if err != nil {
					log.Printf("failed to forward packet: %v\n", err)
				}
			}
		}()
	}

	for {
		n, addr, err := srcConn.ReadFrom(buffer)
		if err != nil {
			log.Printf("failed to read from source: %v\n", err)
			break
		}

		log.Printf("Received %d bytes from %s", n, addr.String())

		pkt := make([]byte, n)
		copy(pkt, buffer[:n])
		packetChan <- pkt

	}
	return nil
}

func main() {

	// Start forwarding RTP packets to two destinations
	err := ForwardRTP2Dest("127.0.0.1:5000", "52.212.107.49:5000", "10.250.195.18:6234")
	if err != nil {
		log.Fatalf("Error starting RTP proxy: %v\n", err)
	}

	// Keep the proxy running indefinitely
	for {
		time.Sleep(time.Second * 10)
	}
}
