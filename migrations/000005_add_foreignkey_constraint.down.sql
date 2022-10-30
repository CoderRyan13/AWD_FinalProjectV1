-- Filename: migrations/000005_add_foreignkey_constraint.down.sql
ALTER TABLE forums
DROP CONSTRAINT FK_forumsusers;