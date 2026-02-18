package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/chehsunliu/poker"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Position    string `long:"pos" short:"p" description:"Position to use for the server, e.g. sb,d,co,hj,lj,utg2,utg1,utg" default:"sb"`
	IncludeTies bool   `long:"include-ties" short:"t" description:"Include ties in the probability calculation"`
}

var posMap = map[string]int{
	"sb":   1,
	"d":    2,
	"co":   3,
	"hj":   4,
	"lj":   5,
	"utg2": 6,
	"utg1": 7,
	"utg":  8,
}

var rankPoints = map[string]float64{
	"A": 10,
	"K": 8,
	"Q": 7,
	"J": 6,
	"T": 5,
	"9": 4.5,
	"8": 4,
	"7": 3.5,
	"6": 3,
	"5": 2.5,
	"4": 2,
	"3": 1.5,
	"2": 1,
}

func main() {
	var opts Options
	parser := flags.NewNamedParser("faucet", flags.Default)
	parser.AddGroup("Options", "Configuration options", &opts)
	if _, err := parser.Parse(); err != nil {
		return
	}

	fmt.Printf("                                      Range Cart for %s \n\n", opts.Position)
	printGrid(strings.ToLower(opts.Position), opts.IncludeTies)
}

func computeWinProbality(c1, c2 poker.Card, includeTies bool) float64 {
	deck1 := poker.NewDeck()
	combos := make(map[string]float64)

	for !deck1.Empty() {
		deck2 := poker.NewDeck()

		card1 := deck1.Draw(1)[0]

		for !deck2.Empty() {
			card2 := deck2.Draw(1)[0]
			if card1.Rank() == card2.Rank() && card1.Suit() == card2.Suit() {
				continue
			}
			if _, ok := combos[card1.String()+card2.String()]; ok {
				continue
			}
			if _, ok := combos[card2.String()+card1.String()]; ok {
				continue
			}
			card1Points := rankPoints[string(card1.String()[0])]
			card2Points := rankPoints[string(card2.String()[0])]

			score := card1Points
			if card2Points > score {
				score = card2Points
			}

			if card1.Rank() == card2.Rank() {
				pairScore := float64(5)
				if score*2 > pairScore {
					pairScore = score * 2
				}
				combos[card1.String()+card2.String()] = pairScore
				continue
			}

			if card1.Suit() == card2.Suit() {
				score += 2
			}

			penalty := gapPenalty(card1.Rank(), card2.Rank())
			score -= penalty

			if penalty < 2 && card1.Rank() < 10 && card2.Rank() < 10 {
				score += 1
			}

			score = math.Round(score)

			combos[card1.String()+card2.String()] = score
		}
	}
	handScore := combos[c1.String()+c2.String()]
	if handScore == 0 {
		handScore = combos[c2.String()+c1.String()]
	}
	betterScores := float64(0)
	equalScores := float64(0)
	counted := 0
	for combo, score := range combos {
		if combo[0:2] == c1.String() || combo[2:4] == c1.String() || combo[0:2] == c2.String() || combo[2:4] == c2.String() {
			continue
		}
		if score > handScore {
			betterScores++
		}
		if score == handScore {
			equalScores++
		}
		counted++
	}
	if !includeTies {
		return 1 - ((betterScores + equalScores) / 1225)
	}
	return 1 - (betterScores / 1225)
}

func gapPenalty(r1, r2 int32) float64 {
	// absolute difference
	diff := r1 - r2
	if diff < 0 {
		diff = -diff
	}

	// convert distance to gap (adjacent = gap 0)
	gap := diff - 1
	if gap < 0 {
		gap = 0
	}

	switch gap {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 4
	default:
		return 5
	}
}

var ranks = []string{
	"A", "K", "Q", "J", "T", "9", "8",
	"7", "6", "5", "4", "3", "2",
}

func printGrid(position string, includeTies bool) {
	cellWidth := 6

	// Top labels
	fmt.Printf(" ")
	for _, r := range ranks {
		fmt.Printf("%*s", cellWidth+1, r)
	}
	fmt.Println()

	// Horizontal border
	printBorder(cellWidth)

	// Rows
	for _, row := range ranks {

		// Left label
		fmt.Printf("%2s |", row)

		unsuited := true
		// Cells
		for _, col := range ranks {
			var prob float64
			if unsuited {
				prob = computeWinProbality(poker.NewCard(col+"c"), poker.NewCard(row+"s"), includeTies)
			} else {
				prob = computeWinProbality(poker.NewCard(col+"c"), poker.NewCard(row+"c"), includeTies)
			}
			if col == row {
				unsuited = false
			}
			exp := 1
			if position != "" {
				exp = posMap[position]
			}
			fmt.Printf(" %4.1f |", math.Pow(prob, float64(exp))*100)
		}
		fmt.Println()

		printBorder(cellWidth)
	}
}

func printBorder(cellWidth int) {
	fmt.Printf("%3s", "")
	for range ranks {
		fmt.Printf("+%s", repeat("-", cellWidth))
	}
	fmt.Println("+")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
