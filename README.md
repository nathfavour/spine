# spine

Core coordination engine and human-mimicry state vault for agentic systems.

## Overview

`spine` is a high-performance, low-power state manager designed to anchor decentralized agents. It provides:

- **Deterministic Pacing**: Human-like interaction rhythms using Poisson distributions.
- **Low-Power Suspension**: Kernel-level parking using `timerfd` and `epoll`.
- **Shared Intention Vault**: Concurrency-safe state management via POSIX `mmap`.
- **Hardware Affinity**: Efficient workload scheduling on Linux E-Cores.

## Integration

Agents should interact with `spine` via the Unix Domain Socket at `/tmp/spine.sock`.

For detailed technical specifications and the IPC protocol, see [AGENTS.md](./AGENTS.md).

## Build & Install

```bash
anyisland install github.com/nathfavour/spine
```

Managed by [anyisland](https://github.com/nathfavour/anyisland).
