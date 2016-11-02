insert into part_types (name) values ('Memory');
insert into part_types (name) values ('Disk');
insert into part_types (name) values ('Mainboard');
insert into part_types (name) values ('Power Supply');
insert into part_types (name) values ('CPU');

ALTER TABLE "vlans" rename to "old_vlans";

DROP TABLE IF EXISTS "vlans";
CREATE TABLE "vlans" (
    vli integer primary key,
    sti integer,
    name integer not null,
    profile string,
    gateway text,
    netmask text,
    route text,
    starting text,
    note text,
    min_ip32 integer,
    max_ip32 integer,
    usr integer,
    ts timestamp DEFAULT CURRENT_TIMESTAMP
);

insert into vlans
    (vli, sti, name, profile, gateway,netmask, route, note, min_ip32, max_ip32, usr, ts)
    select * from old_vlans
    ;

DROP TABLE "old_vlans";
