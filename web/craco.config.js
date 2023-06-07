const CracoLessPlugin = require("craco-less");
const path = require("path")

module.exports = {
  webpack: {
    configure: (webpackConfig) => {
      if(webpackConfig.mode === "production") {
        webpackConfig.devtool = false;
        webpackConfig.output = {
          path: path.resolve(__dirname, './build'),
          filename: 'static/js/[name].[contenthash].js',
          chunkFilename: 'static/js/[name].[chunkhash].chunk.js'
        };
        webpackConfig.optimization = {
          splitChunks: {
            minSize: 30,
            cacheGroups: {
              default: {
                name: 'common',
                chunks: 'initial',
                minChunks: 2,
                priority: -20
              },
              vendors: {
                test: /[\\/]node_modules[\\/]/,
                name: 'vendor',
                chunks: 'initial',
                priority: -10
              },
            }
          }
        }
      }
      return webpackConfig;
    }
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
