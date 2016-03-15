drop table if exists part_types;
CREATE TABLE "part_types" (
    tid integer primary key,
    name text not null,
    user_id integer not null default 0, 
    modified date DEFAULT CURRENT_TIMESTAMP
);

insert into part_types (name) values ('Memory');
insert into part_types (name) values ('Disk');
insert into part_types (name) values ('Mainboard');
insert into part_types (name) values ('Power Supply');
insert into part_types (name) values ('CPU');


alter table skus rename to old_skus;

CREATE TABLE "skus" (
    kid integer primary key AUTOINCREMENT,
    mid integer REFERENCES mfgr(mid) ON DELETE RESTRICT,
    tid integer REFERENCES part_types(tid) ON DELETE RESTRICT,
    description text not null,
    part_no text not null, 
    user_id integer not null, 
    modified date DEFAULT CURRENT_TIMESTAMP
    --unique (mid,part_no)
);

insert into skus 
    (kid,mid,description,part_no,user_id, modified)
    select kid,mid,description,part_no,user_id, modified
    from old_skus;
    ;

drop table old_skus;
