package zonecontroller

var coverageSlots = []struct {
	Name string
	Bits []int
}{
	{"Charm", []int{1}},
	{"Ear", []int{2, 16}},
	{"Head", []int{4}},
	{"Face", []int{8}},
	{"Neck", []int{32}},
	{"Shoulders", []int{64}},
	{"Arms", []int{128}},
	{"Back", []int{256}},
	{"Wrist", []int{512, 1024}},
	{"Range", []int{2048}},
	{"Hands", []int{4096}},
	{"Primary", []int{8192}},
	{"Secondary", []int{16384}},
	{"Ring", []int{32768, 65536}},
	{"Chest", []int{131072}},
	{"Legs", []int{262144}},
	{"Feet", []int{524288}},
	{"Waist", []int{1048576}},
}

var coverageClasses = []struct {
	ID   int
	Name string
	Bit  int
}{
	{1, "WAR", 1},
	{2, "CLR", 2},
	{3, "PAL", 4},
	{4, "RNG", 8},
	{5, "SK", 16},
	{6, "DRU", 32},
	{7, "MNK", 64},
	{8, "BRD", 128},
	{9, "ROG", 256},
	{10, "SHM", 512},
	{11, "NEC", 1024},
	{12, "WIZ", 2048},
	{13, "MAG", 4096},
	{14, "ENC", 8192},
	{15, "BST", 16384},
	{16, "BER", 32768},
}

type CoverageCell struct {
	Slot    string `json:"slot"`
	Class   string `json:"class"`
	ClassID int    `json:"classId"`
	Count   int    `json:"count"`
	Filled  bool   `json:"filled"`
}

func kitCoverage(items []RecipeKitItem) []CoverageCell {
	out := make([]CoverageCell, 0, len(coverageSlots)*len(coverageClasses))
	for _, slot := range coverageSlots {
		for _, class := range coverageClasses {
			n := 0
			for _, it := range items {
				if itemFits(it, slot.Bits, class.Bit) {
					n++
				}
			}
			out = append(out, CoverageCell{
				Slot:    slot.Name,
				Class:   class.Name,
				ClassID: class.ID,
				Count:   n,
				Filled:  n > 0,
			})
		}
	}
	return out
}

func itemFits(it RecipeKitItem, slotBits []int, classBit int) bool {
	if it.Slots == 0 && it.Classes == 0 {
		return false
	}
	okSlot := it.Slots == 0
	for _, bit := range slotBits {
		if it.Slots&bit != 0 {
			okSlot = true
			break
		}
	}
	okClass := it.Classes == 0 || it.Classes&classBit != 0
	return okSlot && okClass
}

func coverageMissing(cells []CoverageCell) int {
	n := 0
	for _, cell := range cells {
		if !cell.Filled {
			n++
		}
	}
	return n
}
