DROP TABLE if exists part_types;

CREATE TABLE "part_types" (
    tid integer primary key AUTOINCREMENT,
    name text not null,
    user_id integer not null, 
    modified date DEFAULT CURRENT_TIMESTAMP
); 

insert into part_types(name, user_id) values('Memory',0);

--ALTER TABLE skus rename to old_skus;

DROP TABLE IF EXISTS skus;
CREATE TABLE "skus" (
    kid integer primary key AUTOINCREMENT,
    mid integer not null REFERENCES mfgr(mid) ON DELETE RESTRICT,
    tid integer not null REFERENCES part_types(tid) ON DELETE RESTRICT,
    description text not null,
    part_no text not null, 
    user_id integer not null, 
    modified date DEFAULT CURRENT_TIMESTAMP,
    unique (mid,part_no)
);

INSERT INTO skus (kid,mid,tid,description,part_no,user_id,modified)
    select kid,mid,last_insert_rowid(),description,part_no,user_id,modified from old_skus;
    
