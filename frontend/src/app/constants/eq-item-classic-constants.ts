// Lookup lists matching the ProjectEQ PHP editor (lib/items.php) item form

export const CLASSIC_ITEM_SIZE = {
  0: "Tiny",
  1: "Small",
  2: "Medium",
  3: "Large",
  4: "Giant",
  5: "Giant - No Container",
}

export const CLASSIC_ITEM_BAG_SIZE = {
  0: "Non-Bag",
  1: "Small",
  2: "Medium",
  3: "Large",
  4: "Giant",
  5: "Giant - Assembly Kit",
}

export const CLASSIC_ITEM_LDON_THEME = {
  0: "None",
  1: "GUK",
  2: "MIR",
  4: "MMC",
  8: "RUJ",
  16: "TAK",
  31: "ALL",
}

export const CLASSIC_ITEM_POINT_TYPE = {
  0: "None",
  1: "LDoN",
  2: "Discord Merchant",
  4: "Norrath Keeper",
  5: "Dark Reign",
}

export const CLASSIC_ITEM_CLICK_TYPE = {
  0: "None",
  1: "Click from inventory w/Lvl",
  3: "Expendable",
  4: "Must equip to click",
  5: "Click from inventory w/Lvl/Class/Race",
}

export const CLASSIC_ITEM_PROC_TYPE = {0: "None/Proc"}
export const CLASSIC_ITEM_WORN_TYPE = {0: "None", 2: "Worn"}
export const CLASSIC_ITEM_FOCUS_TYPE = {0: "None", 6: "Focus"}
export const CLASSIC_ITEM_SCROLL_TYPE = {0: "None", 7: "Scroll"}

export const CLASSIC_NO_YES = {0: "No", 1: "Yes"}

// nodrop / norent are inverted in the database (1 = can drop / can rent)
export const CLASSIC_INVERTED_NO_YES = {1: "No", 0: "Yes"}

export const CLASSIC_SLOT_BITS = [
  [1, "Charm"], [2, "Ear01"], [4, "Head"], [8, "Face"],
  [16, "Ear02"], [32, "Neck"], [64, "Shoulders"], [128, "Arms"],
  [256, "Back"], [512, "Bracer01"], [1024, "Bracer02"], [2048, "Range"],
  [4096, "Hands"], [8192, "Primary"], [16384, "Secondary"], [32768, "Ring01"],
  [65536, "Ring02"], [131072, "Chest"], [262144, "Legs"], [524288, "Feet"],
  [1048576, "Waist"], [2097152, "Powersource"], [4194304, "Ammo"],
]

export const CLASSIC_RACE_BITS = [
  [1, "Human"], [2, "Barbarian"], [4, "Erudite"], [8, "Wood Elf"],
  [16, "High Elf"], [32, "Dark Elf"], [64, "Half Elf"], [128, "Dwarf"],
  [256, "Troll"], [512, "Ogre"], [1024, "Halfling"], [2048, "Gnome"],
  [4096, "Iksar"], [8192, "Vah Shir"], [16384, "Froglok"], [32768, "Drakkin"],
  [65536, "Shroud"],
]

export const CLASSIC_CLASS_BITS = [
  [1, "Warrior"], [2, "Cleric"], [4, "Paladin"], [8, "Ranger"],
  [16, "Shadowknight"], [32, "Druid"], [64, "Monk"], [128, "Bard"],
  [256, "Rogue"], [512, "Shaman"], [1024, "Necromancer"], [2048, "Wizard"],
  [4096, "Magician"], [8192, "Enchanter"], [16384, "Beastlord"], [32768, "Berserker"],
]

export const CLASSIC_DEITY_BITS = [
  [1, "Agnostic"], [2, "Bertoxxulous"], [4, "Brell Serilis"], [8, "Cazic-Thule"],
  [16, "Erollisi Marr"], [32, "Bristlebane"], [64, "Innoruuk"], [128, "Karana"],
  [256, "Mithaniel Marr"], [512, "Prexus"], [1024, "Quellious"], [2048, "Rallos Zek"],
  [4096, "Rodcet Nife"], [8192, "Solusek Ro"], [16384, "The Tribunal"], [32768, "Tunare"],
  [65536, "Veeshan"],
]

export const CLASSIC_AUG_TYPE_BITS = Array.from({length: 30}, (_, i) => [2 ** i, "Type " + (i + 1)])
