package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"syscall"
)

// Constants from <linux/connector.h>
const (
	CN_IDX_PROC             = 0x1
	CN_VAL_PROC             = 0x1
	CN_IDX_CIFS             = 0x2
	CN_VAL_CIFS             = 0x1
	CN_W1_IDX               = 0x3 /* w1 communication */
	CN_W1_VAL               = 0x1
	CN_IDX_V86D             = 0x4
	CN_VAL_V86D_UVESAFB     = 0x1
	CN_IDX_BB               = 0x5 /* BlackBoard, from the TSP GPL sampling framework */
	CN_DST_IDX              = 0x6
	CN_DST_VAL              = 0x1
	CN_IDX_DM               = 0x7 /* Device Mapper */
	CN_VAL_DM_USERSPACE_LOG = 0x1
	CN_IDX_DRBD             = 0x8
	CN_VAL_DRBD             = 0x1
	CN_KVP_IDX              = 0x9 /* HyperV KVP */
)

type ConnMsgid struct {
	Idx uint32
	Val uint32
}

type ConnMsghr struct {
	Id ConnMsgid
	Seq uint32
	Ack uint32
	Len uint32
}

type ConnMessage struct {
	Header syscall.NlMsghdr
	ConnHdr ConnMsghr
	Data []byte
}

// Constants from <linux/cn_proc.h>

const (
	PROC_CN_MCAST_LISTEN = 1
	PROC_CN_MCAST_IGNORE = 2

	PROC_EVENT_NONE = 0
	PROC_EVENT_FORK = 1
	PROC_EVENT_EXEC = 1 << 1
	PROC_EVENT_UID  = 1 << 2
	PROC_EVENT_GID  = 1 << 6
	PROC_EVENT_SID  = 1 << 7
	PROC_EVENT_EXIT = 1 << 31
)

type KernelPID uint32

type ProcEventHdr struct {
	What          uint32
	Cpu           uint32
	TimeStampNano uint64
}

type ProcEventAck struct {
	Header ProcEventHdr
	Err    uint32
}

type ProcEventFork struct {
	Header     ProcEventHdr
	ParentPid  KernelPID // Task ID
	ParentTGid KernelPID // Process ID
	ChildPid   KernelPID
	ChildTGid  KernelPID
}

type ProcEventExec struct {
	Header     ProcEventHdr
	ParentPid  KernelPID
	ParentTGid KernelPID
}

type ProcEventId struct {
	Header      ProcEventHdr
	ParentPid   KernelPID
	ParentTGid  KernelPID
	RealID      uint32 // Task UID or GID
	EffectiveID uint32 // EUID or EGID
}

type ProcEventSid struct {
	Header     ProcEventHdr
	ParentPid  KernelPID
	ParentTGid KernelPID
}

type ProcEventExit struct {
	Header     ProcEventHdr
	ParentPid  KernelPID
	ParentTGid KernelPID
	ExitCode   uint32
	ExitSignal uint32
}

func (msg *ConnMessage) toRawMsg() (rawmsg syscall.NetlinkMessage) {
	rawmsg.Header = msg.Header
	msg.ConnHdr.Len = uint32(len(msg.Data))
	w := bytes.NewBuffer([]byte{})
	binary.Write(w,SystemEndianness,msg.ConnHdr)
	w.Write(msg.Data)
	rawmsg.Data = w.Bytes()
	return
}

func ParseConnMessage(msg syscall.NetlinkMessage) (ParsedNetlinkMessage, error) {
	switch msg.Header.Type {
	case syscall.NLMSG_ERROR:
		return ParseErrorMessage(msg), nil
	}
	var cn_msg ConnMessage
	cn_msg.Header = msg.Header
	r := bytes.NewBuffer(msg.Data)
	binary.Read(r,SystemEndianness,&cn_msg.ConnHdr)
	cn_msg.Data = r.Bytes()[:cn_msg.ConnHdr.Len]
	return cn_msg,nil
}

func MakeProcConnectorMsg() ConnMessage {
	var msg ConnMessage
	msg.Header.Type = syscall.NLMSG_DONE
	msg.Header.Flags = 0
	msg.Header.Pid = uint32(os.Getpid())
	msg.ConnHdr.Id =ConnMsgid{Idx: CN_IDX_PROC, Val: CN_VAL_PROC}
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, int32(PROC_CN_MCAST_LISTEN))
	msg.Data = buf.Bytes()
	return msg
}

func main()  {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	s, err := DialNetlink("conn",CN_IDX_PROC)
	if err != nil{
		log.Println("create proc conn err:",err)
		return
	}
	msg := MakeProcConnectorMsg()
	WriteMessage(s,&msg)
	for {
		resp, err := ReadMessage(s)
		if err != nil {
			log.Println("msg error:",err)
			break
		}

		cnmsg, er := ParseConnMessage(resp)
		if er != nil {
			fmt.Println(er)
			break
		}

		msg_s, er := json.MarshalIndent(cnmsg, "", "  ")
		fmt.Printf("[+]msg:%s\n", msg_s)
		if resp.Header.Type == syscall.NLMSG_DONE {
			return
		}
	}
}

