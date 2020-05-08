//+build linux

package main

import (
	"log"
	"syscall"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}
func main() {
	fd, err := syscall.Socket(syscall.AF_NETLINK,syscall.SOCK_DGRAM,syscall.NETLINK_CONNECTOR)
	if err != nil {
		log.Printf("open socket fd failed:%+v",err)
	}
	lca := syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Pid:uint32(syscall.Getpid()),
		Groups: 1,
	}

	if err:= syscall.Bind(fd,&lca);err != nil{
		log.Printf("bind to addr failed:%+v",err)
		syscall.Close(fd)
	}




}
