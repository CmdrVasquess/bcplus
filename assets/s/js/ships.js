const ships = {
    typeNmMap: {
	adder: "Adder",
	anaconda: "Anaconda",
	asp: "Asp Explorer",
	asp_scout: "Asp Scout",
	belugaliner: "Beluga Liner",
	cobramkiii: "Cobra MkIII",
	cobramkiv: "Cobra MkIV",
	cutter: "Imperial Cutter",
	diamondback: "Diamondback Scout",
	diamondbackxl: "Diamondback Explorer",
	dolphin: "Dolphin",
	empire_courier: "Imperial Courier",
	empire_eagle: "Imperial Eagle",
	empire_fighter: "Gu-97",
	empire_trader: "Imperial Clipper",
	federation_corvette: "Federal Corvette",
	federation_dropship: "Federal Dropship",
	federation_dropship_mkii: "Federal Assault Ship",
	federation_fighter: "F63 Condor",
	federation_gunship: "Federal Gunship",
	ferdelance: "Fer-de-Lance",
	hauler: "Hauler",
	independant_trader: "Keelback",
	independent_fighter: "Taipan",
	krait_mkii: "Krait MkII",
	mamba: "Mamba",
	orca: "Orca",
	python: "Python",
	sidewinder: "Sidewinder",
	testbuggy: "SRV",
	type6: "Type-6 Transporter",
	type7: "Type-7 Transporter",
	type9: "Type-9 Heavy",
	type9_military: "Type-10 Defender",
	typex: "Alliance Chieftain",
	typex_2: "Alliance Crusader",
	typex_3: "Alliance Challenger",
	viper: "Viper MkIII",
	viper_mkiv: "Viper MkIV",
	vulture: "Vulture"
    }
};
Vue.filter('shipTyNm', function(sty) {
    var res = ships.typeNmMap[sty];
    if (!res) { res = "<"+sty+">"; }
    return res;
});
