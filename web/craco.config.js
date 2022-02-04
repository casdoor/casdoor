const CracoLessPlugin = require('craco-less');

module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '/swagger': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '/files': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '/.well-known/openid-configuration': {
        target: 'http://localhost:8000',
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
            modifyVars: {'@primary-color': 'rgb(45,120,213)'},
            javascriptEnabled: true,
          },
        },
      },
    },
  ],
};
