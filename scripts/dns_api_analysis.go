package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/api/types"
	clouddnsv1 "go.anx.io/go-anxcloud/pkg/apis/clouddns/v1"
	"go.anx.io/go-anxcloud/pkg/client"
)

func main() {
	ctx := context.Background()

	// Get environment variables
	token := os.Getenv("ANEXIA_TOKEN")
	baseURL := os.Getenv("ANEXIA_BASE_URL")

	if token == "" {
		log.Fatal("ANEXIA_TOKEN environment variable is required")
	}

	fmt.Println("=== CloudDNS API Response Analysis ===")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Timestamp: %s\n\n", time.Now().Format(time.RFC3339))

	// Create client options
	opts := []client.Option{
		client.TokenFromString(token),
	}
	if baseURL != "" {
		opts = append(opts, client.BaseURL(baseURL))
	}

	// Create API client
	apiClient, err := api.NewAPI(api.WithClientOptions(opts...))
	if err != nil {
		log.Fatalf("Failed to create AnxCloud API client: %v", err)
	}

	// Test 1: List available zones
	fmt.Println("1. Listing available DNS zones...")

	var zonePageIter types.PageInfo
	zones := make([]clouddnsv1.Zone, 0)
	err = apiClient.List(ctx, &clouddnsv1.Zone{}, api.Paged(1, 100, &zonePageIter))
	if err != nil {
		log.Fatalf("Failed to list zones: %v", err)
	}

	var pagedZones []clouddnsv1.Zone
	for zonePageIter.Next(&pagedZones) {
		zones = append(zones, pagedZones...)
	}

	if err := zonePageIter.Error(); err != nil {
		log.Fatalf("Error iterating zones: %v", err)
	}

	if len(zones) == 0 {
		fmt.Println("No DNS zones found. Please create a test zone first.")
		return
	}

	fmt.Printf("Found %d zones:\n", len(zones))
	for i, z := range zones {
		fmt.Printf("  %d. %s (IsMaster: %t, IsEditable: %t)\n", i+1, z.Name, z.IsMaster, z.IsEditable)
	}

	// Use the first editable zone for testing
	var testZone *clouddnsv1.Zone
	for _, z := range zones {
		if z.IsEditable {
			testZone = &z
			break
		}
	}

	if testZone == nil {
		fmt.Println("No editable zones found. Using first zone for read-only testing.")
		testZone = &zones[0]
	}

	fmt.Printf("\nUsing zone: %s (Editable: %t)\n\n", testZone.Name, testZone.IsEditable)

	// Test 2: Get records using v1 API
	fmt.Println("2. Retrieving DNS records using v1 API...")

	var recordPageIter types.PageInfo
	recordList := clouddnsv1.Record{ZoneName: testZone.Name}
	err = apiClient.List(ctx, &recordList, api.Paged(1, 100, &recordPageIter))
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}

	var records []clouddnsv1.Record
	var pagedRecords []clouddnsv1.Record
	for recordPageIter.Next(&pagedRecords) {
		records = append(records, pagedRecords...)
	}

	if err := recordPageIter.Error(); err != nil {
		log.Fatalf("Error iterating records: %v", err)
	}

	fmt.Printf("Found %d records in zone %s\n", len(records), testZone.Name)

	// Print detailed record information
	for i, record := range records {
		fmt.Printf("\n--- Record %d ---\n", i+1)
		fmt.Printf("Identifier: %s\n", record.Identifier)
		fmt.Printf("Name: %s\n", record.Name)
		fmt.Printf("Type: %s\n", record.Type)
		fmt.Printf("RData: %s\n", record.RData)
		fmt.Printf("TTL: %d\n", record.TTL)
		fmt.Printf("Region: %s\n", record.Region)
		fmt.Printf("Immutable: %t\n", record.Immutable)

		// Try to marshal to JSON to see the raw structure
		jsonBytes, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling to JSON: %v\n", err)
		} else {
			fmt.Printf("Raw JSON:\n%s\n", string(jsonBytes))
		}
	}
	// Test 3: Test identifier stability (only if zone is editable)
	if !testZone.IsEditable {
		fmt.Println("\n3. Skipping identifier stability test - zone is not editable")
		return
	}

	fmt.Println("\n3. Testing identifier stability...")

	// Create a test record
	testRecord := clouddnsv1.Record{
		Name:     fmt.Sprintf("test-identifier-%d", time.Now().Unix()),
		Type:     "TXT",
		RData:    fmt.Sprintf("\"test-value-%d\"", time.Now().Unix()),
		ZoneName: testZone.Name,
		TTL:      300,
	}

	fmt.Printf("Creating test record: %s %s %s\n", testRecord.Name, testRecord.Type, testRecord.RData)

	// Create the record
	if err := apiClient.Create(ctx, &testRecord); err != nil {
		log.Fatalf("Failed to create test record: %v", err)
	}

	fmt.Printf("Created record with identifier: %s\n", testRecord.Identifier)

	// Wait a moment
	time.Sleep(2 * time.Second)

	// Find the record again to see if identifier changed
	foundRecord, err := findDNSRecord(ctx, apiClient, testRecord)
	if err != nil {
		log.Fatalf("Failed to find created record: %v", err)
	}

	fmt.Printf("Found record identifier after creation: %s\n", foundRecord.Identifier)
	fmt.Printf("Identifier stable after creation: %t\n", testRecord.Identifier == foundRecord.Identifier)

	// Modify the record
	originalIdentifier := foundRecord.Identifier
	foundRecord.RData = fmt.Sprintf("\"modified-value-%d\"", time.Now().Unix())

	fmt.Printf("\nModifying record RData to: %s\n", foundRecord.RData)

	if err := apiClient.Update(ctx, &foundRecord); err != nil {
		log.Fatalf("Failed to update test record: %v", err)
	}

	fmt.Printf("Updated record identifier: %s\n", foundRecord.Identifier)
	fmt.Printf("Identifier stable after update: %t\n", originalIdentifier == foundRecord.Identifier)

	// Find again after update
	foundAgain, err := findDNSRecord(ctx, apiClient, foundRecord)
	if err != nil {
		log.Fatalf("Failed to find updated record: %v", err)
	}

	fmt.Printf("Found record identifier after update: %s\n", foundAgain.Identifier)
	fmt.Printf("Identifier stable in listing after update: %t\n", foundRecord.Identifier == foundAgain.Identifier)

	// Clean up - delete the test record
	fmt.Printf("\nCleaning up test record...\n")
	if err := apiClient.Destroy(ctx, &foundAgain); err != nil {
		log.Printf("Warning: Failed to delete test record: %v", err)
	}

	fmt.Println("\n=== Analysis Complete ===")
}

// findDNSRecord is a simplified version of the one in the Terraform provider
func findDNSRecord(ctx context.Context, a api.API, r clouddnsv1.Record) (foundRecord clouddnsv1.Record, err error) {
	var pageIter types.PageInfo
	err = a.List(ctx, &r, api.Paged(1, 100, &pageIter))
	if err != nil {
		return
	}

	var pagedRecords []clouddnsv1.Record
	for pageIter.Next(&pagedRecords) {
		// Simple search - look for matching name and type
		for _, record := range pagedRecords {
			if record.Name == r.Name && record.Type == r.Type {
				return record, nil
			}
		}
	}

	return foundRecord, api.ErrNotFound
}
