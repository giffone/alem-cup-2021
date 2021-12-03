package main

import "math/rand"

func main() {
	p := new([]pers)
	addPlayers(p, nil, "new")
	cup(maps(11), *p)
}

type pers struct {
	entType                   string
	pID, x, y, param1, param2 int
}

func addPlayers(p *[]pers, g *game, method string) {
	pXY, mXY := [2]int{}, [][2]int{}
	if method == "continue" {
		if g.player.new != nil {
			(*p)[0].x, (*p)[0].y = g.player.new.xy[0], g.player.new.xy[1]
		} else {
			(*p)[0].x, (*p)[0].y = g.player.current.xy[0], g.player.current.xy[1]
		}
		// (*p)[1].x, (*p)[1].y = g.randomEnemy((*p)[1].x, (*p)[1].y)
		// (*p)[2].x, (*p)[2].y = g.randomEnemy((*p)[2].x, (*p)[2].y)
		// (*p)[3].x, (*p)[3].y = g.randomEnemy((*p)[3].x, (*p)[3].y)
		return
	}
	if method == "new" {
		m := []string{
			".............", // 0
			".###..!..###.", // 1
			".!!!..!..!!!.", // 2
			".............", // 3
			"..#..!!!..#..", // 4
			"#!!!p!.!.!!!#", // 5
			"..#..!!!.....", // 6
			".............", // 7
			".!!!..!..!!!.", // 8
			".###..!..###.", // 9
			".............", // 10
		}

		for i := 0; i < heightGlobal; i++ {
			for j := 0; j < widthGlobal; j++ {
				if m[i][j] == 'p' {
					pXY = [2]int{j, i}
				}
				if m[i][j] == 'm' {
					mXY = append(mXY, [2]int{j, i})
				}
			}
		}

		*p = append(*p, pers{
			entType: "p",
			pID:     1,
			x:       pXY[0],
			y:       pXY[1],
		})
		for _, enemy := range mXY {
			*p = append(*p, pers{
				entType: "m",
				pID:     0,
				x:       enemy[0],
				y:       enemy[1],
			})
		}
	}
}

func maps(num int) []string {
	switch num {
	case 1:
		return []string{
			".............", // 0
			".#!.##!###!#.", // 1
			"............i", // 2
			".##!..!..!...", // 3
			".!#........!.", // 4
			".#!..!!!..!..", // 5
			".!#....#..#!.", // 6
			".##!..!..!##.", // 7
			".............", // 8
			".#!###!###!#.", // 9
			".............", // 10
		}
	case 2:
		return []string{
			".............", // 0
			"..!.......!..", // 1
			".!!!.....!!!.", // 2
			"..!.......!..", // 3
			"......!......", // 4
			".....!!!.....", // 5
			"......!......", // 6
			".#!.......!..", // 7
			".!!!.!!!.!!!.", // 8
			"..!.......!..", // 9
			".............", // 10
		}
	case 3:
		return []string{
			".............", // 0
			"..!.......!..", // 1
			".!!!.!!!.!!!.", // 2
			"..!.......!..", // 3
			"......!......", // 4
			".!!!.!!!.!!!.", // 5
			"......!......", // 6
			"..!.......!..", // 7
			".!!!.!!!.!!!.", // 8
			"..!.......!..", // 9
			".............", // 10
		}
	case 4:
		return []string{
			"......!......", // 0
			".....!!!.....", // 1
			".............", // 2
			".............", // 3
			"..!...!...!..", // 4
			".!!..!.!..!!.", // 5
			"..!...!...!..", // 6
			".............", // 7
			".............", // 8
			".....!!!.....", // 9
			"......!......", // 10
		}
	case 5:
		return []string{
			".............", // 0
			"......!......", // 1
			".!!!..!..!!!.", // 2
			".............", // 3
			".....!!!.....", // 4
			".!!!.!.!.!!!.", // 5
			".....!!!.....", // 6
			".............", // 7
			".!!!..!..!!!.", // 8
			"......!......", // 9
			".............", // 10
		}
	case 6:
		return []string{
			".............", // 0
			".!.........!.", // 1
			"...!.!!!.!...", // 2
			"..!.......!..", // 3
			".....!.!.....", // 4
			"......!......", // 5
			".....!.!.....", // 6
			"..!.......!..", // 7
			"...!.!!!.!...", // 8
			".!.........!.", // 9
			".............", // 10
		}
	case 7:
		return []string{
			".............", // 0
			"..#!.#!#.!...", // 1
			"..!#.!#!.#!..", // 2
			".............", // 3
			".#!..!#!..!..", // 4
			".!#..#!#..#!.", // 5
			"..!..!#!..!#.", // 6
			".............", // 7
			"..!#.!#!.#!..", // 8
			"...!.#!#.!#..", // 9
			".............", // 10
		}
	case 8:
		return []string{
			".............", // 0
			"..!..!!!..!..", // 1
			".!.........!.", // 2
			"....!...!....", // 3
			".!....!....!.", // 4
			".!..!...!..!.", // 5
			".!....!....!.", // 6
			".#..!...!....", // 7
			".!#........!.", // 8
			"..!#.!!!..!..", // 9
			".............", // 10
		}
	case 9:
		return []string{
			".............", // 0
			"...!..!..!...", // 1
			"..!...!...!..", // 2
			"......!...#..", // 3
			"..!.......!..", // 4
			".!!..!!!..!!.", // 5
			"..!.......!..", // 6
			"......!......", // 7
			"..!#..!...!..", // 8
			"...!..!..!...", // 9
			".............", // 10
		}
	case 10:
		return []string{
			"..#..........", // 0
			".#!..###..!#.", // 1
			"#!!..!!!..!!.", // 2
			".............", // 3
			"..!...!#..!..", // 4
			"..!..!!!#.!..", // 5
			".#!...!...!..", // 6
			".............", // 7
			".!!..!!!..!!.", // 8
			".#!..###..!#.", // 9
			"..........#..", // 10
		}
	case 11:
		return []string{
			".............", // 0
			".###..!..###.", // 1
			".!!!..!..!!!.", // 2
			".............", // 3
			"..#..!!!..#..", // 4
			"#!!!.!.!.!!!#", // 5
			"..#..!!!.....", // 6
			".............", // 7
			".!!!..!..!!!.", // 8
			".###..!..###.", // 9
			".............", // 10
		}
	case 100:
		// abcdefghijklmnopqrstuvwxyz
		return []string{
			"abcdefghijklm", // 0
			"nopqrstuvwxyz", // 1
			"0123456789012", // 2
			"abcdefghijklm", // 3
			"nopqrstuvwxyz", // 4
			"0123456789012", // 5
			"abcdefghijklm", // 6
			"nopqrstuvwxyz", // 7
			"0123456789012", // 8
			"abcdefghijklm", // 9
			"nopqrstuvwxyz", // 10
		}
	default:
		return []string{
			"#............", // 0
			".............", // 1
			".............", // 2
			".............", // 3
			".............", // 4
			".............", // 5
			".............", // 6
			".............", // 7
			".............", // 8
			".............", // 9
			".............", // 10
		}
	}
}

func (g *game) randomEnemy(x, y int) (int, int) {
	a := rand.Intn(4)
	for i := a; i < byX; i++ { // loop moving (up-down-left-right) without diagonals
		if in, width, height := inRangeMons(x, y, i, oneStep); in {
			if !scan(width, height, &g.arena, '!') { // if cell = wall
				return width, height
			}
		}
	}
	return x, y
}

func inRangeMons(width, height, method, steps int) (in bool, x int, y int) {
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

func scan(x, y int, arena *[11][13]cells, person ...byte) bool {
	for _, unit := range person {
		if arena[y][x].unit[keyOrigin] == unit {
			return true
		}
	}
	return false
}
