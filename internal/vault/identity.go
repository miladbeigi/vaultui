package vault

import (
	"fmt"
	"sort"
)

// IdentityEntity represents an identity entity.
type IdentityEntity struct {
	ID       string
	Name     string
	Policies []string
}

// IdentityGroup represents an identity group.
type IdentityGroup struct {
	ID              string
	Name            string
	Policies        []string
	MemberEntityIDs []string
	Type            string
}

// ListIdentityEntities lists entities via the identity/entity/id LIST endpoint.
func (c *Client) ListIdentityEntities(mount string) ([]IdentityEntity, error) {
	cacheKey := "identity:entities"
	if mount != "" {
		cacheKey = "identity:entities:" + mount
	}
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]IdentityEntity), nil
	}

	secret, err := c.raw.Logical().List("identity/entity/id")
	if err != nil {
		return nil, fmt.Errorf("listing identity entities: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return []IdentityEntity{}, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []IdentityEntity{}, nil
	}

	var entities []IdentityEntity
	for _, k := range keysRaw {
		id, ok := k.(string)
		if !ok {
			continue
		}
		entity, err := c.ReadIdentityEntity(id)
		if err != nil {
			continue // skip entities we can't read
		}
		if entity != nil {
			entities = append(entities, *entity)
		}
	}

	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Name < entities[j].Name
	})

	c.cache.Set(cacheKey, entities)
	return entities, nil
}

// ListIdentityGroups lists groups via the identity/group/id LIST endpoint.
func (c *Client) ListIdentityGroups() ([]IdentityGroup, error) {
	cacheKey := "identity:groups"
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.([]IdentityGroup), nil
	}

	secret, err := c.raw.Logical().List("identity/group/id")
	if err != nil {
		return nil, fmt.Errorf("listing identity groups: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return []IdentityGroup{}, nil
	}

	keysRaw, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []IdentityGroup{}, nil
	}

	var groups []IdentityGroup
	for _, k := range keysRaw {
		id, ok := k.(string)
		if !ok {
			continue
		}
		group, err := c.ReadIdentityGroup(id)
		if err != nil {
			continue
		}
		if group != nil {
			groups = append(groups, *group)
		}
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	c.cache.Set(cacheKey, groups)
	return groups, nil
}

// ReadIdentityEntity reads entity detail by ID.
func (c *Client) ReadIdentityEntity(id string) (*IdentityEntity, error) {
	cacheKey := "identity:entity:" + id
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.(*IdentityEntity), nil
	}

	secret, err := c.raw.Logical().Read("identity/entity/id/" + id)
	if err != nil {
		return nil, fmt.Errorf("reading identity entity %q: %w", id, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	entity := &IdentityEntity{ID: id}
	if v, ok := secret.Data["name"].(string); ok {
		entity.Name = v
	}
	if v, ok := secret.Data["policies"].([]interface{}); ok {
		for _, p := range v {
			if s, ok := p.(string); ok {
				entity.Policies = append(entity.Policies, s)
			}
		}
	}

	c.cache.Set(cacheKey, entity)
	return entity, nil
}

// ReadIdentityGroup reads group detail by ID.
func (c *Client) ReadIdentityGroup(id string) (*IdentityGroup, error) {
	cacheKey := "identity:group:" + id
	if v, ok := c.cache.Get(cacheKey); ok {
		return v.(*IdentityGroup), nil
	}

	secret, err := c.raw.Logical().Read("identity/group/id/" + id)
	if err != nil {
		return nil, fmt.Errorf("reading identity group %q: %w", id, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	group := &IdentityGroup{ID: id}
	if v, ok := secret.Data["name"].(string); ok {
		group.Name = v
	}
	if v, ok := secret.Data["type"].(string); ok {
		group.Type = v
	}
	if v, ok := secret.Data["policies"].([]interface{}); ok {
		for _, p := range v {
			if s, ok := p.(string); ok {
				group.Policies = append(group.Policies, s)
			}
		}
	}
	if v, ok := secret.Data["member_entity_ids"].([]interface{}); ok {
		for _, m := range v {
			if s, ok := m.(string); ok {
				group.MemberEntityIDs = append(group.MemberEntityIDs, s)
			}
		}
	}

	c.cache.Set(cacheKey, group)
	return group, nil
}
