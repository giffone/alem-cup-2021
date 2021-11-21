package main

import (
	"fmt"
	"os"
)

const (
	heightGlob      = 11
	widthGlob       = 13
	oneStep         = 1
	twoStep         = 2
	byDiagonal      = 8         // steps by lines and diagonales
	byX             = 4         // steps by lines (up-down,left-right)
	pl1             = byte(80)  // 'P' - player without
	pl1_bonus       = byte(81)  // 'Q' - player with bonus
	pl1_knife       = byte(82)  // 'R' - player with knife
	pl1_hero        = byte(83)  // 'S' - player with knife and bonus
	pl2             = byte(112) // 'p' - player without
	monster         = byte(77)  // 'M' - monster
	monsterArea     = byte(111) // 'o' - area around monster
	monsterAreaCell = 2         // how many cells around
)

type game struct {
	id, tick, answer int                    // incomign options of the game
	stackCoin        map[int]map[int][2]int // coordinations of closer coin
	steps            int                    // save minimal steps to found coin
	cellGo           map[byte][2]int        // coordinations next steps
	maps
	player
	monsters
}

type maps struct {
	arena        [11][13]byte // main map
	arenaPersons [11][13]byte // map with persons
	arenaBuf     [11][13]byte // map for steps
}

type player struct {
	super bool
	x     int // width
	y     int // height
}

type monsters struct {
	position [][2]int
}

func (g *game) loop() {
	g.cellGo = make(map[byte][2]int)
	lCoin := len(g.stackCoin)
	limit := 3

	if lCoin == 0 {
		return
	}
	if lCoin < 3 {
		limit = lCoin
	}

	for limit != 0 {
		if stack, ok := g.stackCoin[g.steps]; ok {
			for _, xy := range stack {
				// g.print("buf")
				if g.player.super { // if have knife, ignore monsters (map without monsters)
					g.arenaBuf = g.arena
				} else { // if no knife, should see monsters
					g.arenaBuf = g.arenaPersons
				}
				g.algorithm([][2]int{xy})
			}
			limit--
		}
		g.steps++
	}
}

func (g *game) algorithm(stack [][2]int) {
	count := byte('1') // first step on map
	for {
		stackBuf := [][2]int{}
		for _, position := range stack { // the positions of coins
			for i := 0; i < byX; i++ { // loop moving (up-down-left-right) without diagonals
				in, width, height := g.inRange(position[0], position[1], i, oneStep)
				if !in {
					continue
				}
				if g.scan(width, height, "buf", '.') { // if cell = dot(nil)
					g.rewrite(width, height, count, "buf")
					stackBuf = append(stackBuf, [2]int{width, height}) // save current x,y to the stack
					continue
				}
				if g.checkPlayer(width, height) { // if player on cell
					g.cellGo[count] = position // save current x,y to the stack
					return
				}
			}
		}
		if len(stackBuf) == 0 { // if not founded new coins
			return
		}
		stack = stackBuf // erase previous
		count++
	}
}

func (g *game) inRange(width, height, method, radius int) (in bool, x int, y int) {
	switch method {
	case 0: // up
		x = width
		y = height - radius
	case 1: // down
		x = width
		y = height + radius
	case 2: // left
		x = width - radius
		y = height
	case 3: // right
		x = width + radius
		y = height
	case 4: // up-left
		x = width - radius
		y = height - radius
	case 5: // up-right
		x = width + radius
		y = height - radius
	case 6: // down-left
		x = width - radius
		y = height + radius
	case 7: // down-right
		x = width + radius
		y = height + radius
	}

	if !g.limit(x, y) {
		return false, 0, 0
	}
	return true, x, y
}

func (g *game) limit(x, y int) bool {
	if x < 0 || x >= widthGlob {
		return false
	}
	if y < 0 || y >= heightGlob {
		return false
	}
	return true
}

func (g *game) scan(x, y int, sourse string, symbol ...byte) bool {
	if sourse == "origin" { // scan in a origin map
		for _, unit := range symbol {
			if g.arena[y][x] == unit {
				return true
			}
		}
	}

	if sourse == "person" { // scan in a origin map
		for _, unit := range symbol {
			if g.arenaPersons[y][x] == unit {
				return true
			}
		}
	}

	if sourse == "buf" { // scan in a copy of origin map
		for _, unit := range symbol {
			if g.arenaBuf[y][x] == unit {
				return true
			}
		}
	}
	return false
}

func (g *game) rewrite(x, y int, symbol byte, sourse string) {
	if sourse == "origin" { // rewrite in a origin map
		g.arena[y][x] = symbol
	}

	if sourse == "person" { // rewrite in a copy of origin map
		g.arenaPersons[y][x] = symbol
	}

	if sourse == "buf" { // rewrite in a copy of origin map
		g.arenaBuf[y][x] = symbol
	}
}

func (g *game) checkPlayer(x, y int) bool {
	if g.arenaPersons[y][x] >= pl1 && g.arenaPersons[y][x] <= pl1_hero { // if my player stands on cell
		return true
	}
	return false
}

func (g *game) chooseCell() int {
	min := byte(255)
	for steps := range g.cellGo {
		if steps < min {
			min = steps
		}
	}
	if min != 255 {
		return g.readCell(min)
	}
	x, y := g.randomStep(g.player.x, g.player.y)
	fmt.Fprintf(os.Stderr, "random %d:%d / new %d:%d \n", g.player.x, g.player.y, x, y)
	g.cellGo[0] = [2]int{x, y}
	return g.readCell(0)
}

func (g *game) readCell(key byte) int {
	// 0 "left"
	// 1 "right"
	// 2 "up"
	// 3 "down"i
	if g.cellGo[key][0] == g.player.x { // width equal
		if g.cellGo[key][1] < g.player.y {
			return 2 // up
		}
		return 3 // down
	}
	if g.cellGo[key][1] == g.player.y { // height equal
		if g.cellGo[key][0] < g.player.x {
			return 0 // left
		}
		return 1 // right
	}
	return 4 // stay
}

func (g *game) scanCoins() {
	g.stackCoin = make(map[int]map[int][2]int)
	g.steps = 1000
	count := 0
	for i := 0; i < heightGlob; i++ {
		for j := 0; j < widthGlob; j++ {
			if g.arena[i][j] != '#' && g.arena[i][j] != 'd' && g.arena[i][j] != 'b' {
				continue
			}
			if g.ignoreCoins(j, i) { // ignore coin if closed by walls
				g.arena[i][j] = 'X'
				g.arenaPersons[i][j] = 'X'
				continue
			}
			steps := un(g.player.y-i) + un(g.player.x-j)
			if _, ok := g.stackCoin[steps]; !ok {
				g.stackCoin[steps] = make(map[int][2]int)
			}
			g.stackCoin[steps][count] = [2]int{j, i}
			if g.steps > steps {
				g.steps = steps
			}
			count++
		}
	}
}

func (g *game) ignoreCoins(x, y int) bool {
	wall := 0
	for i := 0; i < byX; i++ { // loop moving (up-down-left-right) without diagonals
		if in, width, height := g.inRange(x, y, i, oneStep); in {
			if g.scan(width, height, "origin", '!') { // if cell = wall
				wall++
			}
		} else {
			wall++
		}
	}
	return wall == 4
}

func (g *game) enterPerson(entType string, pID, x, y, param1, param2 int) {
	//     x y b k
	// p 1 0 0 0 0
	// m 0 3 6 0 0
	// m 0 8 4 0 0
	// p 2 12 10 0 0
	if entType == "p" { // if player
		if pID == g.id {
			if param1 == 0 && param2 == 0 {
				g.arena[y][x] = pl1
				g.arenaPersons[y][x] = pl1
				g.player.super = false
			} else if param1 != 0 && param2 == 0 {
				g.arena[y][x] = pl1_knife
				g.arenaPersons[y][x] = pl1_knife
				g.player.super = true
			} else if param1 == 0 && param2 != 0 {
				g.arena[y][x] = pl1_bonus
				g.arenaPersons[y][x] = pl1_bonus
			} else {
				g.arena[y][x] = pl1_hero
				g.arenaPersons[y][x] = pl1_hero
				g.player.super = true
			}
			g.player.x = x
			g.player.y = y
		} else {
			g.arenaPersons[y][x] = pl2
		}
	}

	if entType == "m" { // if monster
		g.monsters.position = append(g.monsters.position, [2]int{x, y})
		g.arenaPersons[y][x] = monster

		for i := 0; i < byDiagonal; i++ { // loop moving (up-down-left-right) with diagonals - first step in depth
			in, width, height := g.inRange(x, y, i, oneStep)
			if !in {
				continue
			}
			if g.scan(width, height, "person", '!', 'P', 'Q', 'R', 'S', 'p') { // if cell = wall or player
				continue
			}
			g.rewrite(width, height, monsterArea, "person")
			if i >= byX {
				continue
			}
			in2, width2, height2 := g.inRange(x, y, i, twoStep)
			if !in2 {
				continue
			}
			if g.scan(width2, height2, "person", '!', 'P', 'Q', 'R', 'S', 'p') { // if cell = wall or player
				continue
			}
			g.rewrite(width2, height2, monsterArea, "person")
		}
	}
}

func (g *game) lastChance() bool {
	for _, monster := range g.monsters.position {
		for i := 0; i < byX; i++ {
			if !g.diagonalMonster(monster[0], monster[1], i) {
				continue
			}
			g.answer = 4 // stay
			fmt.Fprintf(os.Stderr, "diagonal monster(%d:%d) true, will stay\n", monster[0], monster[1])
			return true
		}
	}
	return false
}

func (g *game) diagonalMonster(width, height, method int) bool {
	x, y := 0, 0
	switch method {
	case 0: // up-left
		x = width - 1
		y = height - 1
	case 1: // up-right
		x = width + 1
		y = height - 1
	case 2: // down-left
		x = width - 1
		y = height + 1
	case 3: // down-right
		x = width + 1
		y = height + 1
	}

	if !g.limit(x, y) {
		return false
	}
	return g.arena[y][x] == 'P'
}

func (g *game) randomStep(x, y int) (int, int) {
	stack := [][2]int{}
	n := 2 // 2 chanses to find way: first find units {!, M, m, p}, if no way - retry with {!, M, p}
	units := []byte{'!', 'M', 'o', 'p'}
	for n != 0 {
		n--
		for i := 0; i < byX; i++ {
			in, width, height := g.inRange(x, y, i, oneStep)
			if !in {
				continue
			}
			if g.scan(width, height, "person", units...) {
				continue
			}
			in2, width2, height2 := g.inRange(width, height, i, twoStep)
			if !in2 {
				stack = append(stack, [2]int{width, height})
				continue
			}
			if !g.scan(width2, height2, "person", units...) {
				return width, height
			}
		}
		for _, position := range stack {
			return position[0], position[1]
		}
		units = []byte{'!', 'M', 'p'}
	}
	return 0, 0 // if no way, it will choose 'stay'
}

// func un makes uint
func un(n int) int {
	if n < 0 {
		return n * -1
	}
	return n
}

func (g *game) printArena() {
	for i, line := range g.arenaPersons {
		fmt.Fprintf(os.Stderr, "%2d %s\n", i, line)
	}
}

func main() {
	for true {
		g := new(game)
		g.answer = 4
		var w, h, playerID, tick int
		fmt.Scan(&w, &h, &playerID, &tick)
		g.id = playerID
		g.tick = tick
		fmt.Fprintf(os.Stderr, "ticks %d / me %d \n", g.tick, playerID)

		// read map
		for i := 0; i < heightGlob; i++ {
			line := ""
			fmt.Scan(&line)
			for j := 0; j < widthGlob; j++ {
				g.arena[i][j] = line[j]
			}
		}

		// number of entities
		var n int
		fmt.Scan(&n)
		g.arenaPersons = g.arena
		// read entities
		for i := 0; i < n; i++ {
			var entType string
			var pID, x, y, param1, param2 int
			fmt.Scan(&entType, &pID, &x, &y, &param1, &param2)
			fmt.Fprintf(os.Stderr, "enttype %s / pID %d / xy %d:%d / param %d:%d\n", entType, pID, x, y, param1, param2)
			g.enterPerson(entType, pID, x, y, param1, param2)
		}
		if !g.lastChance() {
			g.scanCoins()
			g.loop()
		}
		g.printArena()
		// if len(g.stackCoin) == 0 { // game over
		// 	return
		// }
		g.answer = g.chooseCell()

		// this will choose one of random actions
		actions := []string{"left", "right", "up", "down", "stay"}
		// use `os.Stderr` to print for debugging
		fmt.Fprintf(os.Stderr, "%s\n", actions[g.answer])

		// bot action
		fmt.Println(actions[g.answer])
	}
}
