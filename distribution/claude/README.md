# Rapibase Backend

This plugin helps coding agents build and operate projects on Rapibase. Its four skills cover database and auth design, frontend and service deployment, TypeScript functions, and production troubleshooting. They guide the agent through the Rapibase CLI or the connected MCP tools, according to the available session and the user's request.

The plugin connects to the Rapibase MCP server at https://rapibase.com/mcp using browser based OAuth. The person chooses which projects the connection can access and whether it can write. The skills also describe optional CLI commands; the CLI is installed separately according to its package manager instructions. No credential is bundled with this plugin. Keep secrets in the Rapibase console or the server environment and never paste them into chat.

Documentation: https://rapibase.com/docs  
Privacy: https://rapibase.com/privacy  
Terms: https://rapibase.com/terms  
Source: https://github.com/rapibase/skills
