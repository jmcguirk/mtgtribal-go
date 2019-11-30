package main

type TopSetsReportChart struct {
	Labels		[]ReportLabel `json:"labels"`;
	Data		[]TopSetsReportData `json:"series"`;
}

type TopSetsReportData struct {
	Value 		int `json:"value"`;
	Label		string `json:"meta"`;
}

type TribalTrendGraph struct {
	Data		[]TribalTrendData `json:"series"`;
}

type TribalTrendData struct {
	ReleaseTimeStamp 		int `json:"releaseTimeStamp"`;
	SetName					string `json:"setName"`;
	Score		 		int `json:"score"`;
}

type ReportLabel struct {
	Name		string `json:"label"`;
	ImageUrl    string `json:"image"`;
}