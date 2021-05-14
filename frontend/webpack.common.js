const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
    entry: {
	app: './src/js/app.js',
    },
    plugins: [
	new HtmlWebpackPlugin({
	    title: 'Production',
	}),
    ],
    output: {
	filename: '[name].bundle.js',
	path: `${__dirname}/dist`,
	clean: true,
    },
    module: {
	rules: [
	    {
		test: /\.css$/,
		use: [
		    'style-loader',
		    'css-loader',
		],
	    },
	],
    },
};
