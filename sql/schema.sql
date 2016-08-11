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
    address text not null,
    city text not null,
    state text not null,
    phone text not null,
    web text not null,
    dcman text not null,
    remote_addr text not null default '', 
    modified timestamp, 
    user_id int default 0
);

DROP TABLE IF EXISTS "racks";
CREATE TABLE "racks" (
    rid integer primary key,
    rack integer,
    sti int,
    x_pos text default '',
    y_pos text default '',
    rackunits int default 45,
    uid int default 0,
    ts timestamp DEFAULT CURRENT_TIMESTAMP, 
    vendor_id text default '',
    FOREIGN KEY(sti) REFERENCES sites(sti)
);

DROP TABLE IF EXISTS "part_types" ;
CREATE TABLE "part_types" (
    pti integer primary key,
    name text not null COLLATE NOCASE,
    user_id integer not null default 0, 
    modified date DEFAULT CURRENT_TIMESTAMP,
    unique (name)
);

insert into part_types (name) values('misc');


DROP TABLE IF EXISTS "mfgrs" ;
CREATE TABLE "mfgrs" (
    mid integer primary key,
    name text not null COLLATE NOCASE, -- full given name
    aka text , -- shortened and unique
    url text , 
    user_id integer , 
    modified date DEFAULT CURRENT_TIMESTAMP,
    unique (name)
);
-- ensure we have a 'catch all'
insert into mfgrs (name) values('unknown');


DROP TABLE IF EXISTS "skus" ;
CREATE TABLE "skus" (
    kid integer primary key,
    vid integer,
    mid integer,
    pti integer,
    description text not null,
    part_no text , 
    user_id integer , 
    modified date DEFAULT CURRENT_TIMESTAMP,
    unique (mid,description),
    FOREIGN KEY(mid) REFERENCES mfgrs(mid)
    FOREIGN KEY(pti) REFERENCES part_types(pti)
);

CREATE TABLE "vendors" (
    vid integer primary key,
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

DROP TABLE IF EXISTS "parts" ;
CREATE TABLE "parts" (
    pid integer primary key,
    kid integer default 0,
    vid integer default 0,
    did integer default 0,
    sti integer default 0,
    --rma_id integer default 0, -- needs foreign key to rmas
    unused integer default 0, -- boolean (1 is unused)
    bad    integer default 0, -- boolean (1 is bad)
    location text,
    serial_no text,
    asset_tag text, 
    user_id integer not null default 0, 
    modified date DEFAULT CURRENT_TIMESTAMP  ,
    FOREIGN KEY(sti) REFERENCES sites(sti)
    FOREIGN KEY(kid) REFERENCES skus(kid)
    /*
    FOREIGN KEY(vid) REFERENCES vendors(vid)
    */
);


DROP TABLE IF EXISTS "rmas" ;
CREATE TABLE "rmas" (
    rma_id integer primary key,
    sti integer default 0, -- site id
    did integer default 0, -- device id
    vid integer default 0, -- vendor id
    old_pid integer default 0,
    new_pid integer default 0,
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
    user_id integer default 0 ,
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
    kid integer default 0,    -- vendor sku ID
    rid integer default 0,    -- rack ID
    dti integer default 0,    -- device type ID
    tid integer default 0,    -- tag ID
    ru  integer default 0,
    height    int default 1,
    hostname  text not null COLLATE NOCASE,
    alias     text,
    asset_tag text,
    sn        text,
    profile   text,
    assigned  text,
    note      text,
    user_id   integer not null default 0, 
    modified  date DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(rid) REFERENCES racks(rid)
    FOREIGN KEY(dti) REFERENCES device_types(dti)
);

CREATE TABLE "vms" (
    vmi integer primary key,
    did int,
    hostname text,
    profile text default '',
    note text default '',
    user_id int default 0,
    modified timestamp DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(did) REFERENCES devices(did)
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
    vli integer,    -- VLAN ID, for reserved IPs
    ipt integer,    -- ip type ID
    ip32 integer default 0,    -- ip address as integer
    ipv4 text default '',    -- ip address as string
    note text,
    user_id integer not null default 0, 
    modified date DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(ifd) REFERENCES interfaces(ifd)
    FOREIGN KEY(vmi) REFERENCES vms(vmi)
    FOREIGN KEY(ipt) REFERENCES ip_types(ipt)
);

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
    user_id   integer,
    modified  date
);

DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id integer primary key,
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

-- rack specific vlans
DROP TABLE IF EXISTS "vlans";
CREATE TABLE "vlans" (
    vli integer primary key,
    sti integer not null,
    name integer not null,
    profile string not null,
    gateway text not null,
    netmask text not null,
    route text not null,
    note text,
    user_id int not null,
    modified timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE VIRTUAL TABLE notes USING fts4(id, kind, hostname, note);
COMMIT;
