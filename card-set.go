package main

import (
	"strconv"
	"strings"
	"time"
)

type CardSet struct {
	Id 			string  `json:"id"`;
	Code		string  `json:"code"`;
	Name		string `json:"name"`;
	ReleaseDay 	string `json:"released_at"`;
	ReleaseTimestamp time.Time;
	SetType		string `json:"set_type"`;
	Digital		bool `json:"digital"`;
	Icon		string `json:"icon_svg_uri"`;
	FoilOnly	bool `json:"foil_only"`;
	CardCount	int `json:"card_count"`;
	Cards		map[string]*Card;
	CreaturesByType map[string][]*Card;
	LoadedCardCount int;
	SupportedTribes map[string]*CardTribe;
	PercentTribal	float32;
	PercentTribalNormalized float32;
	TribalScore		int;
}

type CardSetCollection struct {
	Sets []CardSet `json:"data"`;
}

func (this *CardSet) IsRelevantSet() bool{
	return (this.SetType == "core" || this.SetType == "expansion") && !this.FoilOnly && !this.Digital;
}

func (this *CardSet) AddCardToSet(card Card){
	ptr := &card;
	this.Cards[card.ScryfallId] = ptr;
	ptr.DoBasicDerivations();
	this.LoadedCardCount++;
}

func (this *CardSet) trackCreatureByType(card *Card, t string){
	existing, exists := this.CreaturesByType[t];
	if(!exists){
		existing = make([]*Card, 0);
	}
	existing = append(existing, card);
	this.CreaturesByType[t] = existing;
}

func (this *CardSet) AddSupportedTribe(tribe *CardTribe){
	tribe.Set = this.Code;
	this.SupportedTribes[tribe.Name] = tribe;
}


func (this *CardSet) IsCardTribalInSet(card *Card) bool{
	for _, tribe := range this.SupportedTribes {
		if(tribe.Contains(card)){
			return true;
		}
	}
	if(card.IsShapeShifter) { // Shape shifters are inherently tribal
		return true;
	}
	if(strings.Contains(card.OracleText, "creature type")){ // Cards that reference a creature type(s) are inherently tribal
		return true;
	}
	if(strings.Contains(card.OracleText, "creature types")){
		return true;
	}
	return false;
}

func (this *CardSet) DoBasicDerivations(){
	parts := strings.Split(this.ReleaseDay, "-"); // Annoying, but golang can't parse this format natively
	year, _ := strconv.ParseInt(parts[0], 10, 32);
	month, _ := strconv.ParseInt(parts[1], 10, 32);
	day, _ := strconv.ParseInt(parts[2], 10, 32);
	this.ReleaseTimestamp = time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	this.CreaturesByType = make(map[string][]*Card);
	this.SupportedTribes = make(map[string]*CardTribe);
	for _, v := range this.Cards {
		if(v.IsCreature){
			for _, t := range v.SubTypes {
				this.trackCreatureByType(v, t);
			}
		}
	}
}