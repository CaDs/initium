import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";
import { server } from "./msw/server";

// MSW lifecycle: start once, reset between tests so per-test overrides don't
// leak, and close on suite exit. Tests that hit the network without an
// explicit handler will fail with a clear "unmatched request" error instead
// of accidentally hitting the real API.
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => {
  server.resetHandlers();
  cleanup();
});
afterAll(() => server.close());
