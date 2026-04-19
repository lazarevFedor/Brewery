INSERT INTO countries (name)
VALUES ($1) 
ON CONFLICT (name) DO UPDATE
SET name = EXCLUDED.name 
RETURNING id;