package main

import (
	"fmt"
)

func greet(name string) string {
	return "Hello, " + name + "!"
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func fibonacci(n int) int {
	var a int = 0
	var b int = 1
	for i := 0; i < n; i++ {
		var tmp int = a + b
		a = b
		b = tmp
	}
	return a
}

func main() {
	fmt.Println("=== goc test suite ===")
	fmt.Println("")

	// strings
	msg := greet("World")
	fmt.Println(msg)

	// arithmetic
	fmt.Println("5 + 3 =", 5+3)
	fmt.Println("10 * 4 =", 10*4)
	fmt.Println("17 % 5 =", 17%5)

	// recursion
	fmt.Println("10! =", factorial(10))

	// loop + fibonacci
	fmt.Println("Fibonacci sequence:")
	for i := 0; i < 10; i++ {
		fmt.Printf("  fib(%d) = %d\n", i, fibonacci(i))
	}

	// conditionals
	var x int = 42
	if x > 100 {
		fmt.Println("big")
	} else if x > 10 {
		fmt.Println("medium")
	} else {
		fmt.Println("small")
	}

	// bool
	var flag bool = true
	fmt.Println("flag is:", flag)

	// break / continue
	fmt.Println("evens 0-9:")
	for i := 0; i < 10; i++ {
		if i%2 != 0 {
			continue
		}
		fmt.Printf("  %d\n", i)
	}

	// nested loops with break
	fmt.Println("primes up to 30:")
	for n := 2; n <= 30; n++ {
		isPrime := true
		for d := 2; d*d <= n; d++ {
			if n%d == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			fmt.Printf("  %d\n", n)
		}
	}
}
