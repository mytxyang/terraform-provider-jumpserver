package jumpserver

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func setPilotList(d *schema.ResourceData, data map[string]interface{}, key string) error {
	raw, ok := data[key]
	if !ok {
		return nil
	}
	if raw == nil {
		return d.Set(key, []string{})
	}
	items, ok := raw.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected API list field %s", key)
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		if v, ok := item.(string); ok {
			values = append(values, v)
			continue
		}
		obj, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unexpected API item in %s", key)
		}
		attr := "id"
		if key == "actions" {
			attr = "value"
		}
		v, ok := obj[attr].(string)
		if !ok {
			return fmt.Errorf("missing string %s in %s", attr, key)
		}
		values = append(values, v)
	}
	return d.Set(key, values)
}
