package main

import (
	"flag"
	"fmt"
	"os"
)

var order = flag.Int("order", 2000, "")

type WisdomNumber struct {
	value, b, a int
}

func main() {
	wisdomNumbers := []WisdomNumber{
		WisdomNumber{1, 1, 0},
	}
	lastWisdomNumber := 1
	next := 2
	nextStartMap := map[int]int{}
	for len(wisdomNumbers) < *order {
		b := next
		a := next - 1
		// b * b - a * a
		minWisdomNumber := 2*next - 1
		if minWisdomNumber < lastWisdomNumber {
			fmt.Printf("Invalid state: %d < %d\n", minWisdomNumber, lastWisdomNumber)
			os.Exit(1)
		}
		for nb, na := range nextStartMap {
			wisdomNumber := nb*nb - na*na
			if wisdomNumber < lastWisdomNumber {
				fmt.Printf("Invalid state: %d < %d\n", wisdomNumber, lastWisdomNumber)
				os.Exit(1)
			}
			if wisdomNumber < minWisdomNumber {
				b = nb
				a = na
				minWisdomNumber = wisdomNumber
			}
		}

		if a > 1 {
			nextStartMap[b] = a - 1
		} else {
			delete(nextStartMap, b)
		}
		if b == next {
			next++
		}

		if minWisdomNumber > lastWisdomNumber {
			lastWisdomNumber = minWisdomNumber
			wisdomNumbers = append(wisdomNumbers, WisdomNumber{minWisdomNumber, b, a})
		}
	}

	for i, wisdomNumber := range wisdomNumbers {
		fmt.Printf("%d: %d = %d * %d - %d * %d\n", i, wisdomNumber.value,
			wisdomNumber.b, wisdomNumber.b, wisdomNumber.a, wisdomNumber.a)
	}
}
