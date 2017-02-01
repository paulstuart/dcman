
-- move to normalized profiles

drop table if exists profiler;

create table profiler (
    id integer primary key,
    map int,
    profile text,
    aka text
);


insert into profiler (profile) select distinct profile from devices where profile > '' order by profile;
update profiler set aka=lower(replace(replace(replace(profile,' ',''),'/',''),'-',''));

-- normalize further
update profiler set aka='hypervisor' where aka in ('hyperviser','hypervisors');
update profiler set aka='statserver' where aka in ('stats');
update profiler set aka='mobile' where aka in ('mob','mobilead');
update profiler set aka='ias' where aka in ('iasvarnish','iasproject');

-- find the first instance of a profile and make it the one for all with same 'aka'
with pp as (select * from profiler)
    update profiler set map=(select id from pp where pp.aka == profiler.aka)
    ;

drop table if exists profiles;

CREATE TABLE "profiles" (
    prd      integer primary key,
    profile  text,
    script   text,
    note     text,
    usr      integer,
    ts       date DEFAULT CURRENT_TIMESTAMP
);

with good(id) as (
       select distinct map from profiler
   ),
   keep(prd, profile) as (
       select id, profile from profiler
        where id in (select * from good)
        order by aka
   )
   insert into profiles (prd, profile) select prd,profile from keep;
   --insert into profiles select prd,profile from keep

-- modify existing system

DROP TABLE IF EXISTS old_audit;
ALTER TABLE audit_devices rename to old_audit;

CREATE TABLE "audit_devices" (
    did integer,
    rid integer,
    dti integer,
    mid integer,
    prd integer,
    ru  integer,
    height    integer,
    hostname  text,
    alias     text,
    model     text,
    asset_tag text,
    sn        text,
    tag       text,
    assigned  text,
    note      text,
    restricted integer,
    version integer,
    usr integer,
    ts  date
);

insert into audit_devices
    (
        did,
        rid,
        dti,
        mid,
        --tid,
        ru,
        height,
        hostname,
        alias,
        model,
        asset_tag,
        sn,
        --profile,
        assigned,
        note,
        version,
        usr,
        ts
    )
    select 
        did,
        rid,
        dti,
        mid,
        --tid,
        ru,
        height,
        hostname,
        alias,
        model,
        asset_tag,
        sn,
        --profile,
        assigned,
        note,
        version,
        usr,
        ts
    from old_audit
    ;
DROP TABLE IF EXISTS old_audit;


DROP TABLE IF EXISTS old_devices;
ALTER TABLE devices rename to old_devices;

drop table if exists devices;
CREATE TABLE "devices" (
    did integer primary key,
    rid integer,    -- rack ID
    dti integer,    -- device type ID
    mid integer,    -- mfgr ID
    prd integer,    -- profile ID
    ru  integer default 0,
    height    int default 1,
    hostname  text not null COLLATE NOCASE,
    alias     text,
    model     text,
    asset_tag text,
    sn        text,
    tag       text,
    assigned  text,
    note      text,
    restricted integer default 0,
    version integer default 0,
    usr integer,
    ts  date DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(rid) REFERENCES racks(rid)
    FOREIGN KEY(dti) REFERENCES device_types(dti)
    FOREIGN KEY(mid) REFERENCES mfgrs(mid)
    FOREIGN KEY(prd) REFERENCES profiles(prd)
);

insert into devices
    (
        did,
        rid,
        dti,
        mid,
        prd,
        ru,
        height,
        hostname,
        alias,
        model,
        asset_tag,
        sn,
        assigned,
        note,
        version,
        usr,
        ts
    )
    select 
        did,
        rid,
        dti,
        mid,
        (select map from profiler where profiler.profile == d.profile limit 1),
        ru,
        height,
        hostname,
        alias,
        model,
        asset_tag,
        sn,
        assigned,
        note,
        version,
        usr,
        ts
    from old_devices d
    ;

drop table old_devices;

.read 'sql/views.sql'
.read 'sql/triggers.sql'

select profile, count(profile) as cnt from devices_view group by profile order by profile;

/*
update vlans set starting = route || '1' where starting is null or length(starting) == 0;

DROP VIEW IF EXISTS vlans_fix;
CREATE VIEW vlans_fix as
    select i.*, v.vli as fixed
    from ips i
    left outer join vlans_first v
                on (i.ip32 >= v.network and i.ip32 < v.broadcast)
    where i.vli is null and i.ip32 > 0
    ;

DROP TRIGGER IF EXISTS vlans_fix_update;
CREATE TRIGGER vlans_fix_update INSTEAD OF UPDATE ON vlans_fix
BEGIN
    update ips set vli=NEW.fixed where ips.iid = NEW.iid;
END;

update vlans_fix set vli=fixed ;
*/
