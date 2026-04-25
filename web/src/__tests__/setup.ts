import "@testing-library/jest-dom/vitest";
// eslint-disable-next-line testing-library/no-manual-cleanup
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";
import { server } from "./msw/server";

// MSW lifecycle: start once, reset between tests so per-test overrides don't
// leak, and close on suite exit. Tests that hit the network without an
// explicit handler will fail with a clear "unmatched request" error instead
// of accidentally hitting the real API.
//
// Manual cleanup() is required: testing-library auto-cleanup only kicks in
// when vitest globals are enabled, and we keep globals: false to avoid
// polluting the production bundle's type space.
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => {
  server.resetHandlers();
  cleanup();
});
afterAll(() => server.close());
