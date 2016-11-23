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

    insert or replace into notes values(NEW.iid, 'IP', NEW.ipv4, NEW.note);
END;

DROP TRIGGER IF EXISTS ips_update;
CREATE TRIGGER ips_update AFTER UPDATE OF ipv4 ON ips 
--when NEW.ip32 is null and NEW.ipv4 > ''
BEGIN
    update ips 
    set ip32 = (select ipcalc from ips_calc where iid = OLD.iid)
    where iid=OLD.iid
    ;

    insert or replace into notes values(NEW.iid, 'IP', NEW.ipv4, NEW.note);
END;


--
-- Devices
--

DROP TRIGGER IF EXISTS devices_view_insert;
CREATE TRIGGER devices_view_insert INSTEAD OF INSERT ON devices_view 
BEGIN
    insert into devices
        (usr, rid, dti, tid, ru, height, hostname, alias, model, sn, profile, asset_tag, assigned, note, mid)
        values
        (NEW.usr, NEW.rid, nullif(NEW.dti,0), NEW.tid, NEW.ru, NEW.height, NEW.hostname, NEW.alias, 
            NEW.model, NEW.sn, NEW.profile, NEW.asset_tag, NEW.assigned, NEW.note,
            (select mid from mfgrs where name=new.make)
        )
        ;
END;


DROP TRIGGER IF EXISTS devices_view_update;
CREATE TRIGGER devices_view_update INSTEAD OF UPDATE ON devices_view 
BEGIN
  update devices set 
	rid = ifnull(nullif(NEW.rid,0),OLD.rid),
	usr = ifnull(nullif(NEW.usr,0),OLD.usr),
	dti = ifnull(nullif(NEW.dti,0),OLD.dti),
	mid = ifnull(nullif(NEW.mid,0),OLD.mid),
	tid = ifnull(nullif(NEW.tid,0),OLD.tid),
    ru =  ifnull(NEW.ru, OLD.ru),
    height = ifnull(NEW.height, OLD.height),
    hostname = ifnull(NEW.hostname, OLD.hostname),
    alias = ifnull(NEW.alias, OLD.alias),
    sn = ifnull(NEW.sn, OLD.sn),
    model = ifnull(NEW.model, OLD.model),
    profile = ifnull(NEW.profile, OLD.profile),
    asset_tag = ifnull(NEW.asset_tag, OLD.asset_tag),
    assigned = ifnull(NEW.assigned, OLD.assigned),
    note = ifnull(NEW.note, OLD.note)
    where did = OLD.did
    ;
END;


-- add data for full text search
DROP TRIGGER IF EXISTS devices_insert;
CREATE TRIGGER devices_insert AFTER INSERT ON devices 
BEGIN
    insert or replace into notes values(NEW.did, 'Device', NEW.hostname, NEW.note);
END;

DROP TRIGGER IF EXISTS devices_update;
CREATE TRIGGER devices_update AFTER UPDATE ON devices 
BEGIN
    insert or replace into notes values(NEW.did, 'Device', NEW.hostname, NEW.note);
END;

-- delete ips / interfaces to remove FK dependencies
DROP TRIGGER IF EXISTS devices_delete_before;
CREATE TRIGGER devices_delete_before BEFORE DELETE ON devices 
BEGIN
    -- VM network --
    delete from ips 
        where iid in (
            select iid from ips 
            where ips.vmi in (
                select vmi from vms where did = OLD.did
            )
        );

    -- VM instances --
    delete from vms 
        where vmi in (
            select vmi from vms 
            where vms.did = OLD.did
        );

    -- Device network --
    delete from ips 
        where ifd in (
            select ifd from interfaces 
             where did = OLD.did
        );

    -- Device hardware --
    delete from interfaces
        where did = OLD.did
        ;
END;

-- remove deleted data from full text search
DROP TRIGGER IF EXISTS devices_delete;
CREATE TRIGGER devices_delete AFTER DELETE ON devices 
BEGIN
    delete from notes where id=OLD.did and kind='Device';
END;

DROP TRIGGER IF EXISTS devices_view_delete;
CREATE TRIGGER devices_view_delete INSTEAD OF DELETE ON devices_view 
BEGIN
    delete from devices where did = OLD.did;
END;

drop trigger if exists devices_adjust_move;
create trigger devices_adjust_move INSTEAD OF UPDATE on devices_adjust
    when NEW.ru != OLD.ru
    and NEW.height == OLD.height
    and not exists (
        select did from devices_adjust 
        where rid = OLD.rid
          and did != OLD.did
          and (NEW.ru + NEW.height -1 ) between ru and space
    )
BEGIN
    update devices set ru = NEW.ru where did = OLD.did;
END;

drop trigger if exists devices_adjust_resize;
create trigger devices_adjust_resize INSTEAD OF UPDATE on devices_adjust
    when NEW.height > 0
     and NEW.height != OLD.height
     and not exists (
        select did from devices_adjust 
        where rid = OLD.rid
          and did != OLD.did
          and (OLD.ru + NEW.height -1 ) between ru and space
    )
BEGIN
    update devices set height = NEW.height where did = OLD.did;
END;

DROP TRIGGER IF EXISTS interfaces_view_update;
CREATE TRIGGER interfaces_view_update INSTEAD OF UPDATE ON interfaces_view 
    when NEW.ipv4 != OLD.ipv4
     and NEW.ipv4 is not null
     and NEW.did is not null
     and NEW.iid is not null
BEGIN
    update ips 
    set ipv4 = NEW.ipv4 
    where iid=(select iid from devices_network where did=OLD.did and ipv4=OLD.ipv4)
    ;
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

DROP TRIGGER IF EXISTS vms_view_delete;
CREATE TRIGGER vms_view_delete INSTEAD OF DELETE ON vms_view 
BEGIN
    delete from ips where vmi = old.vmi;
    delete from vms where vmi = old.vmi;
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
	netmask = ifnull(NEW.netmask,OLD.netmask),
	starting = ifnull(NEW.starting,OLD.starting)
    where vli = OLD.vli
    ;
END;

-- add data for full text search
DROP TRIGGER IF EXISTS racks_insert;
CREATE TRIGGER racks_insert AFTER INSERT ON racks 
BEGIN
    insert or replace into notes values(NEW.rid, 'Rack', NEW.rack, NEW.note);
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

   insert or replace into notes values(NEW.rid, 'Rack', NEW.rack, NEW.note);
END;

DROP TRIGGER IF EXISTS racks_view_delete;
CREATE TRIGGER racks_view_delete INSTEAD OF DELETE ON racks_view 
BEGIN
  delete from racks where rid = OLD.rid ;
END;

DROP TRIGGER IF EXISTS racks_audit;
CREATE TRIGGER racks_audit BEFORE UPDATE
ON racks
BEGIN
       INSERT INTO audit_racks select * from racks where rid=old.rid;
END;

DROP TRIGGER IF EXISTS devices_audit;
CREATE TRIGGER devices_audit BEFORE UPDATE ON devices
BEGIN
    INSERT INTO audit_devices select * from devices where did=old.did;
    update devices set version=version+1 where did=old.did;
END;

DROP TRIGGER IF EXISTS vms_audit;
CREATE TRIGGER vms_audit BEFORE UPDATE
ON vms
BEGIN
    INSERT INTO audit_vms select * from vms where vmi=old.vmi;
    update vms set version=version+1 where vmi=old.vmi;
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
        ifnull(nullif(new.vid,0), (select vid from vendors where name=NEW.vendor)), 
        kid 
        from skus_view 
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
when NEW.sti > 0
BEGIN
  insert into racks (sti, rack, rackunits, vendor_id, note, usr) 
        values (NEW.sti, NEW.rack, NEW.rackunits, NEW.vendor_id, NEW.note, NEW.usr)
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

drop trigger if exists users_view_insert;
CREATE TRIGGER users_view_insert INSTEAD OF INSERT ON users_view 
BEGIN
    insert into users 
        (login, firstname, lastname, email, admin)
        values(NEW.login, NEW.firstname, NEW.lastname, NEW.email, NEW.admin)
        ;
END;

drop trigger if exists users_view_update;
CREATE TRIGGER users_view_update INSTEAD OF UPDATE ON users_view 
BEGIN
    update users set 
        login = ifnull(new.login,old.login), 
        firstname = ifnull(new.firstname,old.lastname),
        lastname = ifnull(new.lastname,old.lastname), 
        email = ifnull(new.email,old.email), 
        admin = ifnull(new.admin,old.admin)
    where usr = OLD.usr
    ;
END;

drop trigger if exists users_view_delete;
CREATE TRIGGER users_view_delete INSTEAD OF DELETE ON users_view 
BEGIN
    delete from users where usr = OLD.usr;
END;
