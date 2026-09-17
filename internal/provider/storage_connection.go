package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func storageFlatTypes() map[string]attr.Type {
	r := map[string]attr.Type{}
	for k, a := range storageFieldCatalog() {
		r[k] = a.GetType()
	}
	return r
}
func storageSelection(c types.Object) (string, int) {
	if c.IsNull() || c.IsUnknown() {
		return "", 0
	}
	selected, count := "", 0
	for kind, v := range c.Attributes() {
		if !v.IsNull() {
			selected = kind
			count++
		}
	}
	return selected, count
}

// Flatten only inside the request/readback mapper; the public schema stays typed.
func storageFlat(c types.Object) types.Object {
	if c.IsNull() {
		return types.ObjectNull(storageFlatTypes())
	}
	if c.IsUnknown() {
		return types.ObjectUnknown(storageFlatTypes())
	}
	if _, ok := c.Attributes()["provider"]; ok {
		return c
	}
	kind, count := storageSelection(c)
	if count != 1 {
		return types.ObjectNull(storageFlatTypes())
	}
	child := c.Attributes()[kind].(types.Object)
	if child.IsUnknown() {
		return types.ObjectUnknown(storageFlatTypes())
	}
	values := map[string]attr.Value{}
	for k, t := range storageFlatTypes() {
		values[k] = nullValue(t)
	}
	for k, v := range child.Attributes() {
		values[k] = v
	}
	values["provider"] = types.StringValue(kind)
	return types.ObjectValueMust(storageFlatTypes(), values)
}
func storageNested(flat types.Object) types.Object {
	if flat.IsNull() {
		return types.ObjectNull(storageConnectionTypes())
	}
	values := map[string]attr.Value{}
	for k, t := range storageConnectionTypes() {
		values[k] = nullValue(t)
	}
	kind := flat.Attributes()["provider"].(types.String).ValueString()
	fields := map[string]attr.Value{}
	fieldTypes := map[string]attr.Type{}
	for k, a := range storageVariantAttributes(kind) {
		fieldTypes[k] = a.GetType()
		fields[k] = flat.Attributes()[k]
	}
	values[kind] = types.ObjectValueMust(fieldTypes, fields)
	return types.ObjectValueMust(storageConnectionTypes(), values)
}
