package charsettings

import (
	"regexp"
	"strings"
)

type pair struct{ en, it string }

// Names of common client windows, as the EVE client shows them. They are game names, so they
// stay in English in every UI language (the Italian client would not match them anyway).
// Unknown ids are shown raw.
var known = map[string]string{
	"overview":                 "Overview",
	"selecteditemview":         "Selected Item",
	"droneview":                "Drones",
	"dronesview":               "Drones",
	"probeScannerWindow":       "Probe Scanner",
	"directionalScannerWindow": "Directional Scanner",
	"directionalScanner":       "Directional Scanner",
	"inventory":                "Inventory",
	"ActiveShipCargo":          "Ship Cargo",
	"fitting":                  "Fitting",
	"fittingWnd":               "Fitting",
	"wallet":                   "Wallet",
	"walletWindow":             "Wallet",
	"neocom":                   "Neocom",
	"chat":                     "Chat",
	"ChatWindowStack":          "Chat",
	"invitestack":              "Chat Invitations",
	"lobbyWnd":                 "Station Services",
	"fleetwindow":              "Fleet",
	"FleetComposition":         "Fleet Composition",
	"watchlistpanel":           "Watch List",
	"marketbuyaction":          "Buy",
	"mail":                     "EVE Mail",
	"SkillPlanner":             "Skill Plans",
	"careerPortal":             "Opportunities",
	"contractdetails":          "Contract",
	"MultiBuy":                 "Multibuy",
	"tradeWnd":                 "Trade",
	"primary_map_panel":        "Map",
	"MapCmdWindow":             "Map",
	"telecom":                  "Telecom",
}

var (
	overviewN = regexp.MustCompile(`^overview_(\d+)$`)
	placesRe  = regexp.MustCompile(`^\('myPlaces'`)
	numericRe = regexp.MustCompile(`^\d+$`)
	// Fixed channels, as named in the game (not translated).
	channelFix = map[string]string{
		"local":    "Local",
		"corp":     "Corp",
		"alliance": "Alliance",
		"fleet":    "Fleet",
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
	if n, ok := channelFix[c.Kind]; ok {
		return n
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
	if n, ok := known[id]; ok && !(len(memberLabels) > 1 && id != "ChatWindowStack") {
		return n
	}
	if m := overviewN.FindStringSubmatch(id); m != nil {
		return "Overview " + m[1]
	}
	if placesRe.MatchString(id) {
		return "Places"
	}
	if len(memberLabels) > 0 {
		return strings.Join(memberLabels, " · ")
	}
	if numericRe.MatchString(id) {
		return pick(pair{"Window group ", "Gruppo finestre "}, lang) + id
	}
	return id
}
