package anxcloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
)

func resourceKubernetesNodePool() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to create and update Kubernetes node pools.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CreateContext: tagsMiddlewareCreate(resourceKubernetesNodePoolCreate),
		ReadContext:   tagsMiddlewareRead(resourceKubernetesNodePoolRead),
		UpdateContext: tagsMiddlewareUpdate(resourceKubernetesNodePoolUpdate),
		DeleteContext: resourceKubernetesNodePoolDelete,
		Schema:        withTagsAttribute(schemaKubernetesNodePool()),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceKubernetesNodePoolCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	nodePool := kubernetesNodePool{requestDefinition: kubernetesNodePoolCreateDefinition(d)}
	if err := a.Create(ctx, &nodePool); err != nil {
		return diag.Errorf("failed to create Kubernetes node pool: %s", err)
	}

	d.SetId(nodePool.Identifier)

	var err error
	nodePool, err = awaitKubernetesNodePoolReconciliation(ctx, a, nodePool.Identifier)
	if err != nil {
		return diag.Errorf("failed awaiting Kubernetes node pool reconciliation: %s", err)
	}

	return setResourceDataFromKubernetesNodePool(d, nodePool)
}

func resourceKubernetesNodePoolRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	nodePool, err := getKubernetesNodePool(ctx, a, d.Id())
	if err != nil {
		if api.IgnoreNotFound(err) == nil {
			d.SetId("")
			return nil
		}
		return diag.Errorf("failed getting Kubernetes node pool: %s", err)
	}

	// Reconciliation failures are remote state, not refresh failures. Keeping
	// them readable lets Terraform plan an update or destroy operation.
	return setResourceDataFromKubernetesNodePool(d, nodePool)
}

func resourceKubernetesNodePoolUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	definition := kubernetesNodePoolUpdateDefinition(d)
	if len(definition) == 0 {
		return resourceKubernetesNodePoolRead(ctx, d, m)
	}

	nodePool := kubernetesNodePool{requestDefinition: definition}
	nodePool.Identifier = d.Id()
	if err := a.Update(ctx, &nodePool); err != nil {
		return diag.Errorf("failed to update Kubernetes node pool: %s", err)
	}

	nodePool, err := awaitKubernetesNodePoolReconciliation(ctx, a, d.Id())
	if err != nil {
		return diag.Errorf("failed awaiting Kubernetes node pool reconciliation: %s", err)
	}

	return setResourceDataFromKubernetesNodePool(d, nodePool)
}

func resourceKubernetesNodePoolDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	nodePool := kubernetesNodePool{}
	nodePool.Identifier = d.Id()

	if err := a.Destroy(ctx, &nodePool); err != nil && api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting Kubernetes node pool: %s", err)
	}

	if err := awaitKubernetesNodePoolDeletion(ctx, a, d.Id()); err != nil {
		return diag.Errorf("failed awaiting Kubernetes node pool deletion: %s", err)
	}

	d.SetId("")
	return nil
}

func kubernetesNodePoolCreateDefinition(d *schema.ResourceData) map[string]any {
	definition := map[string]any{
		"name":                 d.Get("name").(string),
		"replicas":             d.Get("initial_replicas").(int),
		"cpus":                 d.Get("cpus").(int),
		"memory":               int64(d.Get("memory_gib").(int)) * gibiFactor,
		"disk_size":            int64(d.Get("disk").([]any)[0].(map[string]any)["size_gib"].(int)) * gibiFactor,
		"operating_system":     d.Get("operating_system").(string),
		"cluster":              d.Get("cluster").(string),
		"syncsource":           strings.ToLower(d.Get("sync_source").(string)),
		"cpu_performance_type": d.Get("cpu_performance_type").(string),
		"autoscaler_enabled":   d.Get("autoscaler_enabled").(bool),
		"autoscaler_min_nodes": d.Get("autoscaler_min_nodes").(int),
		"autoscaler_max_nodes": d.Get("autoscaler_max_nodes").(int),
		"dns_override_ipv4":    d.Get("dns_override_ipv4").(bool),
		"dns_override_ipv6":    d.Get("dns_override_ipv6").(bool),
	}

	optionalFields := map[string]string{
		"dns_ipv4_1":      "dns_v4_1",
		"dns_ipv4_2":      "dns_v4_2",
		"dns_ipv6_1":      "dns_v6_1",
		"dns_ipv6_2":      "dns_v6_2",
		"ssh_public_keys": "sshpubkeys",
		"taints":          "taints",
		"labels":          "labels",
		"annotations":     "annotations",
	}
	for terraformName, apiName := range optionalFields {
		if value, ok := d.GetOk(terraformName); ok {
			definition[apiName] = value
		}
	}

	primaryDisk := d.Get("disk").([]any)[0].(map[string]any)
	if performanceType := primaryDisk["performance_type"].(string); performanceType != "" {
		definition["disk_performance_type"] = performanceType
	}
	if disks := kubernetesNodePoolAdditionalDisksDefinition(d); len(disks) != 0 {
		definition["additional_disks"] = disks
	}
	if networks := kubernetesNodePoolNetworksDefinition(d); len(networks) != 0 {
		definition["networks"] = networks
	}

	return definition
}

func kubernetesNodePoolUpdateDefinition(d *schema.ResourceData) map[string]any {
	definition := make(map[string]any)
	fields := map[string]string{
		"name":                 "name",
		"initial_replicas":     "replicas",
		"cpus":                 "cpus",
		"operating_system":     "operating_system",
		"cluster":              "cluster",
		"sync_source":          "syncsource",
		"cpu_performance_type": "cpu_performance_type",
		"autoscaler_enabled":   "autoscaler_enabled",
		"autoscaler_min_nodes": "autoscaler_min_nodes",
		"autoscaler_max_nodes": "autoscaler_max_nodes",
		"dns_override_ipv4":    "dns_override_ipv4",
		"dns_ipv4_1":           "dns_v4_1",
		"dns_ipv4_2":           "dns_v4_2",
		"dns_override_ipv6":    "dns_override_ipv6",
		"dns_ipv6_1":           "dns_v6_1",
		"dns_ipv6_2":           "dns_v6_2",
		"ssh_public_keys":      "sshpubkeys",
		"taints":               "taints",
		"labels":               "labels",
		"annotations":          "annotations",
	}
	for terraformName, apiName := range fields {
		if d.HasChange(terraformName) {
			definition[apiName] = d.Get(terraformName)
		}
	}
	if d.HasChange("memory_gib") {
		definition["memory"] = int64(d.Get("memory_gib").(int)) * gibiFactor
	}
	if d.HasChange("disk") {
		primaryDisk := d.Get("disk").([]any)[0].(map[string]any)
		definition["disk_size"] = int64(primaryDisk["size_gib"].(int)) * gibiFactor
		definition["disk_performance_type"] = primaryDisk["performance_type"].(string)
	}
	if d.HasChange("additional_disks") {
		definition["additional_disks"] = kubernetesNodePoolAdditionalDisksDefinition(d)
	}
	if d.HasChange("networks") {
		definition["networks"] = kubernetesNodePoolNetworksDefinition(d)
	}
	return definition
}

func kubernetesNodePoolAdditionalDisksDefinition(d *schema.ResourceData) []map[string]any {
	values := d.Get("additional_disks").([]any)
	disks := make([]map[string]any, 0, len(values))
	for _, value := range values {
		disk := value.(map[string]any)
		disks = append(disks, map[string]any{
			"name":             disk["name"].(string),
			"size_bytes":       int64(disk["size_gib"].(int)) * gibiFactor,
			"performance_type": disk["performance_type"].(string),
		})
	}
	return disks
}

func kubernetesNodePoolNetworksDefinition(d *schema.ResourceData) []map[string]any {
	values := d.Get("networks").([]any)
	networks := make([]map[string]any, 0, len(values))
	for _, value := range values {
		network := value.(map[string]any)
		networks = append(networks, map[string]any{
			"name":            network["name"].(string),
			"bandwidth_limit": network["bandwidth_limit"].(string),
			"vlan":            network["vlan"].(string),
		})
	}
	return networks
}

func setResourceDataFromKubernetesNodePool(d *schema.ResourceData, nodePool kubernetesNodePool) diag.Diagnostics {
	var diags diag.Diagnostics
	set := func(name string, value any) {
		if err := d.Set(name, value); err != nil {
			diags = append(diags, diag.FromErr(fmt.Errorf("set %s: %w", name, err))...)
		}
	}

	set("name", nodePool.Name)
	if !nodePool.AutoscalerEnabled {
		set("initial_replicas", nodePool.Replicas)
	}
	set("cpus", nodePool.CPUs)
	set("memory_gib", nodePool.MemoryBytes/gibiFactor)
	set("operating_system", nodePool.OperatingSystem.ID)
	set("cluster", nodePool.Cluster.Identifier)
	set("sync_source", nodePool.SyncSource.ID)
	set("cpu_performance_type", nodePool.CPUType.ID)
	set("autoscaler_enabled", nodePool.AutoscalerEnabled)
	set("autoscaler_min_nodes", nodePool.AutoscalerMinNodes)
	set("autoscaler_max_nodes", nodePool.AutoscalerMaxNodes)
	set("dns_override_ipv4", nodePool.DNSOverrideIPv4)
	set("dns_ipv4_1", nodePool.DNSv4Entry1)
	set("dns_ipv4_2", nodePool.DNSv4Entry2)
	set("dns_override_ipv6", nodePool.DNSOverrideIPv6)
	set("dns_ipv6_1", nodePool.DNSv6Entry1)
	set("dns_ipv6_2", nodePool.DNSv6Entry2)
	set("ssh_public_keys", nodePool.SSHPubKeys)
	set("taints", nodePool.Taints)
	set("labels", nodePool.Labels)
	set("annotations", nodePool.Annotations)
	set("state", nodePool.State.ID)
	set("state_text", nodePool.State.Text)

	set("disk", []map[string]any{{
		"size_gib":         nodePool.DiskSize / gibiFactor,
		"performance_type": nodePool.DiskPerformanceType.ID,
	}})

	additionalDisks := make([]map[string]any, 0, len(nodePool.AdditionalDisks))
	for _, disk := range nodePool.AdditionalDisks {
		additionalDisks = append(additionalDisks, map[string]any{
			"name":             disk.Name,
			"size_gib":         disk.SizeBytes / gibiFactor,
			"performance_type": disk.PerformanceType.ID,
		})
	}
	set("additional_disks", additionalDisks)

	networks := make([]map[string]any, 0, len(nodePool.Networks))
	for _, network := range nodePool.Networks {
		networks = append(networks, map[string]any{
			"name":            network.Name,
			"bandwidth_limit": network.BandwidthLimit.ID,
			"vlan":            network.VLAN.Identifier,
		})
	}
	set("networks", networks)

	return diags
}

func awaitKubernetesNodePoolReconciliation(
	ctx context.Context,
	a api.API,
	identifier string,
) (kubernetesNodePool, error) {
	for {
		nodePool, err := getKubernetesNodePool(ctx, a, identifier)
		if err != nil {
			return nodePool, err
		}

		switch nodePool.State.ID {
		case "0":
			return nodePool, nil
		case "1":
			return nodePool, fmt.Errorf("node pool reconciliation entered error state %q (%s)", nodePool.State.Text, nodePool.State.ID)
		}

		select {
		case <-ctx.Done():
			return nodePool, ctx.Err()
		case <-time.After(kubernetesReconciliationPollInterval):
		}
	}
}

func awaitKubernetesNodePoolDeletion(
	ctx context.Context,
	a api.API,
	identifier string,
) error {
	for {
		_, err := getKubernetesNodePool(ctx, a, identifier)
		if err != nil {
			if api.IgnoreNotFound(err) == nil {
				return nil
			}
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(kubernetesReconciliationPollInterval):
		}
	}
}

func getKubernetesNodePool(ctx context.Context, a api.API, identifier string) (kubernetesNodePool, error) {
	nodePool := kubernetesNodePool{}
	nodePool.Identifier = identifier
	if err := a.Get(ctx, &nodePool); err != nil {
		return nodePool, err
	}
	return nodePool, nil
}
