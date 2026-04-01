package handlers

import (
	"net/http"
	"sort"
)

var registry = map[string]http.HandlerFunc{}

func Register(name string, h http.HandlerFunc) {
	registry[name] = h
}

func Get(name string) (http.HandlerFunc, bool) {
	h, ok := registry[name]
	return h, ok
}

func List() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
