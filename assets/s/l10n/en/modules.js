var l10n;
if (!l10n) { l10n = {}; }
l10n.modules = {
	"fuelscoop": "Fuel Scoop",
	"cargorack": "Cargo Rack",
	"repairer": "Auto Field Maintainance",
	"guardianfsdbooster": "Guardian FSD Booster",
	"fighterbay": "Fighter Hangar",
	"buggybay": "SRV Hangar",
	"shieldgenerator": "Shield Generator",
	"dronecontrol_repair": "Repair Limpet Controller",
	"dronecontrol_unkvesselresearch": "Research Limpet Controller",
	"detailedsurfacescanner_tiny": "Detailed Surface Scanner",
	"supercruiseassist": "Supercruise Assist"
};
Vue.filter('moduleNm', function(mod) {
    var res = l10n.modules[mod];
    if (!res) { res = "<"+mod+">"; }
    return res;
});
