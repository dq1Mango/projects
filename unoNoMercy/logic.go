package main

import (
	"errors"
	"fmt"
	"math/rand"
	// "time"
)

type card struct {
	color  string
	number int
}

type Deck struct {
	cards []card
	index int
}

var nilCard = card{color: "", number: 0}

func (deck *Deck) Read() card {
	return deck.cards[deck.index]
}

// its almost like its a *stack* of cards
func (deck *Deck) Push(card card) {
	if deck.index >= 147 {
		//fmt.Errorf("cannot push to a full deck of cards")
		return
	}

	deck.index += 1
	deck.cards[deck.index] = card
}

func (deck *Deck) Pop() (card, error) {
	if deck.index < 0 {
		//fmt.Errorf("cannot pop from an empty stack")
		return nilCard, errors.New("cannot pop from an empty stack")
	}

	card := deck.cards[deck.index]
	deck.cards[deck.index] = nilCard
	deck.index -= 1
	return card, nil

}

func MakeDeck() *Deck {
	deck := Deck{
		cards: make([]card, 148),
		index: -1,
	}

	colors := [4]string{"red", "yellow", "green", "blue"}
	numbers := 16

	for _, color := range colors {
		for range 2 {
			for i := range numbers {
				newCard := card{
					color:  color,
					number: i,
				}
				deck.Push(newCard)
			}
		}
	}

	wilds := make(map[int]int)
	wilds[1] = 4
	wilds[4] = 8
	wilds[6] = 4
	wilds[10] = 4

	for name, count := range wilds {
		for range count {
			newCard := card{
				color:  "black",
				number: name,
			}
			deck.Push(newCard)
		}
	}

	return &deck
}

func (deck *Deck) shuffle() {
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range deck.index {
		j := rand.Intn(deck.index + 1)
		deck.cards[i], deck.cards[j] = deck.cards[j], deck.cards[i]
	}
}

type hand struct {
	cards []card
	count int
}

func initializeHand() hand {
	hand := hand{
		cards: make([]card, 25),
		count: 0,
	}

	for i := range 25 {
		hand.cards[i] = nilCard
	}

	return hand
}

func (hand *hand) Add(card card) error {
	if hand.count > 24 {
		fmt.Errorf("should be dead by now")
	}

	for i := range cap(hand.cards) {
		if hand.cards[i] == nilCard {
			hand.cards[i] = card
			break
		}
	}

	hand.count += 1

	if hand.count == 25 {
		return errors.New("death")
	} else {
		return nil
	}

}

func (hand *hand) Remove(index int64) {
	hand.cards[index] = nilCard
	hand.count -= 1
}

// alternate remove using content and not index
//func (hand *hand) Remove(card card) {
//	for i := range 25 {
//		if hand.cards[i] == card {
//			hand.cards[i] = nilCard
//		}
//	}
//}

func Deal(deck *Deck, players int) []hand {

	hands := make([]hand, players)
	for i := range players {
		hands[i] = initializeHand()
	}

	for range 7 {
		for i := range hands {
			newCard, err := deck.Pop()
			if err != nil {
				fmt.Errorf("error could not get new card from deck")
				print("ruh roh")
			}

			hands[i].Add(newCard)
		}
	}

	return hands
}

func ValidStart(card card) bool {
	if card.color == "black" || card.number >= 10 {
		return false
	} else {
		return true
	}
}

func ValidMove(existingCard card, newCard card) bool {

	// shouldnt be needed but gives me piece of mind
	if newCard == nilCard {
		return false
	}

	if newCard.color == "black" {
		return true
	} else {

		if activeColor == newCard.color {
			return true
		}

		if existingCard.number == newCard.number && existingCard.color != "black" {
			return true
		}
	}

	return false
}

// super weird edge case here where there are no cards lefts in either stack (all in player hands)
// probably just gonna limit the players to 6 lol
func RefreshStacks(deck *Deck, discard *Deck) {
	fmt.Println("refreshing stacks...")
	topCard, _ := discard.Pop()

	*deck, *discard = *discard, *deck

	discard.Push(topCard)

	deck.shuffle()
	fmt.Println("new deck; ", deck)
	fmt.Println("new disard; ", discard)
}

func Draw(num int, deck *Deck, hand *hand, discard *Deck) error {
	for range num {
		newCard, err := deck.Pop()
		if err != nil {
			fmt.Println("reacehd the bottom of the barrel")
			RefreshStacks(deck, discard)
			newCard, _ = deck.Pop()
		}

		err = hand.Add(newCard)
		if err != nil {
			return err
		}

	}

	return nil
}

func Abs(num int) int {
	if num <= 0 {
		num *= -1
	}
	return num
}

func KillPlayer(hands []hand, index int, discard *Deck) {
	fmt.Println("killing player ... ", index)

	topCard, err := discard.Pop()
	if err != nil {

	}

	for _, card := range hands[index].cards {
		if card != nilCard {
			discard.Push(card)
		}
	}

	discard.Push(topCard)

	alivePlayers[index] = false

	// and who says go cant do one-liners
	// hands = append(hands[0:index], hands[index+1:]...)

	//newHands := make([]hand, len(hands)-1)

	//for i, hand := range hands {
	//	if i != index {
	//		newHands = append(newHands, hand)
	//	}
	//}

	//hands = newHands

}
