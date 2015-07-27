package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	pp "github.com/paulstuart/ping"
	"github.com/paulstuart/sshclient"
)

/*
http://www.supermicro.com/support/faqs/faq.cfm?faq=12600

To get UID status, please issue: ipmitool raw 0x30 0xC
Returned value: 0 = OFF; 1 = ON

To enable UID, please issue: ipmitool raw 0x30 0xD
To disable UID, please issue: ipmitool raw 0x30 0xE

If successful, the completion Code is 0x00.
*/

var (
	pingTimeout = 3
)

func ping(ip string, timeout int) bool {
	return pp.Ping(ip, timeout)
}

type pingable struct {
	IP string
	OK bool
}

func bulkPing(timeout int, ips ...string) map[string]bool {
	hits := make(map[string]bool)
	c := make(chan pingable)

	for _, ip := range ips {
		go func(addr string) {
			ok := ping(addr, timeout)
			c <- pingable{addr, ok}
		}(ip)
	}
	for range ips {
		r := <-c
		hits[r.IP] = r.OK
	}
	return hits
}

func Blink(ip string, on bool) error {
	cmd := "0xE"
	if on {
		cmd = "0xD"
	}
	rc, _, _, err := ipmicmd(ip, fmt.Sprintf("raw 0x30 %s", cmd))
	if err != nil {
		return err
	}
	if rc > 0 {
		return fmt.Errorf("ipmitool returned: %d", rc)
	}
	return nil
}

func BlinkStatus(ip string) (bool, error) {
	rc, _, _, err := ipmicmd(ip, "raw 0x30 0xC")
	on := false
	if rc == 1 {
		on = true
	}
	return on, err
}

func ipmicmd(ip, input string) (int, string, string, error) {
	if !ping(ip, pingTimeout) {
		return -1, "", "", fmt.Errorf("Cannot ping address: %s", ip)
	}
	args := strings.Fields(input)
	fmt.Println("ARGS:", args)
	cmd := exec.Command("ipmitool", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	rc := status.ExitStatus()

	return rc, stdout.String(), stderr.String(), err
}

func Remote(server, cmd string, timeout int) (rc int, stdout, stderr string, err error) {
	return sshclient.Exec(server+":22", cfg.SSH.Username, cfg.SSH.Password, cmd, timeout)
}
