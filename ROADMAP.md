# Roadmap

tuiweave is working toward a stable v1 component vocabulary. The current
investigation and performance follow-up have scoped evidence requirements:

- define and validate input-method handling from real terminal and application
  evidence in [Issue #19](https://github.com/ishan5ain/tuiweave/issues/19);
- reduce streaming Markdown ANSI writer allocation churn in
  [Issue #49](https://github.com/ishan5ain/tuiweave/issues/49). The
  [Issue #33 investigation](agentic/markdown/STREAMING-PERFORMANCE.md) found
  material cost for long structured responses, with output/writer allocation
  dominating parsing. Keep full-document rendering until a measured, compatible
  optimization is reviewed; an incremental parser is not justified by this data.

The foundations those investigations build on are now in place: bounded and
cell-aware rendering, named interaction scenarios, semantic inspection and
action routing, application-owned mouse hit-testing, a polished non-agentic
reference application, compatibility policy, and paired readable/role-aware
goldens.

Further editor commands, inspection-schema expansion, a machine-readable
scenario format, and persistent agentic session ledgers remain deliberately
deferred. Stabilize those APIs through real application use; open a scoped
issue with an acceptance test, non-goals, and an affected application before
implementation.

Roadmap items are intentions, not commitments or a release schedule. Propose a
new component or public API in an issue before implementation.
