module.exports = {
    entry: './src/js/app.js',
    mode: 'development',
    watch: true,
    output: {
	path: `${__dirname}/dist`,
	filename: 'bundle.js',
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
