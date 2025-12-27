const { merge } = require('webpack-merge');
const common = require('./webpack.common.js');

module.exports = merge(common, {
  mode: 'development',
  devtool: 'inline-source-map',
  devServer: {
    static: './dist',
    port: 3000,
    hot: true,
    historyApiFallback: true,
    proxy: [
      {
        context: ['/api'],
        target: 'http://localhost:4001',
        changeOrigin: true
      },
      {
        context: ['/ws'],
        target: 'http://localhost:4001',
        ws: true
      }
    ]
  }
});
