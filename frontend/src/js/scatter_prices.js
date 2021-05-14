import Chart from 'chart.js/auto';
import 'chartjs-adapter-date-fns';
import { TargetStyle } from './chart_styles.js'

function labelCallback(addrs) {
    return (ctx) => {
	const dataset = ctx.dataset;
	const obj = dataset.data[ctx.dataIndex];
	const idx = obj.addr_idx;
	const addr = addrs[idx];
	return addr.full_txt;
    }
}

function afterBodyCallback(addrs) {
    return (data) => {
	const ctx = data[0];
	const dataset = ctx.dataset;
	const obj = dataset.data[ctx.dataIndex];
	const idx = obj.addr_idx;
	const addr = addrs[idx];

	var o = [`Amount: ${obj.y},-`];

	if (addr.building_size > 0) {
	    o.push(`Size: ${addr.building_size} m²`);
	}

	if (addr.built_year > 0) {
	    o.push(`Build: ${addr.built_year}`);
	}

	if (addr.rooms > 0) {
	    o.push(`Rooms: ${addr.rooms}`);
	}

	if (addr.monthly_owner_expense_dkk > 0) {
	    o.push(`Owner Expense (Monthly): ${addr.monthly_owner_expense_dkk},-`);
	}

	if (addr.energy_marking !== "") {
	    o.push(`Energy Marking: ${addr.energy_marking.toUpperCase()}`);
	}

	return o;
    }
}

function ScatterPricesChart(id) {
    let ctx = document.getElementById(id);
    var plot = new Chart(ctx, {
	type: 'scatter',
	data: {
	    datasets: [],
	},
	options: {
	    scales: {
		x: {
		    type: 'time',
		    time: {
			unit: 'month'
		    },
		},
		y: {
		    title: {
			display: true,
			text: "Salgspris",
		    }
		}
	    },
	}
    });

    let update = (state) => {
	const target_sales = state.sales.filter(s => s.addr_idx === state.primary_idx);
	const nearby_sales = state.sales.filter(s => s.addr_idx !== state.primary_idx);

	const addr = state.addresses[state.primary_idx]
	plot.data.datasets = [{
	    ...TargetStyle,
	    label: addr.full_txt,
	    data: target_sales.map(
		(s) => ({addr_idx: s.addr_idx, x: s.when, y: s.amount})
	    ),
	},{
	    label: "Lokalområdet",
	    data: nearby_sales.map(
		(s) => ({addr_idx: s.addr_idx, x: s.when, y: s.amount})
	    )
	}];

	plot.options.plugins.tooltip.callbacks = {
	    afterBody: afterBodyCallback(state.addresses),
	    label: labelCallback(state.addresses)
	}

	plot.update();
    }

    return update
}

export { ScatterPricesChart }
