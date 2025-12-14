// Load the shell's babel config and extend it
const path = require('path');
const SHELL_DIR = path.resolve(__dirname, '.shell');

try {
  // Try to load shell's babel config
  module.exports = require(path.join(SHELL_DIR, 'babel.config.js'));
} catch (e) {
  // Fallback config if shell config not available
  module.exports = {
    presets: [
      ['@babel/preset-env', { 
        targets: { browsers: ['> 1%', 'last 2 versions'] },
        useBuiltIns: 'usage',
        corejs: 3
      }],
    ],
    plugins: [
      '@babel/plugin-proposal-optional-chaining',
      '@babel/plugin-proposal-nullish-coalescing-operator',
    ],
  };
}
