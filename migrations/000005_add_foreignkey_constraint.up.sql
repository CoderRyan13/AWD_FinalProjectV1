-- Filename: migrations/000005_add_foreignkey_constraint.up.sql
ALTER TABLE forums 
ADD CONSTRAINT FK_forumsusers
FOREIGN KEY (users_id) REFERENCES users(id); --ON DELETE CASCADE ON UPDATE CASCADE;