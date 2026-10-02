package zonecontroller

import (
	"fmt"
	"strconv"
	"strings"
)

type ApplyRequest struct {
	ZoneIDs       []int    `json:"zoneIds"`
	ReloadQuests  bool     `json:"reloadQuests"`
	ReloadWorld   bool     `json:"reloadWorld"`
	RebootZones   bool     `json:"rebootZones"`
	RebootEmpty   *bool    `json:"rebootEmpty"`
	BootIfDown    bool     `json:"bootIfDown"`
	QueueCommands []string `json:"queueCommands"`
}

type ApplyResult struct {
	OK        bool            `json:"ok"`
	Error     string          `json:"error,omitempty"`
	Commands  []string        `json:"commands"`
	Reloaded  []string        `json:"reloaded"`
	Rebooted  []string        `json:"rebooted"`
	Queued    []QueuedCommand `json:"queued,omitempty"`
	Live      []LiveZone      `json:"live,omitempty"`
	Warnings  []string        `json:"warnings,omitempty"`
	WorldNote string          `json:"worldNote,omitempty"`
	WorldOK   bool            `json:"worldOk"`
}

func rebootEmptyValue(v *bool) bool {
	if v == nil {
		return true
	}
	return *v
}

func (s *Service) ApplyLive(req ApplyRequest) ApplyResult {
	out := ApplyResult{Commands: []string{}, Reloaded: []string{}, Rebooted: []string{}}
	ids := req.ZoneIDs
	if len(ids) == 0 {
		if zones, err := s.ListZones(); err == nil {
			for _, z := range zones {
				ids = append(ids, z.ZoneID)
			}
		}
	}
	clean, err := normalizeZoneIDs(ids)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	cmds, err := normalizeCommands(req.QueueCommands)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	for _, id := range clean {
		out.Commands = append(out.Commands, "!initdata "+strconv.Itoa(id))
	}
	live := s.LiveStatus(clean)
	out.Live = live.Zones
	out.WorldOK = live.WorldOK
	if live.Error != "" {
		out.Warnings = append(out.Warnings, live.Error)
	}

	needsWorld := req.ReloadQuests || req.ReloadWorld || req.RebootZones || req.BootIfDown || (rebootEmptyValue(req.RebootEmpty) && len(cmds) > 0)
	if len(cmds) > 0 {
		queued, warns := s.queueCommands(clean, cmds)
		out.Queued = queued
		out.Warnings = append(out.Warnings, warns...)
	}
	if !needsWorld && len(cmds) == 0 {
		out.OK = true
		out.WorldNote = "Copy the commands in-game, or choose a live action."
		return out
	}
	if s.world == nil {
		out.OK = true
		out.WorldNote = "World telnet is not wired. Command files were queued if selected. Use the !initdata list if a zone is already popped."
		return out
	}
	if req.ReloadWorld {
		if _, err := s.world.Reload("quests"); err != nil {
			out.Warnings = append(out.Warnings, "world quest reload: "+err.Error())
		} else {
			out.Reloaded = append(out.Reloaded, "world:quests")
		}
	}
	if req.ReloadQuests {
		for _, id := range clean {
			info, err := s.zoneInfo(id)
			if err != nil {
				out.Warnings = append(out.Warnings, fmt.Sprintf("zone %d: %v", id, err))
				continue
			}
			if err := s.world.ReloadQuestsForZone(info.short); err != nil {
				out.Warnings = append(out.Warnings, fmt.Sprintf("%s quest reload: %v", info.short, err))
				continue
			}
			out.Reloaded = append(out.Reloaded, info.short)
		}
	}

	byID := map[int]LiveZone{}
	for _, z := range live.Zones {
		byID[z.ZoneID] = z
	}
	for _, id := range clean {
		info, err := s.zoneInfo(id)
		if err != nil {
			out.Warnings = append(out.Warnings, fmt.Sprintf("zone %d: %v", id, err))
			continue
		}
		z := byID[id]
		if z.Short == "" {
			z.Short = info.short
		}
		rebootOccupied := req.RebootZones && z.Popped
		rebootEmpty := rebootEmptyValue(req.RebootEmpty) && z.Popped && z.Players == 0 && (len(cmds) > 0 || req.ReloadQuests)
		if rebootOccupied || rebootEmpty {
			if err := s.rebootZone(z, info.short); err != nil {
				out.Warnings = append(out.Warnings, fmt.Sprintf("%s reboot: %v", info.short, err))
				continue
			}
			out.Rebooted = append(out.Rebooted, info.short)
			continue
		}
		if req.BootIfDown && !z.Popped {
			if err := s.bootZone(info.short); err != nil {
				out.Warnings = append(out.Warnings, fmt.Sprintf("%s boot: %v", info.short, err))
				continue
			}
			out.Rebooted = append(out.Rebooted, info.short+" (boot)")
		}
	}

	out.OK = true
	out.WorldNote = applyNote(out, cmds, live.WorldOK)
	return out
}

func applyNote(out ApplyResult, cmds []string, worldOK bool) string {
	parts := []string{}
	if len(out.Reloaded) > 0 {
		parts = append(parts, "Quest scripts reloaded.")
	}
	if len(out.Queued) > 0 {
		parts = append(parts, "Zone controller commands queued for "+strings.Join(cmds, ", ")+".")
		if worldOK {
			parts = append(parts, "A popped controller runs them on its next tick; empty popped zones were rebooted when that option is on.")
		}
	}
	if len(out.Rebooted) > 0 {
		parts = append(parts, "World reboot/boot sent.")
	}
	if len(parts) == 0 {
		if !worldOK {
			return "World did not accept telnet. Use the !initdata list in a popped zone."
		}
		return "No live world action ran."
	}
	return strings.Join(parts, " ")
}

func applyCommands(ids []int) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, "!initdata "+strconv.Itoa(id))
	}
	return out
}
