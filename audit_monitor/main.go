package main

import (
	"fmt"
	"github.com/elastic/go-libaudit/auparse"
	"github.com/pkg/errors"
	"io"
	"log"
	"log/syslog"
	"os"
	"github.com/elastic/go-libaudit"
)

const debug bool = true

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout,"[+]DEBUG[+]: ",log.Lshortfile | log.LstdFlags)
}

func infolog(format string,info...interface{}) {
	if debug {
		logger.Printf(format,info...)
	}
}

func main() {
	if err := read(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func read() error {
	if os.Geteuid() != 0 {
		return errors.New("you must be root to receive audit data")
	}

	var err error
	var client *libaudit.AuditClient

	client, err = libaudit.NewAuditClient(os.Stdout)
	if err != nil {
		return errors.Wrap(err, "failed to create audit client")
	}
	defer client.Close()

	status, err := client.GetStatus()
	if err != nil {
		return errors.Wrap(err, "failed to get audit status")
	}
	infolog("received audit status=%+v", status)

	if status.Enabled == 0 {
		infolog("enabling auditing in the kernel")
		if err = client.SetEnabled(true, libaudit.WaitForReply); err != nil {
			return errors.Wrap(err, "failed to set enabled=true")
		}
	}


	infolog("sending message to kernel registering our PID (%v) as the audit daemon", os.Getpid())
	if err = client.SetPID(libaudit.NoWait); err != nil {
		return errors.Wrap(err, "failed to set audit PID")
	}

	return receive(client)
}

func receive(r *libaudit.AuditClient) error {
	for {
		rawEvent, err := r.Receive(false)
		if err != nil {
			return errors.Wrap(err, "receive failed")
		}

		// Messages from 1300-2999 are valid audit messages.
		if rawEvent.Type < auparse.AUDIT_USER_AUTH ||
			rawEvent.Type > auparse.AUDIT_LAST_USER_MSG2 {
			continue
		}

		fmt.Printf("type=%v msg=%v\n", rawEvent.Type, string(rawEvent.Data))
	}
}

