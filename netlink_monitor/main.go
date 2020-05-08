//+build linux

package main

/*

#include <linux/netlink.h>
#include <linux/connector.h>
#include <linux/cn_proc.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

struct __attribute__ ((aligned(NLMSG_ALIGNTO))) {
	struct nlmsghdr nl_hdr;
	struct __attribute__ ((__packed__)) {
		struct cn_msg cn_msg;
		enum proc_cn_mcast_op cn_mcast;
	};
} Nlcn_msg;

memset(&Nlcn_msg, 0, sizeof(Nlcn_msg));
Nlcn_msg.nl_hdr.nlmsg_len = sizeof(Nlcn_msg);
Nlcn_msg.nl_hdr.nlmsg_pid = getpid();
Nlcn_msg.nl_hdr.nlmsg_type = NLMSG_DONE;

Nlcn_msg.cn_msg.id.idx = CN_IDX_PROC;
Nlcn_msg.cn_msg.id.val = CN_VAL_PROC;
Nlcn_msg.cn_msg.len = sizeof(enum proc_cn_mcast_op);
Nlcn_msg.cn_mcast = PROC_CN_MCAST_LISTEN
*/
import "C"
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
	log.Printf("%+v",C.Nlcn_msg)
	////msg := []byte{}
	//msg := netlink.Message{
	//	Header: netlink.Header{},
	//	Data:   nil,
	//}
	//data :=
	//
	//syscall.Sendto(fd,&msg,0,&lca)


}
