
CREATE TABLE "tags" (
    tid integer primary key AUTOINCREMENT,
    tag text
);

drop table if exists "old_servers";
drop table if exists "old_audit_servers";
drop TRIGGER if exists servers_audit;

alter table servers rename to "old_servers";

CREATE TABLE "servers" (
    id integer primary key AUTOINCREMENT,
    rid int,
    tid int default 0,
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
    mac_eth0 text  default '', 
    mac_eth1 text default '',
    mac_ipmi text default '',
    pdu_a text default '',
    pdu_b text default '',
    outlet text default '',
    note text default '', 
    ip_public text default '', 
    alias text default '', 
    assigned text default '',
    kernel text default '',
    release text default '',
    modified timestamp DEFAULT CURRENT_TIMESTAMP, 
    uid int default 0, 
    remote_addr text default '' /**,
     unique(rid, ru) **/
);

insert into servers
    (
    id,
    rid,
    ru,
    height,
    asset_tag,
    vendor_sku,
    sn,
    profile,
    hostname,
    ip_internal,
    ip_ipmi,
    port_eth0 ,
    port_eth1,
    port_ipmi,
    cable_eth0,
    cable_eth1,
    cable_ipmi,
    cpu,
    memory,
    mac_eth0,
    mac_eth1,
    mac_ipmi,
    pdu_a,
    pdu_b,
    outlet,
    note,
    ip_public,
    alias,
    assigned,
    kernel,
    release,
    modified,
    uid,
    remote_addr
    )
    select
        id,
        rid,
        ru,
        height,
        asset_tag,
        vendor_sku,
        sn,
        profile,
        hostname,
        ip_internal,
        ip_ipmi,
        port_eth0 ,
        port_eth1,
        port_ipmi,
        cable_eth0,
        cable_eth1,
        cable_ipmi,
        cpu text,
        memory,
        mac_eth0,
        mac_eth1,
        mac_ipmi,
        pdu_a,
        pdu_b,
        outlet,
        note,
        ip_public,
        alias,
        assigned,
        kernel,
        release,
        modified,
        uid,
        remote_addr
    from old_servers
    ;

CREATE TRIGGER servers_audit BEFORE UPDATE
ON servers
BEGIN
       INSERT INTO audit_servers select * from servers where id=old.id;
END;

ALTER TABLE "audit_servers" rename to old_audit_servers;
CREATE TABLE "audit_servers" (
    id integer,
    rid int,
    tid int,
    ru int,
    height int,
    asset_tag text,
    vendor_sku text,
    sn text,
    profile text,
    hostname text COLLATE NOCASE,
    ip_internal text,
    ip_ipmi text,
    port_eth0 text,
    port_eth1 text,
    port_ipmi text,
    cable_eth0 text,
    cable_eth1 text,
    cable_ipmi text,
    cpu text,
    memory int,
    mac_eth0 text,
    mac_eth1 text,
    mac_ipmi text,
    pdu_a text,
    pdu_b text,
    outlet text,
    note text,
    ip_public text,
    alias text,
    assigned text,
    kernel text,
    release text,
    modified timestamp DEFAULT CURRENT_TIMESTAMP, 
    uid int,
    remote_addr
);

insert into audit_servers
    (
    id,
    rid,
    ru,
    height,
    asset_tag,
    vendor_sku,
    sn,
    profile,
    hostname,
    ip_internal,
    ip_ipmi,
    port_eth0 ,
    port_eth1,
    port_ipmi,
    cable_eth0,
    cable_eth1,
    cable_ipmi,
    cpu,
    memory,
    mac_eth0,
    mac_eth1,
    mac_ipmi,
    pdu_a,
    pdu_b,
    outlet,
    note,
    ip_public,
    alias,
    assigned,
    kernel,
    release,
    modified,
    uid,
    remote_addr
    )
    select
        id,
        rid,
        ru,
        height,
        asset_tag,
        vendor_sku,
        sn,
        profile,
        hostname,
        ip_internal,
        ip_ipmi,
        port_eth0 ,
        port_eth1,
        port_ipmi,
        cable_eth0,
        cable_eth1,
        cable_ipmi,
        cpu text,
        memory,
        mac_eth0,
        mac_eth1,
        mac_ipmi,
        pdu_a,
        pdu_b,
        outlet,
        note,
        ip_public,
        alias,
        assigned,
        kernel,
        release,
        modified,
        uid,
        remote_addr
    from old_servers
    ;
DROP VIEW if exists sview;

CREATE VIEW sview as
  select d.name as dc, r.rack as rack, s.*, t.tag
  from servers s
  left outer join racks r on s.rid = r.id
  left outer join datacenters d on r.did = d.id
  left outer join tags t on s.tid = t.tid
;

DROP TRIGGER IF EXISTS sview_insert;
CREATE TRIGGER sview_insert INSTEAD OF INSERT ON sview 
BEGIN
  insert into servers (rid, ru, hostname, sn, asset_tag, ip_internal, ip_ipmi, mac_eth0,
	cable_ipmi, port_ipmi, cable_eth0, port_eth0, cable_eth1, port_eth1, pdu_a, pdu_b
	) 
  values ((select id from rview where dc=NEW.dc and rack=NEW.rack),
	NEW.ru, NEW.hostname, NEW.sn, NEW.asset_tag, NEW.ip_internal, NEW.ip_ipmi, NEW.mac_eth0,
	NEW.cable_ipmi, NEW.port_ipmi, NEW.cable_eth0, NEW.port_eth0, NEW.cable_eth1, NEW.port_eth1, 
	NEW.pdu_a, NEW.pdu_b
	);
END;
/*
*/

insert into tags (tag) values ('Normal');
insert into tags (tag) values ('Super');
insert into tags (tag) values ('Unused');
