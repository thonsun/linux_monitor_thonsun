package main

import (
	"fmt"
	"github.com/elastic/go-libaudit"
	"github.com/pkg/errors"
	"log"
	"os"
	"syscall"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
func main() {
	netlink, err := libaudit.NewNetlinkClient(syscall.NETLINK_CONNECTOR, 1, nil, os.Stdout)
	if err != nil {
		switch err {
		case syscall.EINVAL, syscall.EPROTONOSUPPORT, syscall.EAFNOSUPPORT:
			log.Println(err, "not supported by kernel")
		default:
			log.Println(err, "failed to open netlink socket")
		}
	}
	for{
		msg, err := netlink.Receive(false,syscall.ParseNetlinkMessage)
		if err != nil {
			log.Println(err)
		}
		fmt.Printf("%+v",msg)
	}
}
