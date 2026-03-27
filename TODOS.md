# TODOS — AgentFS

## Pre-code (do before any implementation)

### [ ] User validation — 5 developer conversations
**What:** Find 5 developers actively building long-running or autonomous Agents and ask
them one question: *"When you restart your Agent or move it to a new environment, what
do you lose and how do you get it back?"*

**Why:** The entire design is hypothesis. The status quo pain has not been validated
with real users. If fewer than 3/5 describe the "context loss" problem, revisit the
wedge before building.

**Where:** LangChain/CrewAI/AutoGen Discord, AgentFS GitHub Discussions, X/Twitter
technical community. Real conversations, not surveys.

**Go signal:** 3/5 describe "I keep losing context when I restart."
**Stop signal:** <3/5 feel this pain → revisit wedge (enterprise compliance? Skill IP?)

**Depends on:** Nothing. Do this first.

---

### [ ] SKILL.md format compatibility check with Hermes Agent
**What:** Read the Hermes Agent SKILL.md specification. Determine if AgentFS can adopt
it as-is (preferred) or must document a deviation.

**Why:** If AgentFS and Hermes Agent use incompatible SKILL.md formats, the ecosystem
fragments from day 1. Skills built for one won't work in the other.

**How:** Read hermes-agent.org docs + SKILL.md spec. If compatible: adopt it, document
the alignment. If incompatible: document the deviation clearly in the AgentFS spec and
explain why.

**Depends on:** Nothing. Do before implementing the `/skills/` directory handler.

---

## Phase 1 implementation

### [ ] Partial-line crash recovery for episodic.jsonl (CRITICAL)
**What:** On `agentfs mount`, before the FUSE daemon starts serving, open
`episodic.jsonl` and truncate to the last complete newline character.

**Why:** If the FUSE daemon dies mid-write (SIGKILL, OOM, power loss), the last entry
in `episodic.jsonl` may be a partial JSON line. Any tool reading the file will get a
parse error. This silently breaks the memory layer.

**Implementation:** ~20 lines of Go in the mount initialization path.
```go
// on mount: truncate episodic.jsonl to last complete line
f, _ := os.OpenFile(episodicPath, os.O_RDWR, 0644)
stat, _ := f.Stat()
if stat.Size() > 0 {
    // scan backwards for last '\n', truncate there
    buf := make([]byte, 4096)
    offset := stat.Size() - int64(len(buf))
    if offset < 0 { offset = 0 }
    n, _ := f.ReadAt(buf, offset)
    lastNL := bytes.LastIndexByte(buf[:n], '\n')
    if lastNL >= 0 {
        f.Truncate(offset + int64(lastNL) + 1)
    }
}
f.Close()
```
Log a warning to stderr if truncation occurs.

**Depends on:** episodic.jsonl handler implementation.

---

## Phase 2

### [ ] episodic.jsonl pruning / summarization strategy
**What:** Define and implement a strategy for managing episodic.jsonl growth over time.

**Why:** episodic.jsonl grows unboundedly. After months of real use, it could be 10MB+
— far beyond any LLM's usable context window. The "reliable memory" promise breaks
after 3-6 months without this.

**Options to evaluate:**
- (a) Virtual windowing: `/mem/episodic.jsonl?last=100` query param returns last N entries
- (b) Auto-prune: AgentFS daemon prunes entries older than configurable threshold
- (c) `agentfs compress` command: manual trigger that summarizes old entries via LLM

**Data format prerequisite:** Phase 1 structured single-line JSON format (already decided)
makes all three options feasible. This must not be changed.

**Depends on:** Phase 1 shipped + real users reporting context window issues.
