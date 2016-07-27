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

