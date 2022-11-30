package anxcloud

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/apis/common/gs"
	corev1 "go.anx.io/go-anxcloud/pkg/apis/core/v1"
	kubernetesv1 "go.anx.io/go-anxcloud/pkg/apis/kubernetes/v1"
	"go.anx.io/go-anxcloud/pkg/utils/pointer"
)

func resourceKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Description: strings.TrimSpace(`
			Resource to create Kubernetes clusters.
			Updates are currently not supported.
		`),

		CreateContext: tagsMiddlewareCreate(resourceKubernetesClusterCreate),
		ReadContext:   tagsMiddlewareRead(resourceKubernetesClusterRead),
		DeleteContext: resourceKubernetesClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
		Schema: withTagsAttribute(schemaKubernetesCluster()),
	}
}

func resourceKubernetesClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	cluster := kubernetesv1.Cluster{
		Name:              d.Get("name").(string),
		Version:           d.Get("version").(string),
		Location:          corev1.Location{Identifier: d.Get("location").(string)},
		NeedsServiceVMs:   pointer.Bool(d.Get("needs_service_vms").(bool)),
		EnableNATGateways: pointer.Bool(d.Get("enable_nat_gateways").(bool)),
		EnableLBaaS:       pointer.Bool(d.Get("enable_lbaas").(bool)),
	}

	if err := a.Create(ctx, &cluster); err != nil {
		return diag.Errorf("failed to create Kubernetes cluster: %s", err)
	}

	d.SetId(cluster.Identifier)

	return resourceKubernetesClusterRead(ctx, d, m)
}

func resourceKubernetesClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	cluster := kubernetesv1.Cluster{Identifier: d.Id()}
	if err := a.Get(ctx, &cluster); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed getting cluster: %s", err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	if err := gs.AwaitCompletion(ctx, a, &cluster); err != nil {
		return diag.Errorf("failed awaiting Kubernetes cluster completion: %s", err)
	}

	return setResourceDataFromKubernetesCluster(ctx, a, d, cluster)
}

func resourceKubernetesClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	if err := retryKubernetesClusterDeletion(ctx, d, a); err != nil {
		return diag.Errorf("failed deleting cluster: %s", err)
	}

	return nil
}

func retryKubernetesClusterDeletion(ctx context.Context, d *schema.ResourceData, a api.API) error {
	return resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		cluster := kubernetesv1.Cluster{Identifier: d.Id()}
		if err := a.Destroy(ctx, &cluster); api.IgnoreNotFound(err) != nil {
			if errors.Is(err, io.EOF) {
				// if we delete the cluster too soon after node pool deletion we receive a io.EOF error for some reason
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})
}
