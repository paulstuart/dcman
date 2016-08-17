PRAGMA foreign_keys=OFF;
PRAGMA journal_mode = WAL;

BEGIN TRANSACTION;

DROP TABLE IF EXISTS "db_debug";
CREATE TABLE "db_debug" (
    log text,
    ts timestamp DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS sites;
CREATE TABLE "sites" (
    sti integer primary key,
    name text not null,
    address text,
    city text,
    state text,
    phone text,
    web text,
    postal text,
    country text,
    usr integer default 0, 
    ts timestamp DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS "racks";
CREATE TABLE "racks" (
    rid integer primary key,
    sti integer,
    rack integer,
    x_pos text default '',
    y_pos text default '',
    rackunits integer default 45,
    vendor_id text default '',
    note text default '',
    usr integer default 0, 
    ts timestamp DEFAULT CURRENT_TIMESTAMP, 
    FOREIGN KEY(sti) REFERENCES sites(sti)
);

DROP TABLE IF EXISTS "audit_racks";
CREATE TABLE "audit_racks" (
    rid integer,
    sti integer,
    rack integer,
    x_pos text,
    y_pos text,
    rackunits integer,
    vendor_id text,
    note text,
    usr integer default 0, 
    ts timestamp,
    FOREIGN KEY(sti) REFERENCES sites(sti)
);

DROP TABLE IF EXISTS "part_types" ;
CREATE TABLE "part_types" (
    pti integer primary key,
    name text not null COLLATE NOCASE,
    usr integer default 0, 
    ts date DEFAULT CURRENT_TIMESTAMP,
    unique (name)
);

insert into part_types (name) values('misc');


DROP TABLE IF EXISTS "mfgrs" ;
CREATE TABLE "mfgrs" (
    mid integer primary key,
    name text not null COLLATE NOCASE, -- full given name
    aka text, -- shortened and unique (to avoid dupes of different versions of same)
    url text, 
    note text, 
    usr integer default 0, 
    ts date DEFAULT CURRENT_TIMESTAMP,
    unique (name)
);

-- ensure we have a 'catch all'
insert into mfgrs (name) values('unknown');

CREATE TABLE "vendors" (
    vid integer primary key,
    name text,
    www text,
    phone text,
    address text,
    city text,
    state text,
    country text,
    postal text,
    note text,
    usr integer default 0, 
    ts date DEFAULT CURRENT_TIMESTAMP
);


DROP TABLE IF EXISTS "skus" ;
CREATE TABLE "skus" (
    kid integer primary key,
    vid integer,
    mid integer,
    pti integer,
    description text,
    part_no text, 
    sku text, 
    usr integer default 0, 
    ts date DEFAULT CURRENT_TIMESTAMP,
    unique (mid,description),
    FOREIGN KEY(vid) REFERENCES vendors(vid)
    FOREIGN KEY(mid) REFERENCES mfgrs(mid)
    FOREIGN KEY(pti) REFERENCES part_types(pti)
);

DROP TABLE IF EXISTS "parts" ;
CREATE TABLE "parts" (
    pid integer primary key,
    kid integer,
    vid integer,
    did integer,
    sti integer,
    unused integer default 0, -- boolean (1 is unused)
    bad    integer default 0, -- boolean (1 is bad)
    location text,
    serial_no text,
    asset_tag text, 
    cents integer default 0,  -- in cents to avoid floating point
    usr integer default 0, 
    ts date DEFAULT CURRENT_TIMESTAMP  
    ,
    FOREIGN KEY(sti) REFERENCES sites(sti)
    FOREIGN KEY(kid) REFERENCES skus(kid)
    FOREIGN KEY(did) REFERENCES devices(did)
    on update set null
    /*
    FOREIGN KEY(vid) REFERENCES vendors(vid)
    */
);


DROP TABLE IF EXISTS "rmas" ;
CREATE TABLE "rmas" (
    rmd integer primary key,
    sti integer, -- site id
    did integer, -- device id
    vid integer, -- vendor id
    old_pid integer,
    new_pid integer,
    vendor_rma text default '',
    ship_tracking text default '',
    recv_tracking text default '',
    jira text default '',
    dc_ticket text default '',
    dc_receiving text default '',
    note text default '',
    date_shipped date,
    date_received date,
    date_closed date,
    date_created date DEFAULT CURRENT_TIMESTAMP,
    usr integer default 0,
    FOREIGN KEY(sti) REFERENCES sites(sti)
);


DROP TABLE IF EXISTS "tags";
CREATE TABLE "tags" (
    tid integer primary key,
    tag text,
    unique(tag)
);

drop table if exists device_types;
CREATE TABLE "device_types" (
    dti integer primary key,
    name text not null COLLATE NOCASE
);

drop table if exists devices;
CREATE TABLE "devices" (
    did integer primary key,
    rid integer,    -- rack ID
    dti integer,    -- device type ID
    kid integer,    -- vendor sku ID
    tid integer,    -- tag ID
    ru  integer default 0,
    height    int default 1,
    hostname  text not null COLLATE NOCASE,
    alias     text,
    asset_tag text,
    sn        text,
    profile   text,
    assigned  text,
    note      text,
    usr integer default 0,
    ts  date DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(rid) REFERENCES racks(rid)
    FOREIGN KEY(dti) REFERENCES device_types(dti)
    FOREIGN KEY(kid) REFERENCES skus(kid)
    FOREIGN KEY(tid) REFERENCES tags(tid)
);

CREATE TABLE "vms" (
    vmi integer primary key,
    did integer,
    hostname text,
    profile text default '',
    note text default '',
    usr integer default 0,
    ts timestamp DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(did) REFERENCES devices(did)
);

DROP TABLE IF EXISTS "vlans";
CREATE TABLE "vlans" (
    vli integer primary key,
    sti integer,
    name integer not null,
    profile string not null,
    gateway text not null,
    netmask text not null,
    route text not null,
    note text,
    min_ip32 integer,
    max_ip32 integer,
    usr integer default 0,
    ts timestamp DEFAULT CURRENT_TIMESTAMP
);

drop table if exists interfaces;
CREATE TABLE "interfaces" (
    ifd integer primary key,
    did integer,
    mgmt integer default 0,   -- boolean (1 if mgmt port, 0 otherwise)
    port text,    -- eth0, eth1, etc
    mac text default '', 
    cable_tag text default '', 
    switch_port text default '',
    FOREIGN KEY(did) REFERENCES devices(did)
);

drop table if exists ip_types;
create table "ip_types" (
    ipt integer primary key,
    name text
);

drop table if exists ips;
create table "ips" (
    iid integer primary key,
    ifd integer,    -- interface ID
    vmi integer,    -- VM ID, if applicable
    ipt integer,    -- ip type ID
    vli integer,    -- VLAN ID, for reserved IPs
    ip32 integer default 0,    -- ip address as integer
    ipv4 text default '',    -- ip address as string
    note text,
    usr integer default 0, 
    ts date DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(ifd) REFERENCES interfaces(ifd)
    FOREIGN KEY(vmi) REFERENCES vms(vmi)
    FOREIGN KEY(ipt) REFERENCES ip_types(ipt)
    FOREIGN KEY(vli) REFERENCES vlans(vli)
);

drop index if exists ips_ip32;
create index ips_ip32 on ips(ip32);

drop index if exists ips_ipv4;
create index ips_ipv4 on ips(ipv4);

drop table if exists audit_devices;
CREATE TABLE "audit_devices" (
    did integer,
    kid integer,    -- vendor sku ID
    rid integer,    -- rack ID
    dti integer,    -- device type ID
    tid integer,    -- tag ID
    ru  integer,
    height integer,
    hostname  text,
    alias     text,
    asset_tag text,
    sn        text,
    profile   text,
    assigned  text,
    note      text,
    usr   integer default 0,
    ts  date
);

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    usr integer primary key,
    login text, -- optional unix login name
    firstname text,
    lastname text,
    email text,
    salt text,
    admin int default 0, 
    pw_hash text, -- bcrypt hashed password
    pw_salt text default (lower(hex(randomblob(32)))),
    apikey text default (lower(hex(randomblob(32))))
);

DROP TABLE IF EXISTS "providers";
CREATE TABLE "providers" (
    pri integer primary key,
    name text not null,
    contact text,
    phone text,
    email text,
    url text,
    note text,
    usr integer default 0,
    ts timestamp DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS "circuits";
CREATE TABLE "circuits" (
    cid integer primary key,
    sti integer,
    pri integer,
    circuit_id text,
    sub_id text,
    a_side_xcon text,
    a_side_handoff text,
    z_side_location text,
    z_side_xcon text,
    note text
);

DROP TABLE IF EXISTS "sub_circuits";
CREATE TABLE "sub_circuits" (
    sci integer primary key,
    cid integer,
    sub_circuit_id text,
    note text
);

/*

CID sub CID provider    A-side Data Center X-con ID A-side location A-side handoff  panel info  Z-side Data Center X-con ID Z-side location Z-side handoff  panel info  patch cable circuit type    service start   service term    purpose description Notes
10846533        Equinix EU  10846533    AM3                 Cat5    100M    August 2013 unknown OOB 
eqixCID11251637     AMS-IX/Equinix  11251637    AM3                 SC-SC SM    10G December 2013   unknown Peering at AMS-IX 80.249.210.220/21 
GI/Ethernet/00360258        GTT 14315011 needs confirmation, query sent to eqix AM3 "AM3:3:20106:PUBMATIC
PP:0104:170316 Ports 3-4"   dxcon-fg6fapza  AWS EU  unknown SC-SC SM    1G  April 2014  12  Xcon to AWS DirectConnect dxcon-fg6fapza    
GI/Ethernet/00372586        GTT 16516962    AM3 "AM3:3:20106:PUBMATIC
PP:0103:170315 Ports 7-8"   dxcon-fgit2je7  AWS EU  unknown SC-LC SM    10G September 2014  12  Xcon to AWS DirectConnect dxcon-fgit2je7 in London  
GI/X-Connect/00342668       GTT investigating w/ equinix    AM3     transit transit transit         2013??? 10/01/16        
GI/IP  TRANSIT/00337796     GTT investigating w/ equinix    AM3     transit transit transit         August 2014 24 months       
investigating w/ equinix            10866711-A  AM3                         August 2013         
investigating w/ equinix            10142073-A  AM3                         July 2013           

*/
CREATE VIRTUAL TABLE notes USING fts4(id, kind, hostname, note);
COMMIT;
