package simwatch

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

func buildInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		w.WriteHeader(500)
		w.Write([]byte("can't read BuildInfo"))
	}

	for _, kv := range bi.Settings {
		fmt.Fprintf(w, "%s %s\n", kv.Key, kv.Value)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "modules:")

	for _, dep := range bi.Deps {
		fmt.Fprintf(w, "%s %s\n", dep.Path, dep.Version)
	}
}
