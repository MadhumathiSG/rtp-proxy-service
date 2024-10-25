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

	for {
		n, addr, err := srcConn.ReadFrom(buffer)
		if err != nil {
			log.Printf("failed to read from source: %v\n", err)
			break
		}

		log.Printf("Received %d bytes from %s", n, addr.String())

		_, err = destConn1.Write(buffer[:n])
		if err != nil {
			log.Printf("failed to forward packet to destination 1: %v\n", err)
			break
		}

		_, err = destConn2.Write(buffer[:n])
		if err != nil {
			log.Printf("failed to forward packet to destination 2: %v\n", err)
			break
		}
	}
	return nil
}

func ForwardRTP(srcAddr, destAddr string) error {
	srcConn, err := net.ListenPacket("udp", srcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", srcAddr, err)
	}
	defer srcConn.Close()

	destUDPAddr, err := net.ResolveUDPAddr("udp", destAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve destination address %s: %w", destAddr, err)
	}

	buffer := make([]byte, bufferSize)

	destConn, err := net.DialUDP("udp", nil, destUDPAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to destination %s: %w", destAddr, err)
	}
	defer destConn.Close()

	err = srcConn.(*net.UDPConn).SetReadBuffer(8 * 1024 * 1024)
	if err != nil {
		return fmt.Errorf("failed to set read buffer: %w", err)
	}
	err = destConn.SetWriteBuffer(8 * 1024 * 1024) // 4 MB
	if err != nil {
		return fmt.Errorf("failed to set write buffer: %w", err)
	}

	fmt.Printf("Proxying RTP from %s to %s\n", srcAddr, destAddr)

	packetChan := make(chan []byte, maxWorkers) // Buffered channel to handle incoming packets

	for i := 0; i < maxWorkers; i++ {
		go func() {
			for pkt := range packetChan {
				_, err := destConn.Write(pkt)
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
			continue
		}

		log.Printf("Received %d bytes from %s", n, addr.String())

		pkt := make([]byte, n)
		copy(pkt, buffer[:n])
		packetChan <- pkt
	}
}

func workingForwardRTP(srcAddr, destAddr string) error {

	srcConn, err := net.ListenPacket("udp", srcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", srcAddr, err)
	}
	defer srcConn.Close()

	destUDPAddr, err := net.ResolveUDPAddr("udp", destAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve destination address %s: %w", destAddr, err)
	}

	destConn1, err := net.DialUDP("udp", nil, destUDPAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to destination %s: %w", destAddr, err)
	}
	defer destConn1.Close()

	buffer := make([]byte, bufferSize)

	fmt.Printf("Proxying RTP from %s to %s\n", srcAddr, destAddr)

	for {
		n, addr, err := srcConn.ReadFrom(buffer)
		if err != nil {
			log.Printf("failed to read from source: %v\n", err)
			continue
		}

		log.Printf("Received %d bytes from %s", n, addr.String())

		// Forward the RTP packet to the destination
		_, err = destConn1.Write(buffer[:n])
		if err != nil {
			log.Printf("failed to forward packet to destination: %v\n", err)
			continue
		}
	}
}

func main() {
	//if len(os.Args) != 4 {
	//	fmt.Println("Usage: rtp-proxy <srcAddr>:<srcPort> <destAddr1>:<destPort1> <destAddr2>:<destPort2>")
	//	os.Exit(1)
	//}

	//srcAddr := os.Args[1]
	//destAddr1 := os.Args[2]
	//destAddr2 := os.Args[3]

	// Start forwarding RTP packets to two destinations
	//err := ForwardRTP(srcAddr, destAddr1, destAddr2)
	err := ForwardRTP2Dest("127.0.0.1:5000", "52.212.107.49:5000", "10.250.195.18:6234")
	//err := workingForwardRTP("127.0.0.1:5000", "52.212.107.49:5000")
	//err := ForwardRTP("127.0.0.1:5000", "52.212.107.49:5000")
	if err != nil {
		log.Fatalf("Error starting RTP proxy: %v\n", err)
	}

	// Keep the proxy running indefinitely
	for {
		time.Sleep(time.Second * 10)
	}
}
