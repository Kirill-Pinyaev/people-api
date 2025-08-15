CREATE TABLE IF NOT EXISTS people (
    id BIGSERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    middle_name TEXT,
    last_name TEXT NOT NULL,
    gender TEXT,
    nationality TEXT,
    age INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS emails (
    id BIGSERIAL PRIMARY KEY,
    person_id BIGINT NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    email TEXT NOT NULL UNIQUE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_primary_email_per_person
ON emails(person_id) WHERE is_primary = true;

CREATE TABLE IF NOT EXISTS friendships (
    user_id BIGINT NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    friend_id BIGINT NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (user_id < friend_id),
    UNIQUE(user_id, friend_id)
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_people_updated ON people;
CREATE TRIGGER trg_people_updated
BEFORE UPDATE ON people
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
