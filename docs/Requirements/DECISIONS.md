# Decision log

Every ambiguity `@requirements-engineer` resolved on its own authority rather than asking the Owner, one entry each. The authority is granted by `docs/adr/ADR-0001-artifact-driven-agent-org.md § D6`: ordinary ambiguity is settled under the seven-rung precedence ladder and recorded here, and only a genuinely escalation-qualifying question goes to the Owner. **This file is how that delegation stays visible** — the Owner reads it through `@state-reporter`'s digest, which reports every new entry, non-blocking, as a standing invitation to overturn any of it. A decision made autonomously and never written down is indistinguishable, from the Owner's side, from a decision nobody made.

Records: `INDEX.md`. Category register and ID allocation ledger: `README.md`. Story links: `TRACEABILITY.md`. Record format: `requirements-authoring`.

## Entry format

One entry per resolved question, under a `## RD-NNN — <the question, in a line>` heading, with a field block laid out as `requirements-authoring § The canonical record` prescribes — values aligned one space past the longest field name below, continuation lines under them, citations kept whole and never split across a line break.

```
Date:            YYYY-MM-DD
Question:        the question as it was actually asked, not as it reads once answered
Interpretation:  the reading chosen, stated so a later agent can apply it without
                 reconstructing the argument
Rung:            N — the ladder rung it rested on, named
Affects:         the records and documents the interpretation binds
```

**`Rung` is the load-bearing field.** It is what lets a later reader check whether the ladder was followed or whether something merely landed somewhere reasonable, and an entry without it records that a decision happened while withholding the only thing the record exists to prove. The seven rungs, in precedence: explicit specification · architecture constraints · design intent · existing requirements · existing system behavior · established repository conventions · most conservative reasonable interpretation.

**`RD-NNN` IDs are immutable and never reused**, on the same rule as `FR-*` and `NFR-*`. Entries are not rewritten to read better after the fact: a decision that no longer holds is superseded by a **new** entry saying so, never by editing the old one, because a log that can be quietly revised is not evidence of anything. `@requirements-librarian` owns this file's structure, as it owns the rest of the corpus's shape; `@requirements-engineer` writes the entries, and neither judges whether a decision was right — that is the Owner's.

## Entries

No entries yet.
