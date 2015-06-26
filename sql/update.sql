drop table if exists oldservers;
alter table servers rename to oldservers;

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
    assigned text default '',
    unique(rid, ru)
);

delete from oldservers where rowid in (select a.rowid from oldservers a, oldservers b where a.rid=b.rid and a.ru = b.ru and a.rowid < b.rowid order by a.rid,a.ru);

insert into servers select * from oldservers;

DROP TABLE if exists audit_servers;

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

drop trigger if exists servers_audit;
CREATE TRIGGER servers_audit BEFORE UPDATE
ON servers
BEGIN
       INSERT INTO audit_servers select * from servers where id=old.id;
END;

ALTER TABLE datacenters rename to dcX;

CREATE TABLE datacenters (
    id integer primary key AUTOINCREMENT,
    name text,
    location text,
    modified timestamp DEFAULT CURRENT_TIMESTAMP, 
    uid int default 0, 
    remote_addr text default ''
);

insert into datacenters (id,name,location) select id,name,location from dcX;
drop table dcX;

update datacenters set name="NYC2" where name="NY7";

drop view if exists rackunits;
CREATE VIEW rackunits as
select * from (
    select dc, rack, 0 as nid, id as sid, rid, ru, height, hostname, alias, ip_ipmi as ipmi, ip_internal as internal  from sview
    union
    select dc, rack, id as nid, 0 as sid, rid, ru, height,  hostname, '' as alias, '' as ipmi, ip_mgmt as ip_internal from nview
) order by rid, ru desc;


drop table if exists auditing;
create table auditing (
    hostname text,
    remote_addr text,
    ips text,
    eth0 text,
    eth1 text,
    sn text,
    asset text,
    ipmi_ip text,
    ipmi_mac text,
    cpu text,
    mem int,
    ts timestamp DEFAULT CURRENT_TIMESTAMP,
    unique (eth0)
);

