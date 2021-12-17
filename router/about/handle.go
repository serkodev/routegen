package about

import "github.com/serkodev/pbr/route"

func GET(r *route.Route) {
	_ = r
	r.Print("about")
}
