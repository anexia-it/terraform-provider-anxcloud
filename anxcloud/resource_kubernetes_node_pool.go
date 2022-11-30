package anxcloud

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	"go.anx.io/go-anxcloud/pkg/apis/common/gs"
	kubernetesv1 "go.anx.io/go-anxcloud/pkg/apis/kubernetes/v1"
	"go.anx.io/go-anxcloud/pkg/utils/pointer"
)

func resourceKubernetesNodePool() *schema.Resource {
	return &schema.Resource{
		Description: strings.TrimSpace(`
			Resource to create Kubernetes node pools.
			Updates are currently not supported.
		`),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CreateContext: tagsMiddlewareCreate(resourceKubernetesNodePoolCreate),
		ReadContext:   tagsMiddlewareRead(resourceKubernetesNodePoolRead),
		DeleteContext: resourceKubernetesNodePoolDelete,
		Schema:        withTagsAttribute(schemaKubernetesNodePool()),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Read:   schema.DefaultTimeout(time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func resourceKubernetesNodePoolCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	nodePool := kubernetesv1.NodePool{
		Name:            d.Get("name").(string),
		Replicas:        pointer.Int(d.Get("initial_replicas").(int)),
		CPUs:            d.Get("cpus").(int),
		Memory:          d.Get("memory_gib").(int) * gibiFactor,
		DiskSize:        d.Get("disk").([]any)[0].(map[string]any)["size_gib"].(int) * gibiFactor,
		OperatingSystem: kubernetesv1.OperatingSystem(d.Get("operating_system").(string)),
		Cluster:         kubernetesv1.Cluster{Identifier: d.Get("cluster").(string)},
	}

	if err := a.Create(ctx, &nodePool); err != nil {
		return diag.Errorf("failed to create node pool: %s", err)
	}

	d.SetId(nodePool.Identifier)

	return resourceKubernetesNodePoolRead(ctx, d, m)
}

func resourceKubernetesNodePoolRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	nodePool := kubernetesv1.NodePool{Identifier: d.Id()}
	if err := a.Get(ctx, &nodePool); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed getting node pool: %s", err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	if err := gs.AwaitCompletion(ctx, a, &nodePool); err != nil {
		return diag.Errorf("failed awaiting Kubernetes node pool completion: %s", err)
	}

	var diags diag.Diagnostics

	if err := d.Set("name", nodePool.Name); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("initial_replicas", pointer.IntVal(nodePool.Replicas)); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("cpus", nodePool.CPUs); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("memory_gib", nodePool.Memory/gibiFactor); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("operating_system", nodePool.OperatingSystem); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}
	if err := d.Set("cluster", nodePool.Cluster.Identifier); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	disks := []map[string]any{{"size_gib": nodePool.DiskSize / gibiFactor}}
	if err := d.Set("disk", disks); err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	return diags
}

func resourceKubernetesNodePoolDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	if err := a.Destroy(ctx, &kubernetesv1.NodePool{Identifier: d.Id()}); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting node pool: %s", err)
	}

	return nil
}
