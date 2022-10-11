package anxcloud

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func updateKeys[K comparable, V any](m map[K]V, updater func(V), keys ...K) {
	for _, key := range keys {
		updater(m[key])
	}
}

type schemaMap map[string]*schema.Schema

func schemaWith(s schemaMap, modifiers ...func(schemaMap)) schemaMap {
	for _, modifier := range modifiers {
		modifier(s)
	}
	return s
}

func fieldsExactlyOneOf(fields ...string) func(schemaMap) {
	return func(m schemaMap) {
		fieldsOptional(fields...)(m)
		updateKeys(m, func(field *schema.Schema) {
			field.ExactlyOneOf = fields
		}, fields...)
	}
}

func fieldsRequired(fields ...string) func(schemaMap) {
	return func(m schemaMap) {
		updateKeys(m, func(field *schema.Schema) {
			field.Required = true
			field.Optional = false
			field.Computed = false
		}, fields...)
	}
}

func fieldsOptional(fields ...string) func(schemaMap) {
	return func(m schemaMap) {
		updateKeys(m, func(field *schema.Schema) {
			field.Required = false
			field.Optional = true
			field.Computed = false
		}, fields...)
	}
}

func fieldsComputed(fields ...string) func(schemaMap) {
	return func(m schemaMap) {
		updateKeys(m, func(field *schema.Schema) {
			field.Required = false
			field.Optional = false
			field.Computed = true
			field.Default = nil
		}, fields...)
	}
}

func listAllPages[T any](pageRetriever func(int) ([]T, error)) ([]T, error) {
	var allPages = make([]T, 0)

	for page := 1; true; page++ {
		singlePage, err := pageRetriever(page)
		if err != nil {
			return nil, err
		}

		if len(singlePage) == 0 {
			break
		}

		allPages = append(allPages, singlePage...)
	}

	return allPages, nil
}

func sliceSubstract[T comparable](a, b []T) []T {
	out := make([]T, 0, len(a))
outer:
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				continue outer
			}
		}
		out = append(out, a[i])
	}
	return out
}

func mustCastInterfaceArray[T any](in []interface{}) []T {
	out := make([]T, 0, len(in))
	for _, v := range in {
		out = append(out, v.(T))
	}
	return out
}

var kubernetesResourceNameRegexp = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]{0,38}[a-z0-9])?$`)

func validateKubernetesResourceName(val any, key string) (warns []string, errs []error) {
	v := val.(string)
	if !kubernetesResourceNameRegexp.Match([]byte(v)) {
		errs = append(errs, fmt.Errorf("%q isn't a valid name for Anexia Kubernetes Service resources: has to be a lowercase RFC 1123 hostname but with a maximum length of 40 characters", v))
	}
	return
}
