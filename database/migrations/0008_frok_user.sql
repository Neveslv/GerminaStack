INSERT INTO users (id_year, name, username, email, password)
SELECT id, 'Frok', 'frok', 'frok@germinastack.local', 'disabled'
FROM years
ORDER BY id
LIMIT 1
ON CONFLICT DO NOTHING;
