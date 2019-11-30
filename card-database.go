package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

type CardDatabase struct {
	CardCount 		int;
	CardSetsByCode	map[string]*CardSet;
	AllCards		[]Card;
}


func (this *CardDatabase) commonInit() {
	this.AllCards = make([]Card, 0);
	this.CardSetsByCode = make(map[string]*CardSet);
}

func (this *CardDatabase) DescribeTribeWithReferenceCounts(tribe *CardTribe) string {
	buff := tribe.Name;
	buff += " - ";
	set := this.GetSetById(tribe.Set);
	buff += set.Name + " (" + strconv.Itoa(set.ReleaseTimestamp.Year()) + ") - ";
	buff += strconv.Itoa(tribe.CreatureCount) + " Creatures / " + strconv.Itoa(tribe.ReferenceCount) + " References";
	return buff;
}

func (this *CardDatabase) InitializeFromFile(fileName string, setInfo string) error {
	this.commonInit();

	err := this.parseSetsFromFile(setInfo);
	if(err != nil){
		log.Fatal("Failed to parse set info " + err.Error());
	}


	rawJson, err := ioutil.ReadFile(fileName);
	if(err != nil){
		return err;
	}
	fmt.Printf("Loaded raw json - length %d bytes\n", len(rawJson));
	flatParsed := make([]Card,0)
	err = json.Unmarshal(rawJson, &flatParsed);
	fmt.Printf("Completed JSON parse of %d cards\n", len(flatParsed));
	if(err != nil){
		return err;
	}
	this.AllCards = flatParsed;
	for _, card := range this.AllCards{
		set, exists := this.CardSetsByCode[card.Set];
		if(!exists){
			log.Fatal("Card referenced an unknown set - " + card.Set);
		}
		_, exists = set.Cards[card.ScryfallId];
		if(exists){
			log.Fatal("Duplicate card for set " + card.ScryfallId + " set - " + card.Set);
		}
		set.AddCardToSet(card);
	}
	for _, set := range this.CardSetsByCode{
		set.DoBasicDerivations();
	}




	this.CardCount = len(flatParsed);



	return err;
}

func (this *CardDatabase) GetSetById(setFileName string) *CardSet{
	return this.CardSetsByCode[setFileName];
}

func (this *CardDatabase) parseSetsFromFile(setFileName string) error{
	rawJson, err := ioutil.ReadFile(setFileName);
	if(err != nil){
		return err;
	}
	fmt.Printf("Loaded set json - length %d bytes\n", len(rawJson));
	parsedCollection := &CardSetCollection{};
	err = json.Unmarshal(rawJson, &parsedCollection);
	fmt.Printf("Loaded set information, contains %d sets\n", len(parsedCollection.Sets));
	for _, v := range parsedCollection.Sets{
		_, exists := this.CardSetsByCode[v.Code];
		if(exists){
			return errors.New("Encountered duplicate set code - " + v.Code);
		}

		this.initializeSet(v);
	}

	return nil;
}

func (this *CardDatabase) initializeSet(set CardSet) {
	this.CardSetsByCode[set.Code] = &set;
	set.Cards = make(map[string]*Card);
}