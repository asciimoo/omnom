const webpack = require("webpack"),
    path = require("path"),
    fileSystem = require("fs"),
    env = require("./utils/env"),
    CleanWebpackPlugin = require("clean-webpack-plugin").CleanWebpackPlugin,
    CopyWebpackPlugin = require("copy-webpack-plugin"),
    HtmlWebpackPlugin = require("html-webpack-plugin"),
    WriteFilePlugin = require("write-file-webpack-plugin");

// load the secrets
let alias = {};

const secretsPath = path.join(__dirname, ("secrets." + env.NODE_ENV + ".js"));

let fileExtensions = ["jpg", "jpeg", "png", "gif", "eot", "otf", "svg", "ttf"];

if (fileSystem.existsSync(secretsPath)) {
    alias["secrets"] = secretsPath;
}

console.log({ process: process.env.NODE_ENV, env: env.NODE_ENV });
module.exports = {
    mode: env.NODE_ENV,
    entry: {
        omnom: path.join(__dirname, "src", "js", "omnom.js"),
        options: path.join(__dirname, "src", "js", "options.js")
        // background: path.join(__dirname, "src", "js", "background.js")
    },
    output: {
        path: path.join(__dirname, "build"),
        publicPath: '',
        filename: "[name].js"
    },
    module: {
        rules: [
            {
                test: /\.css$/,
                use: ["style-loader", "css-loader"],
                exclude: /node_modules/
            },
            {
                test: /\.(jpe?g|svg|png|gif|ico|eot|ttf|woff2?)(\?v=\d+\.\d+\.\d+)?$/i,
                type: 'asset/resource',
            },
            // {
            //     test: new RegExp('.(' + fileExtensions.join('|') + ')$'),
            //     use: [
            //         {
            //             loader: "file-loader",
            //             options: {
            //                 name: '[name].[ext]',
            //                 outputPath: 'css/webfonts',
            //                 publicPath: '../css/webfonts/'
            //             }
            //         }
            //     ],
            //     exclude: /node_modules/
            // },
            // {
            //     test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
            //     loader: "url-loader",
            //     options: {
            //         limit: 10000,
            //         outputPath: 'css/webfonts',
            //         // publicPath: '../css/webfonts/',
            //         minetype: 'application/font-woff'
            //     }
            // },
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
            cleanStaleWebpackAssets: false
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
                // {
                //     from: "src/css",
                //     to: "css"
                // }
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
        // new HtmlWebpackPlugin({
        //     template: path.join(__dirname, "src", "background.html"),
        //     filename: "background.html",
        //     chunks: ["background"]
        // }),
        new WriteFilePlugin()
    ],
    devtool: 'cheap-module-source-map'
};

// if (env.NODE_ENV === "development") {
//     options.devtool = "eval-cheap-module-source-map";
// }