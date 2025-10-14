package main

import "fmt"
import "jobs"

func main() {
	fmt.Println("Hello World!")
	jobs.RegisterJob("Example Job")
}