// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package skills

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ubuntu-core/snappy/snap"
)

// Repository stores all known snappy skills and slots and types.
type Repository struct {
	// Protects the internals from concurrent access.
	m     sync.Mutex
	types map[string]Type
	// Indexed by [snapName][skillName]
	skills map[string]map[string]*Skill
	slots  map[string]map[string]*Slot
}

// NewRepository creates an empty skill repository.
func NewRepository() *Repository {
	return &Repository{
		types:  make(map[string]Type),
		skills: make(map[string]map[string]*Skill),
		slots:  make(map[string]map[string]*Slot),
	}
}

// Type returns a type with a given name.
func (r *Repository) Type(typeName string) Type {
	r.m.Lock()
	defer r.m.Unlock()

	return r.types[typeName]
}

// AddType adds the provided skill type to the repository.
func (r *Repository) AddType(t Type) error {
	r.m.Lock()
	defer r.m.Unlock()

	typeName := t.Name()
	if err := ValidateName(typeName); err != nil {
		return err
	}
	if _, ok := r.types[typeName]; ok {
		return fmt.Errorf("cannot add skill type: %q, type name is in use", typeName)
	}
	r.types[typeName] = t
	return nil
}

// AllSkills returns all skills of the given type.
// If skillType is the empty string, all skills are returned.
func (r *Repository) AllSkills(skillType string) []*Skill {
	r.m.Lock()
	defer r.m.Unlock()

	var result []*Skill
	for _, skillsForSnap := range r.skills {
		for _, skill := range skillsForSnap {
			if skillType == "" || skill.Type == skillType {
				result = append(result, skill)
			}
		}
	}
	sort.Sort(bySkillSnapAndName(result))
	return result
}

// Skills returns the skills offered by the named snap.
func (r *Repository) Skills(snapName string) []*Skill {
	r.m.Lock()
	defer r.m.Unlock()

	var result []*Skill
	for _, skill := range r.skills[snapName] {
		result = append(result, skill)
	}
	sort.Sort(bySkillSnapAndName(result))
	return result
}

// Skill returns the specified skill from the named snap.
func (r *Repository) Skill(snapName, skillName string) *Skill {
	r.m.Lock()
	defer r.m.Unlock()

	return r.skills[snapName][skillName]
}

// AddSkill adds a skill to the repository.
// Skill names must be valid snap names, as defined by ValidateName.
// Skill name must be unique within a particular snap.
func (r *Repository) AddSkill(skill *Skill) error {
	r.m.Lock()
	defer r.m.Unlock()

	// Reject snaps with invalid names
	if err := snap.ValidateName(skill.Snap); err != nil {
		return err
	}
	// Reject skill with invalid names
	if err := ValidateName(skill.Name); err != nil {
		return err
	}
	t := r.types[skill.Type]
	if t == nil {
		return fmt.Errorf("cannot add skill, skill type %q is not known", skill.Type)
	}
	// Reject skill that don't pass type-specific sanitization
	if err := t.Sanitize(skill); err != nil {
		return err
	}
	if _, ok := r.skills[skill.Snap][skill.Name]; ok {
		return fmt.Errorf("cannot add skill, snap %q already has skill %q", skill.Snap, skill.Name)
	}
	if r.skills[skill.Snap] == nil {
		r.skills[skill.Snap] = make(map[string]*Skill)
	}
	r.skills[skill.Snap][skill.Name] = skill
	return nil
}

// RemoveSkill removes the named skill provided by a given snap.
// The removed skill must exist and must not be used anywhere.
func (r *Repository) RemoveSkill(snapName, skillName string) error {
	r.m.Lock()
	defer r.m.Unlock()

	// TODO: Ensure that the skill is not used anywhere
	if _, ok := r.skills[snapName][skillName]; !ok {
		return fmt.Errorf("cannot remove skill %q from snap %q, no such skill", skillName, snapName)
	}
	delete(r.skills[snapName], skillName)
	if len(r.skills[snapName]) == 0 {
		delete(r.skills, snapName)
	}
	return nil
}

// AllSlots returns all skill slots of the given type.
// If skillType is the empty string, all skill slots are returned.
func (r *Repository) AllSlots(skillType string) []*Slot {
	r.m.Lock()
	defer r.m.Unlock()

	var result []*Slot
	for _, slotsForSnap := range r.slots {
		for _, slot := range slotsForSnap {
			if skillType == "" || slot.Type == skillType {
				result = append(result, slot)
			}
		}
	}
	sort.Sort(bySlotSnapAndName(result))
	return result
}

// Slots returns the skill slots offered by the named snap.
func (r *Repository) Slots(snapName string) []*Slot {
	r.m.Lock()
	defer r.m.Unlock()

	var result []*Slot
	for _, slot := range r.slots[snapName] {
		result = append(result, slot)
	}
	sort.Sort(bySlotSnapAndName(result))
	return result
}

// Slot returns the specified skill slot from the named snap.
func (r *Repository) Slot(snapName, slotName string) *Slot {
	r.m.Lock()
	defer r.m.Unlock()

	return r.slots[snapName][slotName]
}

// AddSlot adds a new slot to the repository.
// Adding a slot with invalid name returns an error.
// Adding a slot that has the same name and snap name as another slot returns an error.
func (r *Repository) AddSlot(slot *Slot) error {
	r.m.Lock()
	defer r.m.Unlock()

	// Reject snaps with invalid names
	if err := snap.ValidateName(slot.Snap); err != nil {
		return err
	}
	// Reject skill with invalid names
	if err := ValidateName(slot.Name); err != nil {
		return err
	}
	// TODO: ensure that apps are correct
	t := r.types[slot.Type]
	if t == nil {
		return fmt.Errorf("cannot add skill slot, skill type %q is not known", slot.Type)
	}
	if _, ok := r.slots[slot.Snap][slot.Name]; ok {
		return fmt.Errorf("cannot add skill slot, snap %q already has slot %q", slot.Snap, slot.Name)
	}
	if r.slots[slot.Snap] == nil {
		r.slots[slot.Snap] = make(map[string]*Slot)
	}
	r.slots[slot.Snap][slot.Name] = slot
	return nil
}

// RemoveSlot removes a named slot from the given snap.
// Removing a slot that doesn't exist returns an error.
// Removing a slot that uses a skill returns an error.
func (r *Repository) RemoveSlot(snapName, slotName string) error {
	r.m.Lock()
	defer r.m.Unlock()

	// TODO: return an error if slot is occupied by at least one capability.
	if _, ok := r.slots[snapName][slotName]; !ok {
		return fmt.Errorf("cannot remove skill slot %q from snap %q, no such slot", slotName, snapName)
	}
	delete(r.slots[snapName], slotName)
	if len(r.slots[snapName]) == 0 {
		delete(r.slots, snapName)
	}
	return nil
}

// Support for sort.Interface

type bySkillSnapAndName []*Skill

func (c bySkillSnapAndName) Len() int      { return len(c) }
func (c bySkillSnapAndName) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c bySkillSnapAndName) Less(i, j int) bool {
	if c[i].Snap != c[j].Snap {
		return c[i].Snap < c[j].Snap
	}
	return c[i].Name < c[j].Name
}

type bySlotSnapAndName []*Slot

func (c bySlotSnapAndName) Len() int      { return len(c) }
func (c bySlotSnapAndName) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c bySlotSnapAndName) Less(i, j int) bool {
	if c[i].Snap != c[j].Snap {
		return c[i].Snap < c[j].Snap
	}
	return c[i].Name < c[j].Name
}