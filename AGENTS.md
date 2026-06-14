Run `just typecheck` to typecheck typescript code.
When modifying `db/schema.sql` or `db/queries.sql`, regenerate SQLC models/queries with `just sqlc`.

# Architecture

Follow MVC-style package boundaries.

- `views`: rendering only. HTML, components, static UI helpers.
- `server`: controller layer. HTTP handlers, request parsing, auth/session flow, orchestration.
- `db`: database layer. SQL queries, transactions, persistence types, row mapping.

Other packages may wrap external APIs or domain-specific helpers, but persistence still belongs in `db`, rendering still belongs in `views`, and request orchestration still belongs in `server`.
