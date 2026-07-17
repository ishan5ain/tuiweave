# Roadmap

tuiweave is working toward a stable v1 component vocabulary. Two active
investigation areas have scoped evidence requirements:

- define and validate input-method handling from real terminal and application
  evidence in [Issue #19](https://github.com/ishan5ain/tuiweave/issues/19);
- measure full-document Markdown rendering under realistic streaming workloads
  before considering an incremental design in
  [Issue #33](https://github.com/ishan5ain/tuiweave/issues/33).

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
