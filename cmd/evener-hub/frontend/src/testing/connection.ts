// Shared connection-store fixtures for web tests: one way to wire a ready
// FakeClient into the connection store, so no test re-declares it.
import { FakeClient } from "@evener/appwire-client/testing/fakeClient";
import { connectionStore } from "../stores/connection";

/** A FakeClient in the "ready" state, connected through the connection store.
 * `configure` registers request handlers before the connect, for tests whose
 * stores issue requests during the handshake. */
export function connectReadyFakeClient(configure?: (fake: FakeClient) => void): FakeClient {
  const fake = new FakeClient("ready");
  configure?.(fake);
  connectionStore.getState().connect(fake);
  return fake;
}
