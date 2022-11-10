package cfcollections

import (
	cloudflare "github.com/cloudflare/cloudflare-go"
)

type AccessGroupCollection []cloudflare.AccessGroup

func (c AccessGroupCollection) Len() int { return len(c) }

func (c AccessGroupCollection) GetByName(name string) *cloudflare.AccessGroup {
	for _, policy := range c {
		if policy.Name == name {
			return &policy
		}
	}

	return nil
}
