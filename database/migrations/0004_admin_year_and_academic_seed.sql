ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE users SET is_admin = FALSE WHERE is_admin IS NULL;
ALTER TABLE users ALTER COLUMN is_admin SET DEFAULT FALSE;
ALTER TABLE users ALTER COLUMN is_admin SET NOT NULL;

ALTER TABLE users ALTER COLUMN id_year DROP NOT NULL;
UPDATE users SET id_year = NULL WHERE is_admin = TRUE;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'users'::regclass
          AND conname = 'users_admin_year_check'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_admin_year_check
            CHECK ((is_admin = TRUE AND id_year IS NULL) OR (is_admin = FALSE AND id_year IS NOT NULL));
    END IF;
END $$;

INSERT INTO years (year) VALUES ('2') ON CONFLICT (year) DO NOTHING;

INSERT INTO subjects (id_year, subject)
SELECT years.id, academic_subjects.subject
FROM years
CROSS JOIN (
    VALUES
        ('Biologia ESG'),
        ('Desenvolvimento'),
        ('DAD'),
        ('Mobile'),
        ('DevOps'),
        ('Educação Física'),
        ('Engenharia e Qualidade de Software'),
        ('Estatística'),
        ('BI'),
        ('Língua Inglesa'),
        ('Chefia e Liderança'),
        ('Língua Portuguesa'),
        ('Matemática'),
        ('Modelagem de Dados'),
        ('Inteligência Artificial'),
        ('Banco de Dados'),
        ('Sociologia'),
        ('UX')
) AS academic_subjects(subject)
WHERE years.year = '2'
ON CONFLICT (id_year, subject) DO NOTHING;

CREATE UNIQUE INDEX IF NOT EXISTS idx_subjects_null_year_subject ON subjects (subject) WHERE id_year IS NULL;

INSERT INTO subjects (id_year, subject)
SELECT NULL, 'Geral'
WHERE NOT EXISTS (
    SELECT 1
    FROM subjects
    WHERE id_year IS NULL AND subject = 'Geral'
);