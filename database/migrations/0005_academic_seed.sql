INSERT INTO years (year)
SELECT '2'
WHERE NOT EXISTS (SELECT 1 FROM years WHERE year = '2');

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
  AND NOT EXISTS (
      SELECT 1
      FROM subjects existing
      WHERE existing.id_year = years.id
        AND existing.subject = academic_subjects.subject
  );
