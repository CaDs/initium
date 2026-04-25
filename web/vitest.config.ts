import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import { resolve } from "path";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "happy-dom",
    setupFiles: ["./src/__tests__/setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "html"],
      // Phased floor: starts at 25% with headroom, ramps to 80% in
      // follow-up PRs as coverage grows. Below this fails the suite.
      thresholds: { lines: 25, branches: 25 },
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/__tests__/**",
        "src/lib/api-types.ts", // generated from openapi.yaml
        "src/**/*.d.ts",
        "src/i18n/**",
        "src/app/**/layout.tsx",
        "src/app/**/page.tsx",
      ],
    },
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "./src"),
    },
  },
});
