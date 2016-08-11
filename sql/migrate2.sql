--.echo on

/*
select count(*) from servers;
.exit
*/

 drop view newservers;

--
-- redo datacenters to use DCD instead of ID
--

PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
DROP TABLE IF EXISTS datacenters;
CREATE TABLE "datacenters" (
    dcd integer primary key,
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
INSERT INTO "datacenters" VALUES(1,'AMS','','Amsterdam','','','','','','2015-07-24 23:04:09',0);
INSERT INTO "datacenters" VALUES(2,'SFO','','San Francisco','','','','','','2015-07-24 23:04:09',0);
INSERT INTO "datacenters" VALUES(3,'NYC','','New York City','','','','','','2015-07-24 23:04:09',0);
INSERT INTO "datacenters" VALUES(4,'NY7','','New Jersey','','','','','10.100.2.224','2016-03-22 17:54:01.770503649',1);
INSERT INTO "datacenters" VALUES(5,'SV3','1735 Lundy Avenue','San Jose','CA','','','','10.100.2.248','2015-09-22 20:37:58.755598077',1);
COMMIT;

--
-- redo racks to use RID instead of ID
--

drop table if exists old_racks;
alter table racks rename to old_racks;

CREATE TABLE "racks" (
    rid integer primary key,
    rack integer,
    dcd int,
    x_pos text default '',
    y_pos text default '',
    rackunits int default 45,
    uid int default 0,
    ts timestamp DEFAULT CURRENT_TIMESTAMP, 
    vendor_id text default ''
);

DROP TRIGGER IF EXISTS racks_audit;

insert into racks select * from old_racks;
drop table if exists old_racks;

CREATE TRIGGER racks_audit BEFORE UPDATE
ON racks
BEGIN
   INSERT INTO audit_racks select * from racks where id=old.id;
END;

--
-- VM audit
--

DROP TABLE IF EXISTS "old_audit_vms";
ALTER TABLE "audit_vms" rename to old_audit_vms;
 
CREATE TABLE "audit_vms" (
        vmi integer,
        sid integer,
        hostname text,
        profile text,
        note text,
        private text,
        public  text,
        vip  text,
        modified timestamp, 
        remote_addr text default '', 
        uid int default 0
);

insert into audit_vms select * from old_audit_vms;
DROP TABLE IF EXISTS "audit_vms";


DROP TABLE IF EXISTS old_vms;
ALTER TABLE "vms" rename to old_vms;
 
CREATE TABLE "vms" (
        vmi integer primary key,
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

insert into vms select * from old_vms;
DROP TABLE IF EXISTS old_vms;

DROP TRIGGER IF EXISTS vm_changes;
CREATE TRIGGER vm_changes BEFORE UPDATE 
ON "vms"
BEGIN
   INSERT INTO audit_vms select * from vms where id=old.id;
END;

DROP TABLE IF EXISTS old_users;
alter table "users" rename to old_users;

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

insert into users (id, login, firstname, lastname, email, admin)
    select id, login, firstname, lastname, email, admin from old_users
    ;

DROP TABLE IF EXISTS old_users;


.read sql/rmas.sql
.read sql/views.sql
.read sql/triggers.sql
--.read sql/devices.sql
