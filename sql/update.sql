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

