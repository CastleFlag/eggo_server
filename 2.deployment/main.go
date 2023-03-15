package main

import (
	"log"
	"net"
	"net/http"
)

func getIP() string {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func echoHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("request")
	w.Write([]byte("hello wolrd " + getIP() + "\n"))
}

func main() {
	http.HandleFunc("/", echoHandler)
	http.ListenAndServe(":8000", nil)
}
