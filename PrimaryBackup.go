package main

import (
	"encoding/binary"
	. "fmt"
	"net"
	"os/exec"
	"time"
)

var count uint16
var buf = make([]byte, 16)

func main() {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	backupConn, _ := net.ListenUDP("udp", addr)
	isPrimary := false

	Println("Backup running")

	// backup will be stuck here until promoted
	for !(isPrimary) {
		backupConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := backupConn.ReadFromUDP(buf)
		//readfromUDP fails: primary down, promote this backup to primary
		if err != nil {
			isPrimary = true
			//else keep backup counter up to date
		} else {
			count = binary.BigEndian.Uint16(buf[:n])
			Println("Backed Up:  ", count)
		}
	}
	backupConn.Close()

	// Backup promoted,  spawning new backup
	Println("Backup promoted xD")
	(exec.Command("gnome-terminal", "-x", "sh", "-c", "go run ex6.go")).Run()

	// counting/broadcast to backup loop
	broadcastConn, _ := net.DialUDP("udp", nil, addr)
	for {
		Println(count)
		count++
		binary.BigEndian.PutUint16(buf, count)
		_, _ = broadcastConn.Write(buf)
		time.Sleep(200 * time.Millisecond)
	}
}
