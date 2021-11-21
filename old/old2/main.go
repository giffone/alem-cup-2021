package main

import (
	"fmt"
	"os"
	"time"
)

const (
	heightGlob   = 11
	widthGlob    = 13
	none         = 0
	oneStep      = 1
	twoStep      = 2
	equal        = true
	notEqual     = false
	byDiagonal   = 8    // steps by lines and diagonales
	byX          = 4    // steps by lines (up-down,left-right)
	maximum      = 1000 // the maximum steps (impossible large)
	tickLife     = 13
	pl1          = byte(80)  // 'P' - player without
	pl1_bonus    = byte(81)  // 'Q' - player with bonus
	pl1_knife    = byte(82)  // 'R' - player with knife
	pl1_hero     = byte(83)  // 'S' - player with knife and bonus
	pl2          = byte(112) // 'p' - player 2
	dots         = byte(46)  // '.' dot nil
	walls        = byte(33)  // '!' - wall
	golds        = byte(35)  // '#' - coin
	bonuses      = byte(98)  // 'b' - bonus
	knifeses     = byte(100) // 'd' - knife
	monster      = byte(77)  // 'M' - monster
	monsterArea1 = byte(111) // 'o' - area around monster
	monsterArea2 = byte(110) // 'n' - area around monster
)

type global struct {
	tickKnifeMy int
	tickKnife   map[int]map[int]int // map[x][y][current tick]
	tickBonus   map[int]map[int]int // map[x][y][current tick]
	yesterday   [11][13]byte
}

type game struct {
	id, tick             int                       // incomign options of the game
	stackUnits           map[byte]map[int][][2]int // coordinations of closer coin      // stackUnits[golds][5 steps][...][x,y]
	steps                map[byte]int              // save minimal steps to found coin  // steps[golds] = 5 steps(min)
	stackWaysToGo        map[byte]map[int][][2]int // coordinations next steps          // stackWaysToGo[golds][5 steps][...][x,y]
	stackWaysToGoExtreme bool
	units                [3]byte //
	maps
	player
	mons
}

type maps struct {
	arena       [11][13]byte // main map
	arenaBuf    [11][13]byte // map for steps
	arenaOrigin [11][13]byte // main map

}

type player struct {
	x, y, bns, knf   int
	answer           int    // choosen step
	newXY            [2]int // coordinates of choosen step
	runAway          []int  // coordinates of monster if need run away
	dead             int    // if dead = 2(monster founded),found money before die
	lastChanseGoAway bool   // if founded two steps go away
	nearEnemy        bool   // if monster one step away
}

type mons struct {
	monsterXY [][2]int
}

func (g *game) scanWays(gl *global) {
	g.stackWaysToGo = make(map[byte]map[int][][2]int)

	g.loop(gl, golds, 3)
	g.loop(gl, knifeses, 3)
	g.loop(gl, bonuses, 3)
	if len(g.stackWaysToGo) > 0 {
		return
	}
	fmt.Fprint(os.Stderr, "extreme loop\n")
	g.stackWaysToGoExtreme = true
	g.loop(gl, golds, 1)
	g.loop(gl, knifeses, 1)
	g.loop(gl, bonuses, 1)
}

func (g *game) loop(gl *global, unit byte, limit int) {
	step := g.steps[unit] // show saved minimal steps for unit

	lStack := len(g.stackUnits[unit])

	if lStack == 0 {
		return
	}
	if lStack < limit {
		limit = lStack
	}

	for limit != 0 {
		if stack, ok := g.stackUnits[unit][step]; ok {
			for _, xy := range stack {
				if g.stackWaysToGoExtreme {
					g.arenaBuf = g.arenaOrigin
				} else {
					g.arenaBuf = g.arena
				}
				g.algorithm(gl, [][2]int{xy}, unit)
			}
			limit--
		}
		step++
	}
	g.arenaBuf = [11][13]byte{} // erase
}

func (g *game) algorithm(gl *global, stack [][2]int, unit byte) {
	count := 1 // first step on map
	s := &settings{
		steps:         oneStep,                                // depth of steps
		mode:          byX,                                    // 4 steps [up-down-left-right]
		eq:            equal,                                  // if founded unit on map,check if equal or not equal to change
		canStepTo:     []byte{dots, bonuses, knifeses, golds}, // unit or value that need to find
		breakIfPlayer: true,                                   // if found player, break search and scan
		player:        [2]byte{pl1, pl1_hero},                 // range of player what can find
	}
	if g.stackWaysToGoExtreme { // if search not found free coins, new search ignore monsters
		s.canStepTo = []byte{dots, bonuses, knifeses, golds, monster, monsterArea1, monsterArea2}
	}
	for {
		s.stack = nil                    // need to erase for new loop!!! (only second step without first)
		for _, position := range stack { // the positions of coins
			s.x = position[0]                        // start position width
			s.y = position[1]                        // start position height
			s.canChangeTo = byte(count + 48)         // unit or value that need to insert if founded
			player := s.markOnMap(&g.arenaBuf, true) // making 4 or 8 steps and adding on map with saving all founded first step
			if player {
				if _, ok := g.stackWaysToGo[unit]; !ok {
					g.stackWaysToGo[unit] = make(map[int][][2]int) // if founded player
				}
				g.stackWaysToGo[unit][count] = append(g.stackWaysToGo[unit][count], position) // save current x,y that founded player
				return
			}
		}
		if len(s.stack) == 0 { // if not founded new coins
			return
		}
		stack = s.stack // erase previous and add new list of steps
		// g.print("buf")
		count++
	}
}

func (g *game) scanWhatHave(gl *global) {
	g.stackUnits = make(map[byte]map[int][][2]int)
	g.steps = make(map[byte]int)

	for i := 0; i < heightGlob; i++ {
		for j := 0; j < widthGlob; j++ {
			unit := g.arena[i][j]

			if unit != golds && unit != knifeses && unit != bonuses {
				continue
			}

			if g.ignoreCoins(j, i) { // ignore coin if closed by walls
				g.arena[i][j] = walls
				continue
			}

			steps := closer(g.player.x, g.player.y, j, i)

			if unit != golds { // add or not bonuses/knifeses to stack
				if g.knifeBonusReset(gl, unit, j, i, steps) {
					continue
				}
			}

			if _, ok := g.stackUnits[unit]; !ok {
				g.stackUnits[unit] = make(map[int][][2]int)
			}
			g.stackUnits[unit][steps] = append(g.stackUnits[unit][steps], [2]int{j, i})

			if _, ok := g.steps[unit]; !ok {
				g.steps[unit] = maximum
			}
			if g.steps[unit] > steps {
				g.steps[unit] = steps
			}
		}
	}
}

func (g *game) knifeBonusReset(gl *global, unit byte, x, y, steps int) bool {
	tick := 0
	if unit == bonuses {
		tick = gl.tickBonus[x][y]
	} else if unit == knifeses {
		tick = gl.tickKnife[x][y]
	}
	fmt.Fprintf(os.Stderr, "knife-bonus(x%d:y%d): tick saved on tick - %d steps to it - %d\n", x, y, tick, steps)
	return 15-(g.tick-tick) < steps
}

func (g *game) ignoreCoins(x, y int) bool {
	wall := 0

	for i := 0; i < byX; i++ { // loop moving (up-down-left-right) without diagonals
		if in, width, height := inRange(x, y, i, oneStep); in {
			if scan(width, height, &g.arena, walls) { // if cell = wall
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
			} else if param1 != 0 && param2 == 0 {
				g.arena[y][x] = pl1_knife
			} else if param1 == 0 && param2 != 0 {
				g.arena[y][x] = pl1_bonus
			} else {
				g.arena[y][x] = pl1_hero
			}
			g.player.x = x
			g.player.y = y
		} else {
			g.arena[y][x] = pl2
		}
	}

	if entType == "m" { // if monster
		g.arena[y][x] = monster
		n := twoStep // number of layers the monster area
		stack := [][2]int{}
		stack = append(stack, [2]int{x, y})
		s := &settings{
			steps:         oneStep,
			mode:          byX,
			eq:            notEqual,
			canStepTo:     []byte{walls, pl1, pl1_bonus, pl1_knife, monster, monsterArea1, monsterArea2},
			canChangeTo:   monsterArea1,
			breakIfPlayer: true,
			player:        [2]byte{pl1, pl1_bonus}, // range of player what can find
		}
		start := true
		for n != 0 {
			s.stack = nil // need to erase for new loop!!! (only second step without first)
			for _, position := range stack {
				s.x = position[0]
				s.y = position[1]
				player := s.markOnMap(&g.arena, true)
				if player {
					if n == twoStep {
						g.nearEnemy = true
					}
					fmt.Fprintf(os.Stderr, "monster(x%d:y%d): in my territory founded player (x%d:y%d)\n", x, y, s.width, s.height)
					if start {
						g.dead++
						start = false
					}

					g.runAway = []int{x, y}
					return
				}
				s.mode = byX
			}
			n--
			stack = s.stack // erase previous and add new list of steps
			if n == oneStep {
				s.canStepTo = []byte{walls, pl1, pl1_bonus, pl1_knife, monster, monsterArea1, monsterArea2, knifeses} // no erase knife!!
				s.canChangeTo = monsterArea2
			}
		}
	}
}

func (g *game) enterMonsters() {
	g.dead = 0 // erase countdown founded monster
	for _, person := range g.monsterXY {
		g.enterPerson("m", 0, person[0], person[1], 0, 0)
	}
}

func (g *game) stepsMain(gl *global) {
	ok := g.chooseStep(gl)
	if !ok {
		fmt.Fprint(os.Stderr, "player: can not choose the way, will stay\n")
		g.answer = 4 // stay
		return
	}
	g.answer = g.readCell()
}

func (g *game) chooseStep(gl *global) bool {
	g.steps = make(map[byte]int) // erase data
	g.sortUnit()                 // make list of most short ways

	if len(g.steps) == 0 {
		return false
	}

	goldSteps := g.steps[golds]
	knifeSteps := g.steps[knifeses]
	bonusSteps := g.steps[bonuses]

	if bonusSteps <= goldSteps && bonusSteps <= knifeSteps {
		if ok := g.readCellGo(bonuses, bonusSteps); ok {
			fmt.Fprint(os.Stderr, "player: 1:, better bonus\n")
			return ok
		}
	}
	if knifeSteps < goldSteps && knifeSteps < bonusSteps && len(g.monsterXY) > 0 {
		if ok := g.readCellGo(knifeses, knifeSteps); ok {
			fmt.Fprintf(os.Stderr, "player: 2:, better knife\n")
			return ok
		}
	}
	if goldSteps > bonusSteps && goldSteps-bonusSteps <= 5 {
		if ok := g.readCellGo(bonuses, bonusSteps); ok {
			fmt.Fprint(os.Stderr, "player: 3:, better bonus\n")
			return ok
		}
	}

	if knifeSteps == goldSteps && len(g.monsterXY) > 0 {
		if ok := g.readCellGo(knifeses, knifeSteps); ok {
			fmt.Fprint(os.Stderr, "player: 4:, better knife\n")
			return ok
		}
	}
	if goldSteps > knifeSteps && goldSteps-knifeSteps <= 2 && len(g.monsterXY) > 0 {
		if ok := g.readCellGo(knifeses, knifeSteps); ok {
			fmt.Fprint(os.Stderr, "player: 5:, better knife\n")
			return ok
		}
	}
	if goldSteps > bonusSteps && goldSteps-bonusSteps <= 2 {
		if ok := g.readCellGo(bonuses, bonusSteps); ok {
			fmt.Fprint(os.Stderr, "player: 6:, better bonus\n")
			return ok
		}
	}
	if bonusSteps == goldSteps {
		if ok := g.readCellGo(bonuses, bonusSteps); ok {
			fmt.Fprint(os.Stderr, "player: 7:, better bonus\n")
			return ok
		}
	}
	fmt.Fprintf(os.Stderr, "player: 8:, last gold\n")
	return g.readCellGo(golds, goldSteps)
}

func (g *game) readCellGo(key byte, step int) bool {
	if _, ok := g.stackWaysToGo[key][step]; !ok {
		return false
	}
	for _, xy := range g.stackWaysToGo[key][step] {
		g.newXY = [2]int{xy[0], xy[1]}
		return true
	}
	return false
}

func (g *game) sortUnit() {
	for _, unit := range g.units {
		g.steps[unit] = g.sortUnitMap(unit)
	}
}

func (g *game) sortUnitMap(unit byte) int {
	min := maximum
	for steps := range g.stackWaysToGo[unit] {
		if steps < min {
			min = steps
		}
	}
	return min
}

func (g *game) playerSaveArea() {
	s := &settings{
		x:         g.x,        // current position player on map width
		y:         g.y,        // current position player on map height
		steps:     oneStep,    // depth of steps
		mode:      byDiagonal, // 4 steps [up-down-left-right]
		eq:        equal,
		canStepTo: []byte{monster, monsterArea1, monsterArea2},
	}
	s.markOnMap(&g.arena, true)
	if len(s.stack) == 0 {
		return
	}
	g.whereMonster()

	if len(g.runAway) == 0 {
		g.runAway = []int{maximum, maximum}
	}
	fmt.Fprintf(os.Stderr, "player(x%d:y%d): in my area founded monster(x%d:y%d)\n", s.x, s.y, g.runAway[0], g.runAway[1])
}

func (g *game) whereMonster() {
	if len(g.monsterXY) == 1 {
		g.runAway = []int{g.monsterXY[0][0], g.monsterXY[0][1]}
		return
	}
	min := maximum
	for _, position := range g.monsterXY {
		if n := closer(g.player.x, g.player.y, position[0], position[1]); n < min {
			min = n
			g.runAway = []int{position[0], position[1]}
		}
	}
}

func (g *game) runRun() {
	fmt.Fprintf(os.Stderr, "player: choosen runRun\n")
	// 0 "up"
	// 1 "down"
	// 2 "left"
	// 3 "right"
	stack := []int{}                                        // list of save sides
	if g.runAway[0] == maximum || g.runAway[1] == maximum { // if monster somewhere closer (without coordination)
		fmt.Fprint(os.Stderr, "player: in my area founded monster, somewhere...!\n")
		stack = []int{0, 1, 2, 3}
	} else {
		if g.x < g.runAway[0] { // if player by left side of the monster
			fmt.Fprint(os.Stderr, "player: in my area founded monster, at right side\n")
			stack = append(stack, 2) // left
		} else if g.x > g.runAway[0] { // if monster coordination is known
			fmt.Fprint(os.Stderr, "player: in my area founded monster, at left side\n")
			stack = append(stack, 3) // right
		} else {
			fmt.Fprint(os.Stderr, "player: in my area founded monster, at middle(left-right) side\n")
			stack = append(stack, 2, 3) // left and right
		}

		if g.y < g.runAway[1] { // if player above of the monster
			fmt.Fprint(os.Stderr, "player: in my area founded monster, at down side\n")
			stack = append(stack, 0) // up
		} else if g.y > g.runAway[1] {
			fmt.Fprint(os.Stderr, "player: in my area founded monster, at up side\n")
			stack = append(stack, 1) // down
		} else {
			fmt.Fprint(os.Stderr, "player: in my area founded monster, at middle(up-down) side\n")
			stack = append(stack, 0, 1) // down
		}
	}
	n := 2
	units := []byte{golds, knifeses, bonuses}
loop:
	for n != 0 {
		for _, side := range stack {
			in2, x2, y2 := inRange(g.x, g.y, side, twoStep)
			if !in2 {
				continue
			}
			ok2 := scan(x2, y2, &g.arena, units...) // check if it free cell at two step
			if !ok2 {
				continue
			}
			in, x, y := inRange(g.x, g.y, side, oneStep) // check if it free cell at one step too
			if !in {
				continue
			}
			ok := scan(x, y, &g.arena, dots, golds, knifeses, bonuses) // check if it free cell at two step
			if !ok {
				continue
			}
			fmt.Fprintf(os.Stderr, "player(x%d:y%d): i see free way at side - %s till TWO step at(x%d:y%d)\n", g.x, g.y, sideName(side), x2, y2)
			g.newXY = [2]int{x, y}
			g.lastChanseGoAway = true
			break loop
		}
		units = []byte{dots}
		n--
	}

	if !g.lastChanseGoAway {
		if g.x == 6 && g.y > 3 && g.y < 7 && len(g.monsterXY) < 3 { // safe area
			if g.nearEnemy {
				g.answer = 4 // stay
				fmt.Fprintf(os.Stderr, "player: choosen stay at 6\n")
				return
			}
		}
		for _, side := range stack {
			in, x, y := inRange(g.x, g.y, side, oneStep)
			if !in {
				continue
			}
			ok := scan(x, y, &g.arena, dots)
			if !ok {
				continue
			}
			fmt.Fprintf(os.Stderr, "player(x%d:y%d): i see free way at side - %s till ONE step at(x%d:y%d)\n", g.x, g.y, sideName(side), x, y)
			g.newXY = [2]int{x, y}
			break
		}
	}
	g.answer = g.readCell()
}

func (g *game) checkGold() {
	if g.lastChanseGoAway {
		return
	}
	s := &settings{
		x:         g.x,     // current position player on map width
		y:         g.y,     // current position player on map height
		steps:     oneStep, // depth of steps
		mode:      byX,     // 4 steps [up-down-left-right] and diagonal
		eq:        equal,
		canStepTo: []byte{golds},
	}
	s.markOnMap(&g.arenaOrigin, true)
	if len(s.stack) == 0 {
		return
	}
	g.newXY = [2]int{s.width, s.height}
	g.answer = g.readCell()
	fmt.Fprintf(os.Stderr, "player(x%d:y%d): i founded gold at(x%d:y%d) before die!\n", g.x, g.y, s.width, s.height)
}

func (g *game) checkKnife(gl *global) {
	s := &settings{
		x:         g.x,     // current position player on map width
		y:         g.y,     // current position player on map height
		steps:     oneStep, // depth of steps
		mode:      byX,     // 4 steps [up-down-left-right] and diagonal
		eq:        equal,
		canStepTo: []byte{knifeses},
	}
	s.markOnMap(&g.arenaOrigin, true)
	if len(s.stack) == 0 {
		return
	}
	g.newXY = [2]int{s.width, s.height}
	g.answer = g.readCell()
	fmt.Fprintf(os.Stderr, "player(x%d:y%d): i founded knife at(x%d:y%d)\n", g.x, g.y, s.width, s.height)
}

func sideName(n int) string {
	// 0 "up"
	// 1 "down"
	// 2 "left"
	// 3 "right"
	switch n {
	case 0:
		return "up"
	case 1:
		return "down"
	case 2:
		return "left"
	default:
		return "right"
	}
}

func (g *game) readCell() int {
	// 0 "left"
	// 1 "right"
	// 2 "up"
	// 3 "down"
	x := g.newXY[0]
	y := g.newXY[1]
	if x == g.player.x { // width equal
		if y < g.player.y { // unit above
			return 2 // up
		}
		return 3 // down
	}
	if y == g.player.y { // height equal
		if x < g.player.x { // unit to the left
			return 0 // left
		}
		return 1 // right
	}
	return 4 // stay
}

func (g *game) printArena() {
	for i, line := range g.arena {
		fmt.Fprintf(os.Stderr, "%2d %s\n", i, line)
	}
}

type settings struct {
	set           string   // sort settings by key if need
	x, y          int      // first point
	steps         int      // how many steps need in depth
	mode          int      // 4 steps (up-down-left-right) or 8 steps by diagonal too
	eq            bool     // found equal or not equal
	canStepTo     []byte   // list of persons need to find
	canChangeTo   byte     // if founded, what to write in cell
	breakIfPlayer bool     // if founded player
	player        [2]byte  // range witch player
	stack         [][2]int // buffer for recieved(x,y) if need
	width, height int      // current scanned coordinations
}

func (s *settings) markOnMap(arena *[11][13]byte, save bool) bool {
	if !save {
		s.stack = nil
	}
	// stackBuf := [][2]int{}
	for j := 0; j < s.mode; j++ { // loop moving (up-down-left-right) with or without diagonals
		in, width, height := inRange(s.x, s.y, j, s.steps)
		if !in {
			continue
		}
		if s.breakIfPlayer {
			if arena[height][width] >= s.player[0] && arena[height][width] <= s.player[1] { // if my player stands on cell
				return true // player founded
			}
		}
		if s.eq {
			if scan(width, height, arena, s.canStepTo...) { // if cell = wall or player
				if s.canChangeTo != none {
					arena[height][width] = s.canChangeTo
				}
				s.stack = append(s.stack, [2]int{width, height})
				s.width, s.height = width, height
			}
		} else {
			if !scan(width, height, arena, s.canStepTo...) { // if cell = wall or player
				if s.canChangeTo != none {
					arena[height][width] = s.canChangeTo
				}
				s.stack = append(s.stack, [2]int{width, height})
				s.width, s.height = width, height
			}
		}
	}
	return false // player not founded
}

func inRange(width, height, method, steps int) (in bool, x int, y int) {
	switch method {
	case 0: // up
		x = width
		y = height - steps
	case 1: // down
		x = width
		y = height + steps
	case 2: // left
		x = width - steps
		y = height
	case 3: // right
		x = width + steps
		y = height
	case 4: // up-left
		x = width - steps
		y = height - steps
	case 5: // up-right
		x = width + steps
		y = height - steps
	case 6: // down-left
		x = width - steps
		y = height + steps
	case 7: // down-right
		x = width + steps
		y = height + steps
	}

	if !limit(x, y) {
		return false, 0, 0
	}
	return true, x, y
}

func limit(x, y int) bool {
	if x < 0 || x >= widthGlob {
		return false
	}
	if y < 0 || y >= heightGlob {
		return false
	}
	return true
}

func scan(x, y int, arena *[11][13]byte, person ...byte) bool {
	for _, unit := range person {
		if arena[y][x] == unit {
			return true
		}
	}
	return false
}

func closer(x, y, x2, y2 int) int {
	return un(x-x2) + un(y-y2)
}

// func un makes uint
func un(n int) int {
	if n < 0 {
		return n * -1
	}
	return n
}

func main() {
	gl := new(global)
	gl.tickKnife = make(map[int]map[int]int)
	gl.tickBonus = make(map[int]map[int]int)
	for true {
		timing := time.Now()
		g := new(game)
		g.units = [3]byte{golds, knifeses, bonuses}

		var w, h, playerID, tick int
		fmt.Scan(&w, &h, &playerID, &tick)
		g.id = playerID
		g.tick = tick
		// read map
		for i := 0; i < heightGlob; i++ {
			line := ""
			fmt.Scan(&line)
			for j := 0; j < widthGlob; j++ {
				g.arena[i][j] = line[j]
				if line[j] == knifeses {
					if gl.yesterday[i][j] == '.' || g.tick == 300 {
						if _, ok := gl.tickKnife[j]; !ok {
							gl.tickKnife[j] = make(map[int]int)
						}
						gl.tickKnife[j][i] = g.tick
					}
				}
				if line[j] == bonuses {
					if gl.yesterday[i][j] == '.' || g.tick == 300 {
						if _, ok := gl.tickBonus[j]; !ok {
							gl.tickBonus[j] = make(map[int]int)
						}
						gl.tickBonus[j][i] = g.tick
					}
				}
			}
		}

		// number of entities
		var n int
		fmt.Scan(&n)

		// read entities
		for i := 0; i < n; i++ {
			var entType string
			var pID, x, y, param1, param2 int
			fmt.Scan(&entType, &pID, &x, &y, &param1, &param2)
			fmt.Fprintf(os.Stderr, "enttype %s / pID %d / xy %d:%d / param %d:%d\n", entType, pID, x, y, param1, param2)
			if entType == "m" {
				g.monsterXY = append(g.monsterXY, [2]int{x, y})
				continue
			}
			if entType == "p" && pID == g.id {
				g.player = player{
					x:   x,
					y:   y,
					knf: param1,
					bns: param2,
				}
				continue
			}
		}
		g.enterPerson("p", g.id, g.x, g.y, g.knf, g.bns)
		g.arenaOrigin = g.arena
		if gl.tickKnifeMy <= 0 && g.knf <= 0 {
			g.enterMonsters()
			g.playerSaveArea()
		}
		g.printArena()
		if len(g.runAway) == 0 {
			g.scanWhatHave(gl)
			g.scanWays(gl)
			g.stepsMain(gl)
		} else {
			g.runRun()
			g.checkGold() // it will run only not founded two step go away
			g.checkKnife(gl)
		}

		gl.tickKnifeMy--
		if g.arena[g.newXY[1]][g.newXY[0]] == knifeses {
			gl.tickKnifeMy = tickLife
		}
		gl.yesterday = g.arena // save last map
		actions := []string{"left", "right", "up", "down", "stay"}
		fmt.Fprintf(os.Stderr, "%s, knife life - %d, ", actions[g.answer], gl.tickKnifeMy)

		// bot action
		fmt.Println(actions[g.answer])
		fmt.Fprintf(os.Stderr, "timing: %s\n", time.Since(timing))
	}
}
