# Rapibase Agent Skills

Agent Skills for Rapibase, the agent-native backend, built on the open Agent Skills format.

> **Status:** in development. The skills ship together with the Rapibase CLI v0.1.

## Installation

```bash
npx skills add rapibase/skills
```

With the Rapibase CLI (it carries these skills inside, so nothing is downloaded; `init` installs them
too):

```bash
npx rapibase@latest skill install
npx rapibase@latest skill update
```

Hermes:

```bash
hermes skills tap add rapibase/skills
```

## Path selection

- Prefer the Rapibase CLI. Start with `npx rapibase@latest init`, then `rapibase whoami` and
  `rapibase status`.
- If the Rapibase MCP server is already connected to your agent, its tools are the fallback path.
- The project for a folder is the one in `.rapibase/project.json`.
- Keys stay in `.env` (written by `rapibase env pull`); agents never read it. Admin calls go through
  `rapibase sql` and `rapibase api`, which add the service key on the server side.

## Skills

- **rapibase-backend**: Postgres database, user auth, file storage, realtime and push notifications.
- **rapibase-deploy**: publish the frontend, deploy your own services in any language, custom domains and CI.
- **rapibase-functions**: server-side TypeScript functions: HTTP endpoints, cron jobs and workers.
- **rapibase-observe**: project status, logs, errors and plan limits.

Each skill is a `SKILL.md` plus a `references/` folder the agent reads only when the task needs it. All
four start with the same Setup block: CLI first, MCP as fallback, the folder's project, and the fixed rules
(no deploy without an explicit request, simple and free first, the person decides about the plan).

## Repository layout

```
skills/
  rapibase-backend/     SKILL.md + references/  design, credentials, auth, rest, rpc, storage, realtime,
                                                rls, pgvector, email, webhooks, push, config
  rapibase-deploy/      SKILL.md + references/  site, github-actions, services, domains
  rapibase-functions/   SKILL.md + references/  runtime
  rapibase-observe/     SKILL.md + references/  errors, config (restart)
tests/                  format and safety checks (go test ./...)
```

## Contributing

`go test ./...` must pass. It checks:

- **Safety lint** over `skills/` and this README: no absolute URLs or domain names (URLs come from
  `rapibase status --json` and `rapibase open` at runtime), no download or pipe-to-shell commands, nothing
  shaped like a key or token (not even as an example), and Markdown only (no scripts in v1).
- **Format** (Agent Skills spec): `name` equals the folder, `description` says what the skill does and
  when to use it (1–1024 characters), `metadata.version` is semver (the CLI's `skill update` compares it),
  `SKILL.md` stays under 500 lines, `references/` is one level deep and every link resolves.
- **Contract:** the exact file list per skill, every reference linked from its `SKILL.md`, and the Setup
  block identical in all four skills.

Bump `metadata.version` in a skill's `SKILL.md` whenever its content changes.
A release is a `vX.Y.Z` tag of this repo (the first one is `v0.1.0`). The CLI embeds the skills of the tag
pinned in its `skills.ref`; installs straight from this repo get the default branch.

## License

MIT
