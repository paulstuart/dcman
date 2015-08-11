// generated by dbgen ; DO NOT EDIT

package main

import (
	"time"
)

//
// Contract DBObject interface functions
//
func (o *Contract) InsertValues() []interface{} {
	return []interface{}{o.VID, o.Policy, o.Phone}
}

func (o *Contract) UpdateValues() []interface{} {
	return []interface{}{o.VID, o.Policy, o.Phone, o.CID}
}

func (o *Contract) MemberPointers() []interface{} {
	return []interface{}{&o.CID, &o.VID, &o.Policy, &o.Phone}
}

func (o *Contract) Key() int64 {
	return o.CID
}

func (o *Contract) SetID(id int64) {
	o.CID = id
}

func (o *Contract) TableName() string {
	return "contracts"
}

func (o *Contract) SelectFields() string {
	return "cid,vid,policy,phone"
}

func (o *Contract) InsertFields() string {
	return "cid,vid,policy,phone"
}

func (o *Contract) KeyField() string {
	return "cid"
}

func (o *Contract) KeyName() string {
	return "CID"
}

func (o *Contract) ModifiedBy(user int64, t time.Time) {
}

//
// Device DBObject interface functions
//
func (o *Device) InsertValues() []interface{} {
	return []interface{}{o.AssetTag, o.Modified, o.RID, o.RU, o.SerialNo, o.UID, o.VID, o.Type, o.MgmtIP, o.PrimaryMac, o.Hostname, o.Model, o.Note, o.Height, o.PrimaryIP, o.MgmtMac}
}

func (o *Device) UpdateValues() []interface{} {
	return []interface{}{o.AssetTag, o.Modified, o.RID, o.RU, o.SerialNo, o.UID, o.VID, o.Type, o.MgmtIP, o.PrimaryMac, o.Hostname, o.Model, o.Note, o.Height, o.PrimaryIP, o.MgmtMac, o.DID}
}

func (o *Device) MemberPointers() []interface{} {
	return []interface{}{&o.DID, &o.AssetTag, &o.Modified, &o.RID, &o.RU, &o.SerialNo, &o.UID, &o.VID, &o.Type, &o.MgmtIP, &o.PrimaryMac, &o.Hostname, &o.Model, &o.Note, &o.Height, &o.PrimaryIP, &o.MgmtMac}
}

func (o *Device) Key() int64 {
	return o.DID
}

func (o *Device) SetID(id int64) {
	o.DID = id
}

func (o *Device) TableName() string {
	return "devices"
}

func (o *Device) SelectFields() string {
	return "did,asset_tag,modified,rid,ru,sn,uid,vid,device_type,mgmt_ip,primary_mac,hostname,model,note,height,primary_ip,mgmt_mac"
}

func (o *Device) InsertFields() string {
	return "did,asset_tag,modified,rid,ru,sn,uid,vid,device_type,mgmt_ip,primary_mac,hostname,model,note,height,primary_ip,mgmt_mac"
}

func (o *Device) KeyField() string {
	return "did"
}

func (o *Device) KeyName() string {
	return "DID"
}

func (o *Device) ModifiedBy(user int64, t time.Time) {
}

//
// Port DBObject interface functions
//
func (o *Port) InsertValues() []interface{} {
	return []interface{}{o.DID, o.PortType, o.MAC, o.CableTag, o.SwitchPort, o.Modified, o.UID}
}

func (o *Port) UpdateValues() []interface{} {
	return []interface{}{o.DID, o.PortType, o.MAC, o.CableTag, o.SwitchPort, o.Modified, o.UID, o.PID}
}

func (o *Port) MemberPointers() []interface{} {
	return []interface{}{&o.PID, &o.DID, &o.PortType, &o.MAC, &o.CableTag, &o.SwitchPort, &o.Modified, &o.UID}
}

func (o *Port) Key() int64 {
	return o.PID
}

func (o *Port) SetID(id int64) {
	o.PID = id
}

func (o *Port) TableName() string {
	return "ports"
}

func (o *Port) SelectFields() string {
	return "pid,did,port_type,mac,cable_tag,switch_port,modified,uid"
}

func (o *Port) InsertFields() string {
	return "pid,did,port_type,mac,cable_tag,switch_port,modified,uid"
}

func (o *Port) KeyField() string {
	return "pid"
}

func (o *Port) KeyName() string {
	return "PID"
}

func (o *Port) ModifiedBy(user int64, t time.Time) {
}

//
// IP DBObject interface functions
//
func (o *IP) InsertValues() []interface{} {
	return []interface{}{o.DID, o.Type, o.Int, o.RemoteAddr, o.Modified, o.UID}
}

func (o *IP) UpdateValues() []interface{} {
	return []interface{}{o.DID, o.Type, o.Int, o.RemoteAddr, o.Modified, o.UID, o.IID}
}

func (o *IP) MemberPointers() []interface{} {
	return []interface{}{&o.IID, &o.DID, &o.Type, &o.Int, &o.RemoteAddr, &o.Modified, &o.UID}
}

func (o *IP) Key() int64 {
	return o.IID
}

func (o *IP) SetID(id int64) {
	o.IID = id
}

func (o *IP) TableName() string {
	return "ips"
}

func (o *IP) SelectFields() string {
	return "iid,did,ip_type,ip_int,remote_addr,modified,uid"
}

func (o *IP) InsertFields() string {
	return "iid,did,ip_type,ip_int,remote_addr,modified,uid"
}

func (o *IP) KeyField() string {
	return "iid"
}

func (o *IP) KeyName() string {
	return "IID"
}

func (o *IP) ModifiedBy(user int64, t time.Time) {
}

//
// User DBObject interface functions
//
func (o *User) InsertValues() []interface{} {
	return []interface{}{o.Last, o.Email, o.Level, o.Login, o.First}
}

func (o *User) UpdateValues() []interface{} {
	return []interface{}{o.Last, o.Email, o.Level, o.Login, o.First, o.ID}
}

func (o *User) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Last, &o.Email, &o.Level, &o.Login, &o.First}
}

func (o *User) Key() int64 {
	return o.ID
}

func (o *User) SetID(id int64) {
	o.ID = id
}

func (o *User) TableName() string {
	return "users"
}

func (o *User) SelectFields() string {
	return "id,lastname,email,admin,login,firstname"
}

func (o *User) InsertFields() string {
	return "id,lastname,email,admin,login,firstname"
}

func (o *User) KeyField() string {
	return "id"
}

func (o *User) KeyName() string {
	return "ID"
}

func (o *User) ModifiedBy(user int64, t time.Time) {
}

//
// Document DBObject interface functions
//
func (o *Document) InsertValues() []interface{} {
	return []interface{}{o.UID, o.DID, o.Filename, o.Title, o.RemoteAddr}
}

func (o *Document) UpdateValues() []interface{} {
	return []interface{}{o.UID, o.DID, o.Filename, o.Title, o.RemoteAddr, o.ID}
}

func (o *Document) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.UID, &o.DID, &o.Filename, &o.Title, &o.RemoteAddr}
}

func (o *Document) Key() int64 {
	return o.ID
}

func (o *Document) SetID(id int64) {
	o.ID = id
}

func (o *Document) TableName() string {
	return "documents"
}

func (o *Document) SelectFields() string {
	return "id,user_id,did,filename,title,remote_addr"
}

func (o *Document) InsertFields() string {
	return "id,user_id,did,filename,title,remote_addr"
}

func (o *Document) KeyField() string {
	return "id"
}

func (o *Document) KeyName() string {
	return "ID"
}

func (o *Document) ModifiedBy(user int64, t time.Time) {
}

//
// Vendor DBObject interface functions
//
func (o *Vendor) InsertValues() []interface{} {
	return []interface{}{o.Country, o.Postal, o.Note, o.RemoteAddr, o.Name, o.WWW, o.Address, o.State, o.Phone, o.City, o.UID, o.Modified}
}

func (o *Vendor) UpdateValues() []interface{} {
	return []interface{}{o.Country, o.Postal, o.Note, o.RemoteAddr, o.Name, o.WWW, o.Address, o.State, o.Phone, o.City, o.UID, o.Modified, o.VID}
}

func (o *Vendor) MemberPointers() []interface{} {
	return []interface{}{&o.VID, &o.Country, &o.Postal, &o.Note, &o.RemoteAddr, &o.Name, &o.WWW, &o.Address, &o.State, &o.Phone, &o.City, &o.UID, &o.Modified}
}

func (o *Vendor) Key() int64 {
	return o.VID
}

func (o *Vendor) SetID(id int64) {
	o.VID = id
}

func (o *Vendor) TableName() string {
	return "vendors"
}

func (o *Vendor) SelectFields() string {
	return "vid,country,postal,note,remote_addr,name,www,address,state,phone,city,user_id,modified"
}

func (o *Vendor) InsertFields() string {
	return "vid,country,postal,note,remote_addr,name,www,address,state,phone,city,user_id,modified"
}

func (o *Vendor) KeyField() string {
	return "vid"
}

func (o *Vendor) KeyName() string {
	return "VID"
}

func (o *Vendor) ModifiedBy(user int64, t time.Time) {
	o.UID = user
	o.Modified = t
}

//
// RMA DBObject interface functions
//
func (o *RMA) InsertValues() []interface{} {
	return []interface{}{o.Description, o.NewSN, o.Ticket, o.Received, o.Replaced, o.Number, o.Tracking, o.Opened, o.Sent, o.SID, o.VID, o.UID, o.Part, o.OldSN, o.Jira}
}

func (o *RMA) UpdateValues() []interface{} {
	return []interface{}{o.Description, o.NewSN, o.Ticket, o.Received, o.Replaced, o.Number, o.Tracking, o.Opened, o.Sent, o.SID, o.VID, o.UID, o.Part, o.OldSN, o.Jira, o.ID}
}

func (o *RMA) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Description, &o.NewSN, &o.Ticket, &o.Received, &o.Replaced, &o.Number, &o.Tracking, &o.Opened, &o.Sent, &o.SID, &o.VID, &o.UID, &o.Part, &o.OldSN, &o.Jira}
}

func (o *RMA) Key() int64 {
	return o.ID
}

func (o *RMA) SetID(id int64) {
	o.ID = id
}

func (o *RMA) TableName() string {
	return "rmas"
}

func (o *RMA) SelectFields() string {
	return "id,description,new_sn,dc_ticket,date_received,date_replaced,rma_no,tracking_no,date_opened,date_sent,sid,vid,user_id,part_no,old_sn,jira"
}

func (o *RMA) InsertFields() string {
	return "id,description,new_sn,dc_ticket,date_received,date_replaced,rma_no,tracking_no,date_opened,date_sent,sid,vid,user_id,part_no,old_sn,jira"
}

func (o *RMA) KeyField() string {
	return "id"
}

func (o *RMA) KeyName() string {
	return "ID"
}

func (o *RMA) ModifiedBy(user int64, t time.Time) {
}

//
// Manufacturer DBObject interface functions
//
func (o *Manufacturer) InsertValues() []interface{} {
	return []interface{}{o.Name, o.URL, o.UID, o.Modified}
}

func (o *Manufacturer) UpdateValues() []interface{} {
	return []interface{}{o.Name, o.URL, o.UID, o.Modified, o.MID}
}

func (o *Manufacturer) MemberPointers() []interface{} {
	return []interface{}{&o.MID, &o.Name, &o.URL, &o.UID, &o.Modified}
}

func (o *Manufacturer) Key() int64 {
	return o.MID
}

func (o *Manufacturer) SetID(id int64) {
	o.MID = id
}

func (o *Manufacturer) TableName() string {
	return "mfgr"
}

func (o *Manufacturer) SelectFields() string {
	return "mid,name,url,user_id,modified"
}

func (o *Manufacturer) InsertFields() string {
	return "mid,name,url,user_id,modified"
}

func (o *Manufacturer) KeyField() string {
	return "mid"
}

func (o *Manufacturer) KeyName() string {
	return "MID"
}

func (o *Manufacturer) ModifiedBy(user int64, t time.Time) {
	o.UID = user
	o.Modified = t
}

//
// Part DBObject interface functions
//
func (o *Part) InsertValues() []interface{} {
	return []interface{}{o.Modified, o.MID, o.Description, o.PartNo, o.UID}
}

func (o *Part) UpdateValues() []interface{} {
	return []interface{}{o.Modified, o.MID, o.Description, o.PartNo, o.UID, o.PID}
}

func (o *Part) MemberPointers() []interface{} {
	return []interface{}{&o.PID, &o.Modified, &o.MID, &o.Description, &o.PartNo, &o.UID}
}

func (o *Part) Key() int64 {
	return o.PID
}

func (o *Part) SetID(id int64) {
	o.PID = id
}

func (o *Part) TableName() string {
	return "parts"
}

func (o *Part) SelectFields() string {
	return "pid,modified,mid,description,part_no,user_id"
}

func (o *Part) InsertFields() string {
	return "pid,modified,mid,description,part_no,user_id"
}

func (o *Part) KeyField() string {
	return "pid"
}

func (o *Part) KeyName() string {
	return "PID"
}

func (o *Part) ModifiedBy(user int64, t time.Time) {
	o.UID = user
	o.Modified = t
}

//
// Stock DBObject interface functions
//
func (o *Stock) InsertValues() []interface{} {
	return []interface{}{o.DID, o.PID, o.VID, o.SN, o.Amount, o.UID, o.Modified}
}

func (o *Stock) UpdateValues() []interface{} {
	return []interface{}{o.DID, o.PID, o.VID, o.SN, o.Amount, o.UID, o.Modified, o.KID}
}

func (o *Stock) MemberPointers() []interface{} {
	return []interface{}{&o.KID, &o.DID, &o.PID, &o.VID, &o.SN, &o.Amount, &o.UID, &o.Modified}
}

func (o *Stock) Key() int64 {
	return o.KID
}

func (o *Stock) SetID(id int64) {
	o.KID = id
}

func (o *Stock) TableName() string {
	return "stock"
}

func (o *Stock) SelectFields() string {
	return "kid,did,pid,vid,sn,amount,user_id,modified"
}

func (o *Stock) InsertFields() string {
	return "kid,did,pid,vid,sn,amount,user_id,modified"
}

func (o *Stock) KeyField() string {
	return "kid"
}

func (o *Stock) KeyName() string {
	return "KID"
}

func (o *Stock) ModifiedBy(user int64, t time.Time) {
	o.UID = user
	o.Modified = t
}

//
// Datacenter DBObject interface functions
//
func (o *Datacenter) InsertValues() []interface{} {
	return []interface{}{o.DCMan, o.PXEUser, o.RemoteAddr, o.City, o.PXEHost, o.Name, o.State, o.PXEPass, o.PXEKey, o.UID, o.Modified, o.Address, o.Phone, o.Web}
}

func (o *Datacenter) UpdateValues() []interface{} {
	return []interface{}{o.DCMan, o.PXEUser, o.RemoteAddr, o.City, o.PXEHost, o.Name, o.State, o.PXEPass, o.PXEKey, o.UID, o.Modified, o.Address, o.Phone, o.Web, o.ID}
}

func (o *Datacenter) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.DCMan, &o.PXEUser, &o.RemoteAddr, &o.City, &o.PXEHost, &o.Name, &o.State, &o.PXEPass, &o.PXEKey, &o.UID, &o.Modified, &o.Address, &o.Phone, &o.Web}
}

func (o *Datacenter) Key() int64 {
	return o.ID
}

func (o *Datacenter) SetID(id int64) {
	o.ID = id
}

func (o *Datacenter) TableName() string {
	return "datacenters"
}

func (o *Datacenter) SelectFields() string {
	return "id,dcman,pxeuser,remote_addr,city,pxehost,name,state,pxepass,pxekey,user_id,modified,address,phone,web"
}

func (o *Datacenter) InsertFields() string {
	return "id,dcman,pxeuser,remote_addr,city,pxehost,name,state,pxepass,pxekey,user_id,modified,address,phone,web"
}

func (o *Datacenter) KeyField() string {
	return "id"
}

func (o *Datacenter) KeyName() string {
	return "ID"
}

func (o *Datacenter) ModifiedBy(user int64, t time.Time) {
	o.UID = user
	o.Modified = t
}

//
// DCView DBObject interface functions
//
func (o *DCView) InsertValues() []interface{} {
	return []interface{}{o.Hostname, o.AssetNumber, o.CPU, o.CPU_Speed, o.MemoryMB, o.Created, o.DID}
}

func (o *DCView) UpdateValues() []interface{} {
	return []interface{}{o.Hostname, o.AssetNumber, o.CPU, o.CPU_Speed, o.MemoryMB, o.Created, o.DID, o.ID}
}

func (o *DCView) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Hostname, &o.AssetNumber, &o.CPU, &o.CPU_Speed, &o.MemoryMB, &o.Created, &o.DID}
}

func (o *DCView) Key() int64 {
	return o.ID
}

func (o *DCView) SetID(id int64) {
	o.ID = id
}

func (o *DCView) TableName() string {
	return "dcview"
}

func (o *DCView) SelectFields() string {
	return "id,hostname,asset_number,cpu_id,cpu_speed,memory,created,datacenter"
}

func (o *DCView) InsertFields() string {
	return "id,hostname,asset_number,cpu_id,cpu_speed,memory,created,datacenter"
}

func (o *DCView) KeyField() string {
	return "id"
}

func (o *DCView) KeyName() string {
	return "ID"
}

func (o *DCView) ModifiedBy(user int64, t time.Time) {
}

//
// ServerVMs DBObject interface functions
//
func (o *ServerVMs) InsertValues() []interface{} {
	return []interface{}{o.DC, o.Hostname, o.VMList, o.IDList}
}

func (o *ServerVMs) UpdateValues() []interface{} {
	return []interface{}{o.DC, o.Hostname, o.VMList, o.IDList, o.ID}
}

func (o *ServerVMs) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.DC, &o.Hostname, &o.VMList, &o.IDList}
}

func (o *ServerVMs) Key() int64 {
	return o.ID
}

func (o *ServerVMs) SetID(id int64) {
	o.ID = id
}

func (o *ServerVMs) TableName() string {
	return "servervms"
}

func (o *ServerVMs) SelectFields() string {
	return "id,dc,hostname,vms,ids"
}

func (o *ServerVMs) InsertFields() string {
	return "id,dc,hostname,vms,ids"
}

func (o *ServerVMs) KeyField() string {
	return "id"
}

func (o *ServerVMs) KeyName() string {
	return "ID"
}

func (o *ServerVMs) ModifiedBy(user int64, t time.Time) {
}

//
// RackUnit DBObject interface functions
//
func (o *RackUnit) InsertValues() []interface{} {
	return []interface{}{o.SerialNo, o.IPMI, o.Note, o.DC, o.NID, o.RID, o.RU, o.Internal, o.Rack, o.Height, o.Hostname, o.AssetTag, o.SID, o.Alias}
}

func (o *RackUnit) UpdateValues() []interface{} {
	return []interface{}{o.SerialNo, o.IPMI, o.Note, o.DC, o.NID, o.RID, o.RU, o.Internal, o.Rack, o.Height, o.Hostname, o.AssetTag, o.SID, o.Alias}
}

func (o *RackUnit) MemberPointers() []interface{} {
	return []interface{}{&o.SerialNo, &o.IPMI, &o.Note, &o.DC, &o.NID, &o.RID, &o.RU, &o.Internal, &o.Rack, &o.Height, &o.Hostname, &o.AssetTag, &o.SID, &o.Alias}
}

func (o *RackUnit) Key() int64 {
	return 0
}

func (o *RackUnit) SetID(id int64) {
}

func (o *RackUnit) TableName() string {
	return "rackunits"
}

func (o *RackUnit) SelectFields() string {
	return "sn,ipmi,note,dc,nid,rid,ru,internal,rack,height,hostname,asset_tag,sid,alias"
}

func (o *RackUnit) InsertFields() string {
	return "sn,ipmi,note,dc,nid,rid,ru,internal,rack,height,hostname,asset_tag,sid,alias"
}

func (o *RackUnit) KeyField() string {
	return ""
}

func (o *RackUnit) KeyName() string {
	return ""
}

func (o *RackUnit) ModifiedBy(user int64, t time.Time) {
}

//
// Server DBObject interface functions
//
func (o *Server) InsertValues() []interface{} {
	return []interface{}{o.IPIpmi, o.MacPort1, o.Alias, o.Assigned, o.IPPublic, o.PartNo, o.CPU, o.UID, o.RU, o.Height, o.AssetTag, o.RID, o.Hostname, o.PortEth1, o.Modified, o.IPInternal, o.Note, o.RemoteAddr, o.PortIpmi, o.CableEth0, o.MacIPMI, o.SerialNo, o.PortEth0, o.CableEth1, o.CableIpmi, o.MacPort0, o.Profile}
}

func (o *Server) UpdateValues() []interface{} {
	return []interface{}{o.IPIpmi, o.MacPort1, o.Alias, o.Assigned, o.IPPublic, o.PartNo, o.CPU, o.UID, o.RU, o.Height, o.AssetTag, o.RID, o.Hostname, o.PortEth1, o.Modified, o.IPInternal, o.Note, o.RemoteAddr, o.PortIpmi, o.CableEth0, o.MacIPMI, o.SerialNo, o.PortEth0, o.CableEth1, o.CableIpmi, o.MacPort0, o.Profile, o.ID}
}

func (o *Server) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.IPIpmi, &o.MacPort1, &o.Alias, &o.Assigned, &o.IPPublic, &o.PartNo, &o.CPU, &o.UID, &o.RU, &o.Height, &o.AssetTag, &o.RID, &o.Hostname, &o.PortEth1, &o.Modified, &o.IPInternal, &o.Note, &o.RemoteAddr, &o.PortIpmi, &o.CableEth0, &o.MacIPMI, &o.SerialNo, &o.PortEth0, &o.CableEth1, &o.CableIpmi, &o.MacPort0, &o.Profile}
}

func (o *Server) Key() int64 {
	return o.ID
}

func (o *Server) SetID(id int64) {
	o.ID = id
}

func (o *Server) TableName() string {
	return "servers"
}

func (o *Server) SelectFields() string {
	return "id,ip_ipmi,mac_eth1,alias,assigned,ip_public,vendor_sku,cpu,uid,ru,height,asset_tag,rid,hostname,port_eth1,modified,ip_internal,note,remote_addr,port_ipmi,cable_eth0,mac_ipmi,sn,port_eth0,cable_eth1,cable_ipmi,mac_eth0,profile"
}

func (o *Server) InsertFields() string {
	return "id,ip_ipmi,mac_eth1,alias,assigned,ip_public,vendor_sku,cpu,uid,ru,height,asset_tag,rid,hostname,port_eth1,modified,ip_internal,note,remote_addr,port_ipmi,cable_eth0,mac_ipmi,sn,port_eth0,cable_eth1,cable_ipmi,mac_eth0,profile"
}

func (o *Server) KeyField() string {
	return "id"
}

func (o *Server) KeyName() string {
	return "ID"
}

func (o *Server) ModifiedBy(user int64, t time.Time) {
	o.UID = user
	o.Modified = t
}

//
// Router DBObject interface functions
//
func (o *Router) InsertValues() []interface{} {
	return []interface{}{o.Hostname, o.Make, o.Model, o.UID, o.RID, o.Note, o.SerialNo, o.Height, o.RU, o.MgmtIP, o.RemoteAddr, o.AssetTag, o.PartNo, o.Modified}
}

func (o *Router) UpdateValues() []interface{} {
	return []interface{}{o.Hostname, o.Make, o.Model, o.UID, o.RID, o.Note, o.SerialNo, o.Height, o.RU, o.MgmtIP, o.RemoteAddr, o.AssetTag, o.PartNo, o.Modified, o.ID}
}

func (o *Router) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Hostname, &o.Make, &o.Model, &o.UID, &o.RID, &o.Note, &o.SerialNo, &o.Height, &o.RU, &o.MgmtIP, &o.RemoteAddr, &o.AssetTag, &o.PartNo, &o.Modified}
}

func (o *Router) Key() int64 {
	return o.ID
}

func (o *Router) SetID(id int64) {
	o.ID = id
}

func (o *Router) TableName() string {
	return "routers"
}

func (o *Router) SelectFields() string {
	return "id,hostname,make,model,uid,rid,note,sn,height,ru,ip_mgmt,remote_addr,asset_tag,sku,modified"
}

func (o *Router) InsertFields() string {
	return "id,hostname,make,model,uid,rid,note,sn,height,ru,ip_mgmt,remote_addr,asset_tag,sku,modified"
}

func (o *Router) KeyField() string {
	return "id"
}

func (o *Router) KeyName() string {
	return "ID"
}

func (o *Router) ModifiedBy(user int64, t time.Time) {
}

//
// Rack DBObject interface functions
//
func (o *Rack) InsertValues() []interface{} {
	return []interface{}{o.Label, o.VendorID, o.XPos, o.YPos, o.UID, o.TS, o.DID, o.RUs}
}

func (o *Rack) UpdateValues() []interface{} {
	return []interface{}{o.Label, o.VendorID, o.XPos, o.YPos, o.UID, o.TS, o.DID, o.RUs, o.ID}
}

func (o *Rack) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Label, &o.VendorID, &o.XPos, &o.YPos, &o.UID, &o.TS, &o.DID, &o.RUs}
}

func (o *Rack) Key() int64 {
	return o.ID
}

func (o *Rack) SetID(id int64) {
	o.ID = id
}

func (o *Rack) TableName() string {
	return "racks"
}

func (o *Rack) SelectFields() string {
	return "id,rack,vendor_id,x_pos,y_pos,uid,ts,did,rackunits"
}

func (o *Rack) InsertFields() string {
	return "id,rack,vendor_id,x_pos,y_pos,uid,ts,did,rackunits"
}

func (o *Rack) KeyField() string {
	return "id"
}

func (o *Rack) KeyName() string {
	return "ID"
}

func (o *Rack) ModifiedBy(user int64, t time.Time) {
}

//
// RackNet DBObject interface functions
//
func (o *RackNet) InsertValues() []interface{} {
	return []interface{}{o.Actual, o.RID, o.CIDR, o.MinIP, o.MaxIP, o.FirstIP, o.LastIP, o.VID, o.Subnet}
}

func (o *RackNet) UpdateValues() []interface{} {
	return []interface{}{o.Actual, o.RID, o.CIDR, o.MinIP, o.MaxIP, o.FirstIP, o.LastIP, o.VID, o.Subnet}
}

func (o *RackNet) MemberPointers() []interface{} {
	return []interface{}{&o.Actual, &o.RID, &o.CIDR, &o.MinIP, &o.MaxIP, &o.FirstIP, &o.LastIP, &o.VID, &o.Subnet}
}

func (o *RackNet) Key() int64 {
	return 0
}

func (o *RackNet) SetID(id int64) {
}

func (o *RackNet) TableName() string {
	return "racknet"
}

func (o *RackNet) SelectFields() string {
	return "actual,rid,cidr,min_ip,max_ip,first_ip,last_ip,vid,subnet"
}

func (o *RackNet) InsertFields() string {
	return "actual,rid,cidr,min_ip,max_ip,first_ip,last_ip,vid,subnet"
}

func (o *RackNet) KeyField() string {
	return ""
}

func (o *RackNet) KeyName() string {
	return ""
}

func (o *RackNet) ModifiedBy(user int64, t time.Time) {
}

//
// VM DBObject interface functions
//
func (o *VM) InsertValues() []interface{} {
	return []interface{}{o.Note, o.Private, o.Public, o.VIP, o.Profile, o.UID, o.SID, o.Hostname, o.Modified, o.RemoteAddr}
}

func (o *VM) UpdateValues() []interface{} {
	return []interface{}{o.Note, o.Private, o.Public, o.VIP, o.Profile, o.UID, o.SID, o.Hostname, o.Modified, o.RemoteAddr, o.ID}
}

func (o *VM) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Note, &o.Private, &o.Public, &o.VIP, &o.Profile, &o.UID, &o.SID, &o.Hostname, &o.Modified, &o.RemoteAddr}
}

func (o *VM) Key() int64 {
	return o.ID
}

func (o *VM) SetID(id int64) {
	o.ID = id
}

func (o *VM) TableName() string {
	return "vms"
}

func (o *VM) SelectFields() string {
	return "id,note,private,public,vip,profile,uid,sid,hostname,modified,remote_addr"
}

func (o *VM) InsertFields() string {
	return "id,note,private,public,vip,profile,uid,sid,hostname,modified,remote_addr"
}

func (o *VM) KeyField() string {
	return "id"
}

func (o *VM) KeyName() string {
	return "ID"
}

func (o *VM) ModifiedBy(user int64, t time.Time) {
}

//
// Orphan DBObject interface functions
//
func (o *Orphan) InsertValues() []interface{} {
	return []interface{}{o.DC, o.Hostname, o.Private, o.Public, o.VIP, o.Note}
}

func (o *Orphan) UpdateValues() []interface{} {
	return []interface{}{o.DC, o.Hostname, o.Private, o.Public, o.VIP, o.Note, o.ID}
}

func (o *Orphan) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.DC, &o.Hostname, &o.Private, &o.Public, &o.VIP, &o.Note}
}

func (o *Orphan) Key() int64 {
	return o.ID
}

func (o *Orphan) SetID(id int64) {
	o.ID = id
}

func (o *Orphan) TableName() string {
	return "vmbad"
}

func (o *Orphan) SelectFields() string {
	return "rowid,dc,hostname,private,public,vip,note"
}

func (o *Orphan) InsertFields() string {
	return "rowid,dc,hostname,private,public,vip,note"
}

func (o *Orphan) KeyField() string {
	return "rowid"
}

func (o *Orphan) KeyName() string {
	return "ID"
}

func (o *Orphan) ModifiedBy(user int64, t time.Time) {
}

//
// Audit DBObject interface functions
//
func (o *Audit) InsertValues() []interface{} {
	return []interface{}{o.FQDN, o.Eth0, o.Hostname, o.Asset, o.IPMI_IP, o.IPMI_MAC, o.VMs, o.IP, o.Mem, o.IPs, o.Eth1, o.SN, o.CPU, o.Kernel, o.Release}
}

func (o *Audit) UpdateValues() []interface{} {
	return []interface{}{o.FQDN, o.Eth0, o.Hostname, o.Asset, o.IPMI_IP, o.IPMI_MAC, o.VMs, o.IP, o.Mem, o.IPs, o.Eth1, o.SN, o.CPU, o.Kernel, o.Release}
}

func (o *Audit) MemberPointers() []interface{} {
	return []interface{}{&o.FQDN, &o.Eth0, &o.Hostname, &o.Asset, &o.IPMI_IP, &o.IPMI_MAC, &o.VMs, &o.IP, &o.Mem, &o.IPs, &o.Eth1, &o.SN, &o.CPU, &o.Kernel, &o.Release}
}

func (o *Audit) Key() int64 {
	return 0
}

func (o *Audit) SetID(id int64) {
}

func (o *Audit) TableName() string {
	return "auditing"
}

func (o *Audit) SelectFields() string {
	return "fqdn,eth0,hostname,asset,ipmi_ip,ipmi_mac,vms,remote_addr,mem,ips,eth1,sn,cpu,kernel,release"
}

func (o *Audit) InsertFields() string {
	return "fqdn,eth0,hostname,asset,ipmi_ip,ipmi_mac,vms,remote_addr,mem,ips,eth1,sn,cpu,kernel,release"
}

func (o *Audit) KeyField() string {
	return ""
}

func (o *Audit) KeyName() string {
	return ""
}

func (o *Audit) ModifiedBy(user int64, t time.Time) {
}

//
// PDU DBObject interface functions
//
func (o *PDU) InsertValues() []interface{} {
	return []interface{}{o.Netmask, o.Gateway, o.DNS, o.AssetTag, o.RID, o.Hostname, o.IP}
}

func (o *PDU) UpdateValues() []interface{} {
	return []interface{}{o.Netmask, o.Gateway, o.DNS, o.AssetTag, o.RID, o.Hostname, o.IP, o.ID}
}

func (o *PDU) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Netmask, &o.Gateway, &o.DNS, &o.AssetTag, &o.RID, &o.Hostname, &o.IP}
}

func (o *PDU) Key() int64 {
	return o.ID
}

func (o *PDU) SetID(id int64) {
	o.ID = id
}

func (o *PDU) TableName() string {
	return "pdus"
}

func (o *PDU) SelectFields() string {
	return "id,netmask,gateway,dns,asset_tag,rid,hostname,ip_address"
}

func (o *PDU) InsertFields() string {
	return "id,netmask,gateway,dns,asset_tag,rid,hostname,ip_address"
}

func (o *PDU) KeyField() string {
	return "id"
}

func (o *PDU) KeyName() string {
	return "ID"
}

func (o *PDU) ModifiedBy(user int64, t time.Time) {
}

//
// VProfile DBObject interface functions
//
func (o *VProfile) InsertValues() []interface{} {
	return []interface{}{o.Name}
}

func (o *VProfile) UpdateValues() []interface{} {
	return []interface{}{o.Name, o.VPID}
}

func (o *VProfile) MemberPointers() []interface{} {
	return []interface{}{&o.VPID, &o.Name}
}

func (o *VProfile) Key() int64 {
	return o.VPID
}

func (o *VProfile) SetID(id int64) {
	o.VPID = id
}

func (o *VProfile) TableName() string {
	return "vlan_profiles"
}

func (o *VProfile) SelectFields() string {
	return "vpid,name"
}

func (o *VProfile) InsertFields() string {
	return "vpid,name"
}

func (o *VProfile) KeyField() string {
	return "vpid"
}

func (o *VProfile) KeyName() string {
	return "VPID"
}

func (o *VProfile) ModifiedBy(user int64, t time.Time) {
}

//
// VLAN DBObject interface functions
//
func (o *VLAN) InsertValues() []interface{} {
	return []interface{}{o.Name, o.Profile, o.Gateway, o.Route, o.Netmask, o.DID}
}

func (o *VLAN) UpdateValues() []interface{} {
	return []interface{}{o.Name, o.Profile, o.Gateway, o.Route, o.Netmask, o.DID, o.ID}
}

func (o *VLAN) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Name, &o.Profile, &o.Gateway, &o.Route, &o.Netmask, &o.DID}
}

func (o *VLAN) Key() int64 {
	return o.ID
}

func (o *VLAN) SetID(id int64) {
	o.ID = id
}

func (o *VLAN) TableName() string {
	return "vlans"
}

func (o *VLAN) SelectFields() string {
	return "id,name,profile,gateway,route,netmask,did"
}

func (o *VLAN) InsertFields() string {
	return "id,name,profile,gateway,route,netmask,did"
}

func (o *VLAN) KeyField() string {
	return "id"
}

func (o *VLAN) KeyName() string {
	return "ID"
}

func (o *VLAN) ModifiedBy(user int64, t time.Time) {
}
