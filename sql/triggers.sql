
DROP TRIGGER IF EXISTS rmaview_insert;
CREATE TRIGGER rmaview_insert INSTEAD OF INSERT ON rmaview 
BEGIN
    insert into rmas (
    dcd,
    sid,
    vid,
    old_pid,
    new_pid,
    vendor_rma,
    ship_tracking,
    recv_tracking,
    jira,
    dc_ticket,
    dc_receiving,
    note,
    date_shipped,
    date_received,
    date_closed
  ) values (
    NEW.dcd,
    ifnull(NEW.sid, (select sid from servers where hostname=NEW.hostname)),
    NEW.vid,
    NEW.old_pid,
    NEW.new_pid,
    NEW.vendor_rma,
    NEW.ship_tracking,
    NEW.recv_tracking,
    NEW.jira,
    NEW.dc_ticket,
    NEW.dc_receiving,
    NEW.note,
    NEW.date_shipped,
    NEW.date_received,
    NEW.date_closed
  ); 
END;

DROP TRIGGER IF EXISTS rmaview_update;
CREATE TRIGGER rmaview_update INSTEAD OF UPDATE ON rmaview 
BEGIN
    update rmas set 
        --sid=coalesce(NEW.sid, OLD.sid, (select sid from servers where hostname=NEW.hostname)),
        sid=2212,
        dcd=ifnull(NEW.dcd, OLD.dcd),
        sid=ifnull(NEW.sid, OLD.sid),
        vid=ifnull(NEW.vid, OLD.vid),
        old_pid=ifnull(NEW.old_pid, OLD.old_pid),
        new_pid=ifnull(NEW.new_pid, OLD.new_pid),
        vendor_rma=ifnull(NEW.vendor_rma, OLD.vendor_rma),
        ship_tracking=ifnull(NEW.ship_tracking, OLD.ship_tracking),
        recv_tracking=ifnull(NEW.recv_tracking, OLD.recv_tracking),
        jira=ifnull(NEW.jira, OLD.jira),
        dc_ticket=ifnull(NEW.dc_ticket, OLD.dc_ticket),
        dc_receiving=ifnull(NEW.dc_receiving, OLD.dc_receiving),
        note=ifnull(NEW.note, OLD.note),
        date_shipped=ifnull(NEW.date_shipped, OLD.date_shipped),
        date_received=ifnull(NEW.date_received, OLD.date_received),
        date_closed=ifnull(NEW.date_closed, OLD.date_closed)
    where rma_id = OLD.rma_id;
END;

DROP TRIGGER IF EXISTS sview_update;
CREATE TRIGGER sview_update INSTEAD OF UPDATE ON sview 
BEGIN
    --insert into logger values('ID:' || OLD.id || 'ATAG:' || NEW.asset_tag);
    insert into logger values('OLD TID:' || OLD.tid || ' NEW TID:' || NEW.tid);
  update servers set 
	rid = ifnull(NEW.rid,OLD.rid),
	tid = ifnull(NEW.tid,OLD.tid),
    ru = ifnull(NEW.ru, OLD.ru),
    hostname = ifnull(NEW.hostname, OLD.hostname),
    alias = ifnull(NEW.alias, OLD.alias),
    sn = ifnull(NEW.sn, OLD.sn),
    profile = ifnull(NEW.profile, OLD.profile),
    asset_tag = ifnull(NEW.asset_tag, OLD.asset_tag),
    assigned = ifnull(NEW.assigned, OLD.assigned),
    ip_internal = ifnull(NEW.ip_internal, OLD.ip_internal),
    ip_public = ifnull(NEW.ip_public, OLD.ip_public),
    ip_ipmi = ifnull(NEW.ip_ipmi, OLD.ip_ipmi),
    mac_ipmi = ifnull(NEW.mac_ipmi, OLD.mac_ipmi),
    mac_eth0 = ifnull(NEW.mac_eth0, OLD.mac_eth0),
	cable_ipmi = ifnull(NEW.cable_ipmi, OLD.cable_ipmi),
    port_ipmi = ifnull(NEW.port_ipmi, OLD.port_ipmi),
    cable_eth0 = ifnull(NEW.cable_eth0, OLD.cable_eth0),
    port_eth0 = ifnull(NEW.port_eth0, OLD.port_eth0),
    cable_eth1 = ifnull(NEW.cable_eth1, OLD.cable_eth1),
    port_eth1 = ifnull(NEW.port_eth1,OLD.port_eth1), 
	pdu_a = ifnull(NEW.pdu_a, OLD.pdu_a),
    pdu_b = ifnull(NEW.pdu_b, OLD.pdu_b),
    note = ifnull(NEW.note, OLD.note)
    where id = OLD.id
    ;
END;

/*
DROP VIEW IF EXISTS vview; 
CREATE VIEW vview as
  select r.dcd, d.name as dc, r.rack as rack, s.rid, s.hostname as server, v.*
  from vms v
  left outer join servers s on v.sid = s.id
  left outer join racks r on s.rid = r.rid
  left outer join datacenters d on r.dcd = d.dcd;

dcd|dc|rack|rid|id|sid|hostname|profile|note|private|public|vip|modified|remote_addr|uid
2|SFO|1|4|1256|2447|APPS33003|||10.100.128.11||162.248.16.38|2015-01-23 19:07:18||0

	VMI        int64     `sql:"vmi" key:"true" table:"vms"`
	SID        int64     `sql:"sid"`
	Hostname   string    `sql:"hostname"`
	Private    string    `sql:"private"`
	Public     string    `sql:"public"`
	VIP        string    `sql:"vip"`
	Profile    string    `sql:"profile"`
	Note       string    `sql:"note"`
	Modified   time.Time `sql:"modified"`
	RemoteAddr string    `sql:"remote_addr"`
	UID        int64     `sql:"uid"`
  */

DROP TRIGGER IF EXISTS vview_update;
CREATE TRIGGER vview_update INSTEAD OF UPDATE ON vview 
BEGIN
  update vms set 
	private = ifnull(NEW.private,OLD.private),
	public = ifnull(NEW.public,OLD.public),
	profile = ifnull(NEW.profile,OLD.profile),
	vip = ifnull(NEW.vip,OLD.vip)
    where vmi = OLD.vmi
    ;
END;

DROP TRIGGER IF EXISTS vlanview_update;
CREATE TRIGGER vlanview_update INSTEAD OF UPDATE ON vlanview 
BEGIN
  update vlans set 
	name = ifnull(NEW.name,OLD.name),
	gateway = ifnull(NEW.gateway,OLD.gateway),
	profile = ifnull(NEW.profile,OLD.profile),
	route = ifnull(NEW.route,OLD.route),
	netmask = ifnull(NEW.netmask,OLD.netmask)
    where id = OLD.id
    ;
    /*
id|did|name|profile|gateway|netmask|route|user_id|modified
1|2|4|public|104.36.112.1|255.255.255.0||1|2015-07-24 23:04:09

    */
END;

/*
dc|rid|rack|dcd|x_pos|y_pos|rackunits|uid|ts|vendor_id
AMS|42|101|1|||45|3|2014-10-07 15:49:51|
*/
DROP TRIGGER IF EXISTS rview_update;
CREATE TRIGGER rview_update INSTEAD OF UPDATE ON rview 
BEGIN
  update racks set 
	rackunits = ifnull(NEW.rackunits,OLD.rackunits),
	vendor_id = ifnull(NEW.vendor_id,OLD.vendor_id),
	dcd = ifnull(NEW.dcd,OLD.dcd)
    where rid = OLD.rid
    ;
END;
