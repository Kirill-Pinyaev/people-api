INSERT INTO people (first_name, last_name, age, gender, nationality) VALUES
  ('Ivan', 'Ivanov', 28, 'male', 'RU'),
  ('Anna', 'Ivanova', 26, 'female', 'RU'),
  ('Petr', 'Petrov', 34, 'male', 'RU'),
  ('Olga', 'Sidorova', 30, 'female', 'RU'),
  ('John', 'Smith', 40, 'male', 'US'),
  ('Maria', 'Garcia', 29, 'female', 'ES'),
  ('Luca', 'Rossi', 31, 'male', 'IT'),
  ('Sofia', 'Martinez', 27, 'female', 'AR'),
  ('Akira', 'Tanaka', 36, 'male', 'JP'),
  ('Emma', 'Johnson', 33, 'female', 'US');

INSERT INTO emails (person_id, email, is_primary) VALUES
  (1, 'ivan.ivanov@example.com', true),
  (2, 'anna.ivanova@example.com', true),
  (3, 'petr.petrov@example.com', true),
  (4, 'olga.sidorova@example.com', true),
  (5, 'john.smith@example.com', true),
  (6, 'maria.garcia@example.com', true),
  (7, 'luca.rossi@example.com', true),
  (8, 'sofia.martinez@example.com', true),
  (9, 'akira.tanaka@example.com', true),
  (10,'emma.johnson@example.com', true);

INSERT INTO friendships (user_id, friend_id) VALUES
  (1,2),(1,3),(2,4),(5,10),(6,7);
