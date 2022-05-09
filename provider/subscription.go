package provider

import "github.com/vatsimnerd/geoidx"

type Subscription struct {
	id     string
	geosub *geoidx.Subscription
}
