DROP TRIGGER IF EXISTS trg_people_updated ON people;
DROP FUNCTION IF EXISTS set_updated_at;
DROP TABLE IF EXISTS friendships;
DROP INDEX IF EXISTS uniq_primary_email_per_person;
DROP TABLE IF EXISTS emails;
DROP TABLE IF EXISTS people;
