const path = require('path');

// Get the shell directory (created as .shell symlink by build-pkg.sh)
const SHELL_DIR = path.resolve(__dirname, '.shell');

module.exports = {
  configureWebpack: {
    resolve: {
      alias: {
        '@shell':      SHELL_DIR,
        '@components': path.join(SHELL_DIR, 'rancher-components'),
        '@pkg':        __dirname,
      },
      extensions: ['.js', '.vue', '.json'],
    },
    // Mark all @shell imports as externals - Rancher provides these at runtime
    externals: [
      // Webpack 4 externals function format
      function(context, request, callback) {
        // Externalize all @shell/* imports
        if (/^@shell\//.test(request)) {
          return callback(null, 'commonjs ' + request);
        }
        callback();
      },
    ],
  },
  css: {
    extract: false,
  },
};
