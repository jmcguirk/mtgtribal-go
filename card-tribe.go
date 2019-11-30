package main

import (
	"fmt"
	"strconv"
	"strings"
)

type CardTribe struct {
	Name		string;
	Set	    	string;
	Creatures		map[string]*Card;
	References		map[string]*Card;
	TokenGenerators		map[string]*Card;
	Score 		int;
	CommonCreatureCount int;
	CreatureCount int;
	ReferenceCount	int;
}

type GlobalCardTribe struct {
	Name		string;
	CardSets 	map[string]*CardSet;
	SetTribes 	map[string]*CardTribe;
	FirstAppearedIn	*CardSet;
	MostRecentAppearedIn *CardSet;
	TotalReferenceCount	int;
	TotalCreatureCount	int;
}

type SnubbedTribe struct {
	Name		string;
	Cards 		[]*Card;
}

func (this *GlobalCardTribe) TrackCardSet(tribe *CardTribe, cardSet *CardSet) {
	this.Name = tribe.Name;
	if(this.CardSets == nil){
		this.CardSets = make(map[string]*CardSet);
	}
	if(this.SetTribes == nil){
		this.SetTribes = make(map[string]*CardTribe);
	}

	this.SetTribes[cardSet.Code] = tribe;
	this.CardSets[cardSet.Code] = cardSet;
	this.TotalReferenceCount += tribe.ReferenceCount;
	this.TotalCreatureCount += tribe.CreatureCount;
	if(this.FirstAppearedIn == nil){
		this.FirstAppearedIn = cardSet;
	}
	this.MostRecentAppearedIn = cardSet;
}

func (this *CardTribe) Init(name string, set string){
	this.Creatures = make(map[string]*Card);
	this.References = make(map[string]*Card);
	this.TokenGenerators = make(map[string]*Card);
	this.Name = name;
	this.Set = set;
}

func (this *CardTribe) AddCreature(card *Card){
	this.Creatures[card.OrcaleId] = card;
	if(card.Rarity == "common"){
		this.CommonCreatureCount++;
	}
	this.CreatureCount++;
}

func (this *CardTribe) AddTokenGenerator(card *Card){
	this.TokenGenerators[card.OrcaleId] = card;
	// For the purposes of sizing, token generators are tracked as creatures
	if(card.Rarity == "common"){
		this.CommonCreatureCount++;
	}
	this.CreatureCount++;
}

func (this *CardTribe) AddReference(card *Card){
	this.References[card.OrcaleId] = card;
	this.ReferenceCount++;
}

func (this *CardTribe) Contains(card *Card) bool{
	_, exists := this.Creatures[card.OrcaleId];
	if(exists){
		return true;
	}
	_, exists = this.References[card.OrcaleId];
	if(exists){
		return true;
	}

	_, exists = this.TokenGenerators[card.OrcaleId];
	if(exists){
		return true;
	}
	return false;
}


func (this *CardTribe) Describe(){
	fmt.Println(this.Set + " - " + this.Name + " - " + strconv.Itoa(this.CreatureCount) + " Creatures - " + strconv.Itoa(this.ReferenceCount) + " References");
}

func (this *CardTribe) GenerateLink() string{
	return "<a href='https://scryfall.com/search?as=grid&order=name&q=type:" + strings.ToLower(this.Name) +"+set:" + strings.ToLower(this.Set)+"'>";
}