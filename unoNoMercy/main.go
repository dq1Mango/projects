package main

import (
	"fmt"
	"os"
	"strconv"
)

var activeColor string

func display(discard card, hands []hand, turn int) {

	colors := make(map[string]string)
	colors["none"] = "\033[0m"
	colors["black"] = "\033[0;40m"
	colors["red"] = "\033[0;31m"
	colors["yellow"] = "\033[0;33m"
	colors["green"] = "\033[0;32m"
	colors["blue"] = "\033[0;34m"

	fmt.Println("current card: ", colors[activeColor], discard, colors["none"])
	fmt.Print("your hand: ")
	for i := range 25 {
		cardToPrint := hands[turn].cards[i]
		if cardToPrint == nilCard {
			continue
		}
		fmt.Print(i+1, ": ", colors[cardToPrint.color], cardToPrint, colors["none"], ", ")
	}
	fmt.Println(hands[turn].count, " cards")
	fmt.Println("turn: ", turn)
}

func getInput() int64 {

	var input string
	for true {

		fmt.Print("Choose a card to play: ")
		fmt.Scan(&input)

		i, err := strconv.ParseInt(input, 10, 64)
		if err == nil && i >= 0 && i <= 25 {
			return i - 1
		}

		fmt.Println("INVALID INPUT (must be a number 0 - 25)")
	}
	return 0
}

func pickColor() string {

	var input string

	for true {

		fmt.Print("What color would you like: ")
		fmt.Scan(&input)

		switch input {
		case // look at this clever case i found on stack overrlow :)
			"red", "yellow", "green", "blue":
			return input
		}

		fmt.Println("Invalid Color!")
	}

	return ""
}

func nextTurn(current int, direction int) int {
	next := Abs(current+direction) % len(alivePlayers)
	if alivePlayers[next] {
		return next
	} else {
		return nextTurn(next, direction)
	}
}

func countAlivePlayers() int {
	sum := 0
	for _, value := range alivePlayers {
		if value {
			sum += 1
		}

	}
	return sum
}

func winMessage(player int) {
	fmt.Println("Contgradulations player ", player, " you win!!!")
	os.Exit(0)
}

func checkWin(hands []hand) {
	if countAlivePlayers() == 1 {
		for i, alive := range alivePlayers {
			if alive {
				winMessage(i)
			}
		}
	}

	for i, hand := range hands {
		if hand.count == 0 && alivePlayers[i] {
			winMessage(i)
		}
	}
}

// kind of an ugly global variable
var alivePlayers = make(map[int]bool)

func main() {
	deck := MakeDeck()
	deck.shuffle()

	players := 2
	// create alive players table
	for i := range players {
		alivePlayers[i] = true
	}
	hands := Deal(deck, players)

	discardPile := &Deck{
		cards: make([]card, 148),
		index: -1,
	}

	turnIndex := 0
	direction := 1

	newCard, err := deck.Pop()
	if err != nil {
		fmt.Println("could not get first card")
	}

	discardPile.Push(newCard)

	for !ValidStart(newCard) {
		newCard, err = deck.Pop()
		if err != nil {
			fmt.Println("could not get first card")
		}

		discardPile.Push(newCard)
	}

	activeColor = newCard.color

	// prolly need to change this when we put it in the web
	for true {
		display(discardPile.Read(), hands, turnIndex)

		var action int64
		var cardToPlay card

		for true {
			action = getInput()

			if action >= 0 && action < 25 {
				cardToPlay = hands[turnIndex].cards[action]
			} else if action == -1 {
				cardToPlay = nilCard
			} else {
				fmt.Println("shoulf have caugt this with prior check")
			}

			if ValidMove(discardPile.Read(), cardToPlay) || action == -1 {
				break
			} else {
				fmt.Println("ILLEGAL MOVE ...")
			}

		}

		if action == -1 {
			drawCard, err := deck.Pop()
			if err != nil {
				RefreshStacks(deck, discardPile)
				drawCard, _ = deck.Pop() // hope theres not an error the second time
			}
			err = hands[turnIndex].Add(drawCard)
			if err != nil {
				KillPlayer(hands, turnIndex, discardPile)
			}

		} else {

			// "play" the card
			hands[turnIndex].Remove(action)
			discardPile.Push(cardToPlay)

			// perform correct action for the card
			if cardToPlay.color == "black" {

				activeColor = pickColor()

				if cardToPlay.number == 1 {
					for true { // i love this control flow idk why
						drawCard, err := deck.Pop()

						if err != nil {
							RefreshStacks(deck, discardPile)
							drawCard, _ = deck.Pop() // hope theres not an error the second time
						}

						fmt.Println("drawing card: ", drawCard) // no color for u lol
						err = hands[nextTurn(turnIndex, direction)].Add(drawCard)
						if err != nil {
							KillPlayer(hands, nextTurn(turnIndex, direction), discardPile)
							break
						}

						if drawCard.color == activeColor {
							break
						}

					}
				} else {

					if cardToPlay.number == 4 {
						direction *= -1
					}

					turnIndex = nextTurn(turnIndex, direction)

					err = Draw(cardToPlay.number, deck, &hands[turnIndex], discardPile)

					if err != nil {
						KillPlayer(hands, turnIndex, discardPile)
					}

				}

			} else { // all the other colors
				activeColor = cardToPlay.color

				// need to change this to implement 0s and 7s rule

				switch cardToPlay.number {

				case 10: // reverse
					direction *= -1

				case 11: // skip
					turnIndex = nextTurn(turnIndex, direction)

				case 12: // skip-all
					turnIndex = nextTurn(turnIndex, direction)

				case 13, 14: // draw 2/4
					turnIndex = nextTurn(turnIndex, direction)

					err = Draw((cardToPlay.number-10)&6, deck, &hands[turnIndex], discardPile)
					// how you like my fancy bit twiddling to advoid repeating code

					if err != nil {
						KillPlayer(hands, turnIndex, discardPile)
					}

				case 15: // put down all of the color

					discardPile.Pop() // unplay the card lol

					for i, card := range hands[turnIndex].cards {
						if card.color == cardToPlay.color {
							discardPile.Push(card)
							hands[turnIndex].Remove(int64(i))
						}
					}

					discardPile.Push(cardToPlay) // un un-play the card
				}
			}
		}

		checkWin(hands)

		turnIndex = nextTurn(turnIndex, direction)
		fmt.Println()

	}
}
