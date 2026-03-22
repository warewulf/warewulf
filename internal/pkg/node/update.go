package node

import (
	"reflect"
)

// UpdateFrom copies fields from src to dst, but only those fields whose
// corresponding cobra flag (identified by the "lopt" struct tag) reports
// as changed. This replaces the old pattern of marshaling to YAML (to strip
// zero values via omitempty) and then merging with mergo.
//
// The changed function should typically be cmd.Flags().Changed.
func (dst *Node) UpdateFrom(src *Node, changed func(string) bool) {
	recursiveUpdateFrom(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), changed)
}

// UpdateFrom copies fields from src to dst for profiles.
func (dst *Profile) UpdateFrom(src *Profile, changed func(string) bool) {
	recursiveUpdateFrom(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem(), changed)
}

func recursiveUpdateFrom(dst, src reflect.Value, changed func(string) bool) {
	srcType := src.Type()

	for i := 0; i < src.NumField(); i++ {
		field := srcType.Field(i)
		srcField := src.Field(i)
		dstField := dst.Field(i)

		if !field.IsExported() {
			continue
		}

		if lopt := field.Tag.Get("lopt"); lopt != "" && field.Tag.Get("comment") != "" {
			// Leaf field with a cobra flag — copy if the flag was changed
			if changed(lopt) {
				dstField.Set(srcField)
			}
		} else if field.Anonymous {
			// Embedded struct (e.g., Profile in Node)
			recursiveUpdateFrom(dstField, srcField, changed)
		} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			// Pointer-to-struct (e.g., *IpmiConf, *KernelConf)
			if srcField.IsNil() {
				continue
			}
			if dstField.IsNil() {
				dstField.Set(reflect.New(field.Type.Elem()))
			}
			recursiveUpdateFrom(dstField.Elem(), srcField.Elem(), changed)
		} else if field.Type.Kind() == reflect.Struct {
			// Direct struct
			recursiveUpdateFrom(dstField, srcField, changed)
		} else if field.Type.Kind() == reflect.Map {
			// Map fields (e.g., NetDevs, Disks, FileSystems)
			switch field.Type.Elem().Kind() {
			case reflect.String, reflect.Interface:
				// Tags map[string]string — handled separately by tag add/del operations
				continue
			case reflect.Pointer:
				// map[string]*Struct — update entries from src into dst
				if srcField.IsNil() {
					continue
				}
				if dstField.IsNil() {
					dstField.Set(reflect.MakeMap(field.Type))
				}
				for _, key := range srcField.MapKeys() {
					srcEntry := srcField.MapIndex(key)
					if srcEntry.IsNil() {
						continue
					}
					dstEntry := dstField.MapIndex(key)
					if !dstEntry.IsValid() || dstEntry.IsNil() {
						// Create new entry in dst
						dstEntry = reflect.New(field.Type.Elem().Elem())
						dstField.SetMapIndex(key, dstEntry)
					}
					// Map values are not addressable, so we need to work with a copy
					tmpDst := reflect.New(field.Type.Elem().Elem()).Elem()
					tmpDst.Set(dstEntry.Elem())
					recursiveUpdateFrom(tmpDst, srcEntry.Elem(), changed)
					dstField.SetMapIndex(key, tmpDst.Addr())
				}
			}
		}
	}
}
