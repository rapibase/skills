# Rapibase Agent Skills

Agent Skills for Rapibase, the agent-native backend, built on the open Agent Skills format.

> **Status:** in development. The skills ship together with the Rapibase CLI v0.1.

## Installation

```bash
npx skills add rapibase/skills
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

## Skills

- **rapibase-backend**: Postgres database, user auth, file storage, realtime and push notifications.
- **rapibase-deploy**: publish the frontend, deploy your own services in any language, custom domains and CI.
- **rapibase-functions**: server-side TypeScript functions: HTTP endpoints, cron jobs and workers.
- **rapibase-observe**: project status, logs, errors and plan limits.

## License

MIT
