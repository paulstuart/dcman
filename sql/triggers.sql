create table if not exists log(event text); 

--
-- IPs
--

DROP TRIGGER IF EXISTS ips_insert;
CREATE TRIGGER ips_insert AFTER INSERT ON ips 
--when NEW.ip32 is null and NEW.ipv4 > ''
BEGIN
    update ips 
    set ip32 = (select ipcalc from ips_calc where iid = NEW.iid)
    where iid=NEW.iid
    ;
END;

DROP TRIGGER IF EXISTS ips_update;
CREATE TRIGGER ips_update AFTER UPDATE OF ipv4 ON ips 
--when NEW.ip32 is null and NEW.ipv4 > ''
BEGIN
    update ips 
    set ip32 = (select ipcalc from ips_calc where iid = OLD.iid)
    where iid=OLD.iid
    ;
END;


--
-- Devices
--

DROP TRIGGER IF EXISTS devices_view_insert;
CREATE TRIGGER devices_view_insert INSTEAD OF INSERT ON devices_view 
BEGIN
    insert into devices
        (rid, dti, tid, ru, height, hostname, alias, sn, profile, asset_tag, assigned, note)
        values
        (NEW.rid, NEW.dti, NEW.tid, NEW.ru, NEW.height, NEW.hostname, NEW.alias, 
            NEW.sn, NEW.profile, NEW.asset_tag, NEW.assigned, NEW.note)
        ;
END;

DROP TRIGGER IF EXISTS devices_view_update;
CREATE TRIGGER devices_view_update INSTEAD OF UPDATE ON devices_view 
BEGIN
  update devices set 
	rid = ifnull(NEW.rid,OLD.rid),
	dti = ifnull(NEW.dti,OLD.dti),
	tid = ifnull(NEW.tid,OLD.tid),
    ru = ifnull(NEW.ru, OLD.ru),
    height = ifnull(NEW.height, OLD.height),
    hostname = ifnull(NEW.hostname, OLD.hostname),
    alias = ifnull(NEW.alias, OLD.alias),
    sn = ifnull(NEW.sn, OLD.sn),
    profile = ifnull(NEW.profile, OLD.profile),
    asset_tag = ifnull(NEW.asset_tag, OLD.asset_tag),
    assigned = ifnull(NEW.assigned, OLD.assigned),
    note = ifnull(NEW.note, OLD.note)
    where did = OLD.did
    ;
END;

DROP TRIGGER IF EXISTS devices_insert;
CREATE TRIGGER devices_insert AFTER INSERT ON devices 
BEGIN
    insert or replace into notes values(NEW.did, 'Device', NEW.hostname, NEW.note);
END;

DROP TRIGGER IF EXISTS devices_delete;
CREATE TRIGGER devices_delete AFTER DELETE ON devices 
BEGIN
    delete from notes where id=OLD.did and kind='VM';
END;

--
-- VMs
--
DROP TRIGGER IF EXISTS vms_insert;
CREATE TRIGGER vms_insert AFTER INSERT ON vms 
BEGIN
    insert or replace into notes values(NEW.vmi, 'VM', NEW.hostname, NEW.note);
END;

DROP TRIGGER IF EXISTS vms_delete;
CREATE TRIGGER vms_delete AFTER DELETE ON vms 
BEGIN
    delete from notes where id=OLD.vmi and kind='VM';
END;


DROP TRIGGER IF EXISTS vms_view_update;
CREATE TRIGGER vms_view_update INSTEAD OF UPDATE ON vms_view 
BEGIN
  update vms set 
    hostname = ifnull(NEW.hostname, OLD.hostname),
    profile = ifnull(NEW.profile, OLD.profile),
    note = ifnull(NEW.note, OLD.note)
    where vmi = OLD.vmi
    ;
END;

DROP TRIGGER IF EXISTS skus_view_insert;
CREATE TRIGGER skus_view_insert INSTEAD OF INSERT ON skus_view 
BEGIN
    insert into log values('skus_view insert - part_type:' || ifnull(new.part_type, 'misc'));
    insert or ignore into part_types (name) values(ifnull(NEW.part_type, 'misc'));
    insert into log values('skus_view insert - mfgr:' || new.mfgr);
    insert or ignore into mfgrs (name) values(ifnull(NEW.mfgr,'unknown'));
    insert or ignore into skus (description, part_no, mid, pti)
        select NEW.description, NEW.part_no, 
            ifnull(NEW.mid, (select mid from mfgrs where name = ifnull(NEW.mfgr,'unknown'))),
            ifnull(NEW.pti, (select pti from part_types where name = ifnull(NEW.part_type,'misc')))
            ;
END;

DROP TRIGGER IF EXISTS skus_view_update;
CREATE TRIGGER skus_view_update INSTEAD OF UPDATE ON skus_view 
BEGIN
    insert or ignore into part_types (name) values(ifnull(NEW.part_type,OLD.part_type));
    insert or ignore into mfgrs (name) values(ifnull(NEW.mfgr,OLD.mfgr));
    /*
    insert into logger values( 'KID: '  || OLD.KID);
    insert into logger values( 'MFGR: ' || ifnull(NEW.mfgr, OLD.mfgr));
    insert into logger values( 'DESC: ' || ifnull(NEW.description, OLD.description));
    */
    update skus set
        description = ifnull(new.description,old.description),
        part_no = ifnull(new.part_no,old.part_no),
        mid = (select mid from mfgrs where name = ifnull(NEW.mfgr,OLD.mfgr)),
        pti = (select pti from part_types where name = ifnull(NEW.part_type,OLD.part_type))
        where kid = old.kid
        ;
END;

DROP TRIGGER IF EXISTS rmas_view_insert;
CREATE TRIGGER rmas_view_insert INSTEAD OF INSERT ON rmas_view 
BEGIN
    insert into rmas (
    sti,
    did,
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
    NEW.sti,
    ifnull(NEW.did, (select did from devices where hostname=NEW.hostname)),
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

DROP TRIGGER IF EXISTS rmas_view_update;
CREATE TRIGGER rmas_view_update INSTEAD OF UPDATE ON rmas_view 
BEGIN
    update rmas set 
        sti=ifnull(NEW.sti, OLD.sti),
        did=ifnull(NEW.did, OLD.did),
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
        date_created=ifnull(NEW.date_created, OLD.date_created),
        date_closed=ifnull(NEW.date_closed, OLD.date_closed)
    where rmd = OLD.rmd;
END;

DROP TRIGGER IF EXISTS rmas_view_delete;
CREATE TRIGGER rmas_view_delete INSTEAD OF DELETE ON rmas_view 
BEGIN
    delete from rmas where rmd = OLD.rmd;
END;



DROP TRIGGER IF EXISTS vlans_view_update;
CREATE TRIGGER vlans_view_update INSTEAD OF UPDATE ON vlans_view 
BEGIN
  update vlans set 
	sti = ifnull(NEW.sti,OLD.sti),
	name = ifnull(NEW.name,OLD.name),
	gateway = ifnull(NEW.gateway,OLD.gateway),
	profile = ifnull(NEW.profile,OLD.profile),
	route = ifnull(NEW.route,OLD.route),
	netmask = ifnull(NEW.netmask,OLD.netmask)
    where vli = OLD.vli
    ;
END;

DROP TRIGGER IF EXISTS racks_view_update;
CREATE TRIGGER racks_view_update INSTEAD OF UPDATE ON racks_view 
BEGIN
  update racks set 
	rackunits = ifnull(NEW.rackunits,OLD.rackunits),
	vendor_id = ifnull(NEW.vendor_id,OLD.vendor_id),
	note = ifnull(NEW.note,OLD.note),
	sti = ifnull(NEW.sti,OLD.sti)
    where rid = OLD.rid
    ;
END;

DROP TRIGGER IF EXISTS racks_audit;
CREATE TRIGGER racks_audit BEFORE UPDATE
ON racks
BEGIN
       INSERT INTO audit_racks select * from racks where rid=old.rid;
END;

DROP TRIGGER IF EXISTS devices_audit;
CREATE TRIGGER devices_audit BEFORE UPDATE
ON devices
BEGIN
   INSERT INTO audit_devices select * from devices where did=old.did;
END;

DROP TRIGGER IF EXISTS parts_view_insert_new_sku;
CREATE TRIGGER parts_view_insert_new_sku INSTEAD OF INSERT ON parts_view 
BEGIN
    insert into log values('KID:' || ifnull(new.kid, 'no kid'));
    insert into log values('part_no:' || ifnull(new.part_no, 'no part_no'));
    insert into log values('part_type:' || ifnull(new.part_type, 'no part_type'));
    insert into skus_view (
        description,
        part_no,
        pti,
        part_type,
        mfgr
    ) values (
        new.description,
        new.part_no,
        new.pti,
        new.part_type,
        new.mfgr
    );

    insert into parts (
        sti,
        serial_no,
        asset_tag,
        unused,
        bad,
        location,
        cents,
        vid,
        kid
        )
        select 
            new.sti, 
            new.serial_no,
            new.asset_tag, 
            new.unused,
            new.bad,
            new.location, 
            new.cents,
            ifnull((select vid from vendors where name=NEW.vendor), 0),
            kid from skus_view 
              where description=new.description
                 and part_no=new.part_no 
                 and part_type=ifnull(new.part_type,'misc') 
                 and mfgr=ifnull(new.mfgr,'unknown')
                ;
END;

DROP TRIGGER IF EXISTS parts_view_update;
CREATE TRIGGER parts_view_update INSTEAD OF UPDATE ON parts_view 
BEGIN
    --insert into db_debug (log) values ('old.did:', ifnull 
    update parts set
        --did = coalesce(NEW.did, OLD.did, (select did from devices where hostname=NEW.hostname)),
        did = ifnull(NEW.did, OLD.did),
        kid = ifnull(NEW.kid, OLD.kid),
        vid = ifnull(NEW.vid, OLD.vid),
        serial_no = ifnull(new.serial_no, old.serial_no),
        asset_tag = ifnull(new.asset_tag, old.asset_tag),
        unused = ifnull(new.unused, old.unused),
        bad = ifnull(new.bad, old.bad),
        location = ifnull(new.location, old.location),
        cents = ifnull(new.cents, old.cents)
        where pid = old.pid
        ;
    update skus_view set 
        description=ifnull(new.description,old.description),
        part_no=ifnull(new.part_no,old.part_no),
        part_type=ifnull(new.part_type,old.part_type),
        mfgr=ifnull(new.mfgr,old.mfgr)
        where kid=old.kid
        ;
END;


DROP TRIGGER IF EXISTS racks_view_insert;
CREATE TRIGGER racks_view_insert INSTEAD OF INSERT ON racks_view 
BEGIN
  insert into racks (rack, sti) values (NEW.rack, (select sti from sites where name=NEW.site));
END;

DROP TRIGGER IF EXISTS rmas_view_insert;
CREATE TRIGGER rmas_view_insert INSTEAD OF INSERT ON rmas_view 
BEGIN
    insert into rmas (
    sti,
    did,
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
    NEW.sti,
    ifnull(NEW.did, (select did from devices where hostname=NEW.hostname)),
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

DROP TRIGGER IF EXISTS rmas_view_update;
CREATE TRIGGER rmas_view_update INSTEAD OF UPDATE ON rmas_view 
BEGIN
    update rmas set 
        did=coalesce(NEW.did, OLD.did, (select did from devices where hostname=NEW.hostname)),
        sti=ifnull(NEW.sti, OLD.sti),
        did=ifnull(NEW.did, OLD.did),
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
    where rmd = OLD.rmd;
END;

/*
DROP TRIGGER IF EXISTS parts_view_insert;
CREATE TRIGGER parts_view_insert INSTEAD OF INSERT ON parts_view 
WHEN NEW.kid > 0
BEGIN
    insert into log values('KID:' || new.kid);
    insert into parts (
        sti,
        kid,
        serial_no,
        asset_tag,
        location,
        unused,
        bad
    ) values (
        new.sti,
        new.kid,
        new.serial_no,
        new.asset_tag, 
        new.location, 
        new.unused,
        new.bad
    ); 
END;
*/

