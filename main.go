package main

import (
	"fmt"
	"os"
)

const (
	height      = 11
	width       = 13
	diagonal    = 4         // steps by diagonal true(8) or false(4)
	pl1         = byte(200) // player without
	pl1_bonus   = byte(201) // player with bonus
	pl1_knife   = byte(202) // player with knife
	pl1_hero    = byte(203) // player with knife and bonus
	pl2         = byte(204) // player without
	pl2_bonus   = byte(205) // player with bonus
	pl2_knife   = byte(206) // player with knife
	pl2_hero    = byte(207) // player with knife and bonus
	monster     = byte(208) // monster
	monsterArea = byte(209) // area around monster
)

type game struct {
	arena            [11][13]byte   // main map
	arenaBuf         [11][13]byte   // main map
	id, tick, answer int            // incomign options of the game
	stackCoin        map[int][2]int // coordinations of closer coin
	steps            int            // save minimal steps to found coin
	bingo            map[int][2]int // coordinations next steps
	coin
}

type coin struct {
	x int // height
	y int // width
}

func (g *game) loop() bool {
	g.bingo = make(map[int][2]int)
	lCoin := len(g.stackCoin)
	limit := 3

	if lCoin == 0 {
		return false
	}
	if lCoin < 3 {
		limit = lCoin
	}

	for limit != 0 {
		if xy, ok := g.stackCoin[g.steps]; ok {
			g.arenaBuf = g.arena
			g.algorithm([][2]int{xy})
		}
		g.steps++
		limit--
	}
	return true
}

func (g *game) algorithm(stack [][2]int) {
	count := byte('1') // first step on map

	for {
		stackBuf := [][2]int{}
		for _, position := range stack {
			for i := 0; i < diagonal; i++ {
				changed, x, y := g.scan(position[0], position[1], i, count, "buf")
				if !changed {
					if g.checkBingo(x, y) {
						g.bingo[int(count-48)] = position
						return
					}
				} else {
					if x+y > 0 {
						stackBuf = append(stackBuf, [2]int{x, y})
					}
				}
			}
		}
		if len(stackBuf) == 0 {
			return
		}
		stack = stackBuf // erase previous
		count++
		// g.print()
	}
}

func (g *game) scan(height, width, diagonal int, symbol byte, sourse string) (changed bool, x int, y int) {
	switch diagonal {
	case 0: // up
		x = height - 1
		y = width
	case 1: // down
		x = height + 1
		y = width
	case 2: // left
		x = height
		y = width - 1
	case 3: // right
		x = height
		y = width + 1
	case 4: // up-left
		x = height - 1
		y = width - 1
	case 5: // up-right
		x = height - 1
		y = width + 1
	case 6: // down-left
		x = height + 1
		y = width - 1
	case 7: // down-right
		x = height + 1
		y = width + 1
	}

	if !g.limit(x, y) {
		return false, 0, 0
	}

	if sourse == "origin" {
		if g.arena[x][y] != '.' {
			return false, x, y
		}
		g.arena[x][y] = symbol
	} else {
		if g.arenaBuf[x][y] != '.' {
			return false, x, y
		}
		g.arenaBuf[x][y] = symbol
	}
	return true, x, y
}

func (g *game) limit(x, y int) bool {
	if x < 0 || x >= height {
		return false
	}
	if y < 0 || y >= width {
		return false
	}
	return true
}

func (g *game) checkBingo(x, y int) bool {
	if g.arena[x][y] >= 200 && g.arena[x][y] <= 203 {
		return true
	}
	return false
}

func (g *game) chooseBingo() int {
	min := 1000
	for steps := range g.bingo {
		if steps < min {
			min = steps
		}
	}
	if min != 1000 {
		return g.stepBingo(min)
	}
	return 4 // stay
}

func (g *game) stepBingo(key int) int {
	// 0 "left"
	// 1 "right"
	// 2 "up"
	// 3 "down"
	// 4 "stay"
	if g.bingo[key][0] == g.coin.x { // height equal
		if g.bingo[key][1] < g.coin.y {
			return 0 // left
		}
		return 1 // right
	}
	if g.bingo[key][1] == g.coin.y { // width equal
		if g.bingo[key][0] < g.coin.x {
			return 3 // down
		}
		return 2 // up
	}
	return 4 // stay
}

func (g *game) scanCoins() {
	g.stackCoin = make(map[int][2]int)
	g.steps = 1000
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if g.arena[i][j] == '#' {
				steps := un(g.coin.x-i) + un(g.coin.y-j)
				g.stackCoin[steps] = [2]int{i, j}
				if g.steps > steps {
					g.steps = steps
				}
			}
		}
	}
}

func (g *game) enterPerson(entType string, pID, x, y, param1, param2 int) {
	//     x y b k
	// p 1 0 0 0 0
	// m 0 3 6 0 0
	// m 0 8 4 0 0
	// p 2 12 10 0 0
	if entType == "p" {
		if pID == g.id {
			if param1 == 0 && param2 == 0 {
				g.arena[x][y] = pl1
			} else if param1 != 0 && param2 == 0 {
				g.arena[x][y] = pl1_bonus
			} else if param1 == 0 && param2 != 0 {
				g.arena[x][y] = pl1_knife
			} else {
				g.arena[x][y] = pl1_hero
			}
			g.coin.x = x
			g.coin.y = y
		} else {
			if param1 == 0 && param2 == 0 {
				g.arena[x][y] = pl2
			} else if param1 != 0 && param2 == 0 {
				g.arena[x][y] = pl2_bonus
			} else if param1 == 0 && param2 != 0 {
				g.arena[x][y] = pl2_knife
			} else {
				g.arena[x][y] = pl2_hero
			}
		}
		return
	}
	if entType == "m" {
		g.arena[x][y] = monster
		g.scan(x, y, diagonal, monsterArea, "origin")
	}
}

// func un makes uint
func un(n int) int {
	if n < 0 {
		return n * -1
	}
	return n
}

func main() {
	for true {
		g := new(game)
		g.answer = 4
		var w, h, playerID, tick int
		fmt.Scan(&w, &h, &playerID, &tick)
		g.id = playerID
		g.tick = tick

		// read map
		for i := 0; i < height; i++ {
			line := ""
			fmt.Scan(&line)
			for j := 0; j < width; j++ {
				g.arena[i][j] = line[j]
			}
			// g.arena[i] = append(g.arena[i], line...)
			// fmt.Fprintf(os.Stderr, "%s\n", line)
		}

		// number of entities
		var n int
		fmt.Scan(&n)
		// fmt.Fprintf(os.Stderr, "%d\n", n)

		// read entities
		for i := 0; i < n; i++ {
			var entType string
			var pID, x, y, param1, param2 int
			fmt.Scan(&entType, &pID, &x, &y, &param1, &param2)
			fmt.Fprintf(os.Stderr, "enttype %s / pID %d / xy %d:%d / param %d:%d\n", entType, pID, x, y, param1, param2)
			g.enterPerson(entType, pID, x, y, param1, param2)
		}
		g.scanCoins()
		g.loop()
		g.answer = g.chooseBingo()

		// this will choose one of random actions
		actions := []string{"left", "right", "up", "down", "stay"}
		// use `os.Stderr` to print for debugging
		fmt.Fprintf(os.Stderr, "%s\n", actions[g.answer])

		// bot action
		fmt.Println(actions[g.answer])
	}
}
