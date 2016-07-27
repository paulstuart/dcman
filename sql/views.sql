
DROP VIEW IF EXISTS skuview;
CREATE VIEW skuview as 
  select k.kid, k.pti, k.mid, t.name as parttype, k.part_no, k.description, m.name as mfgr
  from  skus k
  left outer join mfgrs m on k.mid = m.mid
  left outer join part_types t on k.pti = t.pti
  ;

CREATE TRIGGER sku_in INSTEAD OF INSERT ON skuview 
BEGIN
    insert or ignore into part_types (name) values(NEW.parttype);
    insert or ignore into mfgrs (name) values(NEW.mfgr);
    --insert into logger values(NEW.mfgr);
    insert or ignore into skus (description, mid, pti)
        select NEW.description, m.mid, p.pti
          from mfgrs m, part_types p  
           where m.name = NEW.mfgr
               and p.name = NEW.parttype
            ;
END;

CREATE TRIGGER sku_up INSTEAD OF UPDATE ON skuview 
BEGIN
    insert or ignore into part_types (name) values(ifnull(NEW.parttype,OLD.parttype));
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
        pti = (select pti from part_types where name = ifnull(NEW.parttype,OLD.parttype))
        where kid = old.kid
        ;
    /*
        */
        select * from logger;
END;

DROP VIEW IF EXISTS partview;
CREATE VIEW partview as 
   select p.pid, p.sid, p.dcd, ifnull(r.rma_id, 0) as rma_id, s.*, d.name as dc, h.hostname, p.serial_no, p.asset_tag, p.unused, p.bad, p.location
   from parts p
   left outer join skuview s on p.kid = s.kid
   left outer join rmas r on p.pid = r.old_pid
   left outer join servers h on p.sid = h.id
   left outer join datacenters d on p.dcd = d.dcd
;

CREATE TRIGGER partin INSTEAD OF INSERT ON partview 
WHEN NEW.kid > 0
BEGIN
    insert into parts (
        dcd,
        kid,
        serial_no,
        asset_tag,
        location,
        unused,
        bad
    ) values (
        new.dcd,
        new.kid,
        new.serial_no,
        new.asset_tag, 
        new.location, 
        new.unused,
        new.bad
    ); 
END;

CREATE TRIGGER partin_new INSTEAD OF INSERT ON partview 
WHEN NEW.kid == 0
BEGIN
    insert into skuview (
        description,
        part_no,
        parttype,
        mfgr
    ) values (
        new.description,
        new.part_no,
        new.parttype,
        new.mfgr
    );

    insert into parts (
        serial_no,
        asset_tag,
        unused,
        bad,
        location,
        dcd,
        kid
        )
        select 
            new.serial_no,
            new.asset_tag, 
            new.unused,
            new.bad,
            new.location, 
            new.dcd, 
            kid from skuview 
              where description=new.description
                 and part_no=new.part_no 
                 and parttype=new.parttype 
                 and mfgr=new.mfgr
        ;
END;

CREATE TRIGGER partup INSTEAD OF UPDATE ON partview 
BEGIN
    update parts set
        sid = coalesce(NEW.sid, OLD.sid, (select sid from servers where hostname=NEW.hostname)),
        serial_no = ifnull(new.serial_no, old.serial_no),
        asset_tag = ifnull(new.asset_tag, old.asset_tag),
        unused = ifnull(new.unused, old.unused),
        bad = ifnull(new.bad, old.bad),
        location = ifnull(new.location, old.location)
        where pid = old.pid
        ;
    update skuview set 
        description=ifnull(new.description,old.description),
        part_no=ifnull(new.part_no,old.part_no),
        parttype=ifnull(new.parttype,old.parttype),
        mfgr=ifnull(new.mfgr,old.mfgr)
        where kid=old.kid
        ;
END;
DROP VIEW IF EXISTS rview;
CREATE VIEW rview as
	select d.name as dc, r.*
	from racks r
	left outer join datacenters d on r.dcd=d.dcd
    order by dc, r.rack
    ;

CREATE TRIGGER rview_insert INSTEAD OF INSERT ON rview 
BEGIN
  insert into racks (rack, dcd) values (NEW.rack, (select dcd from datacenters where name=NEW.dc));
END;


drop view if exists sview;

CREATE VIEW sview as
  select d.name as dc, r.rack as rack, r.dcd, s.*, t.tag
  from servers s
  left outer join racks r on s.rid = r.rid
  left outer join datacenters d on r.dcd = d.dcd
  left outer join tags t on s.tid = t.tid
  order by dc, rack, ru desc
;

CREATE TRIGGER sview_insert INSTEAD OF INSERT ON sview 
BEGIN
  insert into servers (rid, ru, hostname, alias, sn, asset_tag, ip_internal, ip_ipmi, mac_eth0,
	cable_ipmi, port_ipmi, cable_eth0, port_eth0, cable_eth1, port_eth1, pdu_a, pdu_b, note
	) 
  values ((select id from rview where dc=NEW.dc and rack=NEW.rack),
	NEW.ru, NEW.hostname, NEW.alias, NEW.sn, NEW.asset_tag, NEW.ip_internal, NEW.ip_ipmi, NEW.mac_eth0,
	NEW.cable_ipmi, NEW.port_ipmi, NEW.cable_eth0, NEW.port_eth0, NEW.cable_eth1, NEW.port_eth1, 
	NEW.pdu_a, NEW.pdu_b, NEW.note
	);
END;

DROP VIEW IF EXISTS "rmaview" ;
CREATE VIEW rmaview as 
    select r.*, p.description, p.serial_no as part_sn, p.part_no, s.hostname, s.sn as server_sn
    from rmas r
    left join servers s on r.sid = s.id
    left join partview p on p.pid = r.old_pid
    ;

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
        sid=coalesce(NEW.sid, OLD.sid, (select sid from servers where hostname=NEW.hostname)),
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

DROP VIEW IF EXISTS rma_report;
CREATE VIEW rma_report as 
  select r.*, u.login, s.dc, s.hostname, s.sn as server_sn, s.rack, s.ru, v.name as vendor_name,
         b.serial_no as bad_serial, b.part_no as bad_partno
  from  rmas r
  left outer join users u on r.user_id = u.id
  left outer join sview s on r.sid = s.id
  left outer join vendors v on r.vid = v.vid
;


DROP VIEW IF EXISTS vview; 
CREATE VIEW vview as
  select r.dcd, d.name as dc, r.rack as rack, s.rid, s.hostname as server, v.*
  from vms v
  left outer join servers s on v.sid = s.id
  left outer join racks r on s.rid = r.rid
  left outer join datacenters d on r.dcd = d.dcd;

DROP VIEW IF EXISTS rackspace; 
create view rackspace as select *,ru+height-1 as top from sview;

DROP VIEW IF EXISTS ipprivate; 
CREATE VIEW ipprivate as 
  select vmi as id, ifnull(dcd, 0) as dcd, dc, 'vm' as kind, 'private' as what,  hostname, private as ip, note
  from vview where private > '';

DROP VIEW IF EXISTS ipinternal; 
CREATE VIEW ipinternal as 
  select id, dcd, dc, 'server' as kind, 'internal' as what, hostname, ip_internal as ip, note
  from sview s
  where ip_internal > '';

/*
CREATE VIEW ipprivate as 
  select id, dcd, 'vm' as kind, 'private' as what,  dc, hostname, private as ip, note
  from vview where private > '';

CREATE VIEW ipinternal as 
  select id, dcd, 'server' as kind, 'internal' as what, dc, hostname, ip_internal as ip, note
  from sview s
  where ip_internal > '';
*/

DROP VIEW IF EXISTS ippublic; 
CREATE VIEW ippublic as 
  select vmi as id, dcd, dc, 'vm' as kind, 'public' as what, hostname, public as ip, note
  from vview where public > ''
  union
  select vmi as id, dcd, dc, 'vm' as kind, 'vip' as what, hostname, vip as ip, note
  from vview where vip > ''
  union
  select id, dcd, dc, 'server' as kind, 'public' as what, hostname, ip_public as ip, note
  from sview where ip_public > '';

DROP VIEW IF EXISTS ipipmi; 
CREATE VIEW ipipmi as 
  select id, dcd, dc, 'server' as kind, 'ipmi' as what, hostname, ip_ipmi as ip, note
  from sview where ip_ipmi > '';

DROP VIEW IF EXISTS ipinside; 
CREATE VIEW ipinside as
  select * from ipinternal
  union
  select * from ipipmi
  union
  select * from ipprivate;

DROP VIEW IF EXISTS ipmstr; 
CREATE VIEW ipmstr as
  select * from ipinside
  union
  select * from ippublic;

DROP VIEW IF EXISTS ippool; 
CREATE VIEW ippool as 
  select ip_internal as ip from servers where ip_internal > ''
  union
  select ip_ipmi as ip from servers where ip_ipmi > ''
  union
  select private as ip from vms where private > '';


DROP VIEW IF EXISTS vlanview;
CREATE VIEW vlanview as
    select d.name as dc, v.*
    from vlans v
    left outer join datacenters d on v.did = d.dcd
    ;

DROP VIEW IF EXISTS rack_vlans;
CREATE VIEW rack_vlans as 
select rid, vid, "start" as action, first_ip as ip from racknet
union
select rid, vid, "stop" as action, last_ip as ip from racknet;


DROP VIEW IF EXISTS nview;
CREATE VIEW nview as
  select d.name as dc, r.dcd, r.rack as rack, n.*
  from routers n
  left outer join racks r on n.rid = r.rid
  left outer join datacenters d on r.dcd = d.dcd
  ;

DROP VIEW IF EXISTS rackunits;
CREATE VIEW rackunits as
select * from (
    select dcd, dc, rack, 0 as nid, id as sid, rid, ru, height, hostname, alias, ip_ipmi as ipmi, ip_internal as internal, asset_tag, sn, note  from sview
    union
    select dcd, dc, rack, id as nid, 0 as sid, rid, ru, height,  hostname, '' as alias, '' as ipmi, ip_mgmt as ip_internal, asset_tag, sn, note from nview
) order by dcd, rack, ru desc;
