# Ultimate for Nicely

This folder adapts Ultimate to Nicely's existing development workflow.

Start `ultimate` in the Nicely clone. Choose this project's setup when Ultimate asks.
The built-in agents remain the source of their instructions. Project overlays add Nicely's worktree and pull-request requirements.

## Worktree setup

The setup commands download dependencies for the root, tools, and standalone fixture Go modules.
Go shares its module cache between worktrees. The commands are safe to repeat and require network access for uncached modules.
Install the Go version required by each module before starting work.

Nicely needs no development database, copied `.env`, or development server port.
The Python checks use the pinned tools and packages in `just python`.
No setup command copies credentials, installs system tools, or runs paid services.

Ultimate's automatic worktree location does not satisfy the workspace rules.
Create worktrees under `~/projects/.worktrees/` instead. Run the setup commands there before checks.
The agent overlays describe this manual path. They do not change Ultimate's automatic worktree location.

## Checks and landing

[AGENTS.md](../AGENTS.md) owns the card workflow.
[Developer preferences](../docs/north-star/dev-preferences.md) own review, macOS checks, and signoff.
Reuse the repository's `verify-nicely` and `release-nicely` skills.

Do not use `workspace submit` or `workspace land` for Nicely.
Open a pull request and follow the repository's independent review and signed-SHA landing process.
Automatic lander checks and push are intentionally not configured. The pull-request workflow owns those steps.

## Models

This setup does not change agent models or logins. The built-in designer requires a local Claude Code login.
Run `claude auth login` in a terminal, then refresh **Claude Code setup** in Ultimate's command palette.
Never copy provider credentials between machines.
