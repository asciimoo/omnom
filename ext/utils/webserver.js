const env = require("./env");
const path = require("path");
const common = require("../webpack.config");
const { merge } = require('webpack-merge');
const webpack = require('webpack');
module.exports = merge(common, {
  mode: 'development',
  devServer: {
    open: true,
    port: env.PORT,
    writeToDisk: true,
    static: path.join(__dirname, "../build"),
  },
  plugins: [
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify('development'),
    }),
    [new webpack.HotModuleReplacementPlugin()].concat(common.plugins || []),
  ],
  devtool: 'eval-cheap-module-source-map',
});