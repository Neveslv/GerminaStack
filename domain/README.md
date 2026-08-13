# Domains

`domain/` concentra os contratos das areas de negocio do monolito:

- `auth`: credenciais e desafios do login em duas etapas;
- `users`: cadastro e resultado publico do cadastro;
- `catalog`: anos e materias;
- `account`: perfil e preferencias;
- `discussion`: posts, comentarios, respostas, reacoes e notificacoes;
- `moderation`: consultas e acoes administrativas;
- `pagination`: paginacao compartilhada pelas consultas.

As interfaces sao definidas por quem consome o comportamento. `database/` fornece as implementacoes PostgreSQL, enquanto `handlers/` traduz HTTP para os contratos do dominio. O `main.go` continua montando todas as dependencias em um unico processo.
