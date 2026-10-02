const path = require("path");
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin;

const sageLocal = process.env.SAGE_LOCAL_DEV === "true";
const sageTarget = sageLocal ? "http://127.0.0.1:4100" : "https://eqsage.vercel.app";

module.exports = {
  publicPath: process.env.VUE_APP_PUBLIC_PATH || "/",
  devServer: {
    host: "0.0.0.0",
    disableHostCheck: true,
    watchOptions: {
      ignored: [/node_modules/, /public/],
    },
    // Sage.vue iframes /eqsage. Production Go proxies that to eqsage.vercel.app.
    // Local vue-cli did not, so the iframe 404'd unless SAGE_LOCAL_DEV pointed at a
    // checkout of https://github.com/knervous/eqsage on port 4100.
    proxy: {
      "^/eqsage": {
        changeOrigin: true,
        secure: true,
        timeout: 120000,
        proxyTimeout: 120000,
        logLevel: "warn",
        target: sageTarget,
        pathRewrite: (p) => {
          const next = p.replace(/^\/eqsage/, "");
          return next.length ? next : "/";
        },
      },
      "^/static": {
        changeOrigin: true,
        secure: true,
        timeout: 120000,
        proxyTimeout: 120000,
        logLevel: "warn",
        target: sageTarget,
      },
      "^/api": {
        changeOrigin: true,
        logLevel: "warn",
        target: process.env.VUE_APP_BACKEND_BASE_URL || "http://127.0.0.1:3010",
        ws: true,
      },
      "^/auth": {
        changeOrigin: true,
        logLevel: "warn",
        target: process.env.VUE_APP_BACKEND_BASE_URL || "http://127.0.0.1:3010",
      },
    },
  },
  // configureWebpack: {
  //   plugins: [
  //     new BundleAnalyzerPlugin({analyzerHost: '0.0.0.0', analyzerPort: 3005})
  //   ]
  // },
  chainWebpack: (config) => {
    config.performance.maxEntrypointSize(40000000).maxAssetSize(40000000);

    // ignore asset preview during development to keep build times down
    if (process.env.NODE_ENV !== "production" || process.env.VUE_APP_DEMO === "true") {
      config.plugin("copy").tap(([options]) => {
        options[0].ignore.push("eq-asset-preview-master/**/*");

        return [options];
      });
    }
    //
    config.output
      .filename("[name].[hash].js")
      .path(path.resolve(__dirname, "dist"))
      .clean(true);

    config.plugins.delete('prefetch')

    //
    // config.optimization.moduleIds    = 'deterministic'
    // config.optimization.runtimeChunk = 'single'
    config.optimization.splitChunks = {
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: "vendors",
          chunks: "all",
        },
      },
    };
    // console.log(config)
  },
  runtimeCompiler: true,
  productionSourceMap: false,
};
