
DROP VIEW IF EXISTS racks_view;
CREATE VIEW racks_view as
	select s.name as site, r.*
	from racks r
	left outer join sites s on r.sti=s.sti
    order by site, r.rack
    ;

DROP VIEW IF EXISTS ips_view;
CREATE VIEW ips_view as
    select i.*, t.name as iptype
    from ips i 
    left outer join ip_types t on i.ipt = t.ipt
    ;

drop view if exists devices_view;
create view devices_view as
    select r.sti, r.site, r.rack, d.*, dt.name as devtype, t.tag
    from devices d
    left outer join racks_view r on d.rid = r.rid
    left outer join device_types dt on d.dti = dt.dti
    left outer join tags t on d.tid = t.tid
    ;

drop view if exists interfaces_view;
create view interfaces_view as 
    select i.*, p.iid, p.ipt, p.ip32, p.ipv4, t.name as iptype
    from interfaces i
    left outer join ips p on p.ifd = i.ifd
    left outer join ip_types t on p.ipt = t.ipt
    ;

DROP VIEW IF EXISTS skus_view;
CREATE VIEW skus_view as 
  select k.kid, k.vid, k.pti, k.mid, t.name as part_type, k.part_no, k.description, v.name as vendor, m.name as mfgr
  from  skus k
  left outer join mfgrs m on k.mid = m.mid
  left outer join part_types t on k.pti = t.pti
  left outer join vendors v on k.vid = v.vid
  ;


DROP VIEW IF EXISTS parts_view;
CREATE VIEW parts_view as 
   select p.pid, p.did, p.sti, ifnull(r.rma_id, 0) as rma_id, k.*, s.name as site, d.hostname, p.serial_no, p.asset_tag, p.unused, p.bad, p.location
   from parts p
   left outer join skus_view k on p.kid = k.kid
   left outer join rmas r on p.pid = r.old_pid
   left outer join devices d on p.did = d.did
   left outer join sites s on p.sti = s.sti
;

drop view if exists inventory;
create view inventory as
    select sti, kid, pti, site, count(kid) as qty, mfgr, part_no, part_type, description 
    from parts_view
    where unused = 1
    and  bad = 0
    group by site, kid, bad
    ;

DROP VIEW IF EXISTS "rmas_view" ;
CREATE VIEW rmas_view as 
    select r.*, p.description, p.serial_no as part_sn, p.part_no, s.hostname, s.sn as device_sn
    from rmas r
    left join devices s on r.did = s.did
    left join parts_view p on p.pid = r.old_pid
    ;


DROP VIEW IF EXISTS rma_report;
CREATE VIEW rma_report as 
  select r.*, u.login, s.site, s.hostname, s.sn as server_sn, s.rack, s.ru, v.name as vendor_name,
         b.serial_no as bad_serial, b.part_no as bad_partno
  from  rmas r
  left outer join users u on r.user_id = u.id
  left outer join devices_view s on r.did = s.did
  left outer join vendors v on r.vid = v.vid
;


DROP VIEW IF EXISTS vms_view; 
CREATE VIEW vms_view as
  select ifnull(r.sti,0) as sti, s.name as site, r.rack as rack, d.rid, d.hostname as server, v.*
  from vms v
  left outer join devices d on v.did = d.did
  left outer join racks r on s.rid = r.rid
  left outer join sites s on r.sti = s.sti
    ;

DROP VIEW IF EXISTS vms_view;
CREATE VIEW vms_view as
    select d.sti, d.rid, d.site, d.rack, d.ru, d.hostname as server, v.*
    from vms v
    left outer join devices_view d on v.did = d.did
;
DROP VIEW IF EXISTS vms_ips; 
CREATE VIEW vms_ips as
    select v.*, i.iid, i.ipt, i.ipv4, i.iptype
    from vms_view v
    left outer join ips_view i on v.vmi = i.vmi
    ;


DROP VIEW IF EXISTS rackspace; 
create view rackspace as select *,ru+height-1 as top from devices_view;


DROP VIEW IF EXISTS vlans_view;
CREATE VIEW vlans_view as
    select s.name as site, v.*
    from vlans v
    left outer join sites s on v.sti = s.sti
    order by v.sti, v.name
    ;

DROP VIEW IF EXISTS rack_vlans;
CREATE VIEW rack_vlans as 
select rid, vid, "start" as action, first_ip as ip from racknet
union
select rid, vid, "stop" as action, last_ip as ip from racknet;


-- 
-- totals for front page
--
drop view if exists summary;
create view summary as
with vcnt as (select sti, count(*) as vms from vms_view group by sti),
     scnt as (select sti, site, count(*) as servers from devices_view group by sti)
   select s.*, ifnull(v.vms,0) as vms from scnt s
  left outer join vcnt v on s.sti = v.sti
  ;

drop view if exists devices_network;
create view devices_network as
    select d.*, i.*
    from devices_view d
    left outer join interfaces_view i on d.did = i.did
    ;

drop view if exists devices_all_ips;
create view devices_all_ips as
    select r.sti, s.name as site, d.*, r.rack, dt.name as devtype, i.ipt, i.ipv4, t.name as iptype 
    from devices d
    left outer join racks r on d.rid = r.rid
    left outer join sites s on r.sti = s.sti
    left outer join device_types dt on d.dti = dt.dti
    left outer join interfaces f on d.did = f.did
    left outer join ips i on f.ifd = i.ifd
    left outer join ip_types t on i.ipt = t.ipt
    where ipv4 > ''
    ;

drop view if exists devices_ips;
create view devices_ips as
    select did, hostname, group_concat(ipv4, ', ') as ips
    from devices_all_ips
    --where iptype not in ('IPMI','VIP')
    where iptype = 'Internal'
    group by did
    ;

drop view if exists devices_mgmt;
create view devices_mgmt as
    select did, group_concat(ipv4, ', ') as mgmt
    from devices_all_ips
    where iptype in ('IPMI')
    group by did
    ;

DROP VIEW IF EXISTS devices_list;
CREATE VIEW devices_list as
  select d.*, i.ips, m.mgmt
  from devices_view d 
    left outer join devices_ips as i on d.did = i.did
    left outer join devices_mgmt as m on d.did = m.did 
    ;


DROP VIEW IF EXISTS ips_vms;
CREATE VIEW ips_vms as
    select i.*, v.hostname, 'VM' as devtype
    from ips_view i 
    left outer join vms v on i.vmi = v.vmi
    left outer join devices d on v.did = d.did
    where i.vmi > 0
;

DROP VIEW IF EXISTS ips_devices;
CREATE VIEW ips_devices as
    select r.sti, i.*, d.hostname, t.name as devtype
    from ips_view i 
    left outer join interfaces f on i.ifd = f.ifd
    left outer join devices d on f.did = d.did
    left outer join device_types t on d.dti = t.dti
    left outer join racks r on d.rid = r.rid
    where i.ifd > 0
    ;

DROP VIEW IF EXISTS ips_reserved;
CREATE VIEW ips_reserved as
    select v.sti, vli, i.ipt, v.site, i.iptype, ipv4, i.note  
        from ips_view i
    left outer join vlans_view v on i.vli = v.vli
        where i.ipv4 > '' 
        and i.ipttype = 'Reserved'
        ;

DROP VIEW IF EXISTS ips_list;
CREATE VIEW ips_list as
    select sti, did as id, rid, ipt, devtype as host, site, rack, ipv4 as ip, iptype, hostname, note  from devices_all_ips where ipv4 > ''
    union
    select sti, vmi as id, rid, ipt, 'VM' as host, site, rack, ipv4 as ip, iptype, hostname, note  from vms_ips where ipv4 > '' and iptype != 'VIP'
    union
    select sti, vli as id, rid, ipt, iptype as host, site, '' as rack, ipv4 as ip, iptype, iptype as hostname, note from ips_reserved 
    ;
