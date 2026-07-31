# Model

O pacote `model` contém as estruturas de dados que representam as tabelas do
GerminaStack. Ele não abre conexões, não executa queries, não aplica defaults do
SQL e não implementa repositórios ou CRUD.

## Models disponíveis

| Struct | Tabela |
|---|---|
| `Year` | `years` |
| `Subject` | `subjects` |
| `User` | `users` |
| `Preference` | `preferences` |
| `Post` | `posts` |
| `Comment` | `comments` |
| `CommentOnComment` | `comments_on_comments` |
| `Reaction` | `reactions` |
| `Notification` | `notifications` |

Todos os IDs usam `int64`, e todos os campos `created_at` usam `*time.Time`
porque o SQL permite `NULL`, apesar de definir `current_timestamp` como default.
Cada campo possui tags `db` e `json` com o nome `snake_case` da respectiva
coluna.

`User.Password` é mapeado para a coluna `password`, mas usa `json:"-"` para
nunca expor a chave nem o valor secreto pela serialização JSON.

## Campos opcionais

Os ponteiros representam colunas que podem ser nulas:

- `Subject.YearID`;
- `User.ProfileImageURL` e `User.ProfileImageDescription`;
- `Preference.ContrastTheme`, `Preference.FontFamily`,
  `Preference.FontSpacing` e `Preference.FontSize`;
- `Post.ImageURL` e `Post.ImageDescription`;
- `Reaction.PostID`, `Reaction.CommentID` e
  `Reaction.CommentOnCommentID`;
- `Notification.PostID`;
- `CreatedAt` em todos os nove models.

Nos pares de imagem, uma string apontada, ainda que vazia, está presente. Essa
semântica corresponde ao `NULL`/`NOT NULL` usado pelos `CHECK` do SQL.

## Enums

Os enums são tipos baseados em `string`, e cada tipo oferece `IsValid() bool`.

- `ContrastTheme`: `normal`, `dark`, `high_contrast`, `black_yellow` e
  `yellow_black`;
- `FontFamily`: `normal`, `arial`, `verdana`, `lexend`,
  `atkinson_hyperlegible` e `open_dyslexic`;
- `FontSpacing`: `normal`, `pequeno` e `grande`;
- `FontSize`: `normal`, `pequeno` e `grande`;
- `ReactionType`: `like` e `dislike`.

As constantes exportadas evitam repetir esses valores literais no restante da
aplicação. Os quatro campos enumerados de `Preference` são ponteiros: `nil`
representa `NULL`, enquanto um valor presente deve pertencer ao enum.

## Validação

As structs que espelham invariantes `CHECK` oferecem `Validate() error`:

- `User.Validate()` exige imagem de perfil e descrição ambas ausentes ou ambas
  presentes;
- `Post.Validate()` aplica a mesma regra à imagem do post;
- `Preference.Validate()` aceita preferências ausentes e rejeita qualquer valor
  presente fora dos quatro enums de acessibilidade;
- `Reaction.Validate()` exige exatamente um alvo entre post, comentário e
  resposta, além de um tipo de reação válido.

Os validadores não alteram os models nem preenchem valores padrão. Restrições de
chave estrangeira, unicidade e obrigatoriedade das demais colunas continuam sob
responsabilidade do banco e das camadas de entrada/persistência.
