package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
)

/*
http://www.supermicro.com/support/faqs/faq.cfm?faq=12600

To get UID status, please issue: ipmitool raw 0x30 0xC
Returned value: 0 = OFF; 1 = ON

To enable UID, please issue: ipmitool raw 0x30 0xD
To disable UID, please issue: ipmitool raw 0x30 0xE

If successful, the completion Code is 0x00.
*/

type Credentials struct {
	Username, Password string
}

var (
	// ErrNoPing - cannot ping address
	ErrNoPing = fmt.Errorf("cannot ping address")
	// ErrBadIPMI - IPMI command failed
	ErrBadIPMI = fmt.Errorf("IPMI command failed")
	// ErrLoginIPMI - unable to log into IPMI
	ErrLoginIPMI = fmt.Errorf("unable to log into IPMI")
	// ErrIncompleteIPMI - incomplete IPMI response
	ErrIncompleteIPMI = fmt.Errorf("incomplete IPMI response")
	// ErrExecFailed - command execution failed
	ErrExecFailed = fmt.Errorf("command execution failed")
	// ErrNoAddress - no address specified
	ErrNoAddress = fmt.Errorf("no address specified")
	// ErrNoUsername - no username specified
	ErrNoUsername = fmt.Errorf("no username specified")
	// ErrNoPassword - no password specified
	ErrNoPassword = fmt.Errorf("no password specified")

	cLock    sync.Mutex
	ipmiCred = make(map[string]Credentials)
)

func blink(ip string, on bool) error {
	cmd := "0xE"
	if on {
		cmd = "0xD"
	}
	u, p, _ := getCredentials(ip)
	rc, _, _, err := ipmicmd(ip, u, p, fmt.Sprintf("raw 0x30 %s", cmd))
	if err != nil {
		return err
	}
	if rc > 0 {
		return fmt.Errorf("ipmitool returned: %d", rc)
	}
	return nil
}

func blinkStatus(ip string) (bool, error) {
	u, p, _ := getCredentials(ip)
	rc, _, _, err := ipmicmd(ip, u, p, "raw 0x30 0xC")
	on := false
	if rc == 1 {
		on = true
	}
	return on, err
}

func ipmiexec(ip, username, password, input string) (int, string, string, error) {
	if len(ip) == 0 {
		return -1, "", "", ErrNoAddress
	}
	if len(username) == 0 {
		return -1, "", "", ErrNoUsername
	}
	if len(password) == 0 {
		return -1, "", "", ErrNoPassword
	}
	args := []string{"-Ilanplus", "-H", ip, "-U", username, "-P", password}
	//args := []string{"-H", ip, "-U", username, "-P", password}
	args = append(args, strings.Fields(input)...)
	cmd := exec.Command("ipmitool", args...)
	//cmd.Stdin = nil
	stdout, err := cmd.Output()
	//fmt.Println("OUT:", string(stdout), "ERR:", err)
	rc := 0
	stderr := ""
	if err != nil {
		switch err := err.(type) {
		case *exec.ExitError:
			stderr = string(err.Stderr)
			status := err.ProcessState.Sys().(syscall.WaitStatus)
			rc = status.ExitStatus()
		}
	}
	return rc, string(stdout), stderr, err
}

func ipmicmd(ip, username, password, input string) (int, string, string, error) {
	if len(ip) == 0 {
		return -1, "", "", ErrNoAddress
	}
	if !ping(ip, pingTimeout) {
		log.Printf("ping failed for: %s (%d)\n", ip, pingTimeout)
		return -1, "", "", ErrNoPing
	}
	return ipmiexec(ip, username, password, input)
}

func ipmichk(ip, username, password string) (string, error) {
	const chkcmd = "mc info"
	rc, stdout, stderr, err := ipmiexec(ip, username, password, chkcmd)
	log.Printf("(%s,%s) ipmichk rc:%d stdout:%s stderr:%s\n", username, password, rc, stdout, stderr)
	if err != nil {
		return "", err
	}
	if rc > 0 {
		return "", ErrExecFailed
	}
	log.Println("ipmichk stdout:", stdout)
	if strings.Contains(stdout, "Manufacturer") {
		i := strings.Index(stdout, ":") + 2
		return stdout[i:], nil
	}
	if len(stdout) > 0 {
		log.Println("unexpected stdout:", stdout)
	}
	if len(stderr) > 0 {
		log.Println("unexpected stderr:", stderr)
	}
	return "", ErrBadIPMI
}

// find the correct credentials
func fixCredentials(ip string) (string, string, error) {
	log.Println("fix credentials for:", ip)
	if !ping(ip, pingTimeout) {
		return "", "", ErrNoPing
	}
	// is this a Dell?
	u := "root"
	p := "calvin"
	if _, err := ipmichk(ip, u, p); err == nil {
		setCredentials(ip, u, p)
		return u, p, nil
	}
	versions := []string{"ADMIN", "Admin", "admin"}
	for _, u = range versions {
		for _, p = range versions {
			if _, err := ipmichk(ip, u, p); err == nil {
				setCredentials(ip, u, p)
				return u, p, nil
			}
		}
	}
	return "", "", ErrLoginIPMI
}

func findMAC(ipmi string) (string, error) {
	const cmd = "raw 0x30 0x21" // supermicro specific
	u, p, err := getCredentials(ipmi)
	if err != nil {
		fmt.Println("ERR:", err)
		return "", err
	}
	// TODO: fix this ugly hack
	if u == "root" && p == "calvin" {
		return dellMAC(u, p, ipmi)
	}
	rc, stdout, _, err := ipmicmd(ipmi, u, p, cmd)
	if err != nil {
		return "", err
	}
	if rc != 0 {
		return "", err
	}
	if len(stdout) < 13 {
		return "", ErrIncompleteIPMI
	}
	lines := strings.Split(stdout, "\n")
	if len(lines) > 1 {
		stdout = lines[2]
	}
	return strings.Replace(stdout[13:], " ", ":", -1), nil
}

// dell is different than supermicro, go figure
func dellMAC(u, p, ipmi string) (string, error) {
	const cmd = "delloem mac"
	rc, stdout, _, err := ipmicmd(ipmi, u, p, cmd)
	if err != nil {
		return "", err
	}
	if rc != 0 {
		return "", err
	}
	lines := strings.Split(stdout, "\n")
	macs := make([]string, 0, len(lines))
	re, err := regexp.Compile("^[0-9].*")
	if err != nil {
		panic(err)
	}
	for _, line := range lines {
		if re.Match([]byte(line)) {
			f := strings.Fields(line)
			macs = append(macs, f[1])
		}
	}
	if len(macs) > 0 {
		sort.Strings(macs)
		return macs[0], nil
	}
	return "", fmt.Errorf("no MAC found")
}

func getCredentials(ipmi string) (string, string, error) {
	cLock.Lock()
	creds, ok := ipmiCred[ipmi]
	cLock.Unlock()
	if ok {
		return creds.Username, creds.Password, nil
	}
	u, p, err := fixCredentials(ipmi)
	if err != nil {
		return "", "", err
	}
	return u, p, nil
}

func setCredentials(ipmi, username, password string) {
	cLock.Lock()
	ipmiCred[ipmi] = Credentials{username, password}
	cLock.Unlock()
}
