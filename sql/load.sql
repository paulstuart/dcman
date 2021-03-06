.bail on

attach database 'inventory.db' as olddb;

insert into users (usr, login, firstname, lastname, email, admin)
    select id, login, firstname, lastname, email, admin from olddb.users
    ;

INSERT INTO sites (sti,name,address,city,state,phone,web,usr)
   select id,name,address,city,state,phone,web,user_id from olddb.datacenters;

INSERT INTO "mfgrs" (name) values('Cisco');
INSERT INTO "mfgrs" (name) values('Juniper');

INSERT INTO "vendors" (name) values('SuperMicro');
INSERT INTO "vendors" (name) values('Amax');
INSERT INTO "vendors" (name) values('Hyve');

insert into device_types (name) values('Server');
insert into device_types (name) values('Switch');
insert into device_types (name) values('Router');
insert into device_types (name) values('Firewall');
insert into device_types (name) values('PDU');
insert into device_types (name) values('Cable Manager');

insert into ip_types (name, mgmt) values('IPMI', 1);
insert into ip_types (name) values('Internal');
insert into ip_types (name) values('Public');
insert into ip_types (name, mgmt) values('Mgmt', 1);
insert into ip_types (name,multi) values('VIP',1);


insert into racks 
    (rid, rack, sti, x_pos, y_pos, rackunits, vendor_id, usr, ts)
    select id, rack, did, x_pos, y_pos, rackunits, vendor_id, uid, ts
    from olddb.racks
    ;

insert into vms (vmi, did, hostname, profile, note, ts, usr) 
    select id, sid, hostname, profile, note, modified, uid from olddb.vms
    ;

DROP TRIGGER IF EXISTS devices_audit;
insert into devices
    (did,
    rid,
    tid,
    ru,
    height,
    hostname,
    alias,
    asset_tag,
    sn,
    profile,
    assigned,
    note,
    dti)
select
    id,
    rid,
    tid,
    ru,
    height,
    hostname,
    alias,
    asset_tag,
    sn,
    profile,
    assigned,
    note,
    (select dti from device_types where name='Server')
from olddb.servers;

insert into devices_view
    (
    rid,
    ru,
    height,
    hostname,
    asset_tag,
    sn,
    make,
    model,
    note,
    dti
)
select
    rid,
    ru,
    height,
    hostname,
    asset_tag,
    sn,
    make,
    model,
    note,
    (select dti from device_types where name='Switch')
from olddb.routers;

update devices set tid=null where tid=0;

.read sql/triggers.sql

create temp view devfix as
  select d.*, f.ifd, f.mgmt, s.ip_ipmi, s.ip_internal, s.ip_public
  from devices d, olddb.servers s, interfaces f
   where d.did = s.id
     and d.did = f.did
    ;

create temp view switch_ips as 
   select d.*, r.ip_mgmt from devices_network d
    left outer join routers r on d.hostname = r.hostname
    where d.devtype = 'Switch'
    ;

-- add IPMI
insert into interfaces (did, mgmt, mac, cable_tag, switch_port)
    select id, 1, mac_ipmi, cable_ipmi, port_ipmi
    from olddb.servers;

insert into ips
     (ifd, ipv4, ipt)
    select ifd, ip_ipmi, (select ipt from ip_types where name='IPMI') 
    from devfix
    where ip_ipmi > ' '
;

-- add internal
insert into interfaces (did, mac, cable_tag, switch_port)
    select id, mac_eth0, cable_eth0, port_eth0
    from olddb.servers
    ;

insert into ips
     (ifd, ipv4, ipt)
    select ifd, ip_internal , (select ipt from ip_types where name='Internal')
    from devfix 
    where mgmt = 0
    and ip_internal > ' '
    ;

 -- public 
insert into ips
     (ifd, ipv4, ipt)
    select ifd, ip_public, (select ipt from ip_types where name='Public') 
    from devfix 
    where mgmt = 0 
    and length(ip_public) > 3
;

-- add eth1
insert into interfaces (did, port, mac, cable_tag, switch_port)
    select id, 1, mac_eth1, cable_eth1, port_eth1
    from olddb.servers
    ;

insert into interfaces (did, mgmt)
    select did, 1
    from devices_view where devtype = 'Switch'
    ;

--
-- VM IPs
--
insert into ips
     (vmi, ipv4, ipt)
    select id, private, (select ipt from ip_types where name='Internal') 
    from olddb.vms 
    where private > ' '
;

insert into ips
     (vmi, ipv4, ipt)
    select id, public, (select ipt from ip_types where name='Public') 
    from olddb.vms 
    where length(public) > 3
;

insert into ips
     (vmi, ipv4, ipt)
    select id, vip, (select ipt from ip_types where name='VIP') 
    from olddb.vms 
    where length(vip) > 3
;

--
-- "Router" IPs
--
insert into ips(ifd, ipv4) select distinct ifd, ip_mgmt 
    from switch_ips 
    where ip_mgmt > ' ';

insert into vlans (vli,sti,name,profile,gateway,netmask,route,usr,ts) select * from olddb.vlans;
insert into tags (tid,tag) select tid,tag from olddb.tags;


-- clean up null entries
update interfaces set switch_port = '' where switch_port is null;

--update ips set ipv4 = ipv4;

detach database olddb;

.exit
.header on
.explain on
--select * from ips_missing limit 10;
select * from ips limit 10;

.echo on
PRAGMA foreign_keys=on;
select * from parts_view;
insert into parts_view (description,bad,asset_tag,unused,hostname,location,serial_no,part_no,mfgr,cents,sti,part_type,sku) 
    values(
        'XEON E5-2640V3, 8C, 2.60 GHZ 20M TRAY', 0, null, 1, null, null, null, 'INT-CM8064401830901','Intel', 85300, 2, 'Processor', '3926170'
    );
select * from parts_view;
select * from log;

.exit

insert into parts_view (description,bad,asset_tag,unused,kid,rmd,hostname,location,serial_no,part_no,mfgr,cents,price,did,sti,part_type,sku) 
    values(
        'XEON E5-2640V3, 8C, 2.60 GHZ 20M TRAY', 0, null, 1, 0, null, null, null, null, 'INT-CM8064401830901','Intel', 85300, 853.00, 0, 2, 'Processor', '3926170'
    );
insert into parts_view (vid,description,bad,asset_tag,unused,kid,rmd,hostname,location,serial_no,part_no,mfgr,cents,price,did,sti,part_type,sku) 
    values(
        0, 'XEON E5-2640V3, 8C, 2.60 GHZ 20M TRAY', 0, null, 1, 0, null, null, null, null, 'INT-CM8064401830901','Intel', 85300, 853.00, 0, 2, 'Processor', '3926170'
    );

--
-- NEW STUFF
--
--.echo on         
drop view if exists partload;
create view partload as
   select p.*, s.pti, d.name as site, s.description, s.part_type, m.name as mfgr
   from parts p
   left join skus_view s on p.kid = s.kid
   left join mfgrs m on m.mid = s.mid
   LEFT JOIN sites d on p.sti = d.sti
    ;

drop table if exists parttmp;
create temp table parttmp (
    p_type text,
    p_desc text,
    p_mfg text
);

.mode tabs
.import sql/ny.tab parttmp


CREATE TEMP TRIGGER partly_in INSTEAD OF INSERT ON partload 
BEGIN
    insert into skus_view (part_type, description, mfgr)
      values(NEW.part_type, NEW.description, NEW.mfgr);

    insert into parts (unused, kid, sti) 
        select 1, sv.kid, d.sti from skus_view sv, sites d
            where NEW.description = sv.description
              and NEW.mfgr = sv.mfgr
              and NEW.part_type = sv.part_type
              and d.name=NEW.site
        ;
END;

/*
--select distinct p_mfg from parttmp ;
insert into mfgrs (name) select distinct p_mfg from parttmp ;
select * from mfgrs;
.exit
*/
insert into partload (part_type, description, mfgr, site) select p_type, p_desc, p_mfg, 'NY7' from parttmp ;





--
--
-- COMMENT EXIT FOR VISUAL VALIDATION OF PART LOAD
--
--
--

.exit

.header on
--select p_type, p_desc, p_mfg, 'NY7' from parttmp ;
.print 'SKUVIEW'
select * from skus_view;

select count(*) as pcnt from partload;
select count(*) as cnt from parttmp;
select count(*) as pcnt from parts;


/*
 * Set some parts bad for ****** TESTING ****** 
 */

update parts set bad=1 where pid % 3 == 0;


select site, qty, mfgr, part_type, description from inventory;
select * from inventory limit 1;

.print TYPES
.print
select * from part_types;
.print
.print MFGR
.print
select * from mfgrs;
.print
.print SKUS
.print
select * from skus;
.print
.print PARTS
select pid, site, mfgr, part_type, description from partload;
/*
.print
.print RAW
select p_mfg from parttmp;
*/
