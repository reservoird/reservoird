//go:generate mkdir -p mocks
//go:generate mockgen -package mocks -destination mocks/bool.go github.com/reservoird/proxy Bool
//go:generate mockgen -package mocks -destination mocks/plugin.go github.com/reservoird/proxy Plugin

package main
