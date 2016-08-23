package main

import (
	"bufio"
	"io"
	"reflect"
	"regexp"
	"strings"
)

var (
	alpha = regexp.MustCompile("^[A-Za-z]+")
	skip  = regexp.MustCompile("^([0-9]+|#|SMBIOS|Table)")
	ws    = regexp.MustCompile("[\t ]+")
)

type memoryDevice struct {
	Size                 string
	Locator              string
	Speed                string
	manufacturer         string
	SerialNumber         string
	AssetTag             string
	PartNumber           string
	ConfiguredClockSpeed string
}

type baseBoardInformation struct {
	manufacturer      string
	ProductName       string
	Version           string
	SerialNumber      string
	AssetTag          string
	LocationInChassis string
	ChassisHandle     string
	Type              string
}

type systemPowerSupply struct {
	Location        string
	manufacturer    string
	SerialNumber    string
	AssetTag        string
	ModelPartNumber string
}

type systemInfo struct {
	Motherboard baseBoardInformation
	Memory      []*memoryDevice
	Power       []*systemPowerSupply
}

func (sys *systemInfo) AddLine(part interface{}, line string) {
	colon := strings.Index(line, ":")
	if colon < 0 {
		return
	}
	label := ws.ReplaceAllString(line[:colon], "")
	colon++
	value := strings.TrimSpace(line[colon:])
	p := reflect.ValueOf(part)
	val := reflect.Indirect(p)
	f := val.FieldByName(label)
	if f.IsValid() {
		f.SetString(value)
	}
}

func parseDMI(src io.Reader) *systemInfo {
	sys := new(systemInfo)
	reader := bufio.NewReader(src)
	var part interface{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		if skip.MatchString(line) {
			continue
		}
		if alpha.MatchString(line) {
			line = strings.TrimSpace(line)
			switch {
			case line == "Memory Device":
				m := new(memoryDevice)
				sys.Memory = append(sys.Memory, m)
				part = m
			case line == "System Power Supply":
				p := new(systemPowerSupply)
				sys.Power = append(sys.Power, p)
				part = p
			case line == "Base Board Information":
				part = &sys.Motherboard
			default:
				part = nil
			}
			continue
		}
		if part != nil {
			sys.AddLine(part, line)
		}
	}
	return sys
}
