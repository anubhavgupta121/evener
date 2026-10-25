import type { QueueState, Thread, ThreadCapabilities } from "@evener/appwire-client";

/**
 * ALL_CAPABILITIES is the every-permission ThreadCapabilities set the thread
 * fixtures grant: a thread that may do anything. The identical literal was
 * copy-pasted into dozens of test files as CAPABILITIES / FULL_CAPABILITIES /
 * CAPS; it lives here once. Nothing mutates it — a test that needs a reduced
 * set spreads it into a new object (`{ ...ALL_CAPABILITIES, send: false }`).
 */
export const ALL_CAPABILITIES: ThreadCapabilities = {
  send: true,
  steer: true,
  interrupt: true,
  compact: true,
  clear: true,
  forkFromTurn: true,
  shutdown: true,
  changeModel: true,
  changeVisionModel: true,
  queue: true,
  goal: true,
  sharedNotes: true,
  rename: true,
};

export type TestThreadOverrides = Omit<Partial<Thread>, "evener"> & {
  evener?: Omit<Thread["evener"], "queue"> & { queue: Partial<QueueState> };
};

/**
 * testThread builds a Thread fixture with stable ids derived from `ref`, the
 * all-permission capability set, and per-field overrides. `evener.queue`
 * overrides merge onto `{ revision: 0 }`; every other field replaces wholesale.
 */
export function testThread(ref: string, overrides: TestThreadOverrides = {}): Thread {
  const { evener, ...threadOverrides } = overrides;
  const threadID = threadOverrides.id ?? `thr_${ref}`;
  return {
    id: threadID,
    sessionId: `sess_${ref}`,
    preview: "test",
    ephemeral: false,
    modelProvider: "anthropic/claude-sonnet-4-5",
    createdAt: 1000,
    updatedAt: 1000,
    status: { type: "idle" },
    cwd: "/tmp/project",
    cliVersion: "1.0.0",
    source: "evener",
    evener: {
      ref,
      instanceId: threadID,
      mutationStateAuthoritative: true,
      capabilities: ALL_CAPABILITIES,
      ...evener,
      queue: { revision: 0, ...evener?.queue },
    },
    ...threadOverrides,
  };
}
