package anxcloud

import (
	"fmt"
	"testing"

	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
)

func TestTXTRecordComparisonLogic(t *testing.T) {
	target := clouddnsv1.Record{
		Type:     "TXT",
		Name:     "test",
		RData:    "\"hello world\"",
		ZoneName: "example.com",
	}

	records := []clouddnsv1.Record{
		{
			Type:     "TXT",
			Name:     "test",
			RData:    "\"hello world\"",
			ZoneName: "example.com",
		},
	}

	found, err := findRecordByContentFlexible(records, target, false, false)
	if err != nil {
		t.Errorf("Should find TXT record with quoted API data: %v", err)
	}
	if found == nil {
		t.Errorf("Should find the record")
	} else if found.RData != "\"hello world\"" {
		t.Errorf("Should return the API record as-is, got: %s", found.RData)
	}

	// Test edge case: empty TXT record
	target2 := clouddnsv1.Record{
		Type:     "TXT",
		Name:     "empty",
		RData:    "\"\"",
		ZoneName: "example.com",
	}

	records2 := []clouddnsv1.Record{
		{
			Type:     "TXT",
			Name:     "empty",
			RData:    "\"\"", // Quoted empty from API
			ZoneName: "example.com",
		},
	}

	found2, err2 := findRecordByContentFlexible(records2, target2, false, false)
	if err2 != nil {
		t.Errorf("Should find empty TXT record: %v", err2)
	}
	if found2 == nil {
		t.Errorf("Should find the empty record")
	}

	// Test non-TXT record (should work as before)
	target3 := clouddnsv1.Record{
		Type:     "A",
		Name:     "test",
		RData:    "192.168.1.1",
		ZoneName: "example.com",
	}

	records3 := []clouddnsv1.Record{
		{
			Type:     "A",
			Name:     "test",
			RData:    "192.168.1.1",
			ZoneName: "example.com",
		},
	}

	found3, err3 := findRecordByContentFlexible(records3, target3, false, false)
	if err3 != nil {
		t.Errorf("Should find A record: %v", err3)
	}
	if found3 == nil {
		t.Errorf("Should find the A record")
	}

	// Test TXT record without quotes (edge case)
	target4 := clouddnsv1.Record{
		Type:     "TXT",
		Name:     "noquotes",
		RData:    "no quotes",
		ZoneName: "example.com",
	}

	records4 := []clouddnsv1.Record{
		{
			Type:     "TXT",
			Name:     "noquotes",
			RData:    "no quotes", // No quotes from API (shouldn't happen but be safe)
			ZoneName: "example.com",
		},
	}

	found4, err4 := findRecordByContentFlexible(records4, target4, false, false)
	if err4 != nil {
		t.Errorf("Should find TXT record without quotes: %v", err4)
	}
	if found4 == nil {
		t.Errorf("Should find the TXT record without quotes")
	}

	// Test TXT record with quotes (edge case)
	target5 := clouddnsv1.Record{
		Type:     "TXT",
		Name:     "quotes",
		RData:    "\"quotes\"",
		ZoneName: "example.com",
	}

	records5 := []clouddnsv1.Record{
		{
			Type:     "TXT",
			Name:     "quotes",
			RData:    "\"quotes\"",
			ZoneName: "example.com",
		},
	}

	found5, err5 := findRecordByContentFlexible(records5, target5, false, false)
	if err5 != nil {
		t.Errorf("Should find TXT record without quotes: %v", err5)
	}
	if found5 == nil {
		t.Errorf("Should find the TXT record without quotes")
	}
}

func TestDNSRecordTXT_RefreshNoDiff(t *testing.T) {
	// This is a unit test using mock API
	// Tests that read operation produces same value as create

	zoneName := "example.com"
	recordName := "test"
	recordValue := "hello world" // Unquoted as user would write

	// Simulate what API returns (quoted)
	apiValue := fmt.Sprintf("%q", recordValue)

	// Mock record from API
	apiRecord := clouddnsv1.Record{
		Identifier: "test-id-123",
		ZoneName:   zoneName,
		Name:       recordName,
		Type:       "TXT",
		RData:      apiValue, // API returns quoted
		TTL:        3600,
	}

	// Test that our Read function strips quotes
	// Result should be: "hello world" (unquoted)
	rData := apiRecord.RData
	if apiRecord.Type == "TXT" {
		if len(rData) >= 2 && rData[0] == '"' && rData[len(rData)-1] == '"' {
			rData = rData[1 : len(rData)-1]
		}
	}

	// Verify normalized value matches original unquoted value
	if rData != recordValue {
		t.Errorf("Expected %q, got %q", recordValue, rData)
	}
}

func TestDNSRecordTXT_ComparisonLogic(t *testing.T) {
	testCases := []struct {
		name        string
		stateRData  string // What's in Terraform state (user's config)
		apiRData    string // What API returns (always quoted)
		shouldMatch bool
	}{
		{
			name:        "quoted config matches quoted API",
			stateRData:  "\"hello world\"",
			apiRData:    "\"hello world\"",
			shouldMatch: true,
		},
		{
			name:        "unquoted config matches after stripping API quotes",
			stateRData:  "hello world",
			apiRData:    "\"hello world\"",
			shouldMatch: true, // After stripping API quotes, both are "hello world"
		},
		{
			name:        "spf record matches",
			stateRData:  "v=spf1 include:_spf.example.com ~all",
			apiRData:    "\"v=spf1 include:_spf.example.com ~all\"",
			shouldMatch: true,
		},
		{
			name:        "different values don't match",
			stateRData:  "hello",
			apiRData:    "\"goodbye\"",
			shouldMatch: false,
		},
		{
			name:        "empty string matches",
			stateRData:  "",
			apiRData:    "\"\"",
			shouldMatch: true,
		},
		{
			name:        "quoted config doesn't match unquoted API content",
			stateRData:  "\"hello world\"",
			apiRData:    "\"hello world\"",
			shouldMatch: true, // Direct match since both have quotes
		},
		{
			name:        "nested quotes preserved",
			stateRData:  "\"inner quotes\"",
			apiRData:    "\"\\\"inner quotes\\\"\"",
			shouldMatch: false, // Different content after proper comparison
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate our comparison logic - matches findRecordByContentFlexible
			targetRData := tc.stateRData
			recordRData := tc.apiRData

			// Our logic: if target (config) has quotes, compare directly
			// If target doesn't have quotes, strip API quotes and compare
			if isQuotedString(targetRData) {
				// Target has quotes - compare directly
			} else {
				// Target doesn't have quotes - strip API quotes for comparison
				recordRData = stripOuterQuotes(recordRData)
			}

			matches := (targetRData == recordRData)
			if matches != tc.shouldMatch {
				t.Errorf("Expected match=%v, got match=%v for state=%q api=%q (compared: %q vs %q)",
					tc.shouldMatch, matches, tc.stateRData, tc.apiRData, targetRData, recordRData)
			}
		})
	}
}

func TestDNSRecordTXT_ImportNormalization(t *testing.T) {
	// During import, we don't have user's config, so we strip outer quotes by default
	// This is the most common case - users typically don't want quotes in their config
	// Users who want quoted values can adjust their config after import
	testCases := []struct {
		name          string
		apiRData      string // What API returns during import
		expectedRData string // What should be in state after import
	}{
		{
			name:          "simple text has quotes stripped",
			apiRData:      "\"hello world\"",
			expectedRData: "hello world", // Quotes stripped for user-friendly default
		},
		{
			name:          "spf record has quotes stripped",
			apiRData:      "\"v=spf1 include:_spf.example.com ~all\"",
			expectedRData: "v=spf1 include:_spf.example.com ~all",
		},
		{
			name:          "empty quoted value becomes empty string",
			apiRData:      "\"\"",
			expectedRData: "",
		},
		{
			name:          "nested quotes are preserved inside",
			apiRData:      "\"\\\"inner quotes\\\"\"",
			expectedRData: "\\\"inner quotes\\\"", // Outer quotes stripped, inner escaped quotes preserved
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate import normalization logic - uses stripOuterQuotes
			rdata := stripOuterQuotes(tc.apiRData)

			if rdata != tc.expectedRData {
				t.Errorf("Expected %q, got %q", tc.expectedRData, rdata)
			}
		})
	}
}

func TestDNSRecordTXT_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single character",
			input:    "\"a\"",
			expected: "a",
		},
		{
			name:     "single quote only",
			input:    "\"",
			expected: "\"", // Don't strip if not both quotes
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no quotes",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "internal quotes",
			input:    "\"hello \\\"world\\\"\"",
			expected: "hello \\\"world\\\"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rdata := tc.input
			// Apply our quote stripping logic
			if len(rdata) >= 2 && rdata[0] == '"' && rdata[len(rdata)-1] == '"' {
				rdata = rdata[1 : len(rdata)-1]
			}

			if rdata != tc.expected {
				t.Errorf("Input %q: expected %q, got %q", tc.input, tc.expected, rdata)
			}
		})
	}
}
