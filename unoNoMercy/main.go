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
	fmt.Println()
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

	fmt.Print("What color would you like: ")
	var input string

	for true {
		fmt.Scan(&input)

		switch input {
		case
			"red", "yellow", "green", "blue":
			return input
		}

		fmt.Println("Invalid Color!")
	}

	return ""
}

func main() {
	deck := MakeDeck()
	deck.shuffle()

	players := 2
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
				KillPlayer(&hands, turnIndex, discardPile)
				players -= 1
			}

		} else {

			// "play" the card
			hands[turnIndex].Remove(action)
			discardPile.Push(cardToPlay)

			// perform correct action for the card
			if cardToPlay.color == "black" {
				if cardToPlay.number == 1 {
					// cry
				} else {

					if cardToPlay.number == 4 {
						direction *= -1
					}

					turnIndex = Abs(turnIndex+direction) % players

					err = Draw(cardToPlay.number, deck, &hands[turnIndex], discardPile)

					if err != nil {
						KillPlayer(&hands, turnIndex, discardPile)
						players -= 1
					}

					activeColor = pickColor()

				}

			} else { // all the other colors
				activeColor = cardToPlay.color

				// need to change this to implement 0s and 7s rule

				switch cardToPlay.number {

				case 10: // reverse
					direction *= -1

				case 11: // skip
					turnIndex = Abs(turnIndex+direction) % players

				case 12: // skip-all
					turnIndex = Abs(turnIndex-direction) % players

				case 13, 14: // draw 2/4
					turnIndex = Abs(turnIndex+direction) % players
					err = Draw((cardToPlay.number-10)&6, deck, &hands[turnIndex], discardPile)
					// how you like my fancy bit twiddling to advoid repeating code

					if err != nil {
						KillPlayer(&hands, turnIndex, discardPile)
						players -= 1
					}

				case 15: // put down all of the color

					discardPile.Pop() // unplay the card lol

					for i, card := range hands[turnIndex].cards {
						if card.color == cardToPlay.color {
							discardPile.Push(card)
							hands[turnIndex].Remove(int64(i))
						}
					}

					discardPile.Push(cardToPlay) // replay the card
				}
			}
		}

		fmt.Println(len(hands), " players left")

		//fmt.Println(hands[turnIndex].cards[action])
		turnIndex = Abs(turnIndex+direction) % players
		fmt.Println()

		if hands[turnIndex].count == 0 || len(hands) == 1 {
			fmt.Println("Contgradulations player ", turnIndex, "you win!!!")
			os.Exit(0)
		}
	}
}
