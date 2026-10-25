import { FakeClient } from "@evener/appwire-client/testing/fakeClient";
import { connectionStore } from "./stores/connection";

/**
 * connectFakeClient connects a fresh ready FakeClient to the connection store
 * and returns it — the way a store or pane test begins a connection lifecycle
 * without a transport. The identical helper was copy-pasted across dozens of
 * test files; it lives here once.
 */
export function connectFakeClient(): FakeClient {
  const fake = new FakeClient("ready");
  connectionStore.getState().connect(fake);
  return fake;
}
