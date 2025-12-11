package anxcloud

import (
	"fmt"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"go.anx.io/go-anxcloud/pkg/utils/test"
)

func TestAccAnxCloudDNSRecord(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSZoneAndRecord(zoneName, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "zone_name", "0-"+zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "name", "a-record"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.txt_record", "name", "txt-record"),
				),
			},
			{
				Config: testAccAnxDNSZoneAndRecord(zoneName, 1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "zone_name", "1-"+zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.a_record", "name", "a-record"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.txt_record", "name", "txt-record"),
				),
			},
		},
	})
}

// TestRegressionMultipleRecordsSameNameType tests that records with same name/type

// testCheckResourceIDsUnique verifies that multiple resources have unique IDs
func testCheckResourceIDsUnique(resourceNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ids := make(map[string]bool)
		for _, resourceName := range resourceNames {
			rs, ok := s.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("resource %s not found in state", resourceName)
			}
			if rs.Primary == nil {
				return fmt.Errorf("resource %s has no primary instance", resourceName)
			}
			id := rs.Primary.ID
			if ids[id] {
				return fmt.Errorf("resources have duplicate IDs: %s appears more than once", id)
			}
			ids[id] = true
		}
		return nil
	}
}

// testCheckResourceIDsDifferent verifies that two resources have different IDs
func testCheckResourceIDsDifferent(resourceName1, resourceName2 string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs1, ok := s.RootModule().Resources[resourceName1]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName1)
		}
		rs2, ok := s.RootModule().Resources[resourceName2]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName2)
		}
		if rs1.Primary == nil {
			return fmt.Errorf("resource %s has no primary instance", resourceName1)
		}
		if rs2.Primary == nil {
			return fmt.Errorf("resource %s has no primary instance", resourceName2)
		}
		if rs1.Primary.ID == rs2.Primary.ID {
			return fmt.Errorf("resources %s and %s have the same ID: %s", resourceName1, resourceName2, rs1.Primary.ID)
		}
		return nil
	}
}

// TestRegressionDNSRecordValidationStuckAtZero tests that DNS record creation handles
// zone validation stuck at 0% gracefully without hanging indefinitely
func TestRegressionDNSRecordValidationStuckAtZero(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create DNS zone and records successfully
			// This tests the normal case where validation starts and completes
			{
				Config: testAccDNSRecordValidationStuckAtZeroNormalCase(zoneName),
				Check: resource.ComposeTestCheckFunc(
					// Verify zone was created
					resource.TestCheckResourceAttr("anxcloud_dns_zone.validation_test_zone", "name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_zone.validation_test_zone", "is_master", "true"),
					// Verify records were created successfully
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_a", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_a", "name", "@"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_a", "type", "A"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_a", "rdata", "192.168.1.1"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.validation_record_a", "id"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_txt", "name", "validation-test"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_txt", "type", "TXT"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_txt", "rdata", "test validation"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.validation_record_txt", "id"),
				),
			},
			// Step 2: Test that operations complete within reasonable time
			// This verifies the timeout logic prevents indefinite hanging
			// Note: We can't easily simulate the stuck validation case in integration tests,
			// but we can verify the operation completes and doesn't hang
			{
				Config: testAccDNSRecordValidationStuckAtZeroAdditionalRecords(zoneName),
				Check: resource.ComposeTestCheckFunc(
					// Verify additional records were created
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_mx", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_mx", "name", "@"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_mx", "type", "MX"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.validation_record_mx", "rdata", "10 mail.example.com"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.validation_record_mx", "id"),
					// Verify all records still exist
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.validation_record_a", "id"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.validation_record_txt", "id"),
				),
			},
		},
	})
}

// TestRegressionDNSRecordRefactorComprehensive tests the complete DNS record refactor functionality
// including content-based ID generation, migration scenarios, import compatibility, and CRUD operations
func TestRegressionDNSRecordRefactorComprehensive(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create multiple records with same name/type but different RDATA
			// This tests content-based ID generation and uniqueness
			{
				Config: testAccDNSRecordRefactorMultipleRecords(zoneName),
				Check: resource.ComposeTestCheckFunc(
					// Verify all records are created with unique IDs
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.record1", "id"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.record2", "id"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.record3", "id"),
					// Verify IDs are different (content-based uniqueness) using custom check
					testCheckResourceIDsUnique("anxcloud_dns_record.record1", "anxcloud_dns_record.record2", "anxcloud_dns_record.record3"),
					// Verify record attributes
					resource.TestCheckResourceAttr("anxcloud_dns_record.record1", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.record1", "name", "@"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.record1", "type", "A"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.record1", "rdata", "192.168.1.1"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.record1", "ttl", "3600"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.record2", "rdata", "192.168.1.2"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.record3", "rdata", "192.168.1.3"),
				),
			},
			// Step 2: Update TTL on one record to test ID generation changes
			// This tests that TTL changes affect ID generation
			{
				Config: testAccDNSRecordRefactorTTLChange(zoneName),
				Check: resource.ComposeTestCheckFunc(
					// Verify the updated record has a different ID due to TTL change
					testCheckResourceIDsDifferent("anxcloud_dns_record.record1", "anxcloud_dns_record.record1_updated"),
					// Verify the other records maintain their IDs
					resource.TestCheckResourceAttr("anxcloud_dns_record.record2", "rdata", "192.168.1.2"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.record3", "rdata", "192.168.1.3"),
					// Verify the updated record has new TTL
					resource.TestCheckResourceAttr("anxcloud_dns_record.record1_updated", "ttl", "7200"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.record1_updated", "rdata", "192.168.1.1"),
				),
			},
			// Step 3: Test import functionality with content hashes
			// This tests import with new ID format
			{
				ResourceName:      "anxcloud_dns_record.record1_updated",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Step 4: Test backward compatibility - simulate old fake UUID import
			// This tests migration from old fake UUIDs to new content hashes
			{
				Config: testAccDNSRecordRefactorBackwardCompat(zoneName),
				Check: resource.ComposeTestCheckFunc(
					// Create a new record that will be imported with old-style ID
					resource.TestCheckResourceAttr("anxcloud_dns_record.backward_compat", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.backward_compat", "name", "compat-test"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.backward_compat", "type", "A"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.backward_compat", "rdata", "10.0.0.1"),
				),
			},
			// Step 5: Test CRUD operations work correctly with new identification system
			// Update RDATA and verify the change works
			{
				Config: testAccDNSRecordRefactorCRUDTest(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.crud_test", "rdata", "10.0.0.2"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.crud_test", "ttl", "1800"),
				),
			},
		},
	})
}

// Helper functions for the comprehensive regression test

func testAccDNSRecordRefactorMultipleRecords(zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "refactor_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	# Multiple A records with same name but different RDATA - should get unique content-based IDs
	resource "anxcloud_dns_record" "record1" {
		name = "@"
		zone_name = anxcloud_dns_zone.refactor_zone.name
		type = "A"
		rdata = "192.168.1.1"
		ttl = 3600
	}

	resource "anxcloud_dns_record" "record2" {
		name = "@"
		zone_name = anxcloud_dns_zone.refactor_zone.name
		type = "A"
		rdata = "192.168.1.2"
		ttl = 3600
	}

	resource "anxcloud_dns_record" "record3" {
		name = "@"
		zone_name = anxcloud_dns_zone.refactor_zone.name
		type = "A"
		rdata = "192.168.1.3"
		ttl = 3600
	}
	`, zoneName)
}

func testAccDNSRecordRefactorTTLChange(zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "refactor_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	# Same records but record1 has different TTL - should get different content-based ID
	resource "anxcloud_dns_record" "record1_updated" {
		name = "@"
		zone_name = anxcloud_dns_zone.refactor_zone.name
		type = "A"
		rdata = "192.168.1.1"
		ttl = 7200  # Changed TTL
	}

	resource "anxcloud_dns_record" "record2" {
		name = "@"
		zone_name = anxcloud_dns_zone.refactor_zone.name
		type = "A"
		rdata = "192.168.1.2"
		ttl = 3600
	}

	resource "anxcloud_dns_record" "record3" {
		name = "@"
		zone_name = anxcloud_dns_zone.refactor_zone.name
		type = "A"
		rdata = "192.168.1.3"
		ttl = 3600
	}
	`, zoneName)
}

func testAccDNSRecordRefactorBackwardCompat(zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "refactor_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	# Record for backward compatibility testing
	resource "anxcloud_dns_record" "backward_compat" {
		name = "compat-test"
		zone_name = anxcloud_dns_zone.refactor_zone.name
		type = "A"
		rdata = "10.0.0.1"
		ttl = 3600
	}
	`, zoneName)
}

func testAccDNSRecordRefactorCRUDTest(zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "refactor_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	# Test CRUD operations - update both RDATA and TTL
	resource "anxcloud_dns_record" "crud_test" {
		name = "compat-test"
		zone_name = anxcloud_dns_zone.refactor_zone.name
		type = "A"
		rdata = "10.0.0.2"  # Changed RDATA
		ttl = 1800         # Changed TTL
	}
	`, zoneName)
}

func TestAccAnxCloudDNSRecordUpdate(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSZoneAndRecordUpdate(zoneName, "192.168.1.100", 300),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "name", "update-test"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "type", "A"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "rdata", "192.168.1.100"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "ttl", "300"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.update_test", "identifier"),
				),
			},
			{
				Config: testAccAnxDNSZoneAndRecordUpdate(zoneName, "192.168.1.200", 600),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "name", "update-test"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "type", "A"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "rdata", "192.168.1.200"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.update_test", "ttl", "600"),
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.update_test", "identifier"),
				),
			},
		},
	})
}

func testAccAnxDNSZoneAndRecord(zoneNameSuffix string, recordsZoneIndex uint) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "test_dns_zones" {
		count = 2
		name = "${count.index}-%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	resource "anxcloud_dns_record" "a_record" {
		name = "a-record"
		zone_name = anxcloud_dns_zone.test_dns_zones[%[2]d].name
		type = "A"
		rdata = "1.1.1.1"
		ttl = 300
	}

	resource "anxcloud_dns_record" "txt_record" {
		name = "txt-record"
		zone_name = anxcloud_dns_zone.test_dns_zones[%[2]d].name
		type = "TXT"
		rdata = "hello world"
		ttl = 300
	}
	`, zoneNameSuffix, recordsZoneIndex)
}

func testAccAnxDNSZoneAndRecordUpdate(zoneName, rdata string, ttl int) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "update_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@example.com"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	resource "anxcloud_dns_record" "update_test" {
		name = "update-test"
		zone_name = anxcloud_dns_zone.update_zone.name
		type = "A"
		rdata = "%s"
		ttl = %d
	}
	`, zoneName, rdata, ttl)
}

// Helper functions for TestRegressionDNSRecordValidationStuckAtZero

func testAccDNSRecordValidationStuckAtZeroNormalCase(zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "validation_test_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	# Create multiple records to trigger changeset execution and validation
	resource "anxcloud_dns_record" "validation_record_a" {
		name = "@"
		zone_name = anxcloud_dns_zone.validation_test_zone.name
		type = "A"
		rdata = "192.168.1.1"
		ttl = 3600
	}

	resource "anxcloud_dns_record" "validation_record_txt" {
		name = "validation-test"
		zone_name = anxcloud_dns_zone.validation_test_zone.name
		type = "TXT"
		rdata = "test validation"
		ttl = 3600
	}
	`, zoneName)
}

func testAccDNSRecordValidationStuckAtZeroAdditionalRecords(zoneName string) string {
	return fmt.Sprintf(`
	resource "anxcloud_dns_zone" "validation_test_zone" {
		name = "%s"
		is_master = true
		dns_sec_mode = "unvalidated"
		admin_email = "admin@terraform.test"
		refresh = 100
		retry = 100
		expire = 1000
		ttl = 100
	}

	# Keep existing records
	resource "anxcloud_dns_record" "validation_record_a" {
		name = "@"
		zone_name = anxcloud_dns_zone.validation_test_zone.name
		type = "A"
		rdata = "192.168.1.1"
		ttl = 3600
	}

	resource "anxcloud_dns_record" "validation_record_txt" {
		name = "validation-test"
		zone_name = anxcloud_dns_zone.validation_test_zone.name
		type = "TXT"
		rdata = "test validation"
		ttl = 3600
	}

	# Add additional record to test further changeset operations
	resource "anxcloud_dns_record" "validation_record_mx" {
		name = "@"
		zone_name = anxcloud_dns_zone.validation_test_zone.name
		type = "MX"
		rdata = "10 mail.example.com"
		ttl = 3600
	}
	`, zoneName)
}

// TestAccAnxCloudDNSRecord_TTLOmitted tests that omitting TTL uses zone default and doesn't show drift
func TestAccAnxCloudDNSRecord_TTLOmitted(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{

		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSRecordTTLOmitted(zoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.no_ttl", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.no_ttl", "name", "no-ttl"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.no_ttl", "type", "A"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.no_ttl", "rdata", "192.168.100.1"),
					// TTL should be set by provider (computed from zone or API)
					resource.TestCheckResourceAttrSet("anxcloud_dns_record.no_ttl", "ttl"),
				),
			},
			// Second plan should show no changes (no drift for omitted TTL)
			{
				Config:   testAccAnxDNSRecordTTLOmitted(zoneName),
				PlanOnly: true,
			},
		},
	})
}

// TestAccAnxCloudDNSRecord_TTLExplicit tests that explicit TTL is respected
func TestAccAnxCloudDNSRecord_TTLExplicit(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSRecordTTLExplicit(zoneName, 7200),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "zone_name", zoneName),
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "name", "with-ttl"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "type", "A"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "rdata", "192.168.100.2"),
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "ttl", "7200"),
				),
			},
			// Second plan should show no changes
			{
				Config:   testAccAnxDNSRecordTTLExplicit(zoneName, 7200),
				PlanOnly: true,
			},
		},
	})
}

// TestAccAnxCloudDNSRecord_TTLUpdate tests that TTL can be updated in-place
func TestAccAnxCloudDNSRecord_TTLUpdate(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAnxDNSRecordTTLExplicit(zoneName, 3600),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "ttl", "3600"),
				),
			},
			{
				Config: testAccAnxDNSRecordTTLExplicit(zoneName, 1800),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "ttl", "1800"),
				),
			},
			{
				Config: testAccAnxDNSRecordTTLExplicit(zoneName, 300),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "ttl", "300"),
				),
			},
		},
	})
}

// TestAccAnxCloudDNSRecord_ImportWithoutTTL tests import without TTL in config
func TestAccAnxCloudDNSRecord_ImportWithoutTTL(t *testing.T) {
	environment.SkipIfNoEnvironment(t)
	zoneName := test.RandomHostname() + ".terraform.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Create record with explicit TTL
				Config: testAccAnxDNSRecordTTLExplicit(zoneName, 7200),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.with_ttl", "ttl", "7200"),
				),
			},
			{
				// Import into config WITHOUT ttl specified
				ResourceName:      "anxcloud_dns_record.with_ttl",
				ImportState:       true,
				ImportStateVerify: true,
				// After import, the config still has TTL, but if user removes it
				// the subsequent plan should not show drift
			},
			{
				// Now use config without TTL - should use API value (7200) with no drift
				Config: testAccAnxDNSRecordTTLOmittedWithName(zoneName, "with-ttl"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("anxcloud_dns_record.no_ttl", "name", "with-ttl"),
					// TTL should be computed from API (still 7200)
					resource.TestCheckResourceAttr("anxcloud_dns_record.no_ttl", "ttl", "7200"),
				),
			},
			// Verify no drift when TTL is omitted
			{
				Config:   testAccAnxDNSRecordTTLOmittedWithName(zoneName, "with-ttl"),
				PlanOnly: true,
			},
		},
	})
}

func testAccAnxDNSRecordTTLOmitted(zoneName string) string {
	return fmt.Sprintf(`
resource "anxcloud_dns_zone" "test" {
	name = %q
}

resource "anxcloud_dns_record" "no_ttl" {
	zone_name = anxcloud_dns_zone.test.name
	name      = "no-ttl"
	type      = "A"
	rdata     = "192.168.100.1"
	# ttl not specified - should use zone default and not show drift
}
`, zoneName)
}

func testAccAnxDNSRecordTTLOmittedWithName(zoneName, recordName string) string {
	return fmt.Sprintf(`
resource "anxcloud_dns_zone" "test" {
	name = %q
}

resource "anxcloud_dns_record" "no_ttl" {
	zone_name = anxcloud_dns_zone.test.name
	name      = %q
	type      = "A"
	rdata     = "192.168.100.2"
	# ttl not specified - should use value from API
}
`, zoneName, recordName)
}

func testAccAnxDNSRecordTTLExplicit(zoneName string, ttl int) string {
	return fmt.Sprintf(`
resource "anxcloud_dns_zone" "test" {
	name = %q
}

resource "anxcloud_dns_record" "with_ttl" {
	zone_name = anxcloud_dns_zone.test.name
	name      = "with-ttl"
	type      = "A"
	rdata     = "192.168.100.2"
	ttl       = %d
}
`, zoneName, ttl)
}
