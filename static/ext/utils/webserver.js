// var WebpackDevServer = require("webpack-dev-server"),
//   webpack = require("webpack"),
//   config = require("../webpack.config"),
//   env = require("./env"),
//   path = require("path");

// var options = (config.chromeExtensionBoilerplate || {});
// var excludeEntriesToHotReload = (options.notHotReload || []);

// for (var entryName in config.entry) {
//   if (excludeEntriesToHotReload.indexOf(entryName) === -1) {
//     config.entry[entryName] =
//       [
//         ("webpack-dev-server/client?http://localhost:" + env.PORT),
//         "webpack/hot/dev-server"
//       ].concat(config.entry[entryName]);
//   }
// }

// config.plugins =
//   [new webpack.HotModuleReplacementPlugin()].concat(config.plugins || []);

// delete config.chromeExtensionBoilerplate;

// var compiler = webpack(config);

// var server =
//   new WebpackDevServer(compiler, {
//     hot: true,
//     static: path.join(__dirname, "../build"),
//     port: env.PORT,
//     headers: {
//       "Access-Control-Allow-Origin": "*"
//     },
//     // disableHostCheck: true
//   });

// server.listen(env.PORT);
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
    // historyApiFallback: {
    //   index: '/', // used for routing (404 response), and address bar routing
    // },
    // onBeforeSetupMiddleware: (server) => {
    //   setupProxy(server.app);
    //   server.app.use(express.json());
    //   server.app.use(express.urlencoded({ extended: true }));
    //   server.app.post('/logging', loggingEndpoint);
    // },
    static: path.join(__dirname, "../build"),
  },
  plugins: [
    new webpack.DefinePlugin({
      'process.env.NODE_ENV': JSON.stringify('development'),
    }),
    // new BundleAnalyzerPlugin({
    //   analyzerMode: 'disabled', // server mode displays report
    // }),
    [new webpack.HotModuleReplacementPlugin()].concat(common.plugins || []),
  ],
  devtool: 'eval-cheap-module-source-map',
});