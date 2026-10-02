package contentfactory

import "fmt"

type CatalogUse struct {
	Kind string `json:"kind"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CatalogRow struct {
	ID      int          `json:"id"`
	Name    string       `json:"name"`
	Kind    string       `json:"kind"`
	Extra   string       `json:"extra,omitempty"`
	Click   int          `json:"click,omitempty"`
	Proc    int          `json:"proc,omitempty"`
	Uses    []CatalogUse `json:"uses,omitempty"`
	UseNote string       `json:"useNote,omitempty"`
}

type Catalog struct {
	OK    bool         `json:"ok"`
	Error string       `json:"error,omitempty"`
	Kind  string       `json:"kind"`
	Floor int          `json:"floor"`
	Rows  []CatalogRow `json:"rows"`
}

func (s *Service) Catalog(kind string, limit int) Catalog {
	if limit <= 0 || limit > 200 {
		limit = catalogLimit
	}
	out := Catalog{Kind: kind, Floor: reservedFloor, Rows: []CatalogRow{}}
	if s.eqemu() == nil {
		out.Error = "PEQ database is not connected"
		return out
	}
	switch kind {
	case "spells", "spells_new":
		out.Kind = "spells"
		type row struct {
			ID   int
			Name string
		}
		var rows []row
		if err := s.eqemu().Table("spells_new").Select("id, name").Where("id >= ?", reservedFloor).Order("id desc").Limit(limit).Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		ids := make([]int, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		itemUses := s.itemSpellUses(ids)
		for _, r := range rows {
			uses := itemUses[r.ID]
			out.Rows = append(out.Rows, CatalogRow{
				ID: r.ID, Name: r.Name, Kind: "spell", Uses: uses, UseNote: useNote(uses, "item"),
			})
		}
	case "loot", "loottable":
		out.Kind = "loottable"
		type row struct {
			ID   int
			Name string
		}
		var rows []row
		if err := s.eqemu().Table("loottable").Select("id, name").Where("id >= ?", reservedFloor).Order("id desc").Limit(limit).Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		ids := make([]int, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		npcUses := s.npcFieldUses("loottable_id", ids)
		for _, r := range rows {
			uses := npcUses[r.ID]
			out.Rows = append(out.Rows, CatalogRow{
				ID: r.ID, Name: r.Name, Kind: "loottable", Uses: uses, UseNote: useNote(uses, "npc"),
			})
		}
	case "npc_spells", "spellsets":
		out.Kind = "npc_spells"
		type row struct {
			ID   int
			Name string
		}
		var rows []row
		if err := s.eqemu().Table("npc_spells").Select("id, name").Where("id >= ?", reservedFloor).Order("id desc").Limit(limit).Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		ids := make([]int, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		npcUses := s.npcFieldUses("npc_spells_id", ids)
		for _, r := range rows {
			uses := npcUses[r.ID]
			out.Rows = append(out.Rows, CatalogRow{
				ID: r.ID, Name: r.Name, Kind: "npc_spells", Uses: uses, UseNote: useNote(uses, "npc"),
			})
		}
	default:
		out.Kind = "items"
		type row struct {
			ID          int
			Name        string `gorm:"column:Name"`
			Clickeffect int
			Proceffect  int
		}
		var rows []row
		if err := s.eqemu().Table("items").Select("id, Name, clickeffect, proceffect").Where("id >= ?", reservedFloor).Order("id desc").Limit(limit).Scan(&rows).Error; err != nil {
			out.Error = err.Error()
			return out
		}
		ids := make([]int, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		merchUses := s.merchantItemUses(ids)
		npcUses := s.lootItemNpcUses(ids)
		for _, r := range rows {
			uses := append(merchUses[r.ID], npcUses[r.ID]...)
			extra := ""
			if r.Clickeffect > 0 {
				extra = fmt.Sprintf("click %d", r.Clickeffect)
			}
			out.Rows = append(out.Rows, CatalogRow{
				ID: r.ID, Name: r.Name, Kind: "item", Extra: extra,
				Click: r.Clickeffect, Proc: r.Proceffect, Uses: uses, UseNote: useNote(uses, "use"),
			})
		}
	}
	out.OK = true
	return out
}

func useNote(uses []CatalogUse, noun string) string {
	if len(uses) == 0 {
		return "unused"
	}
	if len(uses) == 1 {
		return noun + " " + uses[0].Name
	}
	return fmt.Sprintf("%d %ss", len(uses), noun)
}

func (s *Service) itemSpellUses(spellIDs []int) map[int][]CatalogUse {
	out := map[int][]CatalogUse{}
	if len(spellIDs) == 0 {
		return out
	}
	type row struct {
		ID          int
		Name        string `gorm:"column:Name"`
		Clickeffect int
		Proceffect  int
		Worneffect  int
		Focuseffect int
	}
	var rows []row
	_ = s.eqemu().Table("items").Select("id, Name, clickeffect, proceffect, worneffect, focuseffect").
		Where("clickeffect IN ? OR proceffect IN ? OR worneffect IN ? OR focuseffect IN ?", spellIDs, spellIDs, spellIDs, spellIDs).
		Limit(400).Scan(&rows).Error
	want := map[int]bool{}
	for _, id := range spellIDs {
		want[id] = true
	}
	for _, r := range rows {
		for _, sid := range []int{r.Clickeffect, r.Proceffect, r.Worneffect, r.Focuseffect} {
			if want[sid] {
				out[sid] = append(out[sid], CatalogUse{Kind: "item", ID: r.ID, Name: r.Name})
			}
		}
	}
	return out
}

func (s *Service) npcFieldUses(field string, ids []int) map[int][]CatalogUse {
	out := map[int][]CatalogUse{}
	if len(ids) == 0 {
		return out
	}
	type row struct {
		ID    int
		Name  string
		Value int `gorm:"column:value"`
	}
	var rows []row
	_ = s.eqemu().Table("npc_types").Select("id, name, "+field+" AS value").Where(field+" IN ?", ids).Limit(400).Scan(&rows).Error
	for _, r := range rows {
		out[r.Value] = append(out[r.Value], CatalogUse{Kind: "npc", ID: r.ID, Name: r.Name})
	}
	return out
}

func (s *Service) merchantItemUses(itemIDs []int) map[int][]CatalogUse {
	out := map[int][]CatalogUse{}
	if len(itemIDs) == 0 {
		return out
	}
	type row struct {
		Item       int
		MerchantID int `gorm:"column:merchantid"`
	}
	var rows []row
	_ = s.eqemu().Table("merchantlist").Select("item, merchantid").Where("item IN ?", itemIDs).Scan(&rows).Error
	for _, r := range rows {
		out[r.Item] = append(out[r.Item], CatalogUse{Kind: "merchant", ID: r.MerchantID, Name: fmt.Sprintf("merchant %d", r.MerchantID)})
	}
	return out
}

func (s *Service) lootItemNpcUses(itemIDs []int) map[int][]CatalogUse {
	out := map[int][]CatalogUse{}
	if len(itemIDs) == 0 {
		return out
	}
	type row struct {
		ItemID int    `gorm:"column:item_id"`
		NPCID  int    `gorm:"column:npc_id"`
		Name   string `gorm:"column:name"`
	}
	var rows []row
	_ = s.eqemu().Table("lootdrop_entries").
		Select("lootdrop_entries.item_id, npc_types.id AS npc_id, npc_types.name").
		Joins("JOIN loottable_entries ON loottable_entries.lootdrop_id = lootdrop_entries.lootdrop_id").
		Joins("JOIN npc_types ON npc_types.loottable_id = loottable_entries.loottable_id").
		Where("lootdrop_entries.item_id IN ?", itemIDs).
		Limit(400).
		Scan(&rows).Error
	seen := map[string]bool{}
	for _, r := range rows {
		key := fmt.Sprintf("%d:%d", r.ItemID, r.NPCID)
		if seen[key] {
			continue
		}
		seen[key] = true
		out[r.ItemID] = append(out[r.ItemID], CatalogUse{Kind: "npc-loot", ID: r.NPCID, Name: r.Name})
	}
	return out
}
