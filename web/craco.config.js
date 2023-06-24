const CracoLessPlugin = require("craco-less");
const CompressionPlugin = require("compression-webpack-plugin")

module.exports = {
  webpack: {
    configure: (webpackConfig, {env}) => {
      if (env === "production") {
        webpackConfig.devtool = false;
        // Split chunks to separate vendor dependencies from application code
        webpackConfig.optimization.splitChunks = {
          cacheGroups: {
            vendors: {
              test: /[\\/]node_modules[\\/]/,
              name: "vendors",
              priority: -10,
              enforce: true,
              maxSize: 1000000, // 1MB
              minSize: 0, // Set minimum size to 0,
              chunks: "initial", // Split initial chunks
            },
            antdTokenPreviewer: {
              test: /[\\/]node_modules[\\/](antd-token-previewer)[\\/]/,
              name: "antd-token-previewer",
              priority: -5,
              enforce: true,
              chunks: "all",
            },
            codeMirror: {
              test: /[\\/]node_modules[\\/](codemirror)[\\/]/,
              name: "codemirror",
              priority: -5,
              enforce: true,
              chunks: "all",
            },
          },
        };

        // Enable tree shaking
        webpackConfig.optimization.usedExports = true;

        webpackConfig.plugins.push(new CompressionPlugin({
          algorithm: "gzip",
          test: /\.js$|\.css$|\.html$/,
          threshold: 10240,
          minRatio: 0.8,
        }));
      }

      return webpackConfig;
    },
  },
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
};
