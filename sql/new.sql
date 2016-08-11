
.open 'inventory.db'

drop view if exists newservers;

.open 'data.db'

.print 'schema'
.read sql/schema.sql
.print 'views'
.read sql/views.sql
.print 'triggers'
.read sql/triggers.sql
.print 'load'
.read sql/load.sql

