
DROP TABLE IF EXISTS "skus" ;
CREATE TABLE "skus" (
    kid integer primary key,
    vid integer,
    mid integer,
    pti integer,
    description text,
    part_no text, 
    sku text, 
    usr integer, 
    ts date DEFAULT CURRENT_TIMESTAMP,
    unique (mid,description),
    FOREIGN KEY(vid) REFERENCES vendors(vid)
    FOREIGN KEY(mid) REFERENCES mfgrs(mid)
    FOREIGN KEY(pti) REFERENCES part_types(pti)
);

DROP TABLE IF EXISTS "part_types" ;
CREATE TABLE "part_types" (
    pti integer primary key,
    name text not null COLLATE NOCASE,
    usr integer, 
    ts date DEFAULT CURRENT_TIMESTAMP,
    unique (name)
);

insert into part_types (name) values('misc');

insert into part_types (name) values ('Memory');
insert into part_types (name) values ('Disk');
insert into part_types (name) values ('Mainboard');
insert into part_types (name) values ('Power Supply');
insert into part_types (name) values ('CPU');

