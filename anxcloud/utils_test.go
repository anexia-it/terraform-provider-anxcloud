package anxcloud

import (
	"errors"
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

func TestListAllPages(t *testing.T) {
	testPages := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8},
		{},
	}

	actual, err := listAllPages(func(page int) ([]int, error) {
		return testPages[page-1], nil
	})

	if err != nil {
		t.Error(err)
	}

	expected := []int{1, 2, 3, 4, 5, 6, 7, 8}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("(-expected +actual):\n%s", diff)
	}

	var testErr = errors.New("test error")
	_, err = listAllPages(func(page int) ([]int, error) {
		return nil, testErr
	})

	if err != testErr {
		t.Errorf("expected err to be %s, got %s", testErr, err)
	}
}

func TestSliceSubstract(t *testing.T) {
	a := []string{"a", "b", "c", "d", "e"}
	b := []string{"d", "e", "f", "g", "h"}
	c := []string{}

	type testCase struct {
		actual   []string
		expected []string
	}

	testCases := []testCase{
		{sliceSubstract(a, b), []string{"a", "b", "c"}},
		{sliceSubstract(b, a), []string{"f", "g", "h"}},
		{sliceSubstract(a, c), a},
		{sliceSubstract(a, a), c},
		{sliceSubstract(c, a), c},
		{sliceSubstract(c, c), c},
	}

	for _, testCase := range testCases {
		if diff := cmp.Diff(testCase.actual, testCase.expected); diff != "" {
			t.Errorf("(-expected +actual):\n%s", diff)
		}
	}
}
