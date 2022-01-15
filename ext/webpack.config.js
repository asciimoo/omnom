const webpack = require("webpack"),
    path = require("path"),
    fileSystem = require("fs"),
    env = require("./utils/env"),
    CleanWebpackPlugin = require("clean-webpack-plugin").CleanWebpackPlugin,
    MiniCssExtractPlugin = require("mini-css-extract-plugin"),
    CopyWebpackPlugin = require("copy-webpack-plugin"),
    HtmlWebpackPlugin = require("html-webpack-plugin"),
    TerserPlugin = require("terser-webpack-plugin")
WriteFilePlugin = require("write-file-webpack-plugin");

// load the secrets
let alias = {};

const secretsPath = path.join(__dirname, ("secrets." + env.NODE_ENV + ".js"));

if (fileSystem.existsSync(secretsPath)) {
    alias["secrets"] = secretsPath;
}



console.log({ process: process.env.NODE_ENV, env: env.NODE_ENV });
module.exports = {
    mode: env.NODE_ENV,
    entry: {
        omnom: [path.join(__dirname, "src", "js", "omnom.js"), path.join(__dirname, "src", "css", "style.css")],
        options: path.join(__dirname, "src", "js", "options.js"),
        site: path.join(__dirname, "src", "js", "site.js"),
        // background: path.join(__dirname, "src", "js", "background.js")
    },
    output: {
        path: path.join(__dirname, "build"),
        publicPath: '',
        filename: "[name].js"
    },
    optimization: {
        //minimize: env.NODE_ENV === 'production' ? true : false,
        minimize: false,
        minimizer: [
            new TerserPlugin({
                terserOptions: {
                    format: {
                        comments: false
                    }
                },
                extractComments: false
            })
        ]
    },
    module: {
        rules: [
            {
                test: /\.css$/,
                use: [MiniCssExtractPlugin.loader,
                    'css-loader'],
                exclude: /node_modules/
            },
            {
                test: /\.(jpe?g|svg|png|gif|ico|eot|ttf|woff2?)(\?v=\d+\.\d+\.\d+)?$/i,
                type: 'asset/resource',
            },
            {
                test: /\.html$/,
                loader: "html-loader",
                options: {
                    sources: false
                },
                exclude: /node_modules/
            }
        ]
    },
    resolve: {
        alias: alias
    },
    plugins: [
        // clean the build folder
        new CleanWebpackPlugin({
            cleanStaleWebpackAssets: false,
        }),
        new MiniCssExtractPlugin({
            filename: '[name].css',
        }),
        // expose and write the allowed env vars on the compiled bundle
        new webpack.EnvironmentPlugin(["NODE_ENV"]),
        new CopyWebpackPlugin({
            patterns: [
                {
                    from: "src/manifest.json",
                    transform: function (content, path) {
                        // generates the manifest file using the package.json informations
                        return Buffer.from(JSON.stringify({
                            description: process.env.npm_package_description,
                            version: process.env.npm_package_version,
                            ...JSON.parse(content.toString())
                        }))
                    }
                },
                {
                    from: "src/icons",
                    to: "icons"
                },
            ]
        }),
        new HtmlWebpackPlugin({
            template: path.join(__dirname, "src", "popup.html"),
            filename: "popup.html",
            chunks: ["popup"]
        }),
        new HtmlWebpackPlugin({
            template: path.join(__dirname, "src", "options.html"),
            filename: "options.html",
            chunks: ["options"]
        }),
        new WriteFilePlugin()
    ],
    devtool: 'cheap-module-source-map'
};
