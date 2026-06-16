# spine Agent Integration Specification (UDS)

`spine` provides a low-level coordination and human-mimicry engine for agentic systems. Agents interact with `spine` exclusively via a Unix Domain Socket (UDS) located at `/tmp/spine.sock`.

## 1. IPC Binary Framing Protocol

All communication follows a length-prefixed binary layout:

| Component      | Size     | Definition                                |
|----------------|----------|-------------------------------------------|
| Magic Byte     | 1 Byte   | Always `0xAA`                             |
| Opcode         | 1 Byte   | Operation identifier (see below)          |
| Payload Length | 4 Bytes  | Big-Endian uint32 declaring data bounds   |
| Data Payload   | N Bytes  | JSON-encoded structured data              |

## 2. Opcodes & Data Structures

### `0x01: OpPark` (Sleep & Suspension)
Agents can request a low-power "park" state. `spine` will hold the connection open and only return when the timer expires.
- **Request Body:** `SleepRequest`
- **Response:** `OpWake` frame (sent when timer expires).

### `0x04: OpRegister` (OTA Registration)
Register an agent with a specific behavioral profile.
- **Request Body:** `RegistrationRequest`
```json
{
  "agent_id": "threader-v1",
  "namespace": [100, 101, ...],
  "purpose": "social" 
}
```
**Purposes:**
- `social`: Randomized intervals averaging 30s.
- `research`: Deep collection rhythms averaging 5m.
- `burst`: High-intensity activity averaging 5s.
- `idle`: Low-power background check averaging 1h.

### `0x05: OpRequestPace` (Human Mimicry)
Request a calculated sleep interval based on the registered profile.
- **Request Body:** `PaceRequest`
```json
{
  "agent_id": "threader-v1",
  "weight": 1.0
}
```
- **Response:** `PaceResponse`
```json
{
  "interval_ms": 34521
}
```

## 3. Implementation Workflow

1. **Connect:** Open a Unix Domain Socket connection to `/tmp/spine.sock`.
2. **Register:** Send an `OpRegister` frame with your `agent_id` and desired `purpose`.
3. **Pace:** Before performing an action, send `OpRequestPace`. Use the returned `interval_ms` to wait.
4. **Park:** If you need to suspend for a long period (e.g., hours), send `OpPark`. Your thread will be parked by the kernel with zero CPU overhead.

## 4. Why use spine Mimicry?
Traditional `time.Sleep(rand.Intn(n))` produces predictable uniform distributions. `spine` utilizes Poisson distributions and behavioral weighting to simulate human-like engagement patterns, effectively bypassing detection heuristics on major platforms.
