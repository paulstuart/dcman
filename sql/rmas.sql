
drop table if exists logger;
create table logger (
    log text
);

PRAGMA foreign_keys=OFF;

BEGIN TRANSACTION;
DROP TABLE IF EXISTS "tags";
CREATE TABLE "tags" (
    tid integer primary key,
    tag text,
    unique(tag)
);

INSERT INTO "tags" VALUES(1,'Normal');
INSERT INTO "tags" VALUES(2,'Super');
INSERT INTO "tags" VALUES(3,'Unused');
INSERT INTO "tags" VALUES(4,'HBA');
COMMIT;



DROP TRIGGER if exists servers_audit;
CREATE TRIGGER servers_audit BEFORE UPDATE
ON servers
BEGIN
       INSERT INTO audit_servers select * from servers where id=old.id;
END;

DROP TABLE IF EXISTS "part_types" ;
CREATE TABLE "part_types" (
    pti integer primary key,
    name text not null ,
    unique (name)
);

DROP TABLE IF EXISTS "parts" ;
CREATE TABLE "parts" (
    pid integer primary key,
    kid integer not null default 0,
    vid integer not null default 0,
    sid integer not null default 0,
    dcd integer not null default 0,
    --rma_id integer default 0, -- needs foreign key to rmas
    unused integer default 0, -- boolean (1 is unused)
    bad    integer default 0, -- boolean (1 is bad)
    location text,
    serial_no text,
    asset_tag text, 
    user_id integer not null default 0, 
    modified date DEFAULT CURRENT_TIMESTAMP
);

DROP TABLE IF EXISTS "mfgrs" ;
CREATE TABLE "mfgrs" (
    mid integer primary key,
    name text not null, -- full given name
    aka text , -- shortened and unique
    url text , 
    user_id integer , 
    modified date DEFAULT CURRENT_TIMESTAMP,
    unique (name)
);

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
    --,
    --unique (mid,part_no)
    unique (mid,description)
);


DROP TABLE IF EXISTS "rmas" ;
CREATE TABLE "rmas" (
    rma_id integer primary key,
    dcd integer default 0, -- datacenter id
    sid integer default 0, -- server id
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
    user_id integer default 0 
);

--
-- NEW STUFF
--
         
drop view if exists partload;
create view partload as
   select p.*, s.pti, d.name as dc, s.description, s.parttype, m.name as mfgr
   from parts p
   left join skuview s on p.kid = s.kid
   left join mfgrs m on m.mid = s.mid
   LEFT JOIN datacenters d on p.dcd = d.dcd
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
    insert into skuview (parttype, description, mfgr)
      values(NEW.parttype, NEW.description, NEW.mfgr);

    insert into parts (unused, kid, dcd) 
        select 1, sv.kid, d.dcd from skuview sv, datacenters d
            where NEW.description = sv.description
              and NEW.mfgr = sv.mfgr
              and NEW.parttype = sv.parttype
              and d.name=NEW.dc
        ;
END;

/*
--select distinct p_mfg from parttmp ;
insert into mfgrs (name) select distinct p_mfg from parttmp ;
select * from mfgrs;
.exit
*/
insert into partload (parttype, description, mfgr, dc) select p_type, p_desc, p_mfg, 'NY7' from parttmp ;

.header on
--select p_type, p_desc, p_mfg, 'NY7' from parttmp ;
.print 'SKUVIEW'
select * from skuview;

select count(*) as pcnt from partload;
select count(*) as cnt from parttmp;
select count(*) as pcnt from parts;


/*
 * Set some parts bad for ****** TESTING ****** 
 */

update parts set bad=1 where pid % 3 == 0;


select bad, dc, qty, mfgr, parttype, description from inventory;
select * from inventory limit 1;
--.exit

/*
select * from logger;
.exit

.print DCS
.print
select * from datacenters;
.print
*/
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
select pid, dc, mfgr, parttype, description from partload;
/*
.print
.print RAW
select p_mfg from parttmp;
*/
