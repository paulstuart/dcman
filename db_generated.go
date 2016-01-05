// generated by dbgen ; DO NOT EDIT

package main

import (
	"time"
)

//
// Contract DBObject interface functions
//
func (o *Contract) InsertValues() []interface{} {
	return []interface{}{o.Policy, o.Phone, o.VID}
}

func (o *Contract) UpdateValues() []interface{} {
	return []interface{}{o.Policy, o.Phone, o.VID, o.CID}
}

func (o *Contract) MemberPointers() []interface{} {
	return []interface{}{&o.CID, &o.Policy, &o.Phone, &o.VID}
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
	return "cid,policy,phone,vid"
}

func (o *Contract) InsertFields() string {
	return "cid,policy,phone,vid"
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
	return []interface{}{o.PrimaryMac, o.MgmtMac, o.Note, o.UID, o.PrimaryIP, o.MgmtIP, o.SerialNo, o.Model, o.AssetTag, o.Modified, o.RID, o.RU, o.Height, o.Type, o.VID, o.Hostname}
}

func (o *Device) UpdateValues() []interface{} {
	return []interface{}{o.PrimaryMac, o.MgmtMac, o.Note, o.UID, o.PrimaryIP, o.MgmtIP, o.SerialNo, o.Model, o.AssetTag, o.Modified, o.RID, o.RU, o.Height, o.Type, o.VID, o.Hostname, o.DID}
}

func (o *Device) MemberPointers() []interface{} {
	return []interface{}{&o.DID, &o.PrimaryMac, &o.MgmtMac, &o.Note, &o.UID, &o.PrimaryIP, &o.MgmtIP, &o.SerialNo, &o.Model, &o.AssetTag, &o.Modified, &o.RID, &o.RU, &o.Height, &o.Type, &o.VID, &o.Hostname}
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
	return "did,primary_mac,mgmt_mac,note,uid,primary_ip,mgmt_ip,sn,model,asset_tag,modified,rid,ru,height,device_type,vid,hostname"
}

func (o *Device) InsertFields() string {
	return "did,primary_mac,mgmt_mac,note,uid,primary_ip,mgmt_ip,sn,model,asset_tag,modified,rid,ru,height,device_type,vid,hostname"
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
	return []interface{}{o.DID, o.Type, o.Int, o.Modified, o.UID}
}

func (o *IP) UpdateValues() []interface{} {
	return []interface{}{o.DID, o.Type, o.Int, o.Modified, o.UID, o.IID}
}

func (o *IP) MemberPointers() []interface{} {
	return []interface{}{&o.IID, &o.DID, &o.Type, &o.Int, &o.Modified, &o.UID}
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
	return "iid,did,ip_type,ip_int,modified,uid"
}

func (o *IP) InsertFields() string {
	return "iid,did,ip_type,ip_int,modified,uid"
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
// Choice DBObject interface functions
//
func (o *Choice) InsertValues() []interface{} {
	return []interface{}{o.ID, o.Label}
}

func (o *Choice) UpdateValues() []interface{} {
	return []interface{}{o.ID, o.Label}
}

func (o *Choice) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Label}
}

func (o *Choice) Key() int64 {
	return 0
}

func (o *Choice) SetID(id int64) {
}

func (o *Choice) TableName() string {
	return "part_choices"
}

func (o *Choice) SelectFields() string {
	return "pid,label"
}

func (o *Choice) InsertFields() string {
	return "pid,label"
}

func (o *Choice) KeyField() string {
	return ""
}

func (o *Choice) KeyName() string {
	return ""
}

func (o *Choice) ModifiedBy(user int64, t time.Time) {
}

//
// User DBObject interface functions
//
func (o *User) InsertValues() []interface{} {
	return []interface{}{o.Email, o.Level, o.Login, o.First, o.Last}
}

func (o *User) UpdateValues() []interface{} {
	return []interface{}{o.Email, o.Level, o.Login, o.First, o.Last, o.ID}
}

func (o *User) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Email, &o.Level, &o.Login, &o.First, &o.Last}
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
	return "id,email,admin,login,firstname,lastname"
}

func (o *User) InsertFields() string {
	return "id,email,admin,login,firstname,lastname"
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
	return []interface{}{o.DID, o.Filename, o.Title, o.RemoteAddr, o.UID, o.Modified}
}

func (o *Document) UpdateValues() []interface{} {
	return []interface{}{o.DID, o.Filename, o.Title, o.RemoteAddr, o.UID, o.Modified, o.ID}
}

func (o *Document) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.DID, &o.Filename, &o.Title, &o.RemoteAddr, &o.UID, &o.Modified}
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
	return "id,did,filename,title,remote_addr,user_id,modified"
}

func (o *Document) InsertFields() string {
	return "id,did,filename,title,remote_addr,user_id,modified"
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
	return []interface{}{o.Phone, o.Address, o.State, o.Postal, o.Modified, o.Name, o.City, o.Country, o.Note, o.RemoteAddr, o.UID, o.WWW}
}

func (o *Vendor) UpdateValues() []interface{} {
	return []interface{}{o.Phone, o.Address, o.State, o.Postal, o.Modified, o.Name, o.City, o.Country, o.Note, o.RemoteAddr, o.UID, o.WWW, o.VID}
}

func (o *Vendor) MemberPointers() []interface{} {
	return []interface{}{&o.VID, &o.Phone, &o.Address, &o.State, &o.Postal, &o.Modified, &o.Name, &o.City, &o.Country, &o.Note, &o.RemoteAddr, &o.UID, &o.WWW}
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
	return "vid,phone,address,state,postal,modified,name,city,country,note,remote_addr,user_id,www"
}

func (o *Vendor) InsertFields() string {
	return "vid,phone,address,state,postal,modified,name,city,country,note,remote_addr,user_id,www"
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
	return []interface{}{o.DCTicket, o.Receiving, o.DID, o.Number, o.Jira, o.Opened, o.Closed, o.VID, o.UID, o.Note}
}

func (o *RMA) UpdateValues() []interface{} {
	return []interface{}{o.DCTicket, o.Receiving, o.DID, o.Number, o.Jira, o.Opened, o.Closed, o.VID, o.UID, o.Note, o.ID}
}

func (o *RMA) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.DCTicket, &o.Receiving, &o.DID, &o.Number, &o.Jira, &o.Opened, &o.Closed, &o.VID, &o.UID, &o.Note}
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
	return "rma_id,dc_ticket,dc_receiving,did,rma_no,jira,date_opened,date_closed,vid,user_id,note"
}

func (o *RMA) InsertFields() string {
	return "rma_id,dc_ticket,dc_receiving,did,rma_no,jira,date_opened,date_closed,vid,user_id,note"
}

func (o *RMA) KeyField() string {
	return "rma_id"
}

func (o *RMA) KeyName() string {
	return "ID"
}

func (o *RMA) ModifiedBy(user int64, t time.Time) {
}

//
// Carrier DBObject interface functions
//
func (o *Carrier) InsertValues() []interface{} {
	return []interface{}{o.Name, o.URL, o.UID, o.Modified}
}

func (o *Carrier) UpdateValues() []interface{} {
	return []interface{}{o.Name, o.URL, o.UID, o.Modified, o.CarrierID}
}

func (o *Carrier) MemberPointers() []interface{} {
	return []interface{}{&o.CarrierID, &o.Name, &o.URL, &o.UID, &o.Modified}
}

func (o *Carrier) Key() int64 {
	return o.CarrierID
}

func (o *Carrier) SetID(id int64) {
	o.CarrierID = id
}

func (o *Carrier) TableName() string {
	return "carriers"
}

func (o *Carrier) SelectFields() string {
	return "cr_id,name,tracking_url,user_id,modified"
}

func (o *Carrier) InsertFields() string {
	return "cr_id,name,tracking_url,user_id,modified"
}

func (o *Carrier) KeyField() string {
	return "cr_id"
}

func (o *Carrier) KeyName() string {
	return "CarrierID"
}

func (o *Carrier) ModifiedBy(user int64, t time.Time) {
}

//
// Return DBObject interface functions
//
func (o *Return) InsertValues() []interface{} {
	return []interface{}{o.RMAID, o.CarrierID, o.Tracking, o.UID, o.Sent}
}

func (o *Return) UpdateValues() []interface{} {
	return []interface{}{o.RMAID, o.CarrierID, o.Tracking, o.UID, o.Sent, o.ReturnID}
}

func (o *Return) MemberPointers() []interface{} {
	return []interface{}{&o.ReturnID, &o.RMAID, &o.CarrierID, &o.Tracking, &o.UID, &o.Sent}
}

func (o *Return) Key() int64 {
	return o.ReturnID
}

func (o *Return) SetID(id int64) {
	o.ReturnID = id
}

func (o *Return) TableName() string {
	return "rma_returns"
}

func (o *Return) SelectFields() string {
	return "return_id,rma_id,cr_id,tracking_no,user_id,date_sent"
}

func (o *Return) InsertFields() string {
	return "return_id,rma_id,cr_id,tracking_no,user_id,date_sent"
}

func (o *Return) KeyField() string {
	return "return_id"
}

func (o *Return) KeyName() string {
	return "ReturnID"
}

func (o *Return) ModifiedBy(user int64, t time.Time) {
}

//
// Sent DBObject interface functions
//
func (o *Sent) InsertValues() []interface{} {
	return []interface{}{o.ReturnID, o.PID}
}

func (o *Sent) UpdateValues() []interface{} {
	return []interface{}{o.ReturnID, o.PID}
}

func (o *Sent) MemberPointers() []interface{} {
	return []interface{}{&o.ReturnID, &o.PID}
}

func (o *Sent) Key() int64 {
	return 0
}

func (o *Sent) SetID(id int64) {
}

func (o *Sent) TableName() string {
	return "rma_sent"
}

func (o *Sent) SelectFields() string {
	return "return_id,pid"
}

func (o *Sent) InsertFields() string {
	return "return_id,pid"
}

func (o *Sent) KeyField() string {
	return ""
}

func (o *Sent) KeyName() string {
	return ""
}

func (o *Sent) ModifiedBy(user int64, t time.Time) {
}

//
// Received DBObject interface functions
//
func (o *Received) InsertValues() []interface{} {
	return []interface{}{o.TS, o.RMAID, o.PID, o.UID}
}

func (o *Received) UpdateValues() []interface{} {
	return []interface{}{o.TS, o.RMAID, o.PID, o.UID}
}

func (o *Received) MemberPointers() []interface{} {
	return []interface{}{&o.TS, &o.RMAID, &o.PID, &o.UID}
}

func (o *Received) Key() int64 {
	return 0
}

func (o *Received) SetID(id int64) {
}

func (o *Received) TableName() string {
	return "rma_received"
}

func (o *Received) SelectFields() string {
	return "date_received,rma_id,pid,user_id"
}

func (o *Received) InsertFields() string {
	return "date_received,rma_id,pid,user_id"
}

func (o *Received) KeyField() string {
	return ""
}

func (o *Received) KeyName() string {
	return ""
}

func (o *Received) ModifiedBy(user int64, t time.Time) {
}

//
// Manufacturer DBObject interface functions
//
func (o *Manufacturer) InsertValues() []interface{} {
	return []interface{}{o.AKA, o.URL, o.UID, o.Modified, o.Name}
}

func (o *Manufacturer) UpdateValues() []interface{} {
	return []interface{}{o.AKA, o.URL, o.UID, o.Modified, o.Name, o.MID}
}

func (o *Manufacturer) MemberPointers() []interface{} {
	return []interface{}{&o.MID, &o.AKA, &o.URL, &o.UID, &o.Modified, &o.Name}
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
	return "mid,aka,url,user_id,modified,name"
}

func (o *Manufacturer) InsertFields() string {
	return "mid,aka,url,user_id,modified,name"
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
// PartType DBObject interface functions
//
func (o *PartType) InsertValues() []interface{} {
	return []interface{}{o.Name, o.UID, o.Modified}
}

func (o *PartType) UpdateValues() []interface{} {
	return []interface{}{o.Name, o.UID, o.Modified, o.TID}
}

func (o *PartType) MemberPointers() []interface{} {
	return []interface{}{&o.TID, &o.Name, &o.UID, &o.Modified}
}

func (o *PartType) Key() int64 {
	return o.TID
}

func (o *PartType) SetID(id int64) {
	o.TID = id
}

func (o *PartType) TableName() string {
	return "part_types"
}

func (o *PartType) SelectFields() string {
	return "tid,name,user_id,ts"
}

func (o *PartType) InsertFields() string {
	return "tid,name,user_id,ts"
}

func (o *PartType) KeyField() string {
	return "tid"
}

func (o *PartType) KeyName() string {
	return "TID"
}

func (o *PartType) ModifiedBy(user int64, t time.Time) {
	o.UID = user
	o.Modified = t
}

//
// SKU DBObject interface functions
//
func (o *SKU) InsertValues() []interface{} {
	return []interface{}{o.PartNumber, o.Description, o.UID, o.Modified, o.MID, o.TID}
}

func (o *SKU) UpdateValues() []interface{} {
	return []interface{}{o.PartNumber, o.Description, o.UID, o.Modified, o.MID, o.TID, o.KID}
}

func (o *SKU) MemberPointers() []interface{} {
	return []interface{}{&o.KID, &o.PartNumber, &o.Description, &o.UID, &o.Modified, &o.MID, &o.TID}
}

func (o *SKU) Key() int64 {
	return o.KID
}

func (o *SKU) SetID(id int64) {
	o.KID = id
}

func (o *SKU) TableName() string {
	return "skus"
}

func (o *SKU) SelectFields() string {
	return "kid,part_no,description,user_id,modified,mid,tid"
}

func (o *SKU) InsertFields() string {
	return "kid,part_no,description,user_id,modified,mid,tid"
}

func (o *SKU) KeyField() string {
	return "kid"
}

func (o *SKU) KeyName() string {
	return "KID"
}

func (o *SKU) ModifiedBy(user int64, t time.Time) {
	o.UID = user
	o.Modified = t
}

//
// Part DBObject interface functions
//
func (o *Part) InsertValues() []interface{} {
	return []interface{}{o.SID, o.DID, o.Location, o.Serial, o.AssetTag, o.KID, o.RMAID, o.UID, o.Modified}
}

func (o *Part) UpdateValues() []interface{} {
	return []interface{}{o.SID, o.DID, o.Location, o.Serial, o.AssetTag, o.KID, o.RMAID, o.UID, o.Modified, o.PID}
}

func (o *Part) MemberPointers() []interface{} {
	return []interface{}{&o.PID, &o.SID, &o.DID, &o.Location, &o.Serial, &o.AssetTag, &o.KID, &o.RMAID, &o.UID, &o.Modified}
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
	return "pid,sid,did,location,serial_no,asset_tag,kid,rma_id,user_id,modified"
}

func (o *Part) InsertFields() string {
	return "pid,sid,did,location,serial_no,asset_tag,kid,rma_id,user_id,modified"
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
// Datacenter DBObject interface functions
//
func (o *Datacenter) InsertValues() []interface{} {
	return []interface{}{o.Phone, o.DCMan, o.State, o.PXEHost, o.UID, o.Address, o.City, o.Web, o.Modified, o.Name, o.PXEUser, o.PXEPass, o.PXEKey, o.RemoteAddr}
}

func (o *Datacenter) UpdateValues() []interface{} {
	return []interface{}{o.Phone, o.DCMan, o.State, o.PXEHost, o.UID, o.Address, o.City, o.Web, o.Modified, o.Name, o.PXEUser, o.PXEPass, o.PXEKey, o.RemoteAddr, o.ID}
}

func (o *Datacenter) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Phone, &o.DCMan, &o.State, &o.PXEHost, &o.UID, &o.Address, &o.City, &o.Web, &o.Modified, &o.Name, &o.PXEUser, &o.PXEPass, &o.PXEKey, &o.RemoteAddr}
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
	return "id,phone,dcman,state,pxehost,user_id,address,city,web,modified,name,pxeuser,pxepass,pxekey,remote_addr"
}

func (o *Datacenter) InsertFields() string {
	return "id,phone,dcman,state,pxehost,user_id,address,city,web,modified,name,pxeuser,pxepass,pxekey,remote_addr"
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
	return []interface{}{o.CPU, o.CPU_Speed, o.MemoryMB, o.Created, o.DID, o.Hostname, o.AssetNumber}
}

func (o *DCView) UpdateValues() []interface{} {
	return []interface{}{o.CPU, o.CPU_Speed, o.MemoryMB, o.Created, o.DID, o.Hostname, o.AssetNumber, o.ID}
}

func (o *DCView) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.CPU, &o.CPU_Speed, &o.MemoryMB, &o.Created, &o.DID, &o.Hostname, &o.AssetNumber}
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
	return "id,cpu_id,cpu_speed,memory,created,datacenter,hostname,asset_number"
}

func (o *DCView) InsertFields() string {
	return "id,cpu_id,cpu_speed,memory,created,datacenter,hostname,asset_number"
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
	return []interface{}{o.AssetTag, o.SerialNo, o.Internal, o.Rack, o.RID, o.RU, o.Height, o.Alias, o.DC, o.NID, o.Hostname, o.SID, o.IPMI, o.Note}
}

func (o *RackUnit) UpdateValues() []interface{} {
	return []interface{}{o.AssetTag, o.SerialNo, o.Internal, o.Rack, o.RID, o.RU, o.Height, o.Alias, o.DC, o.NID, o.Hostname, o.SID, o.IPMI, o.Note}
}

func (o *RackUnit) MemberPointers() []interface{} {
	return []interface{}{&o.AssetTag, &o.SerialNo, &o.Internal, &o.Rack, &o.RID, &o.RU, &o.Height, &o.Alias, &o.DC, &o.NID, &o.Hostname, &o.SID, &o.IPMI, &o.Note}
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
	return "asset_tag,sn,internal,rack,rid,ru,height,alias,dc,nid,hostname,sid,ipmi,note"
}

func (o *RackUnit) InsertFields() string {
	return "asset_tag,sn,internal,rack,rid,ru,height,alias,dc,nid,hostname,sid,ipmi,note"
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
	return []interface{}{o.RU, o.Note, o.PortEth1, o.CableEth1, o.RemoteAddr, o.CPU, o.Alias, o.AssetTag, o.PartNo, o.PortEth0, o.Height, o.IPPublic, o.CableEth0, o.MacPort0, o.Modified, o.UID, o.IPInternal, o.MacIPMI, o.Assigned, o.SerialNo, o.CableIpmi, o.RID, o.Hostname, o.Profile, o.IPIpmi, o.PortIpmi, o.MacPort1}
}

func (o *Server) UpdateValues() []interface{} {
	return []interface{}{o.RU, o.Note, o.PortEth1, o.CableEth1, o.RemoteAddr, o.CPU, o.Alias, o.AssetTag, o.PartNo, o.PortEth0, o.Height, o.IPPublic, o.CableEth0, o.MacPort0, o.Modified, o.UID, o.IPInternal, o.MacIPMI, o.Assigned, o.SerialNo, o.CableIpmi, o.RID, o.Hostname, o.Profile, o.IPIpmi, o.PortIpmi, o.MacPort1, o.ID}
}

func (o *Server) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.RU, &o.Note, &o.PortEth1, &o.CableEth1, &o.RemoteAddr, &o.CPU, &o.Alias, &o.AssetTag, &o.PartNo, &o.PortEth0, &o.Height, &o.IPPublic, &o.CableEth0, &o.MacPort0, &o.Modified, &o.UID, &o.IPInternal, &o.MacIPMI, &o.Assigned, &o.SerialNo, &o.CableIpmi, &o.RID, &o.Hostname, &o.Profile, &o.IPIpmi, &o.PortIpmi, &o.MacPort1}
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
	return "id,ru,note,port_eth1,cable_eth1,remote_addr,cpu,alias,asset_tag,vendor_sku,port_eth0,height,ip_public,cable_eth0,mac_eth0,modified,uid,ip_internal,mac_ipmi,assigned,sn,cable_ipmi,rid,hostname,profile,ip_ipmi,port_ipmi,mac_eth1"
}

func (o *Server) InsertFields() string {
	return "id,ru,note,port_eth1,cable_eth1,remote_addr,cpu,alias,asset_tag,vendor_sku,port_eth0,height,ip_public,cable_eth0,mac_eth0,modified,uid,ip_internal,mac_ipmi,assigned,sn,cable_ipmi,rid,hostname,profile,ip_ipmi,port_ipmi,mac_eth1"
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
	return []interface{}{o.RU, o.Model, o.Note, o.AssetTag, o.SerialNo, o.Modified, o.Height, o.MgmtIP, o.PartNo, o.RID, o.Hostname, o.Make, o.RemoteAddr, o.UID}
}

func (o *Router) UpdateValues() []interface{} {
	return []interface{}{o.RU, o.Model, o.Note, o.AssetTag, o.SerialNo, o.Modified, o.Height, o.MgmtIP, o.PartNo, o.RID, o.Hostname, o.Make, o.RemoteAddr, o.UID, o.ID}
}

func (o *Router) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.RU, &o.Model, &o.Note, &o.AssetTag, &o.SerialNo, &o.Modified, &o.Height, &o.MgmtIP, &o.PartNo, &o.RID, &o.Hostname, &o.Make, &o.RemoteAddr, &o.UID}
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
	return "id,ru,model,note,asset_tag,sn,modified,height,ip_mgmt,sku,rid,hostname,make,remote_addr,uid"
}

func (o *Router) InsertFields() string {
	return "id,ru,model,note,asset_tag,sn,modified,height,ip_mgmt,sku,rid,hostname,make,remote_addr,uid"
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
	return []interface{}{o.MinIP, o.LastIP, o.RID, o.VID, o.CIDR, o.Actual, o.Subnet, o.MaxIP, o.FirstIP}
}

func (o *RackNet) UpdateValues() []interface{} {
	return []interface{}{o.MinIP, o.LastIP, o.RID, o.VID, o.CIDR, o.Actual, o.Subnet, o.MaxIP, o.FirstIP}
}

func (o *RackNet) MemberPointers() []interface{} {
	return []interface{}{&o.MinIP, &o.LastIP, &o.RID, &o.VID, &o.CIDR, &o.Actual, &o.Subnet, &o.MaxIP, &o.FirstIP}
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
	return "min_ip,last_ip,rid,vid,cidr,actual,subnet,max_ip,first_ip"
}

func (o *RackNet) InsertFields() string {
	return "min_ip,last_ip,rid,vid,cidr,actual,subnet,max_ip,first_ip"
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
	return []interface{}{o.UID, o.Private, o.Public, o.VIP, o.Note, o.Modified, o.RemoteAddr, o.SID, o.Hostname, o.Profile}
}

func (o *VM) UpdateValues() []interface{} {
	return []interface{}{o.UID, o.Private, o.Public, o.VIP, o.Note, o.Modified, o.RemoteAddr, o.SID, o.Hostname, o.Profile, o.ID}
}

func (o *VM) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.UID, &o.Private, &o.Public, &o.VIP, &o.Note, &o.Modified, &o.RemoteAddr, &o.SID, &o.Hostname, &o.Profile}
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
	return "id,uid,private,public,vip,note,modified,remote_addr,sid,hostname,profile"
}

func (o *VM) InsertFields() string {
	return "id,uid,private,public,vip,note,modified,remote_addr,sid,hostname,profile"
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
	return []interface{}{o.VIP, o.Note, o.DC, o.Hostname, o.Private, o.Public}
}

func (o *Orphan) UpdateValues() []interface{} {
	return []interface{}{o.VIP, o.Note, o.DC, o.Hostname, o.Private, o.Public, o.ID}
}

func (o *Orphan) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.VIP, &o.Note, &o.DC, &o.Hostname, &o.Private, &o.Public}
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
	return "rowid,vip,note,dc,hostname,private,public"
}

func (o *Orphan) InsertFields() string {
	return "rowid,vip,note,dc,hostname,private,public"
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
	return []interface{}{o.CPU, o.VMs, o.Kernel, o.IP, o.FQDN, o.Eth0, o.SN, o.Asset, o.Release, o.Hostname, o.Eth1, o.IPMI_MAC, o.Mem, o.IPs, o.IPMI_IP}
}

func (o *Audit) UpdateValues() []interface{} {
	return []interface{}{o.CPU, o.VMs, o.Kernel, o.IP, o.FQDN, o.Eth0, o.SN, o.Asset, o.Release, o.Hostname, o.Eth1, o.IPMI_MAC, o.Mem, o.IPs, o.IPMI_IP}
}

func (o *Audit) MemberPointers() []interface{} {
	return []interface{}{&o.CPU, &o.VMs, &o.Kernel, &o.IP, &o.FQDN, &o.Eth0, &o.SN, &o.Asset, &o.Release, &o.Hostname, &o.Eth1, &o.IPMI_MAC, &o.Mem, &o.IPs, &o.IPMI_IP}
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
	return "cpu,vms,kernel,remote_addr,fqdn,eth0,sn,asset,release,hostname,eth1,ipmi_mac,mem,ips,ipmi_ip"
}

func (o *Audit) InsertFields() string {
	return "cpu,vms,kernel,remote_addr,fqdn,eth0,sn,asset,release,hostname,eth1,ipmi_mac,mem,ips,ipmi_ip"
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
	return []interface{}{o.Hostname, o.IP, o.Netmask, o.Gateway, o.DNS, o.AssetTag, o.RID}
}

func (o *PDU) UpdateValues() []interface{} {
	return []interface{}{o.Hostname, o.IP, o.Netmask, o.Gateway, o.DNS, o.AssetTag, o.RID, o.ID}
}

func (o *PDU) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Hostname, &o.IP, &o.Netmask, &o.Gateway, &o.DNS, &o.AssetTag, &o.RID}
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
	return "id,hostname,ip_address,netmask,gateway,dns,asset_tag,rid"
}

func (o *PDU) InsertFields() string {
	return "id,hostname,ip_address,netmask,gateway,dns,asset_tag,rid"
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
	return []interface{}{o.Netmask, o.DID, o.Name, o.Profile, o.Gateway, o.Route}
}

func (o *VLAN) UpdateValues() []interface{} {
	return []interface{}{o.Netmask, o.DID, o.Name, o.Profile, o.Gateway, o.Route, o.ID}
}

func (o *VLAN) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Netmask, &o.DID, &o.Name, &o.Profile, &o.Gateway, &o.Route}
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
	return "id,netmask,did,name,profile,gateway,route"
}

func (o *VLAN) InsertFields() string {
	return "id,netmask,did,name,profile,gateway,route"
}

func (o *VLAN) KeyField() string {
	return "id"
}

func (o *VLAN) KeyName() string {
	return "ID"
}

func (o *VLAN) ModifiedBy(user int64, t time.Time) {
}
