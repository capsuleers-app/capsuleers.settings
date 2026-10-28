package charsettings

import (
	"regexp"
	"strings"
)

type pair struct{ en, it string }

// Friendly names of common client windows. Unknown ids are shown raw.
var known = map[string]pair{
	"overview":                 {"Overview", "Panoramica"},
	"selecteditemview":         {"Selected item", "Oggetto selezionato"},
	"droneview":                {"Drones", "Droni"},
	"dronesview":               {"Drones", "Droni"},
	"probeScannerWindow":       {"Probe scanner", "Scanner sonde"},
	"directionalScannerWindow": {"D-Scan", "D-Scan"},
	"directionalScanner":       {"D-Scan", "D-Scan"},
	"inventory":                {"Inventory", "Inventario"},
	"ActiveShipCargo":          {"Ship cargo", "Stiva della nave"},
	"fitting":                  {"Fitting", "Fitting"},
	"fittingWnd":               {"Fitting", "Fitting"},
	"wallet":                   {"Wallet", "Portafoglio"},
	"walletWindow":             {"Wallet", "Portafoglio"},
	"neocom":                   {"Neocom", "Neocom"},
	"chat":                     {"Chat", "Chat"},
	"ChatWindowStack":          {"Chat", "Chat"},
	"invitestack":              {"Chat invitations", "Inviti chat"},
	"lobbyWnd":                 {"Station lobby", "Lobby della stazione"},
	"fleetwindow":              {"Fleet", "Flotta"},
	"FleetComposition":         {"Fleet composition", "Composizione flotta"},
	"watchlistpanel":           {"Watch list", "Lista di controllo"},
	"marketbuyaction":          {"Buy from market", "Acquisto dal mercato"},
	"mail":                     {"EVE Mail", "EVE Mail"},
	"SkillPlanner":             {"Skill planner", "Piano abilità"},
	"careerPortal":             {"Career portal", "Portale carriera"},
	"contractdetails":          {"Contract", "Contratto"},
	"MultiBuy":                 {"Multibuy", "Multibuy"},
	"tradeWnd":                 {"Trade", "Scambio"},
	"primary_map_panel":        {"Map panel", "Pannello mappa"},
	"MapCmdWindow":             {"Map commands", "Comandi mappa"},
	"telecom":                  {"Telecom", "Telecom"},
}

var (
	overviewN  = regexp.MustCompile(`^overview_(\d+)$`)
	placesRe   = regexp.MustCompile(`^\('myPlaces'`)
	numericRe  = regexp.MustCompile(`^\d+$`)
	channelFix = map[string]pair{
		"local":    {"Local", "Locale"},
		"corp":     {"Corp", "Corp"},
		"alliance": {"Alliance", "Alleanza"},
		"fleet":    {"Fleet", "Flotta"},
	}
)

func pick(p pair, lang string) string {
	if lang == "it" {
		return p.it
	}
	return p.en
}

// ChannelLabel is the display name of a chat channel: its saved name, else its kind.
func ChannelLabel(c Channel, lang string) string {
	if c.Name != "" {
		return c.Name
	}
	if p, ok := channelFix[c.Kind]; ok {
		return pick(p, lang)
	}
	switch c.Kind {
	case "private":
		return pick(pair{"Private channel", "Canale privato"}, lang)
	case "player":
		return pick(pair{"Player channel", "Canale giocatore"}, lang)
	}
	return c.ID
}

// WindowLabel gives a friendly name for a window id. channels maps chatchannel ids to channels;
// memberLabels are the labels of the tabs of a stack (used when the stack itself has no name).
func WindowLabel(id, lang string, channels map[string]Channel, memberLabels []string) string {
	if strings.HasPrefix(id, "chatchannel_") {
		if c, ok := channels[id]; ok {
			return ChannelLabel(c, lang)
		}
		return ChannelLabel(Channel{ID: id, Kind: channelKind(id)}, lang)
	}
	if p, ok := known[id]; ok && !(len(memberLabels) > 1 && id != "ChatWindowStack") {
		return pick(p, lang)
	}
	if m := overviewN.FindStringSubmatch(id); m != nil {
		return pick(pair{"Overview ", "Panoramica "}, lang) + m[1]
	}
	if placesRe.MatchString(id) {
		return pick(pair{"Places", "Luoghi"}, lang)
	}
	if len(memberLabels) > 0 {
		return strings.Join(memberLabels, " · ")
	}
	if numericRe.MatchString(id) {
		return pick(pair{"Window group ", "Gruppo finestre "}, lang) + id
	}
	return id
}
