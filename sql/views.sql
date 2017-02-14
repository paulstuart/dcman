
DROP VIEW IF EXISTS sessions_view;
CREATE VIEW "sessions_view" as
    select s.*, u.email
    from sessions s
    left outer join users u on s.usr = u.usr
    order by s.ts desc
    ;

DROP VIEW IF EXISTS racks_view;
CREATE VIEW racks_view as
	select s.name as site, r.*
	from racks r
	left outer join sites s on r.sti=s.sti
    order by site, r.rack
    ;

DROP VIEW IF EXISTS vlans_view;
CREATE VIEW vlans_view as
    select s.name as site, v.*
    from vlans v
    left outer join sites s on v.sti = s.sti
    order by v.sti, v.name
    ;

DROP VIEW IF EXISTS vlans_first;
CREATE VIEW vlans_first as
    with ipaddr(ipv4, vli) as (
            select ifnull(starting,route) as ipv4, vli from vlans
        ),
        oct1(ipv4, vli, o1, rem1) as (
            select ipv4, vli, substr(ipv4,0,instr(ipv4,'.')) as o1, 
                substr(ipv4,instr(ipv4,'.')+1) as rem1
                from ipaddr
        ),
        oct2(ipv4, vli, o1, o2, rem2) as (
            select ipv4, vli, o1,
                substr(rem1,0,instr(rem1,'.')) as o2, 
                substr(rem1,instr(rem1,'.')+1) as rem2
                from oct1
        ),
        oct3(ipv4, vli, o1, o2, o3, o4) as (
            select ipv4, vli, o1, o2,
                substr(rem2,0,instr(rem2,'.')) as o3, 
                substr(rem2,instr(rem2,'.')+1) as o4
                from oct2
        ),
        calculated(vli, ipv4, ipcalc) as (
            select
                vli, ipv4, ((o1 << 24) + (o2 << 16) + (o3 << 8) + o4) as minip
            from oct3
        ),
        net1(vli, o1, rem1) as (
            select vli, substr(netmask,0,instr(netmask,'.')) as o1, 
                substr(netmask,instr(netmask,'.')+1) as rem1
                from vlans
        ),
        net2(vli, o1, o2, rem2) as (
            select vli, o1,
                substr(rem1,0,instr(rem1,'.')) as o2, 
                substr(rem1,instr(rem1,'.')+1) as rem2
                from net1
        ),
        net3(vli, o1, o2, o3, o4) as (
            select vli, o1, o2,
                substr(rem2,0,instr(rem2,'.')) as o3, 
                substr(rem2,instr(rem2,'.')+1) as o4
                from net2
        ),
        netasm(vli, ncalc) as (
            select
                vli, ((o1 << 24) + (o2 << 16) + (o3 << 8) + o4) as calc
            from net3
        ),
        netcalc(vli, ncalc, range) as (
            select vli, ncalc, ~(0xffffffff00000000 | ncalc) - 1 as maxip 
            from netasm
        ),
        merge(vli, starting, ipcalc, ncalc, network, range) as (
            select c.*, n.ncalc, (c.ipcalc & n.ncalc) as network, range
               from calculated c
               left outer join netcalc n on c.vli = n.vli
        )
       select v.name, m.*, ipcalc|range as broadcast 
            from merge m
              left outer join vlans v on m.vli = v.vli
    ;


DROP VIEW IF EXISTS vlans_calc;
CREATE VIEW vlans_calc as
    with oct1(ipv4, vli, o1, rem1) as (
            select starting as ipv4, vli, substr(starting,0,instr(starting,'.')) as o1, 
                substr(starting,instr(starting,'.')+1) as rem1
                from vlans
        ),
        oct2(ipv4, vli, o1, o2, rem2) as (
            select ipv4, vli, o1,
                substr(rem1,0,instr(rem1,'.')) as o2, 
                substr(rem1,instr(rem1,'.')+1) as rem2
                from oct1
        ),
        oct3(ipv4, vli, o1, o2, o3, o4) as (
            select ipv4, vli, o1, o2,
                substr(rem2,0,instr(rem2,'.')) as o3, 
                substr(rem2,instr(rem2,'.')+1) as o4
                from oct2
        ),
        calculated(vli, ipv4, ipcalc) as (
            select
                vli, ipv4, ((o1 << 24) + (o2 << 16) + (o3 << 8) + o4) as minip
            from oct3
        ),
        net1(vli, o1, rem1) as (
            select vli, substr(netmask,0,instr(netmask,'.')) as o1, 
                substr(netmask,instr(netmask,'.')+1) as rem1
                from vlans
        ),
        net2(vli, o1, o2, rem2) as (
            select vli, o1,
                substr(rem1,0,instr(rem1,'.')) as o2, 
                substr(rem1,instr(rem1,'.')+1) as rem2
                from net1
        ),
        net3(vli, o1, o2, o3, o4) as (
            select vli, o1, o2,
                substr(rem2,0,instr(rem2,'.')) as o3, 
                substr(rem2,instr(rem2,'.')+1) as o4
                from net2
        ),
        netasm(vli, ncalc) as (
            select
                vli, ((o1 << 24) + (o2 << 16) + (o3 << 8) + o4) as calc
            from net3
        ),
        netcalc(vli, ncalc, range) as (
            select vli, ncalc, ~(0xffffffff00000000 | ncalc) - 1 as maxip 
            from netasm
        ),
        merge(vli, starting, ipcalc,ncalc,network,range) as (
            select c.*, n.ncalc, (c.ipcalc & n.ncalc) as network, range
               from calculated c
               left outer join netcalc n on c.vli = n.vli
        )

       select *, ipcalc|range as broadcast 
            from merge
    ;

DROP VIEW IF EXISTS ips_view;
CREATE VIEW ips_view as
    select i.*, v.name as vlan, t.name as iptype
    from ips i 
    left outer join ip_types t on i.ipt = t.ipt
    left outer join vlans v on i.vli = v.vli
    ;

DROP VIEW IF EXISTS ips_reserved;
CREATE VIEW ips_reserved as
    select i.iid, i.ipt, i.vli, v.sti, null as rid, 
           v.site, v.name as vlan, i.iptype, i.ip32, i.ipv4, i.note, i.usr, i.ts, u.email as username
    from ips_view i 
    left outer join vlans_view v on i.vli = v.vli
    left outer join users u on i.usr = u.usr
	where i.ifd is null 
    and i.vmi is null
    ;

DROP VIEW IF EXISTS ips_calc;
CREATE VIEW ips_calc as
    with oct1(ipv4, iid, o1, rem1) as (
            select ipv4, iid, substr(ipv4,0,instr(ipv4,'.')) as o1, 
                substr(ipv4,instr(ipv4,'.')+1) as rem1
                from ips
        ),
        oct2(iid, ipv4, o1, o2, rem2) as (
            select iid, ipv4, o1,
                substr(rem1,0,instr(rem1,'.')) as o2, 
                substr(rem1,instr(rem1,'.')+1) as rem2
                from oct1
        ),
        oct3(iid, ipv4, o1, o2, o3, o4) as (
            select iid, ipv4, o1, o2,
                substr(rem2,0,instr(rem2,'.')) as o3, 
                substr(rem2,instr(rem2,'.')+1) as o4
                from oct2
        ),
        calculated(iid, ipv4, ipcalc) as (
            select
                iid, ipv4, ((o1 << 24) + (o2 << 16) + (o3 << 8) + o4) as ipcalc
            from oct3
        )
    select * from calculated
    ;

-- create a list of all used IPs
-- merge in ip right before vlan range so we have
-- a starting point if no IPs in that range taken  
drop view if exists ips_taken;
create view ips_taken as
   with used(vli, ip32) as (
       select vli, ip32 from ips where ip32 > 0
   union
       select vli, ipcalc - 1 as ip32 from vlans_first  -- TODO: vlans_first is calc'd, data should be updated in vlans table
   )
   select distinct * from used
   ;

/*
drop view if exists ips_missing;
create view ips_missing as
   select i.*, i.ip32+1 as ip33
   from ips i
   left outer join ips j on j.ip32 = (i.ip32 + 1)
   where i.ip32 > 0
     and j.ip32 is null
    ;
*/

drop view if exists ips_missing;
create view ips_missing as
   select i.*, i.ip32+1 as ip33
   from ips_taken i
   left outer join ips_taken j on j.ip32 = (i.ip32 + 1)
   where i.ip32 > 0
     and j.ip32 is null
    ;

drop view if exists ips_available;
create view ips_available as
   select v.sti, i.vli, v.name as vlan, ip33 as ip32,
    (ip33 >> 24) || '.' || ((ip33 >> 16) & 255) || '.' || ((ip33 >> 8) & 255) || '.' || (ip33 & 255) as ipv4
    from ips_missing i
    left outer join vlans v on i.vli = v.vli
    order by ip32
   ;

drop view if exists ips_next;
create view ips_next as
   with filter(sti, vli, vlan, ip32, ipv4) as (
       select sti, vli, vlan, min(ip32) as ip32, ipv4 
        from ips_available
        where vli > 0
        group by vli
    )
   select sti,vli,vlan,ipv4 from filter
    ;

drop view if exists devices_view;
create view devices_view as
    select r.sti, r.site, r.rack, d.*, dt.name as devtype, p.profile, m.name as make
    from devices d
    left outer join racks_view r on d.rid = r.rid
    left outer join device_types dt on d.dti = dt.dti
    left outer join mfgrs m on d.mid = m.mid
    left outer join profiles p on d.prd = p.prd
    ;

drop view if exists interfaces_view;
create view interfaces_view as 
    select i.*, p.iid, p.ipt, p.vli, p.ip32, p.ipv4, t.name as iptype
    from interfaces i
    left outer join ips p on p.ifd = i.ifd
    left outer join ip_types t on p.ipt = t.ipt
    ;

DROP VIEW IF EXISTS skus_view;
CREATE VIEW skus_view as 
  select k.*, t.name as part_type, v.name as vendor, m.name as mfgr
  from  skus k
  left outer join mfgrs m on k.mid = m.mid
  left outer join part_types t on k.pti = t.pti
  left outer join vendors v on k.vid = v.vid
  ;


DROP VIEW IF EXISTS parts_view;
CREATE VIEW parts_view as 
   select p.pid, p.did, p.sti, ifnull(r.rmd, 0) as rmd, 
        k.kid, p.vid, k.mid, k.pti, k.description, k.part_no, k.sku, k.part_type, v.name as vendor, k.mfgr,
        s.name as site, d.hostname, d.sn as device_sn, p.serial_no, p.asset_tag, p.unused, p.bad, 
        p.location, p.cents, 
        round(p.cents/100.0,2) as price
   from parts p
   left outer join skus_view k on p.kid = k.kid
   left outer join rmas r on p.pid = r.old_pid
   left outer join devices d on p.did = d.did
   left outer join sites s on p.sti = s.sti
   left outer join vendors v on p.vid = v.vid
;

drop view if exists inventory;
create view inventory as
    select sti, kid, pti, site, count(kid) as qty, mfgr, part_no, part_type, description, sum(cents) as cents, sum(price) as price 
    from parts_view
    where unused = 1
    and  bad = 0
    group by site, kid, bad
    ;

DROP VIEW IF EXISTS "rmas_view" ;
CREATE VIEW rmas_view as 
    select r.*, p.description, p.serial_no as part_sn, p.part_no, p.site, s.hostname, s.sn as device_sn, p.vendor
    from rmas r
    left join devices s on r.did = s.did
    left join parts_view p on p.pid = r.old_pid
    ;

DROP VIEW IF EXISTS vms_view; 
CREATE VIEW vms_view as
  select ifnull(r.sti,0) as sti, s.name as site, r.rack as rack, d.rid, d.hostname as server, v.*
  from vms v
  left outer join devices d on v.did = d.did
  left outer join racks r on d.rid = r.rid
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
    select v.*, i.iid, i.ipt, i.ipv4, i.iptype, i.vlan
    from vms_view v
    left outer join ips_view i on v.vmi = i.vmi
    ;

drop view if exists vms_list;
create view vms_list as
    select vmi, did, sti, rid, site, rack, ru, server, hostname, profile, note, usr, ts, 
    group_concat(ipv4, ', ') as ips
    from vms_ips
    where ipt not in (select ipt from ip_types where multi > 0)
    group by vmi
    ;

DROP VIEW IF EXISTS rackspace; 
create view rackspace as select *,ru+height-1 as top from devices_view;

-- 
-- totals for front page
--
drop view if exists summary;
create view summary as
   with vcnt as (select sti, count(*) as vms from vms_view group by sti),
        scnt as (select sti, site, count(*) as servers from devices_view group by sti)
   select s.*, ifnull(v.vms,0) as vms from scnt s
      left outer join vcnt v on s.sti = v.sti
      where s.sti is not null
  ;

drop view if exists devices_network;
create view devices_network as
    select d.*, i.*
    from devices_view d
    left outer join interfaces_view i on d.did = i.did
    ;

drop view if exists devices_all_ips;
create view devices_all_ips as
    select r.sti, s.name as site, d.*, r.rack, dt.name as devtype, i.ipt, i.vli, i.ipv4, t.name as iptype, v.name as vlan 
    from devices d
    left outer join racks r on d.rid = r.rid
    left outer join sites s on r.sti = s.sti
    left outer join device_types dt on d.dti = dt.dti
    left outer join interfaces f on d.did = f.did
    left outer join ips i on f.ifd = i.ifd
    left outer join ip_types t on i.ipt = t.ipt
    left outer join vlans v on i.vli = v.vli
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
    where ipt in (select ipt from ip_types where mgmt > 0)
    group by did
    ;

DROP VIEW IF EXISTS devices_list;
CREATE VIEW devices_list as
  select d.*, i.ips, m.mgmt
  from devices_view d 
    left outer join devices_ips as i on d.did = i.did
    left outer join devices_mgmt as m on d.did = m.did 
    order by sti, rack, ru desc
    ;

drop view if exists devices_public_ips;
create view devices_public_ips as
    select *
    from devices_all_ips
    where iptype in ('Public')
    ;

DROP VIEW IF EXISTS mactable;
CREATE VIEW mactable as
    select a.mac, a.hostname, a.site, replace(ifnull(a.profile,'-'),' ', '_') as profile, a.ipv4 as ip_internal, ifnull(b.ipv4, '-') as ip_public
        from devices_network a
        left outer join devices_public_ips b on a.did = b.did
        where a.iptype = 'Internal'
          and a.iid > 0
          and a.ipv4 > ''
          and a.mac > ''
          and a.port == 0
    ;

-- data needed to pxeboot a server (some is just for confirmation)
DROP VIEW IF EXISTS pxedevice;
CREATE VIEW pxedevice as
    select d.sti, d.did, d.rid, d.site, d.rack, d.ru, d.hostname, d.profile, 
            i.mac, i.ipv4 as ip, m.ipv4 as ipmi, d.note, d.restricted, p.profile, p.script, s.pxehost
    from devices_view d
    left outer join interfaces_view i on d.did = i.did
    left outer join interfaces_view m on d.did = m.did
    left outer join profiles p on d.prd = p.prd
    left outer join sites s on d.sti = s.sti
    where i.ip32 > 0 and i.ip32 < 184549375 -- 10.255.255.255
      and i.port=0
      and i.mgmt=0
      and m.port=0
      and m.mgmt=1
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

DROP VIEW IF EXISTS ips_list;
CREATE VIEW ips_list as
    select vlan, sti, did as id, rid, ipt, devtype as host, site, rack, ipv4 as ip, iptype, hostname, note  from devices_all_ips where ipv4 > ''
    union
    select vlan, sti, vmi as id, rid, ipt, 'VM' as host, site, rack, ipv4 as ip, iptype, hostname, note  from vms_ips where ipv4 > '' and iptype != 'VIP'
    union
    select vlan, sti, vli as id, rid, ipt, iptype as host, site, '' as rack, ipv4 as ip, iptype, iptype as hostname, note from ips_reserved 
    ;

DROP VIEW IF EXISTS "circuits_view";
CREATE VIEW "circuits_view" as
    select s.name as site, p.name as provider, c.*
    from circuits c
    left outer join providers p on c.pri = p.pri
    left outer join sites s on c.sti = s.sti
    ;

DROP VIEW IF EXISTS "circuits_list";
CREATE VIEW "circuits_list" as
    select c.*, s.sub_circuit_id, s.note as sub_note
    from circuits_view c
    left outer join sub_circuits s on s.cid = c.cid
    ;

drop view if exists devices_adjust;
create view devices_adjust as
    select d.*, (d.ru + d.height - 1) as space 
    from devices d
    ;

drop view if exists devices_history;
CREATE VIEW devices_history as 
    with all_devices as (
        select * from devices
        union 
        select * from audit_devices
    ) 
    select d.*, r.sti, r.site, r.rack, dt.name as devtype, t.tag, u.email 
    from all_devices d
    left outer join racks_view r on d.rid = r.rid
    left outer join device_types dt on d.dti = dt.dti
    left outer join tags t on d.tid = t.tid
    left outer join users u on d.usr=u.usr
    order by did asc, version desc
    ;

drop view if exists vms_history;
CREATE VIEW vms_history as 
    with all_vms as (
        select * from vms
        union 
        select * from audit_vms
    ) 
    select v.*, d.sti, d.rid, d.site, d.rack, d.ru, d.hostname as server, u.email 
    from all_vms v
    left outer join devices_view d on v.did = d.did
    left outer join users u on d.usr=u.usr
    order by vmi asc, version desc
    ;

-- view mirrors users table, used to filter updates via view trigger
DROP VIEW IF EXISTS users_view;
CREATE VIEW users_view as select * from users;

drop view if exists ips_vlan;
create view ips_vlan as
    select i.*,v.name as vlan 
    from ips i
    left outer join vlans_first v
        where i.ip32 > v.network
          and i.ip32 < v.broadcast
        ;
    
