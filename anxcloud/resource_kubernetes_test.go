package anxcloud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/anexia-it/terraform-provider-anxcloud/anxcloud/testutils/environment"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
	kubernetesv1 "go.anx.io/go-anxcloud/pkg/apis/kubernetes/v1"
	"go.anx.io/go-anxcloud/pkg/utils/pointer"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Kubernetes Resource", func() {
	var mock ghttpMock

	BeforeEach(func() {
		mock = ghttpMock{ghttp.NewServer()}

		DeferCleanup(func() {
			mock.server.Close()
		})
	})

	It("can Create, Read and Delete Cluster resources", func() {
		resource.ParallelTest(GinkgoT(), resource.TestCase{
			IsUnitTest:        true,
			ProviderFactories: testAccProviderFactoriesWithMockedClient(GinkgoT(), mock.server),
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						// create with await completion
						mock.appendCreateClusterHandler()
						mock.appendGetClusterHandler()
						mock.appendGetClusterHandler()

						// get with tags middleware
						mock.appendGetClusterHandler()
						mock.appendGetClusterHandler()
						mock.appendGetTagsHandler("test-cluster-identifier")

						// delete
						mock.appendDeleteClusterHandler()
					},
					Config: `
					resource "anxcloud_kubernetes_cluster" "foo" {
						name = "foo"
						location = "test-location"
						needs_service_vms = true
					}
					`,
				},
			},
		})
	})

	It("can Create, Read and Delete Node Pool resources", func() {
		resource.ParallelTest(GinkgoT(), resource.TestCase{
			IsUnitTest:        true,
			ProviderFactories: testAccProviderFactoriesWithMockedClient(GinkgoT(), mock.server),
			Steps: []resource.TestStep{
				{
					PreConfig: func() {
						//// create with await completion
						mock.appendCreateNodePoolHandler()
						mock.appendGetNodePoolHandler()
						mock.appendGetNodePoolHandler()

						// get with tags middleware
						mock.appendGetNodePoolHandler()
						mock.appendGetNodePoolHandler()
						mock.appendGetTagsHandler("test-node-pool-identifier")

						// delete
						mock.appendDeleteNodePoolHandler()
					},
					Config: `
					resource "anxcloud_kubernetes_node_pool" "foo" {
						name = "foo"
						initial_replicas = 3
						memory_gib = 4
						cpus = 2
  					operating_system = "Flatcar Linux"
						cluster = "test-cluster"

						disk {
							size_gib = 10
						}
					}
					`,
				},
			},
		})
	})
})

func TestAccAnxCloudKubernetesResourcesCombined(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Kubernetes e2e tests in short mode.")
	}

	environment.SkipIfNoEnvironment(t)
	envInfo := environment.GetEnvInfo(t)

	cluster := kubernetesv1.Cluster{
		Name:            fmt.Sprintf("tf-acc-test-%s", envInfo.TestRunName),
		Location:        corev1.Location{Identifier: envInfo.Location},
		NeedsServiceVMs: pointer.Bool(true),
	}

	nodePool := kubernetesv1.NodePool{
		Name:     fmt.Sprintf("tf-acc-test-%s-np", envInfo.TestRunName),
		Replicas: pointer.Int(1),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"kubernetes": {
				Source:            "hashicorp/kubernetes",
				VersionConstraint: "2.14.0",
			},
		},
		Steps: []resource.TestStep{
			{ // cluster without node pool
				Config: testAccAnxCloudKubernetesCluster(&cluster),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAnxCloudKubernetesClusterExists(),
					resource.TestCheckResourceAttr("anxcloud_kubernetes_cluster.foo", "name", cluster.Name),
				),
			},
			{ // add node pool to cluster
				Config: strings.Join([]string{
					testAccAnxCloudKubernetesCluster(&cluster),
					testAccAnxCloudKubernetesNodePool(&nodePool),
				}, "\n"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAnxCloudKubernetesClusterExists(),
					testAccAnxCloudKubernetesNodePoolExists(),
					resource.TestCheckResourceAttr("anxcloud_kubernetes_node_pool.foo", "initial_replicas", "1"),
				),
			},
			{ // use clusters kubeconfig to create kubernetes resources
				Config: strings.Join([]string{
					testAccAnxCloudKubernetesCluster(&cluster),
					testAccAnxCloudKubernetesNodePool(&nodePool),
					`
				
				data "anxcloud_kubernetes_cluster" "foo" {
					name = anxcloud_kubernetes_cluster.foo.name
				}

				resource "anxcloud_kubernetes_kubeconfig" "foo" {
					cluster = data.anxcloud_kubernetes_cluster.foo.id
				}

				provider "kubernetes" {
					host                   = anxcloud_kubernetes_kubeconfig.foo.host
					token                  = anxcloud_kubernetes_kubeconfig.foo.token
					cluster_ca_certificate = anxcloud_kubernetes_kubeconfig.foo.cluster_ca_certificate
				}
				
				resource "kubernetes_namespace" "foo" {
					metadata {
						name = "foo"
					}
				}

				`}, "\n"),
			},
		},
	})
}

func testAccAnxCloudKubernetesCluster(cluster *kubernetesv1.Cluster) string {
	return fmt.Sprintf(`
	resource "anxcloud_kubernetes_cluster" "foo" {
		name = "%s"
		location = "%s"
		needs_service_vms = %v
	}
	`,
		cluster.Name,
		cluster.Location.Identifier,
		*cluster.NeedsServiceVMs,
	)
}

func testAccAnxCloudKubernetesNodePool(nodePool *kubernetesv1.NodePool) string {
	return fmt.Sprintf(`
	resource "anxcloud_kubernetes_node_pool" "foo" {
		name = "%s"
		initial_replicas = %d
		memory_gib = 4
		cpus = 2
		operating_system = "Flatcar Linux"
		cluster = anxcloud_kubernetes_cluster.foo.id

		disk {
			size_gib = 10
		}
	}
	`,
		nodePool.Name,
		*nodePool.Replicas,
	)
}

func testAccAnxCloudKubernetesClusterExists() resource.TestCheckFunc {
	n := "anxcloud_kubernetes_cluster.foo"

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", n)
		}

		a := apiFromProviderConfig(testAccProvider.Meta())

		engineCluster := kubernetesv1.Cluster{Identifier: rs.Primary.ID}
		if err := a.Get(context.TODO(), &engineCluster); err != nil {
			return fmt.Errorf("failed retrieving kubernetes cluster: %s", err)
		}

		return nil
	}
}

func testAccAnxCloudKubernetesNodePoolExists() resource.TestCheckFunc {
	n := "anxcloud_kubernetes_node_pool.foo"

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", n)
		}

		a := apiFromProviderConfig(testAccProvider.Meta())

		engineNodePool := kubernetesv1.NodePool{Identifier: rs.Primary.ID}
		if err := a.Get(context.TODO(), &engineNodePool); err != nil {
			return fmt.Errorf("failed retrieving kubernetes node pool: %s", err)
		}

		return nil
	}
}
