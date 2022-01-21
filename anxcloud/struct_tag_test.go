package anxcloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.anx.io/go-anxcloud/pkg/core/tags"
)

// expanders tests

// flatteners tests

func TestFlattenTags(t *testing.T) {
	cases := []struct {
		Input          []tags.Summary
		ExpectedOutput []interface{}
	}{
		{
			[]tags.Summary{
				{
					Name:       "tag-1",
					Identifier: "identifier-1",
				},
				{
					Name:       "tag-2",
					Identifier: "identifier-2",
				},
			},
			[]interface{}{
				map[string]interface{}{
					"name":       "tag-1",
					"identifier": "identifier-1",
				},
				map[string]interface{}{
					"name":       "tag-2",
					"identifier": "identifier-2",
				},
			},
		},
		{
			[]tags.Summary{},
			[]interface{}{},
		},
	}

	for _, tc := range cases {
		output := flattenTags(tc.Input)
		if diff := cmp.Diff(tc.ExpectedOutput, output); diff != "" {
			t.Fatalf("Unexpected output from expander: mismatch (-want +got):\n%s", diff)
		}
	}
}
