package main

import (
	"fmt"
)

const (
	keyMiniM     = "miniM"
	keyOrigin    = "origin"
	keyMark      = "mark"
	keyRatio     = "ratio"
	keyUnRatio   = "unRatio"
	keyRisk      = "risk"
	keyUnRisk    = "unRisk"
	heightGlobal = 11
	widthGlobal  = 13
	none         = 0
	oneStep      = 1
	twoStep      = 2
	threeStep    = 3
	equal        = true
	notEqual     = false
	byDiagonal   = 8 // steps by lines and diagonales
	byX          = 4 // steps by lines (up-down,left-right)
	sugarLimit   = 14
	maximum      = 1000      // the maximum steps (impossible large)
	minimum      = 0         // the minimum
	pl1          = byte(80)  // 'P' - player without
	pl1_bonus    = byte(81)  // 'Q' - player with bonus
	pl1_knife    = byte(82)  // 'R' - player with knife
	pl1_hero     = byte(83)  // 'S' - player with knife and bonus
	dots         = byte(46)  // '.' dot nil
	walls        = byte(33)  // '!' - wall
	golds        = byte(35)  // '#' - coin
	bonuses      = byte(98)  // 'b' - bonus
	knifeses     = byte(100) // 'd' - knife
	freezes      = byte(102) // 'f' - freeze
	immunities   = byte(105) // 'f' - freeze
	sugarSteps   = byte(43)  // '+' - freeze
	monsterArea0 = byte(77)  // 'M' - monster
	monsterArea1 = byte(111) // 'o' - area around monster
	monsterArea2 = byte(110) // 'n' - area around monster
	monsterArea3 = byte(115) // 's' - area around monster
	miniM        = byte(109)
	allRatio     = float64(500)
	bonusesRatio = float64(2000)
	lSide        = 0
	rSide        = 1
	uSide        = 2
	dSide        = 3
	ulSide       = 4
	urSide       = 5
	dlSide       = 6
	drSide       = 7
	stayOn       = 8
)

// global regular memory
type global struct {
	tick           int               // current global tick
	gold, goldMine int               // calculate all golds on map and mine
	tickKnifeMine  int               // count mine ticks time left with knife (13 tick--)
	miniM          [][2]int          // monster yesterday
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
			gl.staticMap[i][j].yesterday = arena[i][j].unit[keyOrigin]
		}
	}
}

func (gl *global) miniMSave(cell []*cells) {
	for _, monster := range cell {
		gl.miniM = append(gl.miniM, monster.xy)
	}
}

func (gl *global) miniMAdd(g *game) {
	for _, xy := range gl.miniM {
		g.arena[xy[1]][xy[0]].unit[keyMiniM] = miniM // add monster yesterday
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
	units      []byte
	allList    []*cells // list of all bonuses(knife,bonus)
	ignore     []*cells
	stack      map[int]map[byte]map[int][]*cells // sorted list by steps
	steps      map[int]map[byte]int              // list of minimal steps
	walkClosed map[string]*cells                 // if sugar walked and not founded player, walk on current cell will close
}

type cells struct {
	xy        [2]int             // current coordinates
	unit      map[string]byte    // type of unit // origin unit if changed // mark cell
	ratio     map[string]float64 // ratio by unit
	risk      map[string]int     // cell risk high/low if monster near // if monster and player at same line (x), safe would be (y)
	sugar     bool               // if it is bonus or knife
	sugarLife int                // time before bonus/knife will disappear

	left, right, up, down,
	uLeft, uRight, dLeft, dRight *cells
}

type player struct {
	id, knf, bns int // player info
	current, new *cells
	cell         map[string][]*cells // current coordinations [0] and new coordinations for next step[1]
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
			cell.ratio[keyRatio], cell.sugar = unitRatio(cell.unit[keyOrigin]) // add bonuses ratio
			if cell.sugar {
				g.sugar.allList = append(g.sugar.allList, cell) // add bonuses to list
				if cell.unit[keyOrigin] != golds {
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
		case lSide: // left
			cell.left = neibour
		case rSide: // right
			cell.right = neibour
		case uSide: // up
			cell.up = neibour
		case dSide: // down
			cell.down = neibour
		case ulSide: // up-left
			cell.uLeft = neibour
		case urSide: // up-right
			cell.uRight = neibour
		case dlSide: // down-left
			cell.dLeft = neibour
		case drSide: // down-right
			cell.dRight = neibour
		}
	}
}

// inRangeXY checks coordinations out of range
func (g *game) inRangeXY(cell *cells, method, steps int) (bool, *cells) {
	width, height := cell.xy[0], cell.xy[1]
	switch method {
	case lSide: // left
		width -= steps
	case rSide: // right
		width += steps
	case uSide: // up
		height -= steps
	case dSide: // down
		height += steps
	case ulSide: // up-left
		width -= steps
		height -= steps
	case urSide: // up-right
		width += steps
		height -= steps
	case dlSide: // down-left
		width -= steps
		height += steps
	case drSide: // down-right
		width += steps
		height += steps
	}

	if !limit(width, height) {
		return false, nil
	}
	return true, &g.arena[height][width]
}

// scanWhatHave makes map of bonuses/knifeses/golds
func (g *game) scanWhatHave() {
	g.sugar.stack = make(map[int]map[byte]map[int][]*cells) // sorted list of founded sugar
	g.sugar.steps = make(map[int]map[byte]int)              // list of minimal steps

	for _, cell := range g.sugar.allList {
		if g.ignoreSugar(cell) { // ignore coin if closed by walls
			g.sugar.ignore = append(g.sugar.ignore, cell)
			continue
		}

		risk := cell.risk[keyRisk]

		if _, ok := g.sugar.stack[risk]; !ok { // mininal permitted risk
			g.sugar.stack[risk] = make(map[byte]map[int][]*cells)
			g.sugar.steps[risk] = make(map[byte]int)
		}
		steps := closer(g.player.current.xy, cell.xy)
		unit := cell.unit[keyOrigin]

		if unit != golds { // add or not bonuses/knifeses to stack
			if cell.sugarLife < steps {
				continue
			}
		}

		if _, ok := g.sugar.stack[risk][unit]; !ok {
			g.sugar.stack[risk][unit] = make(map[int][]*cells)
		}
		g.sugar.stack[risk][unit][steps] = append(g.sugar.stack[risk][unit][steps], cell)

		if _, ok := g.sugar.steps[risk][unit]; !ok {
			g.sugar.steps[risk][unit] = maximum
		}
		if g.sugar.steps[risk][unit] > steps {
			g.sugar.steps[risk][unit] = steps
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
				cell.unit[keyOrigin] = pl1
				cell.unit[keyMark] = pl1
			} else if param1 != 0 && param2 == 0 {
				cell.unit[keyOrigin] = pl1_knife
				cell.unit[keyMark] = pl1_knife
			} else if param1 == 0 && param2 != 0 {
				cell.unit[keyOrigin] = pl1_bonus
				cell.unit[keyMark] = pl1_bonus
			} else {
				cell.unit[keyOrigin] = pl1_hero
				cell.unit[keyMark] = pl1_hero
			}
		}
	}

	if entType == "m" { // if monster
		cell.unit[keyOrigin] = monsterArea0
		cell.unit[keyMark] = monsterArea0
		cell.risk[keyRisk] = 100
		n := threeStep // number of layers the monster area
		keyCell := fmt.Sprintf("%v", cell.xy)
		stack := map[string]*cells{keyCell: cell}
		m := &markSettings{
			field:      "risk",
			object:     monsterArea0,
			steps:      oneStep,
			sides:      [2]int{0, byX},
			fnMakeMark: true,

			makeSearch: append([]markSearch{}, markSearch{
				eq: notEqual,
				eqUnits: []byte{walls, pl1, pl1_bonus, pl1_knife,
					monsterArea0, monsterArea1, monsterArea2, monsterArea3, knifeses},
				markTo: 100,
			}),
		}

		for n != 0 {
			m.stackCurrent = nil // need to erase for new loop!!! (only second step without first)
			for _, cell2 := range stack {
				m.markOnMap(cell2, true, true, nil, m.makeMark)
			}
			n--
			stack = m.stackCurrent // erase previous and add new list of steps
			if n == twoStep {
				for i := 0; i < len(m.makeSearch); i++ {
					m.makeSearch[i].markTo = 50
				}
			}
			if n == oneStep {
				for i := 0; i < len(m.makeSearch); i++ {
					m.makeSearch[i].markTo = 10
				}
			}
		}
		g.safeSide(cell)
	}
}

func (g *game) safeSide(cell *cells) {
	mX, mY := cell.xy[0], cell.xy[1]
	pX, pY := g.player.current.xy[0], g.player.current.xy[1]
	distanceX := un(mX - pX)
	distanceY := un(mY - pY)

	// yes x-0 y-2 // x-1 y-2 // x-1 y-1
	// no x-2 y-2 // x-3 y-3
	if (distanceX > threeStep || distanceY > threeStep) ||
		(distanceX > oneStep && distanceY > oneStep) {
		return
	}

	monsterSide1, miXY := g.miniMSide(cell)
	monsterSide2, back, danger := 0, 0, 0

	switch monsterSide1 {
	case lSide: // monster - left
		if miXY[0] > mX && miXY[0] < pX {
			return // monster goes to another side
		}
		monsterSide2 = rSide
	case rSide: // monster - right
		if miXY[0] < mX && miXY[0] > pX {
			return // monster goes to another side
		}
		monsterSide2 = lSide
	case uSide: // monster - up
		if miXY[1] > mY && miXY[1] < pY {
			return // monster goes to another side
		}
		monsterSide2 = dSide
	case dSide: // monster - down
		if miXY[1] < mY && miXY[1] > pY {
			return // monster goes to another side
		}
		monsterSide2 = uSide
	default: // if miniM not on map
		if distanceX < distanceY { // distance by horisontal minimal
			if mX < pX { // monster left side
				monsterSide1 = lSide
				monsterSide2 = rSide
			} else {
				monsterSide1 = rSide
				monsterSide2 = lSide
			}
		} else { // distance by vertical minimal
			if mY < pY { // monster up side
				monsterSide1 = uSide
				monsterSide2 = dSide
			} else {
				monsterSide1 = dSide
				monsterSide2 = uSide
			}
		}
	}

	if wall := g.wallBetween(g.player.current.xy, cell.xy); wall {
		return // if wall between no need to unRatio
	}

	if monsterSide1 == lSide || monsterSide1 == rSide {
		if mY < pY { // if monster up side
			back, danger = dSide, uSide
		} else if mY > pY { // if monster down side
			back, danger = uSide, dSide
		}
	} else if monsterSide1 == uSide || monsterSide1 == dSide {
		if mX < pX { // if monster left side
			back, danger = rSide, lSide
		} else if mX > pX { // if monster right side
			back, danger = lSide, rSide
		}
	} else {
		return
	}

	s := &sideSettings{
		current: g.player.current,
		eq:      notEqual,
		eqUnits: []byte{walls},
	}
	for pay, side := range []int{monsterSide1, monsterSide2, back, danger} {
		if side == maximum {
			continue
		}
		pay := safeSideRatio(pay)
		s.side = side
		ok, neibour := s.neibourCheck()
		if !ok {
			continue
		}
		switch neibour.risk[keyRisk] {
		case 0:
			neibour.risk[keyUnRisk] += pay
		case 10:
			neibour.risk[keyUnRisk] += pay / 2
		case 50:
			neibour.risk[keyUnRisk] += pay / 10
		case 100:
			neibour.risk[keyUnRisk] += 1
		}
	}
}

func (g *game) miniMSide(cell *cells) (side int, xy [2]int) {
	s := &sideSettings{
		current:  cell,
		eq:       equal,
		eqUnits:  miniM,
		keyUnits: keyMiniM, // monster yesterday
	}
	for side := lSide; side <= dSide; side++ {
		s.side = side
		if ok, neibour := s.neibourCheck(); ok {
			return side, neibour.xy
		}
	}
	return maximum, [2]int{}
}

func safeSideRatio(ratio int) int {
	switch ratio {
	case 2:
		return 1000
	case 3:
		return 0
	default:
		return 100
	}
}

func (g *game) wallBetween(obj1, obj2 [2]int) bool {
	rangeX, rangeX2, rangeY, rangeY2 := 0, 0, 0, 0

	if obj1[0] < obj2[0] {
		rangeX, rangeX2 = obj1[0], obj2[0]
	} else {
		rangeX, rangeX2 = obj2[0], obj1[0]
	}
	if obj1[1] < obj2[1] {
		rangeY, rangeY2 = obj1[1], obj2[1]
	} else {
		rangeY, rangeY2 = obj2[1], obj1[1]
	}

	if rangeX2-rangeX < 2 { // if search between Y
		if rangeY2-rangeY < 2 {
			return false // no wall
		}
		rangeY += 1
		rangeY2 -= 1
	} else { // if search between X
		if rangeX2-rangeX < 2 {
			return false // no wall
		}
		rangeX += 1
		rangeX2 -= 1
	}

	for i := rangeY; i <= rangeY2; i++ {
		for j := rangeX; j <= rangeX2; j++ {
			if g.arena[i][j].unit[keyOrigin] == walls {
				return true
			}
		}
	}
	return false
}

// sugarSteps
func (g *game) sugarSteps(risk string) bool {
	riskRange := [2]int{}
	m := &markSettings{
		steps:      oneStep,        // depth of steps
		sides:      [2]int{0, byX}, // 4 steps [up-down-left-right]
		stepsCount: 1,              // one step by default (to get current cell)
		makeSearch: append([]markSearch{}, markSearch{
			eq:       equal,                                       // if founded unit on map,check if equal or not equal to change
			eqUnits:  []byte{pl1, pl1_bonus, pl1_knife, pl1_hero}, // unit or value that need to find
			keyUnits: keyOrigin,
			toDoNext: "return",
		},
		),
	}
	if risk == "low" {
		riskRange = [2]int{0, 10}
		m.field = "ratio" // enter sugar-ratio in cell.ratio
		m.fnMakeMark = true
		m.fnIgnore = true
		m.makeSearch = append(m.makeSearch, markSearch{
			eq:       equal,
			eqUnits:  []byte{dots, bonuses, knifeses, golds, immunities, freezes, monsterArea3, sugarSteps},
			keyUnits: keyMark,
			markTo:   sugarSteps, // mark map where walking
		})
	}
	if risk == "high" {
		riskRange = [2]int{50, 100}
		m.field = "unRatio"
		m.fnMakeMark = false
		m.fnIgnore = false
		m.makeSearch = append(m.makeSearch, markSearch{
			eq:       equal,
			eqUnits:  []byte{dots, bonuses, knifeses, golds, immunities, freezes, monsterArea0}, // ignore monsters
			keyUnits: keyOrigin,
		})
	}

	for _, unit := range g.sugar.units {
		m.object = unit
		g.loop(m, riskRange)
	}
	return m.sucsess > 0 // if founded in limit
}

// loop
func (g *game) loop(m *markSettings, riskRange [2]int) {
	for _, risk := range riskRange {
		step := g.sugar.steps[risk][m.object] // show saved minimal steps for unit

		limit := len(g.sugar.stack[risk][m.object]) // lenght of sorted list founded sugar

		if limit == 0 {
			return // exit
		}

		for limit != 0 {
			if stack, ok := g.sugar.stack[risk][m.object][step]; ok {
				for _, cell := range stack {
					m.parentCell = cell // parent cell
					g.algorithm(m)
					m.stackVisited, m.stepsCount = nil, 1
				}
				limit--
			}
			step++
		}
	}
}

// algorithm
func (g *game) algorithm(m *markSettings) {
	keyCell := fmt.Sprintf("%v", m.parentCell.xy)
	stack := map[string]*cells{keyCell: m.parentCell}
	for {
		m.stackCurrent = nil         // need to erase for new loop!!! (only second step without first)
		for _, cell := range stack { // the positions of coins
			if player := m.markOnMap(cell, true, true, g.ignoreClosed, m.makeMark); player { // making 4 or 8 steps and adding on map with saving all founded first step
				s := &sugarWalk{
					cell:     m.parentCell,
					step:     m.stepsCount,
					ratioKey: keyRatio,
				}
				if m.field == "unRatio" {
					s.ratioKey = keyUnRatio
				}
				s.calcRatio(cell)
				// g.printArena()
				return
			}
		}

		if len(m.stackCurrent) == 0 { // if not founded new coins
			if g.sugar.walkClosed == nil {
				g.sugar.walkClosed = make(map[string]*cells)
			}
			for key, cell2 := range m.stackVisited {
				g.sugar.walkClosed[key] = cell2
			}
			return
		}
		stack = m.stackCurrent // erase previous and add new list of steps
		m.stepsCount++         // will count steps to player
		// g.printArena()
	}
}

func (g *game) stepsMain(gl *global, ok bool) {
	if !ok {
		g.stay()
		return
	}
	g.chooseStep()

	for _, risk := range []string{"low", "middle", "high", "risk"} {
		if len(g.player.cell[risk]) > 0 { // if founded and need sort
			if g.safeCell(gl, risk) {
				return
			}
		}
	}

	g.stay()
}

func (g *game) chooseStep() {
	g.player.cell = make(map[string][]*cells)
	s := &sideSettings{
		current:          g.player.current,
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
		if neibour.risk[keyRisk] > 10 { // only risked steps
			g.player.cell["risk"] = append(g.player.cell["risk"], neibour)
			continue
		}
		if neibour.unit[keyMark] != 0 || neibour.ratio[keyRatio] > 0 { // walked sugar near or is sugar
			if neibour.risk[keyRisk] == 0 {
				g.player.cell["low"] = append(g.player.cell["low"], neibour)
				continue
			}
			g.player.cell["middle"] = append(g.player.cell["middle"], neibour)
			continue
		}
		g.player.cell["high"] = append(g.player.cell["high"], neibour)
	}
}

type choose struct {
	pay, payMore, price float64
	cell                *cells
}

func (g *game) safeCell(gl *global, key string) (choosen bool) {
	pX := g.player.current.xy[0]
	pY := g.player.current.xy[1]

	c := &choose{} // choosen answer

	for _, cell := range g.player.cell[key] {
		if g.player.current.risk[keyRisk] > 10 {
			c.pay = float64(maximum - cell.risk[keyRisk])
		} else {
			c.pay = cell.ratio[keyRatio]
		}

		if cell.risk[keyUnRisk] > 0 { // high unrisk mean monster closer and it safe cell to run away
			c.payMore = float64(cell.risk[keyUnRisk])
			if c.payMore > c.price { // choose more safe
				c.price = c.payMore
				c.cell = cell
				choosen = true
			}
		}

		if cell.unit[keyOrigin] == knifeses && len(g.monsters.cell) > 0 { // if player in danger and need knife
			if g.player.current.risk[keyRisk] > 10 { // if player in risk
				c.payMore = c.pay
				c.price = c.pay // if player in risk and no need to review another cells
				c.cell = cell
				choosen = true
			}
		}

		if c.payMore != 0 { // if have safe cell, do not review another cells
			continue
		}

		if key == "risk" || g.player.current.risk[keyRisk] > 10 {
			if pX == 6 && !gl.staticMap[pY][pX].monsterArea { // if staying in safe area and x6
				if lWall, _ := g.wallsScan(g.player.current); lWall > 1 {
					c.pay = 100
					c.price = c.pay
					c.cell = g.player.current
					choosen = true
					continue
				}
			}

			if (g.cornerScan(g.player.current) == 2 || g.cornerScan(g.player.current) == 3) &&
				!gl.staticMap[pY][pX].monsterArea { // if staying in safe area and in dead end
				c.pay = 50
				c.price = c.pay
				c.cell = g.player.current
				choosen = true
				continue
			}

			if cell.sugar && cell.unit[keyOrigin] != knifeses { // if founded bonus or gold
				if !g.safeSugar(cell) { // if sugar in risk
					continue
				}
			}
		}

		if cell.ratio[keyUnRatio] > 0 && cell.ratio[keyRatio] == 0 { // if founded bonus closed monster but no risk
			g.twoStepToSafe(c, cell, [2]float64{1, 2})
		}
		if cell.risk[keyRisk] > 0 {
			c.pay /= 2
		}
		if c.pay < c.price {
			continue
		}
		c.price = c.pay
		c.cell = cell
		choosen = true

		g.player.cell["buf"] = nil // clear buffer
	}
	if choosen {
		g.player.answer = g.readWay(c.cell, "new")
	}
	return choosen
}

func (g *game) safeSugar(sugar *cells) bool {
	if g.player.current.risk[keyRisk] > 10 { // if near monster from player
		return g.cornerScan(sugar) != 3 // and sugar in dead end - danger, skip it
	}
	return true // not danger
}

func (g *game) twoStepToSafe(c *choose, cell *cells, pay [2]float64) (choosen bool) {
	if cell.risk[keyRisk] > 10 {
		return
	}
	side := g.readWay(cell, "buf") // check cell side for player

	s := &sideSettings{
		current: cell,
		eq:      notEqual,
		eqUnits: []byte{walls, monsterArea0, monsterArea1, monsterArea2, pl1, pl1_bonus, pl1_knife, pl1_hero},
		side:    side, // check same side (two steps straight)
	}
	ok, _ := s.neibourCheck()
	if ok {
		if c.price == pay[0] && cell.unit[keyMark] != '+' { // if have already choosen same risk,then can choose better with mark bonus
			return
		}
		c.price = pay[0]
		c.cell = cell
		choosen = true
		return
	}

	for side2 := 0; side2 < byX; side2++ {
		s.side = side2 // check all sides
		ok, _ := s.neibourCheck()
		if ok {
			if c.price >= pay[1] {
				if c.price == pay[1] && cell.unit[keyMark] != '+' { // if have already choosen same risk,then can choose better with mark bonus
					return
				}
				c.price = pay[1]
				c.cell = cell
				choosen = true
				return
			}
		}
	}
	return choosen
}

func (g *game) readWay(cell *cells, key string) int {
	// 0 "left"
	// 1 "right"
	// 2 "up"
	// 3 "down"
	x := g.player.current.xy[0]
	y := g.player.current.xy[1]
	newX, newY := 0, 0

	if key == "buf" {
		g.player.cell[key] = append(g.player.cell[key], cell)
		newX = g.player.cell[key][0].xy[0]
		newY = g.player.cell[key][0].xy[1]
	} else {
		g.player.new = cell
		newX = g.player.new.xy[0]
		newY = g.player.new.xy[1]
	}

	if newX == x { // width equal
		if newY < y { // unit above
			return uSide // up
		}
		return dSide // down
	}
	if newY == y { // height equal
		if newX < x { // unit to the left
			return lSide // left
		}
		return rSide // right
	}
	return stayOn // stay
}

func (g *game) wallsScan(cell *cells) (lWall int, wall []int) {
	s := &sideSettings{
		current: cell,
		eq:      equal,
		eqUnits: []byte{walls},
	}

	for side := 0; side < byX; side++ {
		s.side = side
		if ok, _ := s.neibourCheck(); ok {
			wall = append(wall, side)
		}
	}

	return len(wall), wall
}

func (g *game) cornerScan(cell *cells) int {
	// [left 0, right 1, up 2, down 3]
	lWall, wall := g.wallsScan(cell)

	if lWall == 3 {
		return 3 // danger - dead end
	}
	if lWall == 2 {
		if (wall[0] == lSide && wall[1] == rSide) || // if walls create corner
			(wall[0] == uSide && wall[1] == dSide) {
			return 2 // danger - dead end (low)
		}
	}
	return 0 // not danger
}

func (g *game) ignoreClosed(key string) bool {
	if _, ok := g.sugar.walkClosed[key]; ok {
		return true
	}
	return false
}

func (g *game) stay() {
	g.player.answer = stayOn // stay
	g.player.new = g.player.current
}

func makeMapsCell(cell *cells) {
	cell.unit = make(map[string]byte)
	cell.ratio = make(map[string]float64)
	cell.risk = make(map[string]int)
}

type sugarWalk struct {
	cell     *cells
	step     int
	ratioKey string
}

func (s *sugarWalk) calcRatio(cell *cells) {
	if cell.unit[keyOrigin] == knifeses || cell.unit[keyOrigin] == bonuses { // check if bonus/knife not actual
		if cell.sugarLife < int(s.step) {
			return
		}
	}
	if s.step == 0 {
		cell.ratio[s.ratioKey] += s.cell.ratio[keyRatio]
	} else {
		cell.ratio[s.ratioKey] += s.cell.ratio[keyRatio] / float64(s.step)
	}
}

type markSettings struct {
	sucsess              int               // if toDoNext="return" and was sucsessfully
	field                string            // sort anonym function fn() by key
	parentCell           *cells            // parent cell
	object               byte              // current unit
	steps, stepsCount    int               // how many steps need in depth, count steps
	sides                [2]int            //[0] = 0; [1] = 4 steps (up-down-left-right) or 8 steps by diagonal too
	stackCurrent         map[string]*cells // buffer for recieved cells
	stackVisited         map[string]*cells // buffer for visited cells
	currentCell          *cells            // current scanned cell
	fnIgnore, fnMakeMark bool
	makeSearch           []markSearch
}

type markSearch struct {
	eq       bool
	eqUnits  []byte
	keyUnits string
	markTo   interface{}
	toDoNext string
}

func (m *markSettings) markOnMap(cell *cells, saveStackCurrent, ignoreVisited bool, fnIgnore func(string) bool, fnMakeMark func(interface{}, *cells)) bool {
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

	if m.fnIgnore { // ignore closed cell
		if fnIgnore(keyCell) {
			return false
		}
	}
	m.stackVisited[keyCell] = cell // write current cell if need ignore

	s := &sideSettings{current: cell}
loop:
	for side := m.sides[0]; side < m.sides[1]; side++ { // loop moving (up-down-left-right) with or without diagonals
		s.side = side
		for _, searching := range m.makeSearch {
			s.eq = searching.eq
			s.eqUnits = searching.eqUnits
			s.keyUnits = searching.keyUnits
			ok, neibour := s.neibourCheck()
			if !ok {
				continue
			}

			if searching.toDoNext == "return" {
				m.sucsess++
				m.currentCell = cell
				return true
			}

			keyNeibour := fmt.Sprintf("%v", neibour.xy)

			if ignoreVisited {
				if _, ok := m.stackVisited[keyNeibour]; ok {
					continue loop
				}
			}

			if m.fnIgnore { // ignore closed cell
				if fnIgnore(keyNeibour) {
					continue loop
				}
			}

			if m.fnMakeMark {
				fnMakeMark(searching.markTo, neibour)
			}

			m.stackCurrent[keyNeibour] = neibour
			m.currentCell = neibour
		}
	}
	return false // player not founded
}

// makeMark writes data to fields
func (m *markSettings) makeMark(value interface{}, cell *cells) {
	switch v := value.(type) {
	case int:
		switch m.field {
		case "risk": // for monsters
			cell.risk[keyRisk] = v
			if v == 10 {
				cell.unit[keyMark] = monsterArea3
				return
			}
			if v == 50 {
				cell.unit[keyMark] = monsterArea2
				return
			}
			cell.unit[keyMark] = monsterArea1
		}
	case byte:
		switch m.field {
		case "ratio": // for sugar
			cell.unit[keyMark] = v
		case "unRatio": // for closed sugar
			return
		}
	}
}

type sideSettings struct {
	current, neibour *cells
	field            string
	eq               bool
	eqUnits          interface{}
	keyUnits         string
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
	case lSide: // left
		s.neibour = s.current.left
	case rSide: // right
		s.neibour = s.current.right
	case uSide: // up
		s.neibour = s.current.up
	case dSide: // down
		s.neibour = s.current.down
	case ulSide: // up-left
		s.neibour = s.current.uLeft
	case urSide: // up-right
		s.neibour = s.current.uRight
	case dlSide: // down-left
		s.neibour = s.current.dLeft
	case drSide: // down-right
		s.neibour = s.current.dRight
	}
	return s.neibour != nil
}

// analysCell show unit in cell or not
func (s *sideSettings) analysCell(cell *cells) bool {
	switch eqUnits := s.eqUnits.(type) {
	case []byte:
		if !s.eq { // for not equal
			for _, unit := range eqUnits {
				if s.keyUnits != "" {
					if cell.unit[s.keyUnits] == unit {
						return false
					}
				} else {
					for _, cellUnit := range cell.unit {
						if cellUnit == unit {
							return false
						}
					}
				}
			}
			return true
		}

		for _, unit := range eqUnits {
			if s.keyUnits != "" {
				if cell.unit[s.keyUnits] == unit {
					return true
				}
			} else {
				for _, cellUnit := range cell.unit {
					if cellUnit == unit {
						return true
					}
				}
			}
		}
		return false
	}
	return false
}

// unitRatio assings ratio to unit
func unitRatio(unit byte) (float64, bool) {
	switch unit {
	case bonuses:
		return bonusesRatio, true
	case golds:
		return allRatio, true
	case knifeses:
		return allRatio, true
	case immunities:
		return allRatio, true
	case freezes:
		return allRatio, true
	default:
		return 0, false
	}
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
		g := new(game)
		g.sugar.units = append(g.sugar.units, golds, knifeses, bonuses, immunities, freezes)
		var w, h, playerID, tick int
		fmt.Scan(&w, &h, &playerID, &tick)
		gl.tick = tick
		g.player.id = playerID
		// read map
		for i := 0; i < heightGlobal; i++ {
			line := ""
			fmt.Scan(&line)
			for j := 0; j < widthGlobal; j++ {
				cell := &g.arena[i][j]
				makeMapsCell(cell)
				cell.unit[keyOrigin] = line[j]
				cell.unit[keyMark] = line[j]
				cell.xy = [2]int{j, i}
			}
		}
		g.analysMap(gl)
		// number of entities
		var n int
		fmt.Scan(&n)

		// read entities
		for i := 0; i < n; i++ {
			var entType string
			var pID, x, y, param1, param2 int
			fmt.Scan(&entType, &pID, &x, &y, &param1, &param2)
			cell := &g.arena[y][x]
			if entType == "m" {
				gl.staticMap[y][x].monsterArea = true
				g.monsters.cell = append(g.monsters.cell, cell)
				continue
			}
			if entType == "p" && pID == g.player.id {
				g.player = player{
					current: cell,
					knf:     param1,
					bns:     param2,
				}
				continue
			}
		}
		g.enterPerson(g.player.current, "p", g.player.id, g.player.knf, g.player.bns)

		if gl.tickKnifeMine <= 0 || g.player.knf <= 0 {
			gl.miniMAdd(g) // save monsters
			g.enterMonsters()
		}
		gl.miniMSave(g.monsters.cell) // save monsters
		g.scanWhatHave()

		ok := false
		if ok = g.sugarSteps("low"); !ok {
			ok = g.sugarSteps("high")
		}

		g.stepsMain(gl, ok)
		gl.tickKnifeMine--
		if g.player.new.unit[keyOrigin] == knifeses {
			gl.tickKnifeMine = 13
		}

		gl.backUpMap(g.arena)
		actions := []string{"left", "right", "up", "down", "ul", "ur", "dl", "dr", "stay"}

		// bot action
		fmt.Println(actions[g.player.answer])
	}
}
