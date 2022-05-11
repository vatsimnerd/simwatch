package provider

import (
	"github.com/vatsimnerd/geoidx"
	"github.com/vatsimnerd/simwatch-providers/merged"
)

func airportFilter(includeUncontrolled bool) geoidx.Filter {
	if !includeUncontrolled {
		return func(obj *geoidx.Object) bool {
			if arpt, ok := obj.Value().(merged.Airport); ok {
				return arpt.IsControlled()
			}
			return true
		}
	}
	return nil
}
