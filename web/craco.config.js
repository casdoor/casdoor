const CracoLessPlugin = require("craco-less");

module.exports = {
  devServer: {
    proxy: {
      "/api": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/swagger": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/files": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/.well-known/openid-configuration": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/cas/serviceValidate": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/cas/proxyValidate": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/cas/proxy": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/cas/validate": {
        target: "http://localhost:8000",
        changeOrigin: true,
      },
      "/scim": {
        target: "http://localhost:8000",
        changeOrigin: true,
      }
    },
  },
  plugins: [
    {
      plugin: CracoLessPlugin,
      options: {
        lessLoaderOptions: {
          lessOptions: {
            modifyVars: {"@primary-color": "rgb(89,54,213)", "@border-radius-base": "5px"},
            javascriptEnabled: true,
          },
        },
      },
    },
  ],
  webpack: {
    configure: {
      // ignore webpack warnings by source-map-loader 
      // https://github.com/facebook/create-react-app/pull/11752#issuecomment-1345231546
      ignoreWarnings: [
        function ignoreSourcemapsloaderWarnings(warning) {
          return (
            warning.module &&
            warning.module.resource.includes('node_modules') &&
            warning.details &&
            warning.details.includes('source-map-loader')
          )
        },
      ],
      // use polyfill Buffer with Webpack 5
      // https://viglucci.io/articles/how-to-polyfill-buffer-with-webpack-5
      // https://craco.js.org/docs/configuration/webpack/
      resolve: {
        fallback: {
          // "process": require.resolve('process/browser'),
          // "util": require.resolve("util/"),
          // "url": require.resolve("url/"),
          // "zlib": require.resolve("browserify-zlib"),
          // "stream": require.resolve("stream-browserify"),
          // "http": require.resolve("stream-http"),
          // "https": require.resolve("https-browserify"),
          // "assert": require.resolve("assert/"),
          "buffer": require.resolve('buffer/'),    
          "process": false,
          "util": false,
          "url": false,
          "zlib": false,
          "stream": false,
          "http": false,
          "https": false,
          "assert": false,
          "buffer": false,
          "crypto": false,
          "os": false,
        },
      }
    },
  }
};
