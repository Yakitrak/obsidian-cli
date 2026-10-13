package obsidian

import "encoding/json"

// vaultRegistryConfig preserves fields owned by Obsidian while the CLI changes
// vault registrations. The read-only config type need not model those fields.
type vaultRegistryConfig struct {
	ObsidianVaultConfig
	fields  map[string]json.RawMessage
	entries map[string]json.RawMessage
}

func (c *vaultRegistryConfig) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &c.ObsidianVaultConfig); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &c.fields); err != nil {
		return err
	}
	if raw, ok := c.fields["vaults"]; ok {
		return json.Unmarshal(raw, &c.entries)
	}
	return nil
}

func (c vaultRegistryConfig) MarshalJSON() ([]byte, error) {
	fields := make(map[string]json.RawMessage, len(c.fields)+1)
	for key, value := range c.fields {
		fields[key] = value
	}
	entries := make(map[string]json.RawMessage, len(c.Vaults))
	for id, vault := range c.Vaults {
		entry := make(map[string]json.RawMessage)
		if raw, ok := c.entries[id]; ok {
			if err := json.Unmarshal(raw, &entry); err != nil {
				return nil, err
			}
			if entry == nil {
				entry = make(map[string]json.RawMessage)
			}
		}
		path, err := json.Marshal(vault.Path)
		if err != nil {
			return nil, err
		}
		entry["path"] = path
		encoded, err := json.Marshal(entry)
		if err != nil {
			return nil, err
		}
		entries[id] = encoded
	}
	vaults, err := json.Marshal(entries)
	if err != nil {
		return nil, err
	}
	fields["vaults"] = vaults
	return json.Marshal(fields)
}
