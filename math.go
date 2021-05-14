package hjem

import (
	"fmt"
	"math"
	"time"
)

type Aggregation struct {
	Mean int `json:"mean"`
	Std  int `json:"std"`
	N    int `json:"n"`
}

func AggregationFromPrices(prices []int) Aggregation {
	var sum int
	n := len(prices)
	for _, p := range prices {
		sum += p
	}

	mean := sum / n

	var std float64
	for _, p := range prices {
		std += math.Pow(float64(p-mean), 2)
	}
	stdv := int(math.Sqrt(std / float64(n)))

	return Aggregation{
		Mean: mean,
		Std:  stdv,
		N:    n,
	}
}

func SeperateOutliers(sales []*JSONSale, prices []int, aggr Aggregation, stdf int) ([]int, []*JSONSale) {
	var normal []int
	var outliers []*JSONSale

	lowb, upperb := aggr.Mean+aggr.Std*-stdf, aggr.Mean+aggr.Std*stdf
	for i, _ := range sales {
		p := prices[i]
		s := sales[i]

		if p < lowb || p > upperb {
			outliers = append(outliers, s)
			continue
		}

		normal = append(normal, p)
	}

	return normal, outliers
}

func SalesStatistics(addrs []*Address, sales []*JSONSale, stdf int) ([]*JSONSale, map[time.Time]Aggregation) {
	type G struct {
		S []*JSONSale
		P []int
	}

	temp := map[int]G{}
	for _, s := range sales {
		year, _, _ := s.When.Date()
		sqMeters := addrs[s.AddrIndex].BoligaBuildingSize
		if sqMeters == 0 {
			continue
		}

		g := temp[year]
		g.S = append(g.S, s)
		g.P = append(g.P, s.Amount/sqMeters)

		temp[year] = g
	}

	out := map[time.Time]Aggregation{}

	outliers := map[*JSONSale]bool{}
	for Y, g := range temp {
		year, _ := time.Parse("2-1-2006", fmt.Sprintf("1-1-%d", Y))
		agg := AggregationFromPrices(g.P)
		if stdf > 0 {
			_, outlz := SeperateOutliers(g.S, g.P, agg, stdf)
			for _, ol := range outlz {
				outliers[ol] = true
			}
		}
		out[year] = agg
	}

	for i := 0; i < len(sales); i++ {
		s := sales[i]
		// primary index: 0
		if outliers[s] && s.AddrIndex != 0 {
			sales = append(sales[:i], sales[i+1:]...)
			i--
		}
	}

	return sales, out
}
