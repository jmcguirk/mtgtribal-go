package main

import (
	"strings"
)

type Card struct {
	ScryfallId string  `json:"id"`;
	OrcaleId	string  `json:"oracle_id"`;
	Name		string `json:"name"`;
	TypeLine 	string `json:"type_line"`;
	Set			string `json:"set"`;
	OracleText	string `json:"oracle_text"`;
	Rarity		string `json:"rarity"`;
	Images 		CardImageSet `json:"image_uris"`;
	CardUrl		string `json:"scryfall_uri"`;
	CardFaces	[]CardFace `json:"card_faces"`
	RarityScore	int;
	SuperTypes []string;
	SubTypes []string;
	IsCreature	bool;
	IsArtifact bool;
	IsLegendary bool;
	ModifiedOracleText string;
	IsShapeShifter bool;
	IsTokenGenerator bool;
	SmallImageUrl string;
	LargeImageUrl string;
}

type CardFace struct {
	Images 		CardImageSet `json:"image_uris"`;
}

type CardImageSet struct {
	Small		string `json:"small"`;
	Normal		string `json:"normal"`;
	Large		string `json:"large"`;
}

func (this *Card) DoBasicDerivations(){
	this.SuperTypes = make([]string, 0);
	this.SubTypes = make([]string, 0);

	if(len(this.CardFaces) > 0){
		this.SmallImageUrl = this.CardFaces[0].Images.Small;
		this.LargeImageUrl = this.CardFaces[0].Images.Large;
	} else{
		this.SmallImageUrl = this.Images.Small;
		this.LargeImageUrl = this.Images.Large;
	}

	switch(this.Rarity){
	case "common":
		this.RarityScore = 1;
	case "uncommon":
		this.RarityScore = 2;
	case "rare":
		this.RarityScore = 3;
	case "mythic":
		this.RarityScore = 4;
	}
	this.ModifiedOracleText = this.OracleText;
	this.ModifiedOracleText = strings.ReplaceAll(this.ModifiedOracleText, this.Name, "~")

	lowered := strings.ToLower(this.ModifiedOracleText);
	if(strings.Contains(lowered, "create") && strings.Contains(lowered, "creature token")){
		this.IsTokenGenerator = true;
	}

	checkSubject := this.TypeLine;
	tribalSubject := "";
	if(strings.Contains(this.TypeLine, "//")){
		checkSubject = this.TypeLine[0:strings.Index(this.TypeLine, "//")];
	}
	if(strings.Contains(checkSubject, "—")){
		pivot := strings.Index(checkSubject, "—");
		//fmt.Println(checkSubject);
		slice := len(checkSubject);
		//fmt.Println("Slicing " + strconv.Itoa(pivot) + " " + strconv.Itoa(slice) + " ");
		tribalSubject = checkSubject[pivot:slice];
		checkSubject = checkSubject[0:pivot];
	}
	checkSubject = strings.TrimSpace(checkSubject);
	parts := strings.Split(checkSubject, " ");
	for _, v := range parts {
		this.SuperTypes = append(this.SuperTypes, v);
		if(strings.ToLower(v) == "creature"){
			this.IsCreature = true;
		}
		if(strings.ToLower(v) == "artifact"){
			this.IsArtifact = true;
		}
		if(strings.ToLower(v) == "legendary"){
			this.IsLegendary = true;
		}
	}
	if(tribalSubject != ""){
		//fmt.Println("Tribe " + tribalSubject)
		tribalSubject = strings.TrimSpace(tribalSubject);
		parts = strings.Split(tribalSubject, " ");
		for _, v := range parts {
			//fmt.Println(" tribe " + v);
			if(v == "—"){
				continue; // First split
			}
			if(v != ""){
				this.SubTypes = append(this.SubTypes, v);
				if(v == "ShapeShifter"){
					this.IsShapeShifter = true;
				}
			}
		}
	}
}

func (this *Card) IsTokenGeneratorForTribe(s string) bool {
	if(this.IsTokenGenerator && !this.IsCreature){ // We special case token generators as these are unlikely to be actually relevant references
		return strings.Contains(this.ModifiedOracleText, s);
	}
	return false;
}


func (this *Card) IsTribalReference(s string) bool {
	if(this.IsTokenGenerator){ // We special case token generators as these are unlikely to be actually relevant references
		return false;
	}
	if(strings.Contains(this.ModifiedOracleText, s)){
		nonEquiv := "non-" + s;
		sub := this.ModifiedOracleText;
		// See if this is an explicit non reference and is the only reference
		if(strings.Contains(sub, nonEquiv)){
			sub = strings.ReplaceAll(sub, nonEquiv, "~");
			if(!strings.Contains(sub, s)){
				return false;
			}
		}
		nonEquiv = "Non-" + s;
		if(strings.Contains(sub, nonEquiv)){
			sub = strings.ReplaceAll(sub, nonEquiv, "~");
			if(!strings.Contains(sub, s)){
				return false;
			}
		}
		return true;
	}
	return false;
}

