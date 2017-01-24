/*
DROP TABLE IF EXISTS old_users;
ALTER TABLE users rename to old_users;
DROP TABLE IF EXISTS users;
CREATE TABLE users (
    usr integer primary key,
    email text,
    firstname text,
    lastname text,
    salt text,
    admin int default 0, 
    local int default 0, 
    pw_hash text, -- bcrypt hashed password
    pw_salt text default (lower(hex(randomblob(32)))),
    apikey text default (lower(hex(randomblob(32))))
);

insert into users select usr,email,firstname,lastname,salt,admin,0,pw_hash,pw_salt,apikey from old_users;
DROP TABLE IF EXISTS old_users;
*/

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
