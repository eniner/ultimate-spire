package contentfactory

type MerchantStock struct {
	Slot        int    `json:"slot"`
	ItemID      int    `json:"itemId"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Probability int    `json:"probability"`
}

type MerchantView struct {
	OK          bool            `json:"ok"`
	Error       string          `json:"error,omitempty"`
	MerchantID  int             `json:"merchantId"`
	NPCName     string          `json:"npcName,omitempty"`
	NPCID       int             `json:"npcId,omitempty"`
	Stock       []MerchantStock `json:"stock"`
	Count       int             `json:"count"`
	NextSlot    int             `json:"nextSlot"`
	ReplaceNote string          `json:"replaceNote"`
}

func (s *Service) Merchant(id int) MerchantView {
	out := MerchantView{MerchantID: id, Stock: []MerchantStock{}}
	if id <= 0 {
		out.Error = "merchant id is required"
		return out
	}
	if s.eqemu() == nil {
		out.Error = "PEQ database is not connected"
		return out
	}
	type npcRow struct {
		ID   int
		Name string
	}
	var npc npcRow
	_ = s.eqemu().Table("npc_types").Select("id, name").Where("merchant_id = ?", id).Limit(1).Scan(&npc).Error
	out.NPCID = npc.ID
	out.NPCName = npc.Name

	type row struct {
		Slot        int
		Item        int
		Probability int
		Name        string `gorm:"column:Name"`
		Price       int
	}
	var rows []row
	err := s.eqemu().Table("merchantlist").
		Select("merchantlist.slot, merchantlist.item, merchantlist.probability, items.Name, items.price").
		Joins("LEFT JOIN items ON items.id = merchantlist.item").
		Where("merchantlist.merchantid = ?", id).
		Order("merchantlist.slot").
		Scan(&rows).Error
	if err != nil {
		out.Error = err.Error()
		return out
	}
	maxSlot := 0
	for _, r := range rows {
		if r.Slot > maxSlot {
			maxSlot = r.Slot
		}
		out.Stock = append(out.Stock, MerchantStock{
			Slot: r.Slot, ItemID: r.Item, Name: r.Name, Price: r.Price, Probability: r.Probability,
		})
	}
	out.Count = len(out.Stock)
	out.NextSlot = maxSlot + 1
	out.ReplaceNote = "Append uses the next free slot. Replace deletes current stock first."
	out.OK = true
	return out
}
