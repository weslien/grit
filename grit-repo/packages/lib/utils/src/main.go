package main

import "fmt"

// StringHelper provides utility functions for string manipulation
type StringHelper struct{}

// Reverse reverses a string
func (s *StringHelper) Reverse(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	helper := &StringHelper{}
	fmt.Println("Utils library loaded:", helper.Reverse("hello"))
}