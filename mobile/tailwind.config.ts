import type { Config } from 'tailwindcss';
// eslint-disable-next-line @typescript-eslint/no-require-imports
const nativewindPreset = require('nativewind/preset');

const config: Config = {
  content: ['./app/**/*.{ts,tsx}', './src/**/*.{ts,tsx}'],
  presets: [nativewindPreset],
  theme: {
    extend: {},
  },
  plugins: [],
};

export default config;
