package main

import (
	"fmt"
	"strings"
)

func greet(name string) string {
	new_str := strings.TrimSpace(name)

	fmt.Println(new_str)

	if len(new_str) == 0 {
		return "Hello, World!"
	}

	return fmt.Sprintf("Hello, %s!", new_str)

}
