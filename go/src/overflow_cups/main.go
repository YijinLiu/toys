package main

import (
	"flag"
	"log"
)

var waterAmount = flag.Float64("water_amount", 100.0, "")
var cupOrder = flag.Int("cup_order", -1, "")

type Cup struct {
	// All are 1-based.
	level int
	row int
	column int
	Water float64
}

var cups = map[int]*Cup{}

func getLevelCups(numLevels int) int {
	if numLevels == 0 {
		return 0
	}
	return numLevels * (numLevels + 1) * (numLevels + 2) / 6
}

func getRowCups(numRows int) int {
	if numRows == 0 {
		return 0
	}
	return numRows * (numRows + 1) / 2
}

func getOrder(level, row, column int) int {
	return getLevelCups(level - 1) + getRowCups(row - 1) + column
}

func GetCup(level, row, column int) *Cup {
	if level <= 0 || row <= 0 || column <= 0 || row > level || column > row {
		log.Fatalf("Invalid cup param: %d, %d, %d!\n", level, row, column)
	}
	cup := &Cup{level, row, column, 0.0}
	order := cup.GetOrder()
	if existingCup, found := cups[order]; found {
		return existingCup
	}
	cups[order] = cup
	return cup
}

func (c *Cup) GetOrder() int {
	return getOrder(c.level, c.row, c.column)
}

func (c *Cup) AddWater(water float64) {
	c.Water += water
	if c.Water > 1.0 {
		firstChild := GetCup(c.level + 1, c.row, c.column)
		secondChild := GetCup(c.level + 1, c.row + 1, c.column)
		thirdChild := GetCup(c.level + 1, c.row + 1, c.column + 1)
		firstChild.AddWater((c.Water - 1.0) / 3.0)
		secondChild.AddWater((c.Water - 1.0) / 3.0)
		thirdChild.AddWater((c.Water - 1.0) / 3.0)
		c.Water = 1.0
	}
}



func main() {
	flag.Parse()
	cup := GetCup(1, 1, 1)
	cup.AddWater(*waterAmount)

	if cup, ok := cups[*cupOrder]; ok {
		log.Printf("Cup %d: %f\n", cup.GetOrder(), cup.Water)
	} else {
		level := 1
		for {
			if _, found := cups[getOrder(level, 1, 1)]; !found {
				break
			}
			log.Printf("[level %d]\n", level)
			for row := 1; row <= level; row++ {
				log.Printf("row %d:", row)
				for column := 1; column <= row; column++ {
					order := getOrder(level, row, column)
					if cup, found := cups[order]; found {
						log.Printf(" %f", cup.Water)
					} else {
						log.Fatal("MUST BE WRONG!")
					}
				}
				log.Println()
			}
			level++
		}
	}
}
