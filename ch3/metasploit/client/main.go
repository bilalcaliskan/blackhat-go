package main

import (
	"fmt"
	"github.com/bilalcaliskan/blackhat-go/ch3/metasploit/rpc"
	"log"
)

func main() {
	host := "127.0.0.1:55552"
	pass := "UmHioSkY"
	user := "msf"

	if host == "" || pass == "" {
		log.Fatalln("Missing required environment variable MSFHOST or MSFPASS")
	}
	msf, err := rpc.New(host, user, pass)
	if err != nil {
		log.Panicln(err)
	}
	defer msf.Logout()

	sessions, err := msf.SessionList()
	if err != nil {
		log.Panicln(err)
	}
	fmt.Println("Sessions:")
	for _, session := range sessions {
		fmt.Printf("%5d  %s\n", session.ID, session.Info)
	}
}
