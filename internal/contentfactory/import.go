package contentfactory

type ZoneImport struct {
	OK        bool          `json:"ok"`
	Error     string        `json:"error,omitempty"`
	ZoneID    int           `json:"zoneId"`
	Short     string        `json:"short"`
	Cells     []KitCell     `json:"cells"`
	SpellSets []CensusSpellSet `json:"spellSets"`
	Items     []CensusItem  `json:"items"`
	Filled    int           `json:"filled"`
	Note      string        `json:"note,omitempty"`
}

func (s *Service) ImportZone(zone string) ZoneImport {
	out := ZoneImport{Cells: []KitCell{}, SpellSets: []CensusSpellSet{}, Items: []CensusItem{}}
	census := s.Census(zone)
	if !census.OK {
		out.Error = census.Error
		return out
	}
	out.ZoneID = census.ZoneID
	out.Short = census.Short
	out.SpellSets = census.SpellSets
	out.Items = census.Items
	out.Cells = pickKitCells(census.Items)
	out.Filled = len(out.Cells)
	out.Note = "Filled the first unused class/slot cell for each gear drop. Empty slots were skipped."
	out.OK = true
	return out
}

func pickKitCells(items []CensusItem) []KitCell {
	used := map[string]bool{}
	out := []KitCell{}
	for _, it := range items {
		if it.ID <= 0 || it.Slots <= 0 {
			continue
		}
		slot := firstNonEmpty(it.Slot, slotName(it.Slots))
		if slot == "" {
			continue
		}
		placed := false
		for _, cls := range kitClasses {
			if it.Classes > 0 && it.Classes&cls.Bit == 0 {
				continue
			}
			key := slot + "|" + cls.Name
			if used[key] {
				continue
			}
			used[key] = true
			out = append(out, KitCell{
				Slot: slot, Class: cls.Name, SourceID: it.ID, Name: it.Name,
				Slots: slotBits(slot), Classes: cls.Bit,
				ClickSpellID: it.Click, ProcSpellID: it.Proc, WornSpellID: it.Worn, FocusSpellID: it.Focus,
			})
			placed = true
			break
		}
		if !placed && !used[slot+"|ALL"] {
			used[slot+"|ALL"] = true
			out = append(out, KitCell{
				Slot: slot, Class: "", SourceID: it.ID, Name: it.Name,
				Slots: slotBits(slot), Classes: it.Classes,
				ClickSpellID: it.Click, ProcSpellID: it.Proc,
			})
		}
	}
	return out
}
