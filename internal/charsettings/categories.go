package charsettings

import "strings"

// Window categories used by the preview filter.
const (
	CatCombat     = "combat"
	CatChat       = "chat"
	CatFleet      = "fleet"
	CatShip       = "ship"
	CatMarket     = "market"
	CatNavigation = "navigation"
	CatOther      = "other"
)

// Explicit table first; heuristics only for ids not listed here.
var categoryOf = map[string]string{
	"overview": CatCombat, "selecteditemview": CatCombat, "droneview": CatCombat, "dronesview": CatCombat,
	"directionalScannerWindow": CatCombat, "probeScannerWindow": CatCombat,
	"chat": CatChat, "ChatWindowStack": CatChat, "invitestack": CatChat,
	"fleetwindow": CatFleet, "FleetComposition": CatFleet, "watchlistpanel": CatFleet,
	"inventory": CatShip, "ActiveShipCargo": CatShip, "fitting": CatShip, "fittingWnd": CatShip,
	"marketbuyaction": CatMarket, "MultiBuy": CatMarket, "tradeWnd": CatMarket, "contractdetails": CatMarket,
	"wallet": CatMarket, "walletWindow": CatMarket, "SellItemsWindow": CatMarket,
	"primary_map_panel": CatNavigation, "MapCmdWindow": CatNavigation,
	"StoreFleetSetupWnd": CatOther, "TaskConversationWindow": CatOther,
}

func idCategory(id string) string {
	if c, ok := categoryOf[id]; ok {
		return c
	}
	switch {
	case overviewN.MatchString(id), strings.HasPrefix(id, "directionalScanner"):
		return CatCombat
	case strings.HasPrefix(id, "chatchannel_"):
		return CatChat
	case placesRe.MatchString(id):
		return CatNavigation
	}
	l := strings.ToLower(id)
	has := func(subs ...string) bool {
		for _, s := range subs {
			if strings.Contains(l, s) {
				return true
			}
		}
		return false
	}
	switch {
	case has("inventory", "cargo", "hangar"):
		return CatShip
	case has("market", "contract", "sell", "buy"):
		return CatMarket
	case has("map"):
		return CatNavigation
	}
	return CatOther
}

// WindowCategory classifies a window for the preview filter. A stack (non-empty members)
// takes the category of its first member that is not "other".
func WindowCategory(id string, members []string) string {
	if len(members) > 0 {
		for _, m := range members {
			if c := idCategory(m); c != CatOther {
				return c
			}
		}
		return CatOther
	}
	return idCategory(id)
}
