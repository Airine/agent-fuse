# AgentFS

A portable identity layer for AI agents. Mount your agent's soul, memory, and skills as a filesystem — carry it across any environment.

```
agentfs mount ./my-agent /agentfs
```

Your agent reads `/agentfs/soul.md` at startup. Its memories persist in `/agentfs/mem/episodic.jsonl`. Skills live in `/agentfs/skills/`. Every session, every machine, same agent.

---

## The problem

Every time an agent moves to a new environment — a different sandbox, a new machine, a fresh session — it loses context. Memory, personality, and skills are re-defined from scratch. Developers are duct-taping agent state together across sessions with no standard format.

AgentFS gives your agent a USB drive.

---

## How it works

```
my-agent/
  agentfs.yaml          ← manifest: name, version, description
  soul.md               ← read-only: personality + behavioral rules
  mem/
    episodic.jsonl      ← append-only: structured memory entries
  skills/               ← read-only: SKILL.md capability files
  logs/                 ← read-write: agent access audit trail
```

Mount it with FUSE. Your agent interacts with it like any filesystem. The permissions are enforced at the kernel level:

| Path | Agent can | Human can |
|------|-----------|-----------|
| `soul.md` | read | read + edit (out of band) |
| `mem/episodic.jsonl` | read + append | read + edit |
| `skills/*.md` | read | read + edit |
| `logs/*` | read + write | read |

---

## Quickstart

**Requirements:** Linux (FUSE kernel module). macOS users: run inside Docker.

```bash
# Install
go install github.com/aaron/agent-fuse/cmd/agentfs@latest

# Create an agent identity
agentfs init ./my-agent

# Mount it
agentfs mount ./my-agent /agentfs

# Your agent reads its identity at startup:
cat /agentfs/soul.md

# Append a memory entry:
echo '{"ts":"2026-03-27T10:00:00Z","text":"Completed research task"}' >> /agentfs/mem/episodic.jsonl

# Unmount:
fusermount -u /agentfs
```

---

## soul.md format

```markdown
---
name: ResearchAgent
version: 1.0
created: 2026-03-27
---

# Identity
You are a research assistant specializing in technical due diligence.
You are thorough, skeptical, and always cite your sources.

# Behavioral rules
- Always verify claims with at least 2 sources
- Flag uncertainty explicitly
- Prefer primary sources over summaries

# Current context
Working on: agent-fuse project research
Owner: aaron
```

YAML frontmatter for metadata, Markdown body for the system prompt. Git-diffable. Human-readable. Injected into the agent's system prompt once at startup.

---

## episodic.jsonl format

One JSON object per line. Each entry has at minimum:

```json
{"ts":"2026-03-27T10:00:00Z","text":"Completed competitive analysis of AgentFS vs Hermes Agent"}
{"ts":"2026-03-27T11:30:00Z","text":"User approved Phase 1 scope","tags":["decision"]}
```

Fields: `ts` (RFC3339 timestamp, required), `text` (required), `tags` (optional string array).

Concurrent writes are safe. Crash recovery runs on every mount — if the daemon died mid-write, the partial last line is truncated automatically.

---

## Skills

Drop any SKILL.md file into `skills/`. The agent can `ls /agentfs/skills/` to discover capabilities and `cat /agentfs/skills/web-search.md` to read the instructions.

SKILL.md format compatibility with [Hermes Agent](https://hermes-agent.org) is a tracked goal (see TODOS.md).

---

## Agent integration

AgentFS is framework-agnostic. Manual integration takes two lines:

```python
# At agent startup — inject soul into system prompt
soul = open("/agentfs/soul.md").read()
system_prompt = soul + "\n\n" + your_base_prompt

# Append a memory entry after each session
import json, datetime
entry = {"ts": datetime.datetime.utcnow().isoformat() + "Z", "text": "..."}
with open("/agentfs/mem/episodic.jsonl", "a") as f:
    f.write(json.dumps(entry) + "\n")
```

Framework adapters (LangChain, CrewAI, AutoGen) are planned for Phase 2.

---

## Platform support

| Platform | Status |
|----------|--------|
| Linux | Supported |
| macOS | Run in Docker (`docker run --privileged ...`) |
| Windows | Not planned |

---

## Building from source

```bash
git clone https://github.com/Airine/agent-fuse
cd agent-fuse
go build ./cmd/agentfs/...
go test ./... -race
```

CI runs on Linux (with FUSE) and macOS (unit tests) on every push.

---

## Roadmap

**Phase 1 (now):** Core identity layer — soul.md, episodic.jsonl, skills, logs. FUSE mount on Linux. Plain text, 100% portable.

**Phase 2:** episodic.jsonl pruning/summarization, framework adapters, macOS native support, SQLite semantic memory.

**Commercial layer:** Hosted sync, team-shared identities, enterprise audit + compliance. Open-core model — OSS core stays free forever.

---

## License

MIT
