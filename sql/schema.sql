PRAGMA foreign_keys=OFF;
PRAGMA journal_mode = WAL;

BEGIN TRANSACTION;

CREATE TABLE "datacenters" (
    id integer primary key AUTOINCREMENT,
    name text not null,
    address text not null,
    city text not null,
    state text not null,
    phone text not null,
    web text not null,
    dcman text not null,
    pxehost text not null,
    pxeuser text not null,
    pxepass text not null,
    pxekey text not null,
    remote_addr text not null default '', 
    modified timestamp, 
    user_id int default 0
    );


CREATE TABLE "vendors" (
    vid integer primary key AUTOINCREMENT,
    name text not null ,
    www text not null default '',
    phone text not null default '',
    address text not null default '',
    city text not null default '',
    state text not null default '',
    country text not null default '',
    postal text not null default '',
    note text not null default '',
    remote_addr text not null default '', 
    user_id int default 0,
    modified date DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS "rmas" ;
CREATE TABLE "rmas" (
    id integer primary key AUTOINCREMENT,
    sid integer not null, -- server id
    vid integer not null, -- server id
    user_id integer not null, 
    rma_no text not null default '',
    description text not null default '',
    old_sn text not null default '',
    new_sn text not null default '',
    part_no text not null default '',
    tracking_no text not null default '',
    dc_ticket text not null default '',
    date_opened date DEFAULT CURRENT_TIMESTAMP,
    date_sent date,
    date_received date,
    date_replaced date 
);

DROP VIEW IF EXISTS rma_report;
CREATE VIEW rma_report as 
  select r.*, u.login, s.dc, s.hostname, s.sn as server_sn, s.rack, s.ru, v.name as vendor_name
  from  rmas r
  left outer join users u on r.user_id = u.id
  left outer join sview s on r.sid = s.id
  left outer join vendors v on r.vid = v.vid
;

CREATE TABLE master (
    rack    text,
    ru	    int,
    profile text,
    hostname text,
    sn text default '',
    ip_ipmi	text,
    ip_internal text
);

CREATE TABLE "audit_log" (
    id integer primary key AUTOINCREMENT,
    ts timestamp DEFAULT CURRENT_TIMESTAMP, 
    uid int,
    action text,
    ip text,
    msg text
);

CREATE TABLE "racks" (
    id integer primary key AUTOINCREMENT,
    rack integer,
    did int,
    x_pos text default '',
    y_pos text default '',
    rackunits int default 45,
    uid int default 0,
    ts timestamp DEFAULT CURRENT_TIMESTAMP, 
    vendor_id text default ''
);

CREATE TABLE pdus (
  id integer primary key AUTOINCREMENT,
  rid int,
  hostname text default '',
  asset_tag text default '',
  ip_address text default '',
  netmask text default '',
  gateway text default '',
  dns text default ''
);

CREATE TABLE "kinds" (
    kid integer primary key AUTOINCREMENT,
    name text not null,
    tbl text not null,
    fld text not null
);

CREATE TABLE "ipaddr" (
    iid integer primary key AUTOINCREMENT,
    kid integer,
    old_id integer, -- we'll kill this later
    ip32 integer,
    ipv4 text,
    mac text default '',
    ts timestamp DEFAULT CURRENT_TIMESTAMP, 
    uid int default 0
);

CREATE TABLE "racknet" (
    rid integer,
    vid integer,
    cidr text default '',
    actual text default '',
    subnet int default 24,
    min_ip int default 0,
    max_ip int default 0,
    first_ip text default '',
    last_ip text default '',
    unique(rid, vid)
);

CREATE TABLE "audit_racks" (
    id integer,
    rack integer,
    did int,
    x_pos text default '',
    y_pos text default '',
    rackunits int default 45,
    uid int default 0,
    ts timestamp DEFAULT CURRENT_TIMESTAMP, 
    vendor_id text default ''
);

CREATE TABLE users (
    id integer primary key AUTOINCREMENT,
    login text,
    firstname text not null,
    lastname text not null,
    email text not null,
    password text not null,
    admin int default 0
);

CREATE TABLE "vms" (
    id integer primary key AUTOINCREMENT,
    sid int,
    hostname text,
    profile text default '',
    note text default '', 
    private text default '', 
    public  text default '', 
    vip  text default '',
    modified timestamp DEFAULT CURRENT_TIMESTAMP, 
    remote_addr text default '', 
    uid int default 0
);

CREATE TABLE "audit_vms" (
    id integer,
    sid int,
    hostname text default '',
    profile text default '',
    note text default '', 
    private text default '', 
    public  text default '', 
    vip  text default '',
    modified timestamp, 
    remote_addr text default '', 
    uid int default 0
);

CREATE TABLE audit_servers (
    id integer,
    rid int,
    ru int,
    height int default 1,
    asset_tag text default '',
    vendor_sku text default '',
    sn text default '',
    profile default '', 
    hostname text not null COLLATE NOCASE,
    ip_internal text default '',
    ip_ipmi text default '',
    port_eth0 text default '',
    port_eth1 text default '',
    port_ipmi text default '',
    cable_eth0 text default '',
    cable_eth1 text default '',
    cable_ipmi text default '',
    cpu text default '',
    memory int default 0,  -- what unit should this be in?
    mac_port0 text  default '', 
    mac_port1 text default '',
    mac_ipmi text default '',
    note text default '', 
    modified timestamp CURRENT_TIMESTAMP, 
    uid int default 0, 
    remote_addr text default '', 
    ip_public text default '', 
    alias text default '', 
    assigned text default ''
);

CREATE TRIGGER servers_audit BEFORE UPDATE
ON servers
BEGIN
   INSERT INTO audit_servers select * from servers where id=old.id;
END;

CREATE TABLE "routers" (
    id integer primary key AUTOINCREMENT,
    rid int,
    ru int,
    height int default 1,
    make text,
    model text,
    asset_tag text,
    sku text,
    sn text,
    hostname text,
    ip_mgmt text default '',
    note text default '',
    modified timestamp DEFAULT CURRENT_TIMESTAMP, 
    uid int default 0, 
    remote_addr text default ''
);

CREATE TABLE "audit_routers" (
    id integer,
    rid int,
    ru int,
    height int default 1,
    make text,
    model text,
    asset_tag text,
    sku text,
    sn text,
    hostname text,
    ip_mgmt text default '',
    note text default '',
    modified timestamp DEFAULT CURRENT_TIMESTAMP, 
    uid int default 0, 
    remote_addr text default ''
);

CREATE TABLE "servers" (
    id integer primary key AUTOINCREMENT,
    rid int,
    ru int,
    height int default 1,
    asset_tag text default '',
    vendor_sku text default '',
    sn text default '',
    profile default '', 
    hostname text not null COLLATE NOCASE,
    ip_internal text default '',
    ip_ipmi text default '',
    port_eth0 text default '',
    port_eth1 text default '',
    port_ipmi text default '',
    cable_eth0 text default '',
    cable_eth1 text default '',
    cable_ipmi text default '',
    cpu text default '',
    memory int default 0,  -- what unit should this be in?
    mac_port0 text  default '', 
    mac_port1 text default '',
    mac_ipmi text default '',
    note text default '', 
    modified timestamp DEFAULT CURRENT_TIMESTAMP, 
    uid int default 0, 
    remote_addr text default '', 
    ip_public text default '', 
    alias text default '', 
    assigned text default ''
);

CREATE TABLE "vlans" (
    id integer primary key not null,
    did integer not null,
    name integer not null,
    profile string not null,
    gateway text not null,
    netmask text not null,
    route text not null,
    user_id int not null,
    modified timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE vmtmp (
    dc text,
    hostname text,
    vm1 text,
    vm2 text,
    vm3 text,
    vm4 text,
    vm5 text,
    vm6 text
    );

CREATE TABLE vmdetail (
    dc text,
    hostname text,
    private text,
    public text,
    vip text,
    note text
    );

CREATE TABLE vmorphans(
  rowid INT,
  dc TEXT,
  hostname TEXT,
  private TEXT,
  public TEXT,
  vip TEXT,
  note TEXT
);

CREATE VIEW vview as
  select d.name as dc, r.rack as rack, s.rid, v.*
  from vms v
  left outer join servers s on v.sid = s.id
  left outer join racks r on s.rid = r.id
  left outer join datacenters d on r.did = d.id;

CREATE VIEW sview as
  select d.name as dc, r.rack as rack, s.*
  from servers s
  left outer join racks r on s.rid = r.id
  left outer join datacenters d on r.did = d.id;

CREATE VIEW ipprivate as 
  select id, 'vm' as kind, 'private' as what,  dc, hostname, private as ip, note
  from vview where private > '';

CREATE VIEW ipinternal as 
  select id, 'server' as kind, 'internal' as what, dc, hostname, ip_internal as ip, note
  from sview s
  where ip_internal > '';

CREATE VIEW ippublic as 
  select id, 'vm' as kind, 'public' as what, dc, hostname, public as ip, note
  from vview where public > ''
  union
  select id, 'vm' as kind, 'vip' as what, dc, hostname, vip as ip, note
  from vview where vip > ''
  union
  select id, 'server' as kind, 'public' as what, dc, hostname, ip_public as ip, note
  from sview where ip_public > '';

CREATE VIEW ipipmi as 
  select id, 'server' as kind, 'ipmi' as what, dc, hostname, ip_ipmi as ip, note
  from sview where ip_ipmi > '';

CREATE VIEW ipinside as
  select * from ipinternal
  union
  select * from ipipmi
  union
  select * from ipprivate;

CREATE VIEW ipmstr as
  select * from ipinside
  union
  select * from ippublic;

CREATE VIEW ippool as 
  select ip_internal as ip from servers where ip_internal > ''
  union
  select ip_ipmi as ip from servers where ip_ipmi > ''
  union
  select private as ip from vms where private > '';

DELETE FROM sqlite_sequence;

CREATE INDEX iplookup on "ipaddr" (ip32);

CREATE INDEX slookup on "ipaddr" (old_id);

CREATE INDEX kname on kinds (name);

CREATE INDEX ikid on "ipaddr" (kid);

CREATE INDEX oldid on "ipaddr"(old_id);

CREATE INDEX range on racknet (min_ip,max_ip);

CREATE VIEW server_totals as select dc,count(*) as servers from sview group by dc order by dc;

CREATE VIEW server_summary as select * from server_totals union select '~TOTAL~' as datacenter, count(*) as servers from servers group by datacenter order by datacenter;

CREATE VIEW vmlist as
  select s.id as sid, v.id as vid, s.dc, s.hostname as server,v.hostname as vm, v.profile, v.private, v.public, v.vip, v.note
  from sview s, vms as v
    where s.id = v.sid;

CREATE VIEW piview as 
select b.old_id as id, b.ipv4 as ip_internal
from kinds a, ipaddr b
where a.name='internal'
  and a.kid=b.kid;

CREATE VIEW ipmiv as 
select b.old_id as id, b.ipv4 as ip_ipmi
from kinds a, ipaddr b
where a.name='ipmi'
  and a.kid=b.kid;

CREATE VIEW iview as 
select a.kid, b.id as old_id, b.ip_ipmi as ip
from kinds a, servers b
where a.name='ipmi';

CREATE VIEW fix1 as select iid, substr(ipv4,0,instr(ipv4,'.')) as fix, substr(ipv4,instr(ipv4,'.')+1) as remainder from ipaddr;

CREATE VIEW fix2 as select iid, substr(remainder,0,instr(remainder,'.')) as fix, substr(remainder,instr(remainder,'.')+1) as remainder from fix1;

CREATE VIEW fix34 as select iid, substr(remainder,0,instr(remainder,'.')) as fix, substr(remainder,instr(remainder,'.')+1) as remainder from fix2;

CREATE VIEW fixbinary as
select a.iid, (a.fix * 16777216)+(b.fix * 65536)+(c.fix * 256) + c.remainder as ip32
  from fix1 a, fix2 b, fix34 c
  where a.iid = b.iid
    and a.iid = c.iid;

CREATE VIEW rnet as
select c.name as dc, b.rack, a.*
from racknet a, racks b, datacenters c
where a.rid = b.id
  and b.did = c.id;

CREATE VIEW vlanview as 
select rid, vid, "start" as action, first_ip as ip from racknet
union
select rid, vid, "stop" as action, last_ip as ip from racknet;

CREATE TRIGGER ipadd AFTER INSERT ON "ipaddr"
BEGIN
    UPDATE IPADDR set ip32=(SELECT ipcalc FROM ipsub i where i.IID=NEW.IID) where iid=NEW.IID;
END;

CREATE VIEW lim1 as select rid, vid, action, substr(ip,0,instr(ip,'.')) as lim, substr(ip,instr(ip,'.')+1) as remainder, ip from vlanview;

CREATE VIEW lim2 as select rid, vid, action, substr(remainder,0,instr(remainder,'.')) as lim, substr(remainder,instr(remainder,'.')+1) as remainder from lim1;

CREATE VIEW lim34 as select rid, vid, action, substr(remainder,0,instr(remainder,'.')) as lim, substr(remainder,instr(remainder,'.')+1) as remainder from lim2;

CREATE VIEW lims as
select a.rid, a.vid, a.action, (a.lim * 16777216)+(b.lim * 65536)+(c.lim * 256) + c.remainder as ip32, a.ip
  from lim1 a, lim2 b, lim34 c
  where a.rid = b.rid
    and a.vid = b.vid
    and a.action = b.action
    and a.rid = c.rid
    and a.vid = c.vid
    and a.action = c.action;
    

CREATE VIEW limbinary as
select a.rid, a.vid, a.action, (a.lim * 16777216)+(b.lim * 65536)+(c.lim * 256) + c.remainder as ip32, a.ip
  from lim1 a, lim2 b, lim34 c
  where a.rid = b.rid
    and a.vid = b.vid
    and a.action = b.action
    and a.rid = c.rid
    and a.vid = c.vid
    and a.action = c.action;

CREATE VIEW binary_ips as
select a.rid, a.vid, a.ip32 as min_ip, b.ip32 as max_ip
  from limbinary a, limbinary b
  where a.rid = b.rid
    and a.vid = b.vid
    and a.action = "start"
    and b.action = "stop";

CREATE TRIGGER backfit_racknet 
INSTEAD OF UPDATE ON binary_ips 
BEGIN
  update racknet set min_ip=OLD.min_ip, max_ip=OLD.max_ip where rid=OLD.rid and vid=OLD.vid;
END;

CREATE VIEW rackips as
 select b.rack, b.did, a.*
 from racknet a, racks b
   where a.rid = b.id;


CREATE VIEW ipkinds as
select b.name as kind, a.*
from ipaddr a, kinds b
where a.kid = b.kid;

CREATE VIEW dcracks as
  select d.name as dc, r.did, r.rack 
  from racks r, datacenters d
  where r.did = d.id
  order by dc,rack;

CREATE VIEW nview as
  select d.name as dc, r.rack as rack, n.*
  from routers n
  left outer join racks r on n.rid = r.id
  left outer join datacenters d on r.did = d.id;

CREATE VIEW rackunits as
select * from (
    select dc, rack, 0 as nid, id as sid, rid, ru, height, hostname, alias, ip_ipmi as ipmi, ip_internal as internal, asset_tag, sn, note  from sview
    union
    select dc, rack, id as nid, 0 as sid, rid, ru, height,  hostname, '' as alias, '' as ipmi, ip_mgmt as ip_internal, asset_tag, sn, note from nview
) order by rid, ru desc;

CREATE VIEW inside_ip as
  select b.name as kind, a.* from ipaddr a, kinds b
  where a.kid in (select kid from kinds where name in ('internal','ipmi'))
  and b.kid = a.kid;

CREATE VIEW ip4server as
  select b.name as kind, a.* from ipaddr a, kinds b
  where a.kid in (select kid from kinds where name in ('internal','ipmi'))
  and b.kid = a.kid;

CREATE VIEW ipsub1  as select iid, substr(ipv4,0,instr(ipv4,'.')) as ipsub, substr(ipv4,instr(ipv4,'.')+1) as remainder from ipaddr;

CREATE VIEW ipsub2  as select iid, substr(remainder,0,instr(remainder,'.')) as ipsub, substr(remainder,instr(remainder,'.')+1) as remainder from ipsub1;

CREATE VIEW ipsub34 as select iid, substr(remainder,0,instr(remainder,'.')) as ipsub, substr(remainder,instr(remainder,'.')+1) as remainder from ipsub2;

CREATE VIEW ipvminternal as
  select * from ipaddr where kid=(select kid from kinds where name='vminternal');

CREATE VIEW ip4vm as
  select b.name as kind, a.* from ipaddr a, kinds b
  where a.kid in (select kid from kinds where name in ('vminternal'))
  and b.kid = a.kid;

CREATE VIEW vm_group_totals as select dc,profile, count(*) as VMs from vview group by dc,profile order by dc,profile;

CREATE VIEW vm_totals as select dc, '~TOTAL~' as profile, count(*) as VMs from vview group by dc order by dc;

CREATE VIEW vm_summary as select * from vm_group_totals union select * from vm_totals order by dc,profile;

CREATE VIEW servervms as
  select s.id, s.dc, s.hostname,
  group_concat(v.hostname) as vms,
  group_concat(v.id) as ids
  from sview as s
    left outer join vms as v on v.sid = s.id
      group by(s.id);

CREATE TRIGGER ipcalc AFTER UPDATE OF ipv4 ON "ipaddr"
BEGIN
    UPDATE IPADDR set ip32=(SELECT ipcalc FROM ipsub i where i.IID=OLD.IID) where iid=OLD.IID;
END
;

CREATE VIEW ipsub as
select d.*, ((a.ipsub * 16777216)+(b.ipsub * 65536)+(c.ipsub * 256) + c.remainder) as ipcalc
  from ipsub1 a, ipsub2 b, ipsub34 c, ipaddr d
  where a.iid = b.iid
    and a.iid = c.iid
    and a.iid = d.iid
;

CREATE VIEW bips as
select a.rid, a.vid, a.ip32 as min_ip, b.ip32 as max_ip, a.ip as addr_start, b.ip as addr_stop
  from limbinary a, limbinary b
  where a.rid = b.rid
    and a.vid = b.vid
    and a.action = "start"
    and b.action = "stop";

CREATE TRIGGER racks_audit BEFORE UPDATE
ON racks
BEGIN
   INSERT INTO audit_racks select * from racks where id=old.id;
END;

CREATE VIEW vlanrange as
select a.*, b.ip32 from
vlanview a, limbinary b
  where a.rid = b.rid
    and a.vid = b.vid
    and a.action = b.action;

CREATE TRIGGER startstop AFTER UPDATE OF actual ON racknet
BEGIN
update racknet set
first_ip = substr(actual,0,instr(actual,"-")),
last_ip = rtrim(substr(actual,0,instr(actual,"-")),'0123456789') || substr(actual,instr(actual,"-")+1)
where rid=old.rid;
END;

CREATE TRIGGER net_first AFTER UPDATE of first_ip ON racknet
BEGIN
    UPDATE racknet set min_ip=(SELECT min_ip FROM binary_ips i where i.RID=NEW.RID and i.VID=NEW.VID) where rid=NEW.RID and vid=NEW.VID;
END;

CREATE TRIGGER net_last AFTER UPDATE of last_ip ON racknet
BEGIN
    UPDATE racknet set max_ip=(SELECT max_ip FROM binary_ips i where i.RID=NEW.RID and i.VID=NEW.VID) where rid=NEW.RID and vid=NEW.VID;
END;

CREATE TRIGGER vm_changes BEFORE UPDATE 
ON "vms"
BEGIN
   INSERT INTO audit_vms select * from vms where id=old.id;
END;

CREATE VIEW vms_history as 
select s.*, u.login from (
select rowid,* from vms
union 
select rowid,* from audit_vms
order by id asc,modified desc) as s
left outer join users u on s.uid=u.id
order by rowid desc;

CREATE VIEW servers_history as 
select s.*, u.login from (
select rowid,* from servers
union 
select rowid,* from audit_servers
order by id asc,modified desc) as s
left outer join users u on s.uid=u.id
order by rowid desc;

CREATE VIEW routers_history as 
select s.*, u.login from (
select rowid,* from routers
union 
select rowid,* from audit_routers
order by id asc,modified desc) as s
left outer join users u on s.uid=u.id
order by rowid desc;

CREATE TRIGGER router_changes BEFORE UPDATE 
ON routers
BEGIN
   INSERT INTO audit_routers select * from routers where id=old.id;
END;

CREATE VIEW dcvlans as
 select d.name as dc, v.* 
 from datacenters d, vlans v
 where d.id = v.did;

CREATE TRIGGER vlan_load 
INSTEAD OF INSERT ON dcvlans 
BEGIN
  insert into vlans (did,name,gateway,netmask,route)
  values ((select id from datacenters where name like new.dc), 
  new.name, new.gateway, new.netmask, new.route)
  ;
END;

CREATE VIEW profiles as
select id,'server' as kind, dc, hostname, profile, ip_internal as ip, ip_public as public from sview
union
select id,'vm' as kind, dc, hostname, profile, private as ip, public from vview;

CREATE VIEW audit_view as select b.login,a.* 
  from audit_log a
  LEFT OUTER JOIN users b on a.uid=b.id;

CREATE VIEW vmbase as
   select a.dc, a.id as sid, b.vm1 as hostname from sview a, vmtmp b
   where a.dc = b.dc 
    and a.hostname = b.hostname
    and b.vm1 > ''
union
   select a.dc, a.id as sid, b.vm2 as hostname from sview a, vmtmp b
   where a.dc = b.dc 
    and a.hostname = b.hostname
    and b.vm2 > ''
union
   select a.dc, a.id as sid, b.vm3 as hostname from sview a, vmtmp b
   where a.dc = b.dc 
    and a.hostname = b.hostname
    and b.vm3 > ''
union
   select a.dc, a.id as sid, b.vm4 as hostname from sview a, vmtmp b
   where a.dc = b.dc 
    and a.hostname = b.hostname
    and b.vm4 > ''
union
   select a.dc, a.id as sid, b.vm5 as hostname from sview a, vmtmp b
   where a.dc = b.dc 
    and a.hostname = b.hostname
    and b.vm5 > ''
union
   select a.dc, a.id as sid, b.vm6 as hostname from sview a, vmtmp b
   where a.dc = b.dc 
    and a.hostname = b.hostname
    and b.vm6 > '';

CREATE VIEW vmload as
    select b.rowid as did, a.*, b.private, b.public, b.vip, b.note
    from vmbase a, vmdetail b
    where a.dc = b.dc
      and a.hostname = b.hostname;

CREATE VIEW vmbad as 
select rowid,* from vmdetail where rowid not in (select did from vmload);

drop view if exists server_dupes;
create view server_dupes as
select a.id as id, a.dc as dc, a.rack as rack, a.ru as ru,
        a.hostname as hostname ,a.alias as alias, a.profile as profile, 
        a.assigned, a.ip_ipmi as ip_ipmi, a.ip_internal as ip_internal, 
        a.ip_public as ip_putlic,a.asset_tag,a.vendor_sku as vender_sku,
        a.sn  as sn, b.id as dupe
	from sview a, sview b
	where a.rid = b.rid
	  and a.ru  = b.ru
	    and a.id != b.id
        ;

COMMIT;
