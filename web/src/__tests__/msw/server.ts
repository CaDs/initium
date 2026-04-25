import { setupServer } from "msw/node";
import { defaultHandlers } from "./handlers";

// Single MSW server shared across the suite. setup.ts hooks listen/reset/close.
// Per-test overrides go through server.use(...) and are auto-reset between
// tests by the resetHandlers() in afterEach.
export const server = setupServer(...defaultHandlers);
