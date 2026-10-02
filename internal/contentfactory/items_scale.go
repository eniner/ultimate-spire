package contentfactory

import (
	"math"

	"github.com/EQEmu/spire/internal/models"
)

func scaleItemStats(item *models.Item, mult float64) {
	if item == nil || math.Abs(mult-1) <= 0.0001 {
		return
	}
	scalePositive(&item.Ac, mult)
	scalePositive(&item.Hp, mult)
	scalePositive(&item.Mana, mult)
	scalePositive(&item.Endur, mult)
	scalePositive(&item.Astr, mult)
	scalePositive(&item.Asta, mult)
	scalePositive(&item.Aagi, mult)
	scalePositive(&item.Adex, mult)
	scalePositive(&item.Awis, mult)
	scalePositive(&item.Aint, mult)
	scalePositive(&item.Acha, mult)
	scalePositive(&item.Attack, mult)
	scalePositive(&item.Accuracy, mult)
	scalePositive(&item.Avoidance, mult)
	scalePositive(&item.Haste, mult)
	scalePositive(&item.Damage, mult)
	scalePositive(&item.Extradmgamt, mult)
	scaleInt16(&item.Backstabdmg, mult)
	scaleInt16(&item.Spelldmg, mult)
	scalePositive(&item.Banedmgamt, mult)
	scalePositive(&item.Banedmgraceamt, mult)
	scalePositive(&item.Elemdmgamt, mult)
	scalePositive(&item.Regen, mult)
	scalePositive(&item.Manaregen, mult)
	scalePositive(&item.Enduranceregen, mult)
	scalePositive(&item.Cr, mult)
	scalePositive(&item.Dr, mult)
	scalePositive(&item.Fr, mult)
	scalePositive(&item.Mr, mult)
	scalePositive(&item.Pr, mult)
	scalePositive(&item.Shielding, mult)
	scalePositive(&item.Spellshield, mult)
	scalePositive(&item.Strikethrough, mult)
	scalePositive(&item.Stunresist, mult)
	scalePositive(&item.Dotshielding, mult)
	scalePositive(&item.Damageshield, mult)
	scaleInt16(&item.Healamt, mult)
	scaleInt16(&item.Clairvoyance, mult)
	scaleInt16(&item.Dsmitigation, mult)
	scalePositive(&item.Purity, mult)
	if item.Price > 0 {
		item.Price = int(math.Round(float64(item.Price) * mult))
		if item.Price < 1 {
			item.Price = 1
		}
	}
}
