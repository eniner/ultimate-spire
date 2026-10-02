package contentfactory

import "strings"

var kitSlots = []struct {
	Name string
	Bits int
}{
	{"Charm", 1},
	{"Ear", 2 | 16},
	{"Head", 4},
	{"Face", 8},
	{"Neck", 32},
	{"Shoulders", 64},
	{"Arms", 128},
	{"Back", 256},
	{"Wrist", 512 | 1024},
	{"Range", 2048},
	{"Hands", 4096},
	{"Primary", 8192},
	{"Secondary", 16384},
	{"Ring", 32768 | 65536},
	{"Chest", 131072},
	{"Legs", 262144},
	{"Feet", 524288},
	{"Waist", 1048576},
}

var kitClasses = []struct {
	Name string
	Bit  int
}{
	{"WAR", 1},
	{"CLR", 2},
	{"PAL", 4},
	{"RNG", 8},
	{"SK", 16},
	{"DRU", 32},
	{"MNK", 64},
	{"BRD", 128},
	{"ROG", 256},
	{"SHM", 512},
	{"NEC", 1024},
	{"WIZ", 2048},
	{"MAG", 4096},
	{"ENC", 8192},
	{"BST", 16384},
	{"BER", 32768},
}

func slotBits(name string) int {
	name = strings.TrimSpace(strings.ToLower(name))
	for _, slot := range kitSlots {
		if strings.ToLower(slot.Name) == name {
			return slot.Bits
		}
	}
	return 0
}

func classBit(name string) int {
	name = strings.TrimSpace(strings.ToUpper(name))
	if name == "ALL" {
		return 65535
	}
	for _, class := range kitClasses {
		if class.Name == name {
			return class.Bit
		}
	}
	return 0
}

func slotName(bits int) string {
	for _, slot := range kitSlots {
		if bits&slot.Bits != 0 {
			return slot.Name
		}
	}
	return ""
}

func className(bits int) string {
	for _, class := range kitClasses {
		if bits&class.Bit != 0 {
			return class.Name
		}
	}
	return ""
}

type Meta struct {
	Slots          []MetaRow `json:"slots"`
	Classes        []MetaRow `json:"classes"`
	SpellIDCap     int       `json:"npcSpellIdCap"`
	NpcCastFloor   int       `json:"npcCastSpellFloor"`
	Reserved       int       `json:"reserved"`
}

type MetaRow struct {
	Name string `json:"name"`
	Bits int    `json:"bits"`
}

func FactoryMeta() Meta {
	out := Meta{Slots: []MetaRow{}, Classes: []MetaRow{}, SpellIDCap: npcSpellIDCap, NpcCastFloor: npcCastSpellFloor, Reserved: reservedFloor}
	for _, slot := range kitSlots {
		out.Slots = append(out.Slots, MetaRow{Name: slot.Name, Bits: slot.Bits})
	}
	for _, class := range kitClasses {
		out.Classes = append(out.Classes, MetaRow{Name: class.Name, Bits: class.Bit})
	}
	return out
}
