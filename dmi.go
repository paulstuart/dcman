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

type MemoryDevice struct {
	Size                 string
	Locator              string
	Speed                string
	Manufacturer         string
	SerialNumber         string
	AssetTag             string
	PartNumber           string
	ConfiguredClockSpeed string
}

type BaseBoardInformation struct {
	Manufacturer      string
	ProductName       string
	Version           string
	SerialNumber      string
	AssetTag          string
	LocationInChassis string
	ChassisHandle     string
	Type              string
}

type SystemPowerSupply struct {
	Location        string
	Manufacturer    string
	SerialNumber    string
	AssetTag        string
	ModelPartNumber string
}

type System struct {
	Motherboard BaseBoardInformation
	Memory      []*MemoryDevice
	Power       []*SystemPowerSupply
}

func (sys *System) AddLine(part interface{}, line string) {
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

func ParseDMI(src io.Reader) *System {
	sys := new(System)
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
				m := new(MemoryDevice)
				sys.Memory = append(sys.Memory, m)
				part = m
			case line == "System Power Supply":
				p := new(SystemPowerSupply)
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
