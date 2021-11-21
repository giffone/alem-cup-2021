package main

import (
	"fmt"
	"os"
	"time"
)

func (g *game) printArena() {
	// for i, line := range g.arena {
	// 	fmt.Fprintf(os.Stderr, "%2d  ", i)
	// 	for _, cell := range line {
	// 		fmt.Fprintf(os.Stderr, "%s ", string(cell.unit))
	// 	}
	// 	fmt.Fprint(os.Stderr, "\n")
	// }
	fmt.Fprint(os.Stderr, "       0      1      2      3      4      5"+
		"      6      7      8      9     10     11     12\n")
	for i, line := range g.arena {
		fmt.Fprintf(os.Stderr, "%2d ", i)
		for _, cell := range line {
			if cell.unit == walls {
				fmt.Fprint(os.Stderr, "|XXXXXX")
				continue
			}
			if cell.unit == dots && cell.ratio == 0 {
				if cell.mark == 0 {
					fmt.Fprint(os.Stderr, "|      ")
					continue
				}
				fmt.Fprintf(os.Stderr, "| %1s    ", string(cell.mark))
				continue
			}
			fmt.Fprintf(os.Stderr, "| %1s%3d ", string(cell.unit), int(cell.ratio))
		}
		fmt.Fprint(os.Stderr, "|\n")
	}
}

func (g *game) printSugar() {
	for _, list := range []string{"high", "low", "risk"} {
		for _, cell := range g.player.cell[list] {
			for name, sugar := range cell.sugarAIO {
				for bonus, step := range sugar {
					fmt.Fprintf(os.Stderr, "cell (x%d:y%d in %s list, ratio %.2f) bonus %s step %d (x%d:y%d)\n",
						cell.xy[0], cell.xy[1], list, cell.ratio, string(name), step, bonus.xy[0], bonus.xy[1])
				}
			}
		}
	}
}

func (gl *global) printSaveArea() {
	fmt.Fprint(os.Stderr, "    0 1 2 3 4 5 6 7 8 9 10 11 12\n")
	for i, line := range gl.staticMap {
		fmt.Fprintf(os.Stderr, "%2d  ", i)
		for _, cell := range line {
			if cell.monsterArea {
				fmt.Fprint(os.Stderr, "+ ")
				continue
			}
			fmt.Fprint(os.Stderr, "  ")
		}
		fmt.Fprint(os.Stderr, "\n")
	}
}

const (
	heightGlobal  = 11
	widthGlobal   = 13
	none          = 0
	oneStep       = 1
	twoStep       = 2
	equal         = true
	notEqual      = false
	byDiagonal    = 8 // steps by lines and diagonales
	byX           = 4 // steps by lines (up-down,left-right)
	sugarLimit    = 14
	maximum       = 1000      // the maximum steps (impossible large)
	pl1           = byte(80)  // 'P' - player without
	pl1_bonus     = byte(81)  // 'Q' - player with bonus
	pl1_knife     = byte(82)  // 'R' - player with knife
	pl1_hero      = byte(83)  // 'S' - player with knife and bonus
	dots          = byte(46)  // '.' dot nil
	walls         = byte(33)  // '!' - wall
	golds         = byte(35)  // '#' - coin
	bonuses       = byte(98)  // 'b' - bonus
	knifeses      = byte(100) // 'd' - knife
	monster       = byte(77)  // 'M' - monster
	monsterArea1  = byte(111) // 'o' - area around monster
	monsterArea2  = byte(110) // 'n' - area around monster
	goldsRatio    = float64(50)
	knifesesRatio = float64(100)
	bonusesRatio  = float64(200)
)

// global regular memory
type global struct {
	tick           int               // current global tick
	gold, goldMine int               // calculate all golds on map and mine
	tickKnifeMine  int               // count mine ticks time left with knife (13 tick--)
	staticMap      [11][13]cellsStat // static map
}

// cellStat cells of static map
type cellsStat struct {
	monsterArea bool // save cell that was visited monster
	yesterday   byte // save unit type that was in past tick
	sugarLife   int  // count time left of bonus
}

// enterLimit scans new sugar and calculate time remain before sugar disappear
func (gl *global) enterLimit(cell *cells) {
	x, y := cell.xy[0], cell.xy[1]

	if gl.staticMap[y][x].yesterday == dots || gl.tick == 300 { // if before was dot
		gl.staticMap[y][x].sugarLife = gl.tick // remember tick at now
	}
	cell.sugarLife = sugarLimit - (gl.tick - gl.staticMap[y][x].sugarLife) // or calculate time remain 15-(295-290)=10
}

// backUpMap saves old map for compare
func (gl *global) backUpMap(arena [11][13]cells) {
	for i := 0; i < heightGlobal; i++ {
		for j := 0; j < widthGlobal; j++ {
			gl.staticMap[i][j].yesterday = arena[i][j].origin
		}
	}
}

type game struct {
	arena     [11][13]cells // main map
	countGold int
	player    player
	monsters  monsters
	sugar     sugar
}

type sugar struct {
	list  []*cells                  // list of all bonuses(knife,bonus)
	stack map[byte]map[int][]*cells // sorted list by steps
	steps map[byte]int              // list of minimal steps
}

type cells struct {
	xy                 [2]int                  // current coordinates
	ratio, unRatio     float64                 // ratio by unit
	unit, origin, mark byte                    // type of unit // origin unit if changed // mark cell
	sugar              bool                    // if it is bonus or knife
	sugarAIO           map[byte]map[*cells]int // how many bonuses can get stepping on that cell
	sugarLife          int                     // time before bonus/knife will disappear
	risk               int                     // cell risk high/low if monster near
	hide               bool

	left, right, up, down        *cells
	uLeft, uRight, dLeft, dRight *cells
}

type player struct {
	cell         map[string][]*cells // current coordinations [0] and new coordinations for next step[1]
	id, knf, bns int                 // player info
	answer       int                 // new step
}

type monsters struct {
	cell []*cells
}

// analysMap recognize sugar, calc limits and makes link between cells
func (g *game) analysMap(gl *global) {
	for i := 0; i < heightGlobal; i++ {
		for j := 0; j < widthGlobal; j++ {
			cell := &g.arena[i][j]
			cell.ratio, cell.sugar = unitRatio(cell.unit) // add bonuses ratio
			if cell.sugar {
				g.sugar.list = append(g.sugar.list, cell) // add bonuses to list
				if cell.unit != golds {
					gl.enterLimit(cell) // for knife/bonus make limit time
				} else {
					g.countGold++
					cell.sugarLife = 0 // erase old limits
				}
			}
			g.neibourGet(cell) // make communications between cells
		}
	}
}

// neibourGet connect cells
func (g *game) neibourGet(cell *cells) {
	for side := 0; side < byDiagonal; side++ {
		in, neibour := g.inRangeXY(cell, side, oneStep)
		if !in {
			continue
		}

		switch side {
		case 0: // left
			cell.left = neibour
		case 1: // right
			cell.right = neibour
		case 2: // up
			cell.up = neibour
		case 3: // down
			cell.down = neibour
		case 4: // up-left
			cell.uLeft = neibour
		case 5: // up-right
			cell.uRight = neibour
		case 6: // down-left
			cell.dLeft = neibour
		case 7: // down-right
			cell.dRight = neibour
		}
	}
}

// inRangeXY checks coordinations out of range
func (g *game) inRangeXY(cell *cells, method, steps int) (bool, *cells) {
	width, height := cell.xy[0], cell.xy[1]
	switch method {
	case 0: // left
		width -= steps
	case 1: // right
		width += steps
	case 2: // up
		height -= steps
	case 3: // down
		height += steps
	case 4: // up-left
		width -= steps
		height -= steps
	case 5: // up-right
		width += steps
		height -= steps
	case 6: // down-left
		width -= steps
		height += steps
	case 7: // down-right
		width += steps
		height += steps
	}

	if !limit(width, height) {
		return false, nil
	}
	return true, &g.arena[height][width]
}

// scanWhatHave makes map of bonuses/knifeses/golds
func (g *game) scanWhatHave(risk int) {
	g.sugar.stack = make(map[byte]map[int][]*cells) // sorted list of founded sugar
	g.sugar.steps = make(map[byte]int)              // list of minimal steps

	for _, cell := range g.sugar.list {
		if g.ignoreSugar(cell) { // ignore coin if closed by walls
			continue
		}
		if cell.risk > risk {
			continue
		}
		steps := closer(g.player.cell["player"][0].xy, cell.xy)
		unit := cell.origin

		if unit != golds { // add or not bonuses/knifeses to stack
			if cell.sugarLife < steps {
				continue
			}
		}

		if _, ok := g.sugar.stack[unit]; !ok {
			g.sugar.stack[unit] = make(map[int][]*cells)
		}
		g.sugar.stack[unit][steps] = append(g.sugar.stack[unit][steps], cell)

		if _, ok := g.sugar.steps[unit]; !ok {
			g.sugar.steps[unit] = maximum
		}
		if g.sugar.steps[unit] > steps {
			g.sugar.steps[unit] = steps
		}
	}
}

// ignoreSugar checks if it closed by walls around
func (g *game) ignoreSugar(cell *cells) bool {
	s := &sideSettings{
		current:          cell,
		eq:               equal,
		eqUnits:          []byte{walls},
		neibourNilEqUnit: true,
	}
	wall := 0
	for i := 0; i < byX; i++ {
		s.side = i
		if ok, _ := s.neibourCheck(); ok {
			wall++
		}
	}
	return wall == 4
}

func (g *game) enterMonsters() {
	for _, cell := range g.monsters.cell {
		g.enterPerson(cell, "m", 0, 0, 0)
	}
}

// enterPerson
func (g *game) enterPerson(cell *cells, entType string, pID, param1, param2 int) {
	//     x y b k
	// p 1 0 0 0 0
	// m 0 3 6 0 0
	// m 0 8 4 0 0
	// p 2 12 10 0 0
	if entType == "p" { // if player
		if pID == g.player.id {
			if param1 == 0 && param2 == 0 {
				cell.unit = pl1
			} else if param1 != 0 && param2 == 0 {
				cell.unit = pl1_knife
			} else if param1 == 0 && param2 != 0 {
				cell.unit = pl1_bonus
			} else {
				cell.unit = pl1_hero
			}
		}
	}

	if entType == "m" { // if monster
		cell.unit = monster
		cell.risk = 100
		n := twoStep // number of layers the monster area
		keyCell := fmt.Sprintf("%v", cell.xy)
		stack := map[string]*cells{keyCell: cell}
		m := &markSettings{
			set:   "risk",
			unit:  monster,
			steps: oneStep,
			mode:  [2]int{0, byX},

			makeSearch: append([]markSearch{}, markSearch{
				eq:        notEqual,
				eqUnits:   []byte{walls, pl1, pl1_bonus, pl1_knife, monster, monsterArea1, monsterArea2, knifeses},
				markMapTo: 100,
			}),
		}

		for n != 0 {
			m.stackCurrent = nil // need to erase for new loop!!! (only second step without first)
			for _, cell2 := range stack {
				m.markOnMap(cell2, true, true, enterField)
			}
			n--
			stack = m.stackCurrent // erase previous and add new list of steps
			if n == oneStep {
				for i := 0; i < len(m.makeSearch); i++ {
					m.makeSearch[i].markMapTo = 50
				}
			}
		}
		m = &markSettings{
			set:   "monster diagonal",
			unit:  monster,
			steps: oneStep,
			mode:  [2]int{4, byDiagonal},
			makeSearch: append([]markSearch{}, markSearch{
				eq:        notEqual,
				eqUnits:   []byte{walls},
				markMapTo: true,
			}),
		}
		m.markOnMap(cell, false, false, enterField)
	}
}

// sugarSteps
func (g *game) sugarSteps() {
	m := &markSettings{
		set:        "ratio",        // enter sugar-ratio in cell.ratio
		steps:      oneStep,        // depth of steps
		mode:       [2]int{0, byX}, // 4 steps [up-down-left-right]
		stepsCount: 1,              // one step by default (to get current cell)
		makeSearch: append([]markSearch{}, markSearch{
			eq:        equal,                                  // if founded unit on map,check if equal or not equal to change
			eqUnits:   []byte{dots, bonuses, knifeses, golds}, // unit or value that need to find
			markMapTo: byte('+'),                              // mark map where walking
		}, markSearch{
			eq:       equal,
			eqUnits:  []byte{pl1, pl1_bonus, pl1_knife, pl1_hero},
			toDoNext: "return",
		},
		),
	}
	for _, unit := range []byte{golds, knifeses, bonuses} {
		m.unit = unit
		g.loop(m, 3) // loop limit = 3
	}
}

// loop
func (g *game) loop(m *markSettings, limit int) {
	step := g.sugar.steps[m.unit] // show saved minimal steps for unit

	lStack := len(g.sugar.stack[m.unit]) // lenght of sorted list founded sugar

	if lStack == 0 {
		return // exit
	}
	if lStack < limit {
		limit = lStack
	}

	for limit != 0 {
		if stack, ok := g.sugar.stack[m.unit][step]; ok {
			for _, cell := range stack {
				m.parentCell = cell // parent cell
				// fmt.Fprintf(os.Stderr, "sugar %q(x%d:y%d): start\n", s.unit, cell.xy[0], cell.xy[1])
				g.algorithm(m)
				m.stackVisited, m.stepsCount = nil, 1
			}
			limit--
		}
		step++
	}
}

// algorithm
func (g *game) algorithm(m *markSettings) {
	notFounded := 0
	keyCell := fmt.Sprintf("%v", m.parentCell.xy)
	stack := map[string]*cells{keyCell: m.parentCell}
	for {
		m.stackCurrent = nil         // need to erase for new loop!!! (only second step without first)
		for _, cell := range stack { // the positions of coins
			if player := m.markOnMap(cell, true, true, enterField); player { // making 4 or 8 steps and adding on map with saving all founded first step
				if m.set == "ratio" || m.set == "unRatio" {
					if cell.sugarAIO == nil {
						cell.sugarAIO = make(map[byte]map[*cells]int)
					}

					if _, ok := cell.sugarAIO[m.unit]; !ok {
						cell.sugarAIO[m.unit] = make(map[*cells]int)
					}

					cell.sugarAIO[m.unit][m.parentCell] = m.stepsCount
				}

				if m.set == "unRatio" {
					cell.unRatio = -50
				}
				// g.printArena()
				// fmt.Fprintf(os.Stderr, "sugar %q(x%d:y%d): founded player\n", s.unit, s.current.xy[0], s.current.xy[1])
				return
			}
		}

		if len(m.stackCurrent) == 0 { // if not founded new coins
			notFounded++
			if notFounded > 3 {
				return
			}
			m.set = "unRatio" // continue search but with unRatio ignoring monsters
			m.makeSearch[0].eqUnits = []byte{dots, bonuses, knifeses, golds, monster, monsterArea1, monsterArea2}
			for i := 0; i < len(m.makeSearch); i++ {
				m.makeSearch[i].markMapTo = nil
			}
			continue
		}
		stack = m.stackCurrent // erase previous and add new list of steps
		m.stepsCount++         // will count steps to player
		// g.printArena()
	}
}

func (g *game) stepsMain(gl *global) {
	g.chooseStep()

	if len(g.player.cell["high"]) == 1 { // if founded just 1 step
		g.player.answer = g.readWay(g.player.cell["high"][0], "new")
		return
	}

	if len(g.player.cell["high"]) > 1 { // if founded and need sort
		index := g.sortSteps("high")
		g.player.answer = g.readWay(g.player.cell["high"][index], "new")
		return
	}

	if len(g.player.cell["low"]) > 0 { // if founded something
		g.saveCell(gl, "low")
		return
	}

	if len(g.player.cell["risk"]) > 0 { // if founded something
		g.saveCell(gl, "risk")
		return
	}

	fmt.Fprintf(os.Stderr, "player: RANDOM STAY!!!!!!!!!!!\n")
	g.stay()
}

func (g *game) chooseStep() {
	s := &sideSettings{
		current:          g.player.cell["player"][0],
		eq:               equal,
		eqUnits:          []byte{walls},
		neibourNilEqUnit: true,
	}

	for side := 0; side < byX; side++ {
		s.side = side
		ok, neibour := s.neibourCheck()
		if ok {
			continue
		}
		if neibour.risk > 0 { // only risked steps
			g.player.cell["risk"] = append(g.player.cell["risk"], neibour)
			continue
		}
		if neibour.mark != 0 || neibour.ratio > 0 { // walked sugar near or is sugar
			g.player.cell["high"] = append(g.player.cell["high"], neibour)
			continue
		}
		g.player.cell["low"] = append(g.player.cell["low"], neibour)
	}
}

func (g *game) sortSteps(key string) (index int) {
	var maxRatio float64

	for i, cell := range g.player.cell[key] {
		sugar := cell.sugarAIO

		if cell.sugar { // if current is sugar
			return i
		}

		for _, unit := range []byte{golds, knifeses, bonuses} {
			for cell2, step := range sugar[unit] {
				if unit == bonuses || unit == knifeses {
					if cell2.sugarLife < step {
						continue
					}
				}
				ratio := cell2.ratio // ratio of current sugar
				if step > 0 {        // no divide by 0
					ratio /= float64(step) // divide by distance
				}
				cell.ratio += ratio
			}
		}

		if cell.ratio > maxRatio {
			index = i
			maxRatio = cell.ratio
		}
	}
	return index
}

type choose struct {
	risk int
	cell *cells
}

func (g *game) saveCell(gl *global, key string) {
	pX := g.player.cell["player"][0].xy[0]
	pY := g.player.cell["player"][0].xy[1]

	c := &choose{} // choosen answer
	c.risk = maximum

	for _, cell := range g.player.cell[key] {
		if cell.unit == knifeses && len(g.monsters.cell) > 0 && cell.risk < 100 { // if founded knife and monsters on map and risk less
			fmt.Fprintf(os.Stderr, "player in %s: choosen knife (x%d:y%d)\n", key, cell.xy[0], cell.xy[1])
			c.risk = 0
			c.cell = cell
			continue
		}

		if key == "low" {
			if cell.sugar { // if founded bonus or gold
				fmt.Fprintf(os.Stderr, "player in %s: choosen sugar (x%d:y%d)\n", key, cell.xy[0], cell.xy[1])
				c.risk = 1
				c.cell = cell
			}

			if cell.unRatio < 0 { // if founded bonus closed monster but no risk
				if c.risk >= 2 {
					fmt.Fprintf(os.Stderr, "player in %s: choosen sugar with unRatio (x%d:y%d)\n", key, cell.xy[0], cell.xy[1])
					c.risk = 2
					c.cell = cell
				}
			}

			if c.risk >= 3 {
				fmt.Fprintf(os.Stderr, "player in %s: choosen free cell (x%d:y%d)\n", key, cell.xy[0], cell.xy[1])
				c.risk = 3
				c.cell = cell
			}
		}
		if key == "risk" {
			if pX == 6 && !gl.staticMap[pY][pX].monsterArea { // if staying in safe area and x6
				fmt.Fprintf(os.Stderr, "player in %s: choosen stay at x6 recognized\n", key)
				c.risk = 1
				c.cell = g.player.cell["player"][0]
			}

			g.twoStepToSave(c, cell) // 2, 3

			if cell.sugar { // if founded bonus or gold
				if c.risk >= 4 {
					fmt.Fprintf(os.Stderr, "player in %s: choosen sugar (x%d:y%d)\n", key, cell.xy[0], cell.xy[1])
					c.risk = 4
					c.cell = cell
				}
			}

			if cell.hide { // if not founded then choose monster diagonal
				if c.risk >= 5 {
					fmt.Fprintf(os.Stderr, "player in %s: choosen monster diagonal (x%d:y%d)\n", key, cell.xy[0], cell.xy[1])
					c.risk = 5
					c.cell = cell
				}
			}

			if cell.risk < 100 { // if not founded then choose less risk
				if c.risk >= 6 {
					fmt.Fprintf(os.Stderr, "player in %s: choosen low risk - 50 (x%d:y%d)\n", key, cell.xy[0], cell.xy[1])
					c.risk = 6
					c.cell = cell
				}
			}

			if c.risk >= 7 {
				if pX == 6 { // if not founded and staying at x6
					fmt.Fprintf(os.Stderr, "player in %s: choosen stay at x6\n", key)
					c.risk = 7
					c.cell = g.player.cell["player"][0]
				}

				if !gl.staticMap[pY][pX].monsterArea { // if staying in safe area
					fmt.Fprintf(os.Stderr, "player in %s: choosen stay at safe recognized\n", key)
					c.risk = 7
					c.cell = g.player.cell["player"][0]
				}
			}
		}
	}
	g.player.answer = g.readWay(c.cell, "new")
}

func (g *game) twoStepToSave(c *choose, cell *cells) {
	if cell.unit == monsterArea1 || cell.unit == monster {
		return
	}
	side := g.readWay(cell, "buf") // check cell side for player

	s := &sideSettings{
		current: cell,
		eq:      notEqual,
		eqUnits: []byte{walls, monster, monsterArea1, monsterArea2},
		side:    side, // check same side (two steps straight)
	}
	ok, _ := s.neibourCheck()
	if ok {
		c.risk = 2
		c.cell = cell
		g.player.cell["buf"] = nil // clear buffer
		return
	}

	for side2 := 0; side < byX; side2++ {
		s.side = side2 // check all sides
		ok, _ := s.neibourCheck()
		if ok {
			c.risk = 3
			c.cell = cell
			g.player.cell["buf"] = nil // clear buffer
			return
		}
	}
	g.player.cell["buf"] = nil // clear buffer
}

func (g *game) readWay(cell *cells, key string) int {
	// 0 "left"
	// 1 "right"
	// 2 "up"
	// 3 "down"
	x := g.player.cell["player"][0].xy[0]
	y := g.player.cell["player"][0].xy[1]

	g.player.cell[key] = append(g.player.cell[key], cell)
	newX := g.player.cell[key][0].xy[0]
	newY := g.player.cell[key][0].xy[1]

	if newX == x { // width equal
		if newY < y { // unit above
			return 2 // up
		}
		return 3 // down
	}
	if newY == y { // height equal
		if newX < x { // unit to the left
			return 0 // left
		}
		return 1 // right
	}
	return 4 // stay
}

func (g *game) stay() {
	g.player.answer = 4 // stay
	g.player.cell["new"] = append(g.player.cell["new"], g.player.cell["player"]...)
}

type markSettings struct {
	set               string            // sort anonym function fn() by key
	parentCell        *cells            // parent cell
	unit              byte              // current unit
	steps, stepsCount int               // how many steps need in depth, count steps
	mode              [2]int            //[0] = 0; [1] = 4 steps (up-down-left-right) or 8 steps by diagonal too
	stackCurrent      map[string]*cells // buffer for recieved cells
	stackVisited      map[string]*cells // buffer for visited cells
	current           *cells            // current scanned cell
	makeSearch        []markSearch
}

type markSearch struct {
	eq        bool
	eqUnits   []byte
	markMapTo interface{}
	toDoNext  string
}

func (m *markSettings) markOnMap(cell *cells, saveStackCurrent, ignoreVisited bool, fn func(*markSettings, interface{}, *cells)) bool {
	if cell == nil { // if current cell = nil, take from parent
		cell = m.parentCell
	}
	if !saveStackCurrent {
		m.stackCurrent = nil
	}
	if m.stackCurrent == nil {
		m.stackCurrent = make(map[string]*cells)
	}
	if m.stackVisited == nil {
		m.stackVisited = make(map[string]*cells)
	}

	keyCell := fmt.Sprintf("%v", cell.xy)
	m.stackVisited[keyCell] = cell // write current cell if need ignore

	s := &sideSettings{current: cell}
loop:
	for side := m.mode[0]; side < m.mode[1]; side++ { // loop moving (up-down-left-right) with or without diagonals
		s.side = side
		for _, searching := range m.makeSearch {
			s.eq = searching.eq
			s.eqUnits = searching.eqUnits

			ok, neibour := s.neibourCheck()
			if !ok {
				continue
			}

			if searching.toDoNext == "return" {
				m.current = cell
				return true
			}

			keyNeibour := fmt.Sprintf("%v", neibour.xy)

			if ignoreVisited {
				if _, ok := m.stackVisited[keyNeibour]; ok {
					continue loop
				}
			}

			if fn != nil {
				fn(m, searching.markMapTo, neibour)
			}

			m.stackCurrent[keyNeibour] = neibour
			m.current = neibour
		}
	}
	return false // player not founded
}

type sideSettings struct {
	current, neibour *cells
	eq               bool
	eqUnits          []byte
	side             int
	neibourNilEqUnit bool // if cell=nil, what answer return unit=equal or unit=notEqual
}

// neibourCheck checks if linked cell closed by wall
func (s *sideSettings) neibourCheck() (bool, *cells) {
	in := s.inRangeSide()
	if !in {
		if s.neibour == nil {
			if s.neibourNilEqUnit {
				return true, nil // if nil = return true (founded unit)
			}
		}
		return false, nil // if out of range
	}
	if s.eqUnits == nil { // if need just to know in range or not
		return true, s.neibour
	}
	return s.analysCell(s.neibour), s.neibour
}

// inRangeSide returns linked cell by incoming int side
func (s *sideSettings) inRangeSide() bool {
	switch s.side {
	case 0: // left
		s.neibour = s.current.left
	case 1: // right
		s.neibour = s.current.right
	case 2: // up
		s.neibour = s.current.up
	case 3: // down
		s.neibour = s.current.down
	case 4: // up-left
		s.neibour = s.current.uLeft
	case 5: // up-right
		s.neibour = s.current.uRight
	case 6: // down-left
		s.neibour = s.current.dLeft
	case 7: // down-right
		s.neibour = s.current.dRight
	}
	return s.neibour != nil
}

// analysCell show unit in cell or not
func (s *sideSettings) analysCell(cell *cells) bool {
	if !s.eq { // for not equal
		for _, unit := range s.eqUnits {
			if cell.unit == unit {
				return false
			}
		}
		return true
	}

	for _, unit := range s.eqUnits {
		if cell.unit == unit {
			return true
		}
	}
	return false
}

// enterField writes data to fields
func enterField(m *markSettings, value interface{}, cell *cells) {
	switch v := value.(type) {
	case int:
		switch m.set {
		case "risk": // for monsters
			cell.risk = v
			if v == 50 {
				cell.unit = monsterArea2
				return
			}
			cell.unit = monsterArea1
		}
	case byte:
		switch m.set {
		case "ratio": // for sugar
			cell.mark = v
		case "unRatio": // for closed sugar
			return
		}
	case bool:
		switch m.set {
		case "monster diagonal": // for save from monster
			cell.hide = v
		}
	}
}

// unitRatio assings ratio to unit
func unitRatio(unit byte) (float64, bool) {
	if unit == golds {
		return goldsRatio, true
	}
	if unit == knifeses {
		return knifesesRatio, true
	}
	if unit == bonuses {
		return bonusesRatio, true
	}
	return 0, false
}

// limit checks out of range
func limit(x, y int) bool {
	if x < 0 || x >= widthGlobal {
		return false
	}
	if y < 0 || y >= heightGlobal {
		return false
	}
	return true
}

// closer calcs possible steps between player and sugar
func closer(pos1 [2]int, pos2 [2]int) int {
	return un(pos1[0]-pos2[0]) + un(pos1[1]-pos2[1])
}

// func un makes positive int
func un(n int) int {
	if n < 0 {
		return n * -1
	}
	return n
}

func main() {
	gl := new(global)
	for true {
		timing := time.Now()
		g := new(game)

		var w, h, playerID, tick int
		fmt.Scan(&w, &h, &playerID, &tick)
		gl.tick = tick
		g.player.id = 1
		// read map
		for i := 0; i < heightGlobal; i++ {
			line := ""
			fmt.Scan(&line)
			for j := 0; j < widthGlobal; j++ {
				cell := &g.arena[i][j]
				cell.unit = line[j]
				cell.origin = line[j]
				cell.xy = [2]int{j, i}
			}
		}
		g.analysMap(gl)
		g.player.cell = make(map[string][]*cells)
		// number of entities
		var n int
		fmt.Scan(&n)

		// read entities
		for i := 0; i < n; i++ {
			var entType string
			var pID, x, y, param1, param2 int
			fmt.Scan(&entType, &pID, &x, &y, &param1, &param2)
			fmt.Fprintf(os.Stderr, "enttype %s / pID %d / xy %d:%d / param %d:%d\n", entType, pID, x, y, param1, param2)
			cell := &g.arena[y][x]
			if entType == "m" {
				gl.staticMap[y][x].monsterArea = true
				g.monsters.cell = append(g.monsters.cell, cell)
				continue
			}
			if entType == "p" && pID == g.player.id {
				g.player = player{
					cell: map[string][]*cells{
						"player": append(g.player.cell["player"], cell),
					},
					knf: param1,
					bns: param2,
				}
				continue
			}
		}
		g.enterPerson(g.player.cell["player"][0], "p", g.player.id, g.player.knf, g.player.bns)

		if gl.tickKnifeMine <= 0 || g.player.knf <= 0 {
			g.enterMonsters()
		}
		g.scanWhatHave(0)
		if len(g.sugar.stack) == 0 {
			g.scanWhatHave(100)
		}
		g.sugarSteps()
		g.stepsMain(gl)
		g.printArena()
		g.printSugar()
		gl.printSaveArea()
		gl.tickKnifeMine--
		if g.player.cell["new"][0].unit == knifeses {
			gl.tickKnifeMine = 13
		}

		gl.backUpMap(g.arena)
		actions := []string{"left", "right", "up", "down", "stay"}
		fmt.Fprintf(os.Stderr, "%s, knife life - %d, ", actions[g.player.answer], gl.tickKnifeMine)

		// bot action
		fmt.Println(actions[g.player.answer])
		fmt.Fprintf(os.Stderr, "timing: %s\n", time.Since(timing))
	}
}
