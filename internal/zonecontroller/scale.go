package zonecontroller

import (
	"math"
	"strconv"
	"strings"

	"github.com/EQEmu/spire/internal/models"
)

var basedataScaleKeys = []string{"level", "max_hp", "min_hit", "max_hit", "atk", "ac", "cash"}

func scaleBasedata(src map[string]map[string]string, mult float64) map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, kind := range []string{"trash", "boss", "raid"} {
		row := map[string]string{}
		for k, v := range src[kind] {
			row[k] = v
		}
		if math.Abs(mult-1) > 0.0001 {
			for _, key := range basedataScaleKeys {
				if scaled, ok := scaleNumberString(row[key], mult); ok {
					row[key] = scaled
				}
			}
		}
		out[kind] = row
	}
	return out
}

func scaleNumberString(raw string, mult float64) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "-1" {
		return raw, false
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return raw, false
	}
	if n == 0 {
		return raw, false
	}
	return strconv.FormatInt(int64(math.Round(n*mult)), 10), true
}

func scaleItemStats(item *models.Item, mult float64) {
	if item == nil || math.Abs(mult-1) <= 0.0001 {
		return
	}
	scaleInt(&item.Ac, mult)
	scaleInt(&item.Hp, mult)
	scaleInt(&item.Mana, mult)
	scaleInt(&item.Endur, mult)
	scaleInt(&item.Astr, mult)
	scaleInt(&item.Asta, mult)
	scaleInt(&item.Aagi, mult)
	scaleInt(&item.Adex, mult)
	scaleInt(&item.Awis, mult)
	scaleInt(&item.Aint, mult)
	scaleInt(&item.Acha, mult)
	scaleInt(&item.Attack, mult)
	scaleInt(&item.Accuracy, mult)
	scaleInt(&item.Avoidance, mult)
	scaleInt(&item.Haste, mult)
	scaleInt(&item.Damage, mult)
	scaleInt(&item.Extradmgamt, mult)
	scaleInt16(&item.Backstabdmg, mult)
	scaleInt16(&item.Spelldmg, mult)
	scaleInt(&item.Banedmgamt, mult)
	scaleInt(&item.Banedmgraceamt, mult)
	scaleInt(&item.Elemdmgamt, mult)
	scaleInt(&item.Regen, mult)
	scaleInt(&item.Manaregen, mult)
	scaleInt(&item.Enduranceregen, mult)
	scaleInt(&item.Cr, mult)
	scaleInt(&item.Dr, mult)
	scaleInt(&item.Fr, mult)
	scaleInt(&item.Mr, mult)
	scaleInt(&item.Pr, mult)
	scaleInt(&item.Shielding, mult)
	scaleInt(&item.Spellshield, mult)
	scaleInt(&item.Strikethrough, mult)
	scaleInt(&item.Stunresist, mult)
	scaleInt(&item.Dotshielding, mult)
	scaleInt(&item.Damageshield, mult)
	scaleInt16(&item.Healamt, mult)
	scaleInt16(&item.Clairvoyance, mult)
	scaleInt16(&item.Dsmitigation, mult)
	scaleInt(&item.Purity, mult)
	if item.Price > 0 {
		item.Price = int(math.Round(float64(item.Price) * mult))
	}
}

func scaleInt(v *int, mult float64) {
	if v == nil || *v <= 0 {
		return
	}
	n := int(math.Round(float64(*v) * mult))
	if n < 1 {
		n = 1
	}
	*v = n
}

func scaleInt16(v *int16, mult float64) {
	if v == nil || *v <= 0 {
		return
	}
	n := int(math.Round(float64(*v) * mult))
	if n < 1 {
		n = 1
	}
	if n > 32767 {
		n = 32767
	}
	*v = int16(n)
}

func prefixedName(prefix, name string) string {
	name = strings.TrimSpace(name)
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || name == "" {
		return name
	}
	if strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
		return name
	}
	return strings.TrimSpace(prefix + " " + name)
}

func ladderMultiplier(step float64, index int) float64 {
	if step <= 0 {
		step = 1
	}
	if index <= 0 {
		return 1
	}
	return math.Pow(step, float64(index))
}
