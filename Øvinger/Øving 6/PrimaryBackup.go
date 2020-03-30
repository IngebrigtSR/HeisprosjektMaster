package main

import (
	"encoding/binary"
	. "fmt"
	"net"
	"os/exec"
	"time"
)

var count uint16
var port = "3000" //endre til noe som gir mening?
var buf = make([]byte, 16)

func Backup() {

	//runs a parallell main as backup
	(exec.Command("go run PrimaryBackup.go")).Run() //no clue if this works

	Println("Backup running")
}

func main() {
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:"+port)
	backupConn, _ := net.ListenUDP("udp", addr)
	isPrimary := false

	Println("Backup running")

	// backup will be stuck here until promoted
	for !(isPrimary) {
		n, _, err := backupConn.ReadFromUDP(buf)
		//readfromUDP fails: primary down, promote this backup to primary
		if err != nil {
			isPrimary = true
			//else keep backup counter up to date
		} else {
			count = binary.BigEndian.Uint16(buf[:n])
		}
	}
	backupConn.Close()

	// Backup promoted, new primary spawns new backup
	Println("Backup promoted xD")
	Backup()

	// counting/broadcast to backup loop
	broadcastConn, _ := net.DialUDP("udp", nil, addr)
	for {
		Println(count)
		count++
		binary.BigEndian.PutUint16(buf, count)
		_, _ = broadcastConn.Write(buf)
		time.Sleep(100 * time.Millisecond)
	}
}
