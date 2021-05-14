import { ScatterPricesChart } from './scatter_prices.js';
import { SquareMeterPricesChart } from './sqmeter_prices.js';

const updates = [
    ScatterPricesChart('prices'),
    SquareMeterPricesChart('sqmeters')
]

const loader = document.getElementById( "loader-icon" );
const errorbox = document.getElementById( "error-msg" );
const datasets = document.getElementById( "datasets" );
const csvlink = document.getElementById( "csvlink" );
const errorTrans = {
    "non-unique address": "Der findes flere addresser med den beskrivelse, vær mere præcis",
    "no found address": "Kunne ikke finde nogen addresser udfra den søgning"
}

const endpoint = '';

function performSearch() {
    loader.style.display = '';
    errorbox.style.display = 'none';
    datasets.style.display = 'none';

    const XHR = new XMLHttpRequest();
    const fd = new FormData(form);
    const query = fd.get("query");
    const filter = Number(fd.get("filter"));
    const range = Number(fd.get("range"));


    XHR.addEventListener( "load", function(event) {
	loader.style.display = 'none';
	const resp = JSON.parse(event.target.responseText);

	if (resp.error !== undefined) {
	    const err = resp.error;
	    errorbox.innerHTML = err;

	    for (const s in errorTrans) {
		if(err.toLowerCase().includes(s)) {
		    errorbox.innerHTML = errorTrans[s];
		    break
		}
	    }

	    errorbox.style.display = '';
	    return
	}

	var csvUrl = endpoint + "/download/csv?q=" +
	    encodeURIComponent(query) + "&range=" +
	    encodeURIComponent(range);
	csvlink.href = csvUrl;

	datasets.style.display = '';

	for (const f of updates) {
	    f(resp);
	}
    });

    XHR.addEventListener( "error", function( event ) {
	loader.style.display = 'none';
	errorbox.innerHTML = "Kan ikke forbinde til backend";
	errorbox.style.display = '';
    } );

    XHR.open( "POST", endpoint + "/api/lookup" );
    XHR.setRequestHeader("Content-Type", "application/json");
    XHR.send(JSON.stringify({
	"q": query,
	"ranges": [range],
	"filter_below_std": filter,
    }));
}

const form = document.getElementById( "search" );

form.addEventListener( "submit", function ( event ) {
    event.preventDefault();
    performSearch();
});
