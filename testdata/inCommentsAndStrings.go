package main

import "fmt"

func main() {
	// This will not be replaced: $"I will stay the {same}"
	/*
		Neither will this:
		$"Hello {name}"
	*/
	_ = `This will not be touched: $"How are you today, {name}?"`
	hello := "hi"
	fmt.Sprintf("But this will! %s", hello)
}
