package anxcloud

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSchemaWith(t *testing.T) {
	actual := schemaWith(testSchema(),
		fieldsExactlyOneOf("exactly_one_of_0", "exactly_one_of_1"),
		fieldsRequired("required"),
		fieldsOptional("optional"),
		fieldsComputed("computed"),
	)

	expected := testSchema()
	expectedExactlyOneOf := []string{"exactly_one_of_0", "exactly_one_of_1"}
	expected["exactly_one_of_0"].Optional = true
	expected["exactly_one_of_0"].ExactlyOneOf = expectedExactlyOneOf
	expected["exactly_one_of_1"].Optional = true
	expected["exactly_one_of_1"].ExactlyOneOf = expectedExactlyOneOf
	expected["required"].Required = true
	expected["optional"].Optional = true
	expected["computed"].Computed = true

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected +actual):\n%s", diff)
	}
}

func testSchema() schemaMap {
	return schemaMap{
		"exactly_one_of_0": {Description: "Exactly one of #0"},
		"exactly_one_of_1": {Description: "Exactly one of #1"},
		"required":         {Description: "Required Field"},
		"optional":         {Description: "Optional Field"},
		"computed":         {Description: "Computed Field"},
	}
}
