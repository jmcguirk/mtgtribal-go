### MTG: Tribal Analysis

This repo hosts the golang source code for some basic ad hoc analysis for *Tribes* in the popular collectable trading card game; Magic: The Gathering. It leverages a large data export from Scryfall and a charting Library (Chartist.js) to produce a simple static export detailing some basic analysis on the cards contained therin.

Take a look at the [generated report ](https://jmcguirk.github.io/mtgtribal-go/ "generated report ")for more details on approach and findings

#### Getting Started
To use this report (and regenerate the report.) You will need to fetch the[ latest exports from Scryfal](https://scryfall.com/docs/api/bulk-data " latest exports from Scryfal")l. Grab the latest set meta data and card exports and save the JSONs into the source code directory.

#### Caveats 
The HTML generation code in this repo is very embarassing and should not be looked at or appreciated by anyone