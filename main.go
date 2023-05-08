package main

import (
	context "GoContext/mycontext"
)

func main() {
	bg := context.Background()
	ctx, cancel := context.WithCancel(bg)
	cancel()
	add(ctx, 1, 2)
}
func add(ctx context.Context, a, b int) int {
	return a + b
}
