package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MinTribeSize = 10;
const MinCommonRequired = 3;
const MinReferenceRequired = 5;
const MinSetSize = 90;

const ReportTopSize = 20;


func main() {
	fmt.Println("Starting tribal analysis read out");



	db := &CardDatabase{};
	start := time.Now();
	err := db.InitializeFromFile("scryfall-default-cards.json", "sets-info.json");
	dur := time.Now().Sub(start).Milliseconds();
	if(err != nil){
		log.Fatal("Failed to initialize card database - " + err.Error());
	}

	reportBytes, err := ioutil.ReadFile("index.template.html");
	if(err != nil){
		log.Fatal("Failed to read source html - " + err.Error());
	}
	reportHTML := string(reportBytes);

	fmt.Printf("Card database intialized successfully in %dms - contains %d cards\n", dur, db.CardCount);

	var peakTribeScore float32 = -1;
	//var peakTribeSet *CardSet;
	for _, set := range db.CardSetsByCode {
		if(!set.IsRelevantSet()){
			//log.Println("skipping " + set.Name + " because it was a digital or foil only set");
			continue;
		}
		if(len(set.Cards) < MinSetSize){
			continue;
		}

		for tribeName, cards := range set.CreaturesByType {
			candidateTribe := &CardTribe{};
			candidateTribe.Init(tribeName, set.Name);

			for _, c := range cards{
				candidateTribe.AddCreature(c);
			}
			for _, c := range set.Cards {
				if(c.IsTribalReference(tribeName)){
					candidateTribe.AddReference(c);
				}
				if(c.IsTokenGeneratorForTribe(tribeName)){
					candidateTribe.AddTokenGenerator(c);
				}
			}

			if(candidateTribe.ReferenceCount >= MinReferenceRequired &&
				candidateTribe.CreatureCount >= MinTribeSize &&
				candidateTribe.CommonCreatureCount >= MinCommonRequired) {
				set.AddSupportedTribe(candidateTribe);
			}

			//if(commonCount > MinCommonRequired && > referenceCount)
		}

		tribalCards := 0;
		for _, card := range set.Cards {
			if(set.IsCardTribalInSet(card)){
				tribalCards++;
			}
		}
		set.PercentTribal = float32(tribalCards) / float32(len(set.Cards));
		if(set.PercentTribal > peakTribeScore){
			peakTribeScore = set.PercentTribal;
			//peakTribeSet = set;
		}
	}

	allSets := make([]*CardSet, 0);
	for _, set := range db.CardSetsByCode {
		set.PercentTribalNormalized = set.PercentTribal / peakTribeScore;
		set.TribalScore = int(math.Round(float64(set.PercentTribalNormalized * 100)));
		allSets = append(allSets, set);

	}
	// Sort descending by normalized score
	sort.SliceStable(allSets, func(i, j int) bool {
		return allSets[i].PercentTribalNormalized > allSets[j].PercentTribalNormalized;
	})

	chart := &TopSetsReportChart{};
	chart.Data = make([]TopSetsReportData, 0);
	series := make([]TopSetsReportData, 0);

	chart.Labels = make([]ReportLabel, 0);

	relevantSets := make([]*CardSet, 0)

	for _, set := range allSets {
		if (!set.IsRelevantSet()) {
			//log.Println("skipping " + set.Name + " because it was a digital or foil only set");
			continue;
		}
		if (len(set.Cards) < MinSetSize) {
			continue;
		}

		relevantSets = append(relevantSets, set);
	}

	for _, set := range relevantSets {
		//log.Printf("%s - Tribal Score - %.2f%%\n", set.Name, set.PercentTribalNormalized * 100);

		data := &TopSetsReportData{};
		data.Label = set.Name;
		data.Value = set.TribalScore;
		series = append(series, *data);
		lbl := &ReportLabel{};
		lbl.Name = set.Name;
		lbl.ImageUrl = set.Icon;
		chart.Labels = append(chart.Labels, *lbl);
		if(len(series) >= ReportTopSize){
			break;
		}
	}
	chart.Data = series;//append(chart.Data, series);
	chartJson, err := json.MarshalIndent(chart, "", "    ");
	if(err != nil){
		log.Fatal("Encountered error while serializing out chart data " + err.Error());
	}
	//pretty :=
	reportHTML = strings.Replace(reportHTML,"%TOP_SETS_CHART_JSON%", string(chartJson), -1);



	sort.SliceStable(relevantSets, func(i, j int) bool {
		return relevantSets[i].ReleaseTimestamp.Unix() < relevantSets[j].ReleaseTimestamp.Unix();
	})

	trendGraph := &TribalTrendGraph{};
	trendGraph.Data = make([]TribalTrendData, 0);
	for _, set := range relevantSets {
		dataPoint := &TribalTrendData{};
		dataPoint.ReleaseTimeStamp = int(set.ReleaseTimestamp.Unix());
		dataPoint.Score = set.TribalScore;
		dataPoint.SetName = set.Name;
		trendGraph.Data = append(trendGraph.Data, *dataPoint);
	}

	chartJson, err = json.MarshalIndent(trendGraph, "", "    ");
	if(err != nil){
		log.Fatal("Encountered error while serializing out chart data " + err.Error());
	}
	reportHTML = strings.Replace(reportHTML,"%HISTORIC_TREND_CHART_JSON%", string(chartJson), -1);



	globalTribes := make(map[string]*GlobalCardTribe);
	trendGraph = &TribalTrendGraph{};
	trendGraph.Data = make([]TribalTrendData, 0);
	uniqueCount := 0;
	for _, set := range relevantSets {
		for key, tribe := range set.SupportedTribes {
			_, exists := globalTribes[key];
			if (!exists) {
				globalTribes[key] = &GlobalCardTribe{};
				uniqueCount++;
			}
			global, _ := globalTribes[key];
			global.TrackCardSet(tribe, set);
		}
		dataPoint := &TribalTrendData{};
		dataPoint.ReleaseTimeStamp = int(set.ReleaseTimestamp.Unix());
		dataPoint.Score = uniqueCount;
		dataPoint.SetName = set.Name;
		trendGraph.Data = append(trendGraph.Data, *dataPoint);
	}

	chartJson, err = json.MarshalIndent(trendGraph, "", "    ");
	if(err != nil){
		log.Fatal("Encountered error while serializing out chart data " + err.Error());
	}
	reportHTML = strings.Replace(reportHTML,"%HISTORIC_CUMULATIVE_CHART_JSON%", string(chartJson), -1);


	allTribes := make([]*CardTribe, 0);



	for _, set := range relevantSets {
		for _, tribe := range set.SupportedTribes {
			allTribes = append(allTribes, tribe);
		}
	}

	sort.SliceStable(allTribes, func(i, j int) bool {
		return allTribes[i].ReferenceCount +  allTribes[i].CreatureCount > allTribes[j].ReferenceCount + allTribes[j].CreatureCount;
	})

	buff := "";
	buff += "<div class='winner-header'>1st Place: " + allTribes[0].GenerateLink() + db.DescribeTribeWithReferenceCounts(allTribes[0]) + "</a></div>"
	buff += GenerateSampleCardHTML(allTribes[0], false);
	buff += "<div class='second-place-header'>2nd: " + allTribes[1].GenerateLink() + db.DescribeTribeWithReferenceCounts(allTribes[1]) + "</a></div>"
	buff += GenerateSampleCardHTML(allTribes[1], true);
	buff += "<div class='third-place-header'>3rd: " + allTribes[2].GenerateLink() + db.DescribeTribeWithReferenceCounts(allTribes[2]) + "</a></div>"
	buff += GenerateSampleCardHTML(allTribes[2], true);

	reportHTML = strings.Replace(reportHTML,"%MOST_SUPPORTED_TRIBE_HTML%", string(buff), -1);

	// Generate frequent tribes
	buff = "";

	allGlobalTribes := make([]*GlobalCardTribe, 0);
	for _, tribe := range globalTribes{
		allGlobalTribes = append(allGlobalTribes, tribe);
	}

	sort.SliceStable(allGlobalTribes, func(i, j int) bool {
		return len(allGlobalTribes[i].CardSets) > len(allGlobalTribes[j].CardSets);
	})

	buff += "<div class='winner-header'>1st Place: " + allGlobalTribes[0].Name + " - Appears in " + strconv.Itoa(len(allGlobalTribes[0].CardSets)) + " sets</div>"
	buff += GenerateSampleCardHTMLForGlobalTribe(allGlobalTribes[0], false);
	buff += "<div class='second-place-header'>2nd: " + allGlobalTribes[1].Name + "- Appears in " + strconv.Itoa(len(allGlobalTribes[1].CardSets)) + " sets</div>"
	buff += GenerateSampleCardHTMLForGlobalTribe(allGlobalTribes[1], true);
	buff += "<div class='third-place-header'>3rd: " + allGlobalTribes[2].Name + "- Appears in " + strconv.Itoa(len(allGlobalTribes[2].CardSets)) + " sets</div>"
	buff += GenerateSampleCardHTMLForGlobalTribe(allGlobalTribes[2], true);

	reportHTML = strings.Replace(reportHTML,"%MOST_FREQUENT_TRIBE_HTML%", string(buff), -1);

	// Generate the snubbed section html
	buff = "";

	snubbedTribes := make(map[string]*SnubbedTribe);
	allSnubbed := make([]*SnubbedTribe, 0);
	for _, set := range relevantSets {
		for _, card := range set.Cards {
			if(card.IsCreature){
				for _, subType := range card.SubTypes {
					_, exists := globalTribes[subType];
					if(!exists){
						_, snubbedExists := snubbedTribes[subType];
						if(!snubbedExists){
							sbt := &SnubbedTribe{};
							sbt.Cards = make([]*Card, 0);
							sbt.Name = subType;
							snubbedTribes[subType] = sbt;
							allSnubbed = append(allSnubbed, sbt);
						}
						snubbed, _ := snubbedTribes[subType];
						snubbed.Cards = append(snubbed.Cards, card);
					}
				}
			}
		}
	}



	sort.SliceStable(allSnubbed, func(i, j int) bool {
		return len(allSnubbed[i].Cards) > len(allSnubbed[j].Cards);
	})


	buff += "<div class='winner-header'>1st Place: " + allSnubbed[0].Name + " - " + strconv.Itoa(len(allSnubbed[0].Cards)) + " Cards</div>"
	buff += GenerateSampleCardHTMLFromRawCardPool(allSnubbed[0].Cards, false);
	buff += "<div class='second-place-header'>2nd: " + allSnubbed[1].Name + " - " + strconv.Itoa(len(allSnubbed[1].Cards)) + " Cards</div>"
	buff += GenerateSampleCardHTMLFromRawCardPool(allSnubbed[1].Cards, true);
	buff += "<div class='third-place-header'>3rd: " + allSnubbed[2].Name + " - " + strconv.Itoa(len(allSnubbed[2].Cards)) + " Cards</div>"
	buff += GenerateSampleCardHTMLFromRawCardPool(allSnubbed[2].Cards, true);

	reportHTML = strings.Replace(reportHTML,"%MOST_SNUBBED_TRIBE_HTML%", string(buff), -1);


	// Generate all sets report HTML
	buff = "";
	sort.SliceStable(relevantSets, func(i, j int) bool {
		return relevantSets[i].ReleaseTimestamp.Unix() > relevantSets[j].ReleaseTimestamp.Unix();
	})

	for _, set := range relevantSets {
		if(len(set.SupportedTribes) > 0){
			buff += GenerateTribesDescriptionHTML(set);
		}
	}
	reportHTML = strings.Replace(reportHTML,"%ALL_SETS_HTML%", string(buff), -1);


	buff = "";
	buff += "Generated On - " + time.Now().Format("Jan _2 2006") + " - " + strconv.Itoa(len(db.AllCards)) + " Cards in " + strconv.Itoa(len(db.CardSetsByCode)) + " Sets";

	reportHTML = strings.Replace(reportHTML,"%GENERATION_DESCRIPTION%", string(buff), -1);

	reportHTML = strings.Replace(reportHTML,"%CREATURE_COUNT%", string(strconv.Itoa(MinTribeSize)), -1);
	reportHTML = strings.Replace(reportHTML,"%COMMON_CREATURE_COUNT%", string(strconv.Itoa(MinCommonRequired)), -1);
	reportHTML = strings.Replace(reportHTML,"%REFERENCE_COUNT%", string(strconv.Itoa(MinReferenceRequired)), -1);


	ioutil.WriteFile("index.html", []byte(reportHTML), 0644);
	//tribalCandidates := make(map[string]map[string]*Card);

	//

	/**
	for _, card := range set.Cards {
		fmt.Println(card.Name + " - " + card.TypeLine);
	}**/
}

// Warning - this is some terrible code :o)
func GenerateSampleCardHTML(cardTribe *CardTribe, small bool) string {
	buff := "";
	buff += "<table class='CardTable'>";
	buff += "<tr>"

	allCreatures := make([]*Card, 0);
	for _, card := range cardTribe.Creatures {
		allCreatures = append(allCreatures, card);
	}
	sort.SliceStable(allCreatures, func(i, j int) bool {
		return allCreatures[i].RarityScore > allCreatures[j].RarityScore;
	})

	allReferences := make([]*Card, 0);
	for _, card := range cardTribe.References {
		allReferences = append(allReferences, card);
	}
	sort.SliceStable(allReferences, func(i, j int) bool {
		return allReferences[i].RarityScore > allReferences[j].RarityScore;
	})

	first := allCreatures[0];
	second := allCreatures[1];
	third := allCreatures[2];

	buff += "<td>"+GenerateCardHTML(allCreatures[0], small)+"</td>"
	buff += "<td>"+GenerateCardHTML(allCreatures[1], small)+"</td>"
	buff += "<td>"+GenerateCardHTML(allCreatures[2], small)+"</td>"

	var fourth *Card = nil;
	for _, card := range allReferences {
		if(card != first && card != second && card != third){
			fourth = card;
			break;
		}
	}

	if(fourth != nil){
		buff += "<td>"+GenerateCardHTML(fourth, small)+"</td>"
		var fifth *Card = nil;
		for _, card := range allReferences {
			if(card != first && card != second && card != third && card != fourth){
				fifth = card;
				break;
			}
		}
		if(fifth != nil){
			buff += "<td>"+GenerateCardHTML(fifth, small)+"</td>"
		}
	}

	buff += "</tr>";
	buff += "</table>"
	return buff;
}

func GenerateCardHTML(card *Card, small bool) string {
	buff := "";
	buff += "<table class='card-details'>";

	buff += "<tr>"
	buff += "<td>"
	buff += "<a href='" + card.CardUrl + "'>";
	if(small) {
		buff += "<img class='card-image-small' src='" + card.SmallImageUrl + "' border=0>"
	} else{
		buff += "<img class='card-image-large' src='" + card.LargeImageUrl + "' border=0>"
	}

	buff += "</a>"
	//buff += card.Name
	buff += "</td>"
	buff += "</tr>"

	buff += "</table>"
	return buff;
}


func GenerateTribesDescriptionHTML(cardSet *CardSet) string {
	buff := "";
	buff += "<div class='card-set-details-name'>" + cardSet.Name + " (" + strconv.Itoa(cardSet.ReleaseTimestamp.Year()) +") - " + strconv.Itoa(cardSet.TribalScore)+ "% Tribal</div>"
	buff += "<ul class='card-set-details'>";

	tribes := make([]*CardTribe, 0);
	for _, tribe := range cardSet.SupportedTribes{
		tribes = append(tribes, tribe);
	}
	sort.SliceStable(tribes, func(i, j int) bool {
		return tribes[i].CreatureCount + tribes[i].ReferenceCount > tribes[j].CreatureCount + tribes[j].ReferenceCount;
	})

	for _, tribe := range tribes {
		buff += "<li>"+tribe.GenerateLink() + tribe.Name+"</a> - " + strconv.Itoa(tribe.CreatureCount) + " Creatures / " + strconv.Itoa(tribe.ReferenceCount) + " References </li>"
	}

	buff += "</ul>"
	return buff;
}

func GenerateSampleCardHTMLForGlobalTribe(tribe *GlobalCardTribe, small bool) string{
	combinedPool := make([]*Card, 0);
	for _, set := range tribe.SetTribes{
		for _, card := range set.Creatures {
			combinedPool = append(combinedPool, card);
		}
	}
	return GenerateSampleCardHTMLFromRawCardPool(combinedPool, small);
}

func GenerateSampleCardHTMLFromRawCardPool(input []*Card, small bool) string{
	buff := "";
	buff += "<table class='CardTable'>";
	buff += "<tr>"

	dedupeByName := make(map[string]*Card);
	allCreatures := make([]*Card, 0);
	for _, card := range input{
		_, exists := dedupeByName[card.Name];
		if(!exists){
			dedupeByName[card.Name] = card;
			allCreatures = append(allCreatures, card);
		}
	}

	sort.SliceStable(allCreatures, func(i, j int) bool {
		return allCreatures[i].RarityScore > allCreatures[j].RarityScore;
	})


	buff += "<td>"+GenerateCardHTML(allCreatures[0], small)+"</td>"
	buff += "<td>"+GenerateCardHTML(allCreatures[1], small)+"</td>"
	buff += "<td>"+GenerateCardHTML(allCreatures[2], small)+"</td>"
	buff += "<td>"+GenerateCardHTML(allCreatures[3], small)+"</td>"
	buff += "<td>"+GenerateCardHTML(allCreatures[4], small)+"</td>"
	buff += "</tr>";
	buff += "</table>"
	return buff;
}