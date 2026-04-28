const expoConfig = require('eslint-config-expo/flat');

module.exports = [
  ...expoConfig,
  {
    ignores: ['dist/*', '.expo/*', 'coverage/*', 'node_modules/*', 'jest.config.js'],
  },
];
