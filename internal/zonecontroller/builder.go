package zonecontroller

import (
	"fmt"
	"strconv"
	"strings"
)

type IncludeFlags struct {
	Basedata bool `json:"basedata"`
	Loot     bool `json:"loot"`
	Items    bool `json:"items"`
	Custom   bool `json:"custom"`
	Ignore   bool `json:"ignore"`
	Depop    bool `json:"depop"`
	Info     bool `json:"info"`
}

type CreateRequest struct {
	ZoneIDs   []int        `json:"zoneIds"`
	Source    string       `json:"source"`
	CloneFrom int          `json:"cloneFrom"`
	Include   IncludeFlags `json:"include"`
	Name      string       `json:"name"`
	Objective string       `json:"objective"`
	Tip       string       `json:"tip"`
	Respawn   string       `json:"respawn"`
	Overwrite bool         `json:"overwrite"`
	DryRun    *bool        `json:"dryRun"`
}

type TierRequest struct {
	ZoneIDs  []int             `json:"zoneIds"`
	FromZone int               `json:"fromZone"`
	Trash    map[string]string `json:"trash"`
	Boss     map[string]string `json:"boss"`
	Raid     map[string]string `json:"raid"`
	CopyLoot bool              `json:"copyLoot"`
	DryRun   *bool             `json:"dryRun"`
}

type MobsRequest struct {
	ZoneIDs []int    `json:"zoneIds"`
	Type    string   `json:"type"`
	Names   []string `json:"names"`
	Static  string   `json:"static"`
	DryRun  *bool    `json:"dryRun"`
}

func defaultCreateInclude(source string) IncludeFlags {
	if source == "clone" {
		return IncludeFlags{Basedata: true, Loot: true, Items: true, Custom: true, Ignore: true, Depop: true, Info: true}
	}
	return IncludeFlags{Basedata: true, Loot: true, Items: true, Custom: false, Ignore: true, Depop: true, Info: true}
}

func (s *Service) CreateZones(req CreateRequest) WritePlan {
	plan := WritePlan{DryRun: dryRunValue(req.DryRun)}
	ids, err := normalizeZoneIDs(req.ZoneIDs)
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	source := strings.ToLower(strings.TrimSpace(req.Source))
	if source == "" {
		source = "template"
	}
	include := req.Include
	if source == "clone" && req.CloneFrom <= 0 {
		plan.Error = "cloneFrom is required when source is clone"
		return plan
	}
	if !include.Basedata && !include.Loot && !include.Items && !include.Custom && !include.Ignore && !include.Depop && !include.Info {
		include = defaultCreateInclude(source)
	}

	stamp := backupStamp()
	for _, id := range ids {
		if err := s.createOne(id, source, req, include, stamp, &plan); err != nil {
			plan.Error = err.Error()
			return plan
		}
	}
	plan.OK = plan.Error == ""
	return plan
}

func (s *Service) createOne(id int, source string, req CreateRequest, include IncludeFlags, stamp string, plan *WritePlan) error {
	var mobSrc, lootSrc, itemSrc map[string]interface{}
	var err error
	if source == "clone" {
		mobSrc, err = s.readKindRaw(req.CloneFrom, "mob")
		if err != nil {
			return fmt.Errorf("clone source %d mob: %w", req.CloneFrom, err)
		}
		if include.Loot {
			lootSrc, _ = s.readKindRaw(req.CloneFrom, "loot")
		}
		if include.Items {
			itemSrc, _ = s.readKindRaw(req.CloneFrom, "item")
		}
	} else {
		mobSrc, err = s.readTemplateRaw("mob")
		if err != nil {
			return fmt.Errorf("template mob: %w", err)
		}
		if include.Loot {
			lootSrc, _ = s.readTemplateRaw("loot")
		}
		if include.Items {
			itemSrc, _ = s.readTemplateRaw("item")
		}
	}

	mob := remapTopKey(mobSrc, id)
	zone := pickZoneMap(mob, id)
	if !include.Custom {
		zone["custom"] = map[string]interface{}{}
	}
	if !include.Depop {
		zone["depop"] = map[string]interface{}{"all": []interface{}{}}
	}
	if include.Ignore {
		ensureIgnoreController(zone)
	}
	setInfoFields(zone, req.Name, req.Objective, req.Tip, req.Respawn)
	if !include.Basedata && source == "clone" {
		if tmpl, err := s.readTemplateRaw("mob"); err == nil {
			tmplZone := pickZoneMap(remapTopKey(tmpl, id), id)
			zone["basedata"] = deepCopy(tmplZone["basedata"])
		}
	}
	mob[strconv.Itoa(id)] = zone

	if err := s.commitKind(id, "mob", mob, req.Overwrite, plan.DryRun, stamp, plan); err != nil {
		return err
	}
	if include.Loot {
		if lootSrc == nil {
			lootSrc = map[string]interface{}{"TEMPZONEID": map[string]interface{}{}}
		}
		if err := s.commitKind(id, "loot", lootSrc, req.Overwrite, plan.DryRun, stamp, plan); err != nil {
			return err
		}
	}
	if include.Items {
		if itemSrc == nil {
			itemSrc = map[string]interface{}{"TEMPZONEID": map[string]interface{}{}}
		}
		if err := s.commitKind(id, "item", itemSrc, req.Overwrite, plan.DryRun, stamp, plan); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ApplyTier(req TierRequest) WritePlan {
	plan := WritePlan{DryRun: dryRunValue(req.DryRun)}
	ids, err := normalizeZoneIDs(req.ZoneIDs)
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	var sourceBasedata interface{}
	var sourceLoot, sourceItem map[string]interface{}
	if req.FromZone > 0 {
		raw, err := s.readKindRaw(req.FromZone, "mob")
		if err != nil {
			plan.Error = fmt.Sprintf("source zone %d: %v", req.FromZone, err)
			return plan
		}
		sourceBasedata = pickZoneMap(raw, req.FromZone)["basedata"]
		if req.CopyLoot {
			sourceLoot, _ = s.readKindRaw(req.FromZone, "loot")
			sourceItem, _ = s.readKindRaw(req.FromZone, "item")
		}
	}
	stamp := backupStamp()
	for _, id := range ids {
		raw, err := s.readKindRaw(id, "mob")
		if err != nil {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d has no mob json; use Create first", id))
			continue
		}
		mob := remapTopKey(raw, id)
		zone := pickZoneMap(mob, id)
		if sourceBasedata != nil {
			zone["basedata"] = deepCopy(sourceBasedata)
		}
		patchBasedataFields(zone, "trash", req.Trash)
		patchBasedataFields(zone, "boss", req.Boss)
		patchBasedataFields(zone, "raid", req.Raid)
		mob[strconv.Itoa(id)] = zone
		if err := s.commitKind(id, "mob", mob, true, plan.DryRun, stamp, &plan); err != nil {
			plan.Error = err.Error()
			return plan
		}
		if req.CopyLoot && sourceLoot != nil {
			if err := s.commitKind(id, "loot", sourceLoot, true, plan.DryRun, stamp, &plan); err != nil {
				plan.Error = err.Error()
				return plan
			}
		}
		if req.CopyLoot && sourceItem != nil {
			if err := s.commitKind(id, "item", sourceItem, true, plan.DryRun, stamp, &plan); err != nil {
				plan.Error = err.Error()
				return plan
			}
		}
	}
	plan.OK = plan.Error == ""
	return plan
}

func (s *Service) AddMobs(req MobsRequest) WritePlan {
	plan := WritePlan{DryRun: dryRunValue(req.DryRun)}
	ids, err := normalizeZoneIDs(req.ZoneIDs)
	if err != nil {
		plan.Error = err.Error()
		return plan
	}
	names := make([]string, 0, len(req.Names))
	seen := map[string]bool{}
	for _, n := range req.Names {
		n = strings.TrimSpace(n)
		n = strings.ReplaceAll(n, "_", " ")
		if n == "" || seen[strings.ToLower(n)] {
			continue
		}
		seen[strings.ToLower(n)] = true
		names = append(names, n)
	}
	if len(names) == 0 {
		plan.Error = "no mob names"
		return plan
	}
	stamp := backupStamp()
	for _, id := range ids {
		raw, err := s.readKindRaw(id, "mob")
		if err != nil {
			plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d has no mob json; use Create first", id))
			continue
		}
		mob := remapTopKey(raw, id)
		zone := pickZoneMap(mob, id)
		custom := asObject(zone["custom"])
		added := 0
		for _, name := range names {
			if existing := asObject(custom[name]); existing["type"] != nil && asString(existing["type"]) != "" {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("zone %d already has %s", id, name))
				continue
			}
			entry := defaultCustomMob(req.Type)
			if req.Static == "1" {
				entry["static"] = "1"
			}
			custom[name] = entry
			added++
		}
		zone["custom"] = custom
		mob[strconv.Itoa(id)] = zone
		if added == 0 {
			plan.Writes = append(plan.Writes, WriteFile{ZoneID: id, Kind: "mob", RelPath: s.zoneRel(id, "mob"), Action: "skip"})
			continue
		}
		if err := s.commitKind(id, "mob", mob, true, plan.DryRun, stamp, &plan); err != nil {
			plan.Error = err.Error()
			return plan
		}
	}
	plan.OK = plan.Error == ""
	return plan
}

