import Chart from 'chart.js/auto';
import 'chartjs-adapter-date-fns';
import { TargetStyle } from './chart_styles.js'

const _tension = 0.3;

const uncertainty_style = {
    pointBorderColor: 'rgb(0, 0, 0)',
    pointRadius: 0,
    pointHitRadius: 0,
    borderWidth: 0,
    tooltips: false,
}
const unit = "kr/m²";

const callbacks = {
    title: function() { return "" },
    label: function(ctx) {
	const dataset = ctx.dataset;
	const label = dataset.label === undefined ? "" : dataset.label;
	const obj = dataset.data[ctx.dataIndex];
	if (label.toLowerCase().includes('gennemsnit')) {
	    return obj.x.substring(0,4);
	}

	return `Salg (${obj.x.substring(0,4)})`

    },
    afterBody: function(data) {
	const ctx = data[0];
	const dataset = ctx.dataset;
	const label = dataset.label === undefined ? "" : dataset.label;
	const obj = dataset.data[ctx.dataIndex];
	if (label.toLowerCase().includes('gennemsnit')) {
	    return [
		`${obj.y} ${unit} ± ${obj.std}`,
		`Antal salg: ${obj.n}`,
	    ];
	}

	const projection = dataset.data[dataset.data.length - 1];

	return [
	    `${obj.y} ${unit}`,
	    `${projection.x.substring(0,4)} ~ ${projection.y} ${unit}`
	]
    }
}

function SquareMeterPricesChart(id) {
    let ctx = document.getElementById(id);
    var plot = new Chart(ctx, {
	type: 'line',
	data: {
	    datasets: [],
	},
	options: {
	    scales: {
		x: {
		    type: 'time',
		    time: {
			unit: 'year',
		    },
		},
		y: {
		    title: {
			display: true,
			text: "Kvadratmeterpris",
		    },
		}
	    },
	    plugins: {
		tooltip: {
		    callbacks: callbacks,
		},
		legend: {
		    labels: {
			filter: function(item, chart) {
			    return item.text !== undefined
			}
		    }
		}
	    },
	}
    });

    let update = (state) => {
	const sqElems = Object.entries(state.sqmeters.global)
	const sqmeters = sqElems.map(e => ({ ...e[1] }));

	plot.data.datasets = [{
	    label: "Gennemsnit",
	    tension: _tension,
	    data: sqElems.map(
		e => ({x: e[0], y: e[1].mean, std: e[1].std, n: e[1].n})
	    )
	},{
	    ...uncertainty_style,
	    fill: "+1",
	    data: sqElems.map(
		e => ({x: e[0], y: e[1].mean - e[1].std})
	    )
	},{
	    ...uncertainty_style,
	    data: sqElems.map(
		e => ({x: e[0], y: e[1].mean + e[1].std})
	    )
	}];

	for (const proj of state.sqmeters.projections) {
	    let radiuses = new Array(Object.entries(proj).length).fill(0)
	    radiuses[0] = TargetStyle.pointRadius;

	    plot.data.datasets.push({
		...TargetStyle,
		tension: _tension,
		borderDash: [5, 2],
		borderColor: TargetStyle.backgroundColor,
		pointRadius: radiuses,
		data: Object.entries(proj).map(
		    e => ({x: e[0], y: e[1]})
		)
	    });
	}

	plot.update();
    }

    return update
}

export { SquareMeterPricesChart }
