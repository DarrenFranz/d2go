package memory

import (
    "github.com/hectorgimenez/d2go/pkg/data"
    "github.com/hectorgimenez/d2go/pkg/data/area"
)

// NOTE: This implementation is based on the upstream d2go getRoster logic,
// extended to fill data.RosterMember.Relation and IsHostileToMe. The actual
// hostile/party bitmasks depend on current D2R roster structure and may need
// adjustment if offsets change.

// These masks are intentionally conservative defaults; adjust them to match your findings.
const (
    rosterFlagParty   = 0x02 // example: in-party bit (placeholder, verify)
    rosterFlagHostile = 0x04 // example: hostile-to-me bit (placeholder, verify)
)

func (gd *GameReader) getRoster(rawPlayerUnits RawPlayerUnits) (roster []data.RosterMember) {
    partyStruct := uintptr(gd.Process.ReadUInt(gd.Process.moduleBaseAddressPtr+gd.offset.RosterOffset, Uint64))

    // We skip the first position because it's the main player, and we already have the information
    // (+0x148 is the next party/roster entry)
    partyStruct = uintptr(gd.Process.ReadUInt(partyStruct+0x148, Uint64))

    for partyStruct > 0 {
        name := gd.Process.ReadStringFromMemory(partyStruct, 16)
        a := area.ID(gd.Process.ReadUInt(partyStruct+0x5C, Uint32))
        xPos := int(gd.Process.ReadUInt(partyStruct+0x60, Uint32))
        yPos := int(gd.Process.ReadUInt(partyStruct+0x64, Uint32))

        // Suggested: read flags from the roster entry (verify this offset & values against live data)
        flags := uint32(gd.Process.ReadUInt(partyStruct+0x90, Uint32)) // TODO: confirm correct offset

        rel := data.RelationNeutral
        if flags&rosterFlagParty != 0 {
            rel |= data.RelationParty
        }
        if flags&rosterFlagHostile != 0 {
            rel |= data.RelationHostile
        }
        if rel == 0 {
            rel = data.RelationNeutral
        }

        // When the player is in town, roster data is not updated, so we need to get
        // the area from the player unit that matches the same name.
        for _, pu := range rawPlayerUnits {
            if pu.Name == name {
                xPos = pu.Position.X
                yPos = pu.Position.Y
                a = pu.Area
                break
            }
        }

        roster = append(roster, data.RosterMember{
            Name:          name,
            Area:          a,
            Position:      data.Position{X: xPos, Y: yPos},
            Relation:      rel,
            IsHostileToMe: rel&data.RelationHostile != 0,
        })

        partyStruct = uintptr(gd.Process.ReadUInt(partyStruct+0x148, Uint64))
    }

    mainPlayerUnit := rawPlayerUnits.GetMainPlayer()
    // Main player is first entry; never hostile to self.
    selfMember := data.RosterMember{
        Name:          mainPlayerUnit.Name,
        Area:          mainPlayerUnit.Area,
        Position:      mainPlayerUnit.Position,
        Relation:      data.RelationNeutral,
        IsHostileToMe: false,
    }

    return append([]data.RosterMember{selfMember}, roster...)
}
