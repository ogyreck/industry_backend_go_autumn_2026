package main

import (
	"fmt"
)

func main() {
	c := NewCache[string, int](1)
	fmt.Println(c.Set("a", 1), c.Set("b", 0), c.Set("a", 0))
	fmt.Println(c.Get("v"))
	fmt.Println(c.Get("v"))
}
