INSERT INTO users (id_year, name, username, email, password, profile_image_url, profile_image_description)
SELECT id, 'Frok', 'frok', 'frok@germinastack.local', 'disabled', '/static/images/frok-profile.jpeg', 'Foto de perfil do Frok: androide metálico com olhos vermelhos e óculos.'
FROM years
ORDER BY id
LIMIT 1
ON CONFLICT (username) DO UPDATE SET
    profile_image_url = EXCLUDED.profile_image_url,
    profile_image_description = EXCLUDED.profile_image_description;
