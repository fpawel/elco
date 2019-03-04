package daemon

import (
	"fmt"
	"github.com/fpawel/elco/internal/data"
	"github.com/fpawel/elco/pkg/intrng"
	"sort"
)

func formatProducts(products []data.Product) (s string){
	var places []int
	for _, p := range products {
		places = append(places, p.Place)
	}
	sort.Ints(places)
	intrng.IterateOverRanges(places, func(n int, k int) {
		if s != "" {
			s += " "
		}
		if n == k {
			s += data.FormatPlace(n)
		} else {
			s += fmt.Sprintf("%s-%s", data.FormatPlace(n), data.FormatPlace(k))
		}
	})
	return s
}