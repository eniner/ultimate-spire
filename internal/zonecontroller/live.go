package zonecontroller

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type LiveZone struct {
	ZoneID       int    `json:"zoneId"`
	Short        string `json:"short,omitempty"`
	Long         string `json:"long,omitempty"`
	Popped       bool   `json:"popped"`
	Players      int    `json:"players"`
	InstanceID   int    `json:"instanceId,omitempty"`
	Static       bool   `json:"static"`
	ZoneServerID int    `json:"zoneServerId,omitempty"`
}

type LiveStatus struct {
	OK        bool       `json:"ok"`
	WorldOK   bool       `json:"worldOk"`
	Error     string     `json:"error,omitempty"`
	WorldNote string     `json:"worldNote,omitempty"`
	Zones     []LiveZone `json:"zones"`
}

func (s *Service) LiveStatus(ids []int) LiveStatus {
	out := LiveStatus{Zones: []LiveZone{}}
	if s.world == nil {
		out.Error = "World telnet is not wired"
		out.WorldNote = "Spire cannot see popped zones until World is accepting telnet on 127.0.0.1:9000."
		for _, id := range ids {
			out.Zones = append(out.Zones, s.offlineLiveZone(id))
		}
		return out
	}
	list, err := s.world.GetZoneList()
	if err != nil {
		out.Error = "World telnet: " + err.Error()
		out.WorldNote = "World is down or telnet is disabled. Command files can still be queued."
		for _, id := range ids {
			out.Zones = append(out.Zones, s.offlineLiveZone(id))
		}
		return out
	}
	out.OK = true
	out.WorldOK = true
	popped := map[int]LiveZone{}
	for _, z := range list.Data {
		row := LiveZone{
			ZoneID:       z.ZoneID,
			Short:        strings.TrimSpace(z.ZoneName),
			Long:         strings.TrimSpace(z.ZoneLongName),
			Popped:       true,
			Players:      z.NumberPlayers,
			InstanceID:   z.InstanceID,
			Static:       z.IsStaticZone,
			ZoneServerID: z.ID,
		}
		if prev, ok := popped[z.ZoneID]; ok && prev.InstanceID == 0 && z.InstanceID != 0 {
			continue
		}
		popped[z.ZoneID] = row
	}
	if len(ids) == 0 {
		for _, z := range popped {
			out.Zones = append(out.Zones, z)
		}
		sort.Slice(out.Zones, func(i, j int) bool { return out.Zones[i].ZoneID < out.Zones[j].ZoneID })
		return out
	}
	for _, id := range ids {
		if z, ok := popped[id]; ok {
			out.Zones = append(out.Zones, z)
			continue
		}
		out.Zones = append(out.Zones, s.offlineLiveZone(id))
	}
	return out
}

func (s *Service) offlineLiveZone(id int) LiveZone {
	row := LiveZone{ZoneID: id}
	if info, err := s.zoneInfo(id); err == nil {
		row.Short = info.short
		row.Long = info.long
	}
	return row
}

func (s *Service) rebootZone(live LiveZone, short string) error {
	if s.world == nil {
		return fmt.Errorf("world telnet is not wired")
	}
	target := strings.TrimSpace(short)
	if live.ZoneServerID > 0 {
		target = strconv.Itoa(live.ZoneServerID)
	}
	if target == "" {
		target = strings.TrimSpace(live.Short)
	}
	if target == "" {
		return fmt.Errorf("empty zone")
	}
	if err := s.world.ZoneShutdown(target); err != nil {
		if short != "" && target != short {
			if err2 := s.world.ZoneShutdown(short); err2 != nil {
				return err
			}
		} else {
			return err
		}
	}
	time.Sleep(1500 * time.Millisecond)
	if short == "" {
		short = strings.TrimSpace(live.Short)
	}
	if short == "" {
		return nil
	}
	if err := s.world.ZoneBootup(short); err != nil {
		again := s.LiveStatus([]int{live.ZoneID})
		for _, z := range again.Zones {
			if z.ZoneID == live.ZoneID && z.Popped {
				return nil
			}
		}
		return err
	}
	return nil
}

func (s *Service) bootZone(short string) error {
	if s.world == nil {
		return fmt.Errorf("world telnet is not wired")
	}
	short = strings.TrimSpace(short)
	if short == "" {
		return fmt.Errorf("empty zone short name")
	}
	return s.world.ZoneBootup(short)
}
