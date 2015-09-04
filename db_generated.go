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
	return []interface{}{o.Type, o.MgmtIP, o.MgmtMac, o.Hostname, o.AssetTag, o.Height, o.RID, o.RU, o.VID, o.PrimaryMac, o.Modified, o.UID, o.PrimaryIP, o.SerialNo, o.Note, o.Model}
}

func (o *Device) UpdateValues() []interface{} {
	return []interface{}{o.Type, o.MgmtIP, o.MgmtMac, o.Hostname, o.AssetTag, o.Height, o.RID, o.RU, o.VID, o.PrimaryMac, o.Modified, o.UID, o.PrimaryIP, o.SerialNo, o.Note, o.Model, o.DID}
}

func (o *Device) MemberPointers() []interface{} {
	return []interface{}{&o.DID, &o.Type, &o.MgmtIP, &o.MgmtMac, &o.Hostname, &o.AssetTag, &o.Height, &o.RID, &o.RU, &o.VID, &o.PrimaryMac, &o.Modified, &o.UID, &o.PrimaryIP, &o.SerialNo, &o.Note, &o.Model}
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
	return "did,device_type,mgmt_ip,mgmt_mac,hostname,asset_tag,height,rid,ru,vid,primary_mac,modified,uid,primary_ip,sn,note,model"
}

func (o *Device) InsertFields() string {
	return "did,device_type,mgmt_ip,mgmt_mac,hostname,asset_tag,height,rid,ru,vid,primary_mac,modified,uid,primary_ip,sn,note,model"
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
	return []interface{}{o.Modified, o.UID, o.DID, o.PortType, o.MAC, o.CableTag, o.SwitchPort}
}

func (o *Port) UpdateValues() []interface{} {
	return []interface{}{o.Modified, o.UID, o.DID, o.PortType, o.MAC, o.CableTag, o.SwitchPort, o.PID}
}

func (o *Port) MemberPointers() []interface{} {
	return []interface{}{&o.PID, &o.Modified, &o.UID, &o.DID, &o.PortType, &o.MAC, &o.CableTag, &o.SwitchPort}
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
	return "pid,modified,uid,did,port_type,mac,cable_tag,switch_port"
}

func (o *Port) InsertFields() string {
	return "pid,modified,uid,did,port_type,mac,cable_tag,switch_port"
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
// User DBObject interface functions
//
func (o *User) InsertValues() []interface{} {
	return []interface{}{o.First, o.Last, o.Email, o.Level, o.Login}
}

func (o *User) UpdateValues() []interface{} {
	return []interface{}{o.First, o.Last, o.Email, o.Level, o.Login, o.ID}
}

func (o *User) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.First, &o.Last, &o.Email, &o.Level, &o.Login}
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
	return "id,firstname,lastname,email,admin,login"
}

func (o *User) InsertFields() string {
	return "id,firstname,lastname,email,admin,login"
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
	return []interface{}{o.DID, o.Filename, o.Title, o.RemoteAddr, o.UID}
}

func (o *Document) UpdateValues() []interface{} {
	return []interface{}{o.DID, o.Filename, o.Title, o.RemoteAddr, o.UID, o.ID}
}

func (o *Document) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.DID, &o.Filename, &o.Title, &o.RemoteAddr, &o.UID}
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
	return "id,did,filename,title,remote_addr,user_id"
}

func (o *Document) InsertFields() string {
	return "id,did,filename,title,remote_addr,user_id"
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
	return []interface{}{o.Postal, o.UID, o.Name, o.WWW, o.City, o.State, o.Country, o.Note, o.RemoteAddr, o.Modified, o.Phone, o.Address}
}

func (o *Vendor) UpdateValues() []interface{} {
	return []interface{}{o.Postal, o.UID, o.Name, o.WWW, o.City, o.State, o.Country, o.Note, o.RemoteAddr, o.Modified, o.Phone, o.Address, o.VID}
}

func (o *Vendor) MemberPointers() []interface{} {
	return []interface{}{&o.VID, &o.Postal, &o.UID, &o.Name, &o.WWW, &o.City, &o.State, &o.Country, &o.Note, &o.RemoteAddr, &o.Modified, &o.Phone, &o.Address}
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
	return "vid,postal,user_id,name,www,city,state,country,note,remote_addr,modified,phone,address"
}

func (o *Vendor) InsertFields() string {
	return "vid,postal,user_id,name,www,city,state,country,note,remote_addr,modified,phone,address"
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
	return []interface{}{o.Note, o.Closed, o.UID, o.Number, o.Jira, o.DCTicket, o.Opened, o.DID, o.VID}
}

func (o *RMA) UpdateValues() []interface{} {
	return []interface{}{o.Note, o.Closed, o.UID, o.Number, o.Jira, o.DCTicket, o.Opened, o.DID, o.VID, o.ID}
}

func (o *RMA) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Note, &o.Closed, &o.UID, &o.Number, &o.Jira, &o.DCTicket, &o.Opened, &o.DID, &o.VID}
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
	return "rma_id,note,date_closed,user_id,rma_no,jira,dc_ticket,date_opened,did,vid"
}

func (o *RMA) InsertFields() string {
	return "rma_id,note,date_closed,user_id,rma_no,jira,dc_ticket,date_opened,did,vid"
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
	return []interface{}{o.Tracking, o.UID, o.Sent, o.RMAID, o.CarrierID}
}

func (o *Return) UpdateValues() []interface{} {
	return []interface{}{o.Tracking, o.UID, o.Sent, o.RMAID, o.CarrierID, o.ReturnID}
}

func (o *Return) MemberPointers() []interface{} {
	return []interface{}{&o.ReturnID, &o.Tracking, &o.UID, &o.Sent, &o.RMAID, &o.CarrierID}
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
	return "return_id,tracking_no,user_id,date_sent,rma_id,cr_id"
}

func (o *Return) InsertFields() string {
	return "return_id,tracking_no,user_id,date_sent,rma_id,cr_id"
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
	return []interface{}{o.RMAID, o.PID, o.UID, o.TS}
}

func (o *Received) UpdateValues() []interface{} {
	return []interface{}{o.RMAID, o.PID, o.UID, o.TS}
}

func (o *Received) MemberPointers() []interface{} {
	return []interface{}{&o.RMAID, &o.PID, &o.UID, &o.TS}
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
	return "rma_id,pid,user_id,date_received"
}

func (o *Received) InsertFields() string {
	return "rma_id,pid,user_id,date_received"
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
	return []interface{}{o.UID, o.Modified, o.Name, o.AKA, o.URL}
}

func (o *Manufacturer) UpdateValues() []interface{} {
	return []interface{}{o.UID, o.Modified, o.Name, o.AKA, o.URL, o.MID}
}

func (o *Manufacturer) MemberPointers() []interface{} {
	return []interface{}{&o.MID, &o.UID, &o.Modified, &o.Name, &o.AKA, &o.URL}
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
	return "mid,user_id,modified,name,aka,url"
}

func (o *Manufacturer) InsertFields() string {
	return "mid,user_id,modified,name,aka,url"
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
// SKU DBObject interface functions
//
func (o *SKU) InsertValues() []interface{} {
	return []interface{}{o.MID, o.PartNumber, o.Description, o.UID, o.Modified}
}

func (o *SKU) UpdateValues() []interface{} {
	return []interface{}{o.MID, o.PartNumber, o.Description, o.UID, o.Modified, o.KID}
}

func (o *SKU) MemberPointers() []interface{} {
	return []interface{}{&o.KID, &o.MID, &o.PartNumber, &o.Description, &o.UID, &o.Modified}
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
	return "kid,mid,part_no,description,user_id,modified"
}

func (o *SKU) InsertFields() string {
	return "kid,mid,part_no,description,user_id,modified"
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
	return []interface{}{o.AssetTag, o.Modified, o.UID, o.KID, o.SID, o.DID, o.RMAID, o.Location, o.Serial}
}

func (o *Part) UpdateValues() []interface{} {
	return []interface{}{o.AssetTag, o.Modified, o.UID, o.KID, o.SID, o.DID, o.RMAID, o.Location, o.Serial, o.PID}
}

func (o *Part) MemberPointers() []interface{} {
	return []interface{}{&o.PID, &o.AssetTag, &o.Modified, &o.UID, &o.KID, &o.SID, &o.DID, &o.RMAID, &o.Location, &o.Serial}
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
	return "pid,asset_tag,modified,user_id,kid,sid,did,rma_id,location,serial_no"
}

func (o *Part) InsertFields() string {
	return "pid,asset_tag,modified,user_id,kid,sid,did,rma_id,location,serial_no"
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
	return []interface{}{o.Web, o.PXEPass, o.RemoteAddr, o.Name, o.Phone, o.PXEKey, o.City, o.DCMan, o.PXEHost, o.UID, o.Address, o.State, o.PXEUser, o.Modified}
}

func (o *Datacenter) UpdateValues() []interface{} {
	return []interface{}{o.Web, o.PXEPass, o.RemoteAddr, o.Name, o.Phone, o.PXEKey, o.City, o.DCMan, o.PXEHost, o.UID, o.Address, o.State, o.PXEUser, o.Modified, o.ID}
}

func (o *Datacenter) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Web, &o.PXEPass, &o.RemoteAddr, &o.Name, &o.Phone, &o.PXEKey, &o.City, &o.DCMan, &o.PXEHost, &o.UID, &o.Address, &o.State, &o.PXEUser, &o.Modified}
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
	return "id,web,pxepass,remote_addr,name,phone,pxekey,city,dcman,pxehost,user_id,address,state,pxeuser,modified"
}

func (o *Datacenter) InsertFields() string {
	return "id,web,pxepass,remote_addr,name,phone,pxekey,city,dcman,pxehost,user_id,address,state,pxeuser,modified"
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
	return []interface{}{o.DID, o.Hostname, o.AssetNumber, o.CPU, o.CPU_Speed, o.MemoryMB, o.Created}
}

func (o *DCView) UpdateValues() []interface{} {
	return []interface{}{o.DID, o.Hostname, o.AssetNumber, o.CPU, o.CPU_Speed, o.MemoryMB, o.Created, o.ID}
}

func (o *DCView) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.DID, &o.Hostname, &o.AssetNumber, &o.CPU, &o.CPU_Speed, &o.MemoryMB, &o.Created}
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
	return "id,datacenter,hostname,asset_number,cpu_id,cpu_speed,memory,created"
}

func (o *DCView) InsertFields() string {
	return "id,datacenter,hostname,asset_number,cpu_id,cpu_speed,memory,created"
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
	return []interface{}{o.SID, o.Hostname, o.Note, o.Height, o.SerialNo, o.IPMI, o.Rack, o.RID, o.RU, o.AssetTag, o.Internal, o.DC, o.NID, o.Alias}
}

func (o *RackUnit) UpdateValues() []interface{} {
	return []interface{}{o.SID, o.Hostname, o.Note, o.Height, o.SerialNo, o.IPMI, o.Rack, o.RID, o.RU, o.AssetTag, o.Internal, o.DC, o.NID, o.Alias}
}

func (o *RackUnit) MemberPointers() []interface{} {
	return []interface{}{&o.SID, &o.Hostname, &o.Note, &o.Height, &o.SerialNo, &o.IPMI, &o.Rack, &o.RID, &o.RU, &o.AssetTag, &o.Internal, &o.DC, &o.NID, &o.Alias}
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
	return "sid,hostname,note,height,sn,ipmi,rack,rid,ru,asset_tag,internal,dc,nid,alias"
}

func (o *RackUnit) InsertFields() string {
	return "sid,hostname,note,height,sn,ipmi,rack,rid,ru,asset_tag,internal,dc,nid,alias"
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
	return []interface{}{o.RU, o.IPInternal, o.IPIpmi, o.MacPort0, o.MacIPMI, o.PortEth1, o.CableEth1, o.CableIpmi, o.MacPort1, o.RID, o.Height, o.Hostname, o.Assigned, o.PartNo, o.PortIpmi, o.Alias, o.AssetTag, o.IPPublic, o.RemoteAddr, o.Profile, o.SerialNo, o.CableEth0, o.CPU, o.PortEth0, o.UID, o.Note, o.Modified}
}

func (o *Server) UpdateValues() []interface{} {
	return []interface{}{o.RU, o.IPInternal, o.IPIpmi, o.MacPort0, o.MacIPMI, o.PortEth1, o.CableEth1, o.CableIpmi, o.MacPort1, o.RID, o.Height, o.Hostname, o.Assigned, o.PartNo, o.PortIpmi, o.Alias, o.AssetTag, o.IPPublic, o.RemoteAddr, o.Profile, o.SerialNo, o.CableEth0, o.CPU, o.PortEth0, o.UID, o.Note, o.Modified, o.ID}
}

func (o *Server) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.RU, &o.IPInternal, &o.IPIpmi, &o.MacPort0, &o.MacIPMI, &o.PortEth1, &o.CableEth1, &o.CableIpmi, &o.MacPort1, &o.RID, &o.Height, &o.Hostname, &o.Assigned, &o.PartNo, &o.PortIpmi, &o.Alias, &o.AssetTag, &o.IPPublic, &o.RemoteAddr, &o.Profile, &o.SerialNo, &o.CableEth0, &o.CPU, &o.PortEth0, &o.UID, &o.Note, &o.Modified}
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
	return "id,ru,ip_internal,ip_ipmi,mac_eth0,mac_ipmi,port_eth1,cable_eth1,cable_ipmi,mac_eth1,rid,height,hostname,assigned,vendor_sku,port_ipmi,alias,asset_tag,ip_public,remote_addr,profile,sn,cable_eth0,cpu,port_eth0,uid,note,modified"
}

func (o *Server) InsertFields() string {
	return "id,ru,ip_internal,ip_ipmi,mac_eth0,mac_ipmi,port_eth1,cable_eth1,cable_ipmi,mac_eth1,rid,height,hostname,assigned,vendor_sku,port_ipmi,alias,asset_tag,ip_public,remote_addr,profile,sn,cable_eth0,cpu,port_eth0,uid,note,modified"
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
	return []interface{}{o.Make, o.Note, o.MgmtIP, o.UID, o.RID, o.Hostname, o.PartNo, o.SerialNo, o.Height, o.RU, o.AssetTag, o.Model, o.RemoteAddr, o.Modified}
}

func (o *Router) UpdateValues() []interface{} {
	return []interface{}{o.Make, o.Note, o.MgmtIP, o.UID, o.RID, o.Hostname, o.PartNo, o.SerialNo, o.Height, o.RU, o.AssetTag, o.Model, o.RemoteAddr, o.Modified, o.ID}
}

func (o *Router) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Make, &o.Note, &o.MgmtIP, &o.UID, &o.RID, &o.Hostname, &o.PartNo, &o.SerialNo, &o.Height, &o.RU, &o.AssetTag, &o.Model, &o.RemoteAddr, &o.Modified}
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
	return "id,make,note,ip_mgmt,uid,rid,hostname,sku,sn,height,ru,asset_tag,model,remote_addr,modified"
}

func (o *Router) InsertFields() string {
	return "id,make,note,ip_mgmt,uid,rid,hostname,sku,sn,height,ru,asset_tag,model,remote_addr,modified"
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
	return []interface{}{o.FirstIP, o.LastIP, o.Actual, o.MinIP, o.CIDR, o.Subnet, o.MaxIP, o.RID, o.VID}
}

func (o *RackNet) UpdateValues() []interface{} {
	return []interface{}{o.FirstIP, o.LastIP, o.Actual, o.MinIP, o.CIDR, o.Subnet, o.MaxIP, o.RID, o.VID}
}

func (o *RackNet) MemberPointers() []interface{} {
	return []interface{}{&o.FirstIP, &o.LastIP, &o.Actual, &o.MinIP, &o.CIDR, &o.Subnet, &o.MaxIP, &o.RID, &o.VID}
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
	return "first_ip,last_ip,actual,min_ip,cidr,subnet,max_ip,rid,vid"
}

func (o *RackNet) InsertFields() string {
	return "first_ip,last_ip,actual,min_ip,cidr,subnet,max_ip,rid,vid"
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
	return []interface{}{o.Profile, o.Modified, o.UID, o.SID, o.Hostname, o.Public, o.RemoteAddr, o.Private, o.VIP, o.Note}
}

func (o *VM) UpdateValues() []interface{} {
	return []interface{}{o.Profile, o.Modified, o.UID, o.SID, o.Hostname, o.Public, o.RemoteAddr, o.Private, o.VIP, o.Note, o.ID}
}

func (o *VM) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.Profile, &o.Modified, &o.UID, &o.SID, &o.Hostname, &o.Public, &o.RemoteAddr, &o.Private, &o.VIP, &o.Note}
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
	return "id,profile,modified,uid,sid,hostname,public,remote_addr,private,vip,note"
}

func (o *VM) InsertFields() string {
	return "id,profile,modified,uid,sid,hostname,public,remote_addr,private,vip,note"
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
	return []interface{}{o.Hostname, o.FQDN, o.Eth1, o.SN, o.IPMI_IP, o.Release, o.IP, o.IPs, o.Eth0, o.IPMI_MAC, o.Asset, o.CPU, o.Mem, o.VMs, o.Kernel}
}

func (o *Audit) UpdateValues() []interface{} {
	return []interface{}{o.Hostname, o.FQDN, o.Eth1, o.SN, o.IPMI_IP, o.Release, o.IP, o.IPs, o.Eth0, o.IPMI_MAC, o.Asset, o.CPU, o.Mem, o.VMs, o.Kernel}
}

func (o *Audit) MemberPointers() []interface{} {
	return []interface{}{&o.Hostname, &o.FQDN, &o.Eth1, &o.SN, &o.IPMI_IP, &o.Release, &o.IP, &o.IPs, &o.Eth0, &o.IPMI_MAC, &o.Asset, &o.CPU, &o.Mem, &o.VMs, &o.Kernel}
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
	return "hostname,fqdn,eth1,sn,ipmi_ip,release,remote_addr,ips,eth0,ipmi_mac,asset,cpu,mem,vms,kernel"
}

func (o *Audit) InsertFields() string {
	return "hostname,fqdn,eth1,sn,ipmi_ip,release,remote_addr,ips,eth0,ipmi_mac,asset,cpu,mem,vms,kernel"
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
	return []interface{}{o.RID, o.Hostname, o.IP, o.Netmask, o.Gateway, o.DNS, o.AssetTag}
}

func (o *PDU) UpdateValues() []interface{} {
	return []interface{}{o.RID, o.Hostname, o.IP, o.Netmask, o.Gateway, o.DNS, o.AssetTag, o.ID}
}

func (o *PDU) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.RID, &o.Hostname, &o.IP, &o.Netmask, &o.Gateway, &o.DNS, &o.AssetTag}
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
	return "id,rid,hostname,ip_address,netmask,gateway,dns,asset_tag"
}

func (o *PDU) InsertFields() string {
	return "id,rid,hostname,ip_address,netmask,gateway,dns,asset_tag"
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
	return []interface{}{o.DID, o.Name, o.Profile, o.Gateway, o.Route, o.Netmask}
}

func (o *VLAN) UpdateValues() []interface{} {
	return []interface{}{o.DID, o.Name, o.Profile, o.Gateway, o.Route, o.Netmask, o.ID}
}

func (o *VLAN) MemberPointers() []interface{} {
	return []interface{}{&o.ID, &o.DID, &o.Name, &o.Profile, &o.Gateway, &o.Route, &o.Netmask}
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
	return "id,did,name,profile,gateway,route,netmask"
}

func (o *VLAN) InsertFields() string {
	return "id,did,name,profile,gateway,route,netmask"
}

func (o *VLAN) KeyField() string {
	return "id"
}

func (o *VLAN) KeyName() string {
	return "ID"
}

func (o *VLAN) ModifiedBy(user int64, t time.Time) {
}
