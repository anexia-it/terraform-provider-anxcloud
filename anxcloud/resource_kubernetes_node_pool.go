package anxcloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	a := newKubernetesServiceAPI[kubernetesNodePool](m.(providerContext).legacyClient, "node_pool")

	nodePool, err := a.Create(ctx, kubernetesNodePoolCreateDefinition(d))
	if err != nil {
		return diag.Errorf("failed to create Kubernetes node pool: %s", err)
	}

	d.SetId(nodePool.Identifier)
	if definition := kubernetesNodePoolCreateUpdateDefinition(d); len(definition) != 0 {
		nodePool, err = a.Update(ctx, nodePool.Identifier, definition)
		if err != nil {
			return diag.Errorf("failed to apply Kubernetes node pool v2 fields after creation: %s", err)
		}
	}

	nodePool, err = awaitKubernetesNodePoolReconciliation(ctx, a, nodePool.Identifier)
	if err != nil {
		return diag.Errorf("failed awaiting Kubernetes node pool reconciliation: %s", err)
	}

	return setResourceDataFromKubernetesNodePool(d, nodePool)
}

func resourceKubernetesNodePoolRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := newKubernetesServiceAPI[kubernetesNodePool](m.(providerContext).legacyClient, "node_pool")

	nodePool, err := a.Get(ctx, d.Id())
	if isLegacyNotFoundError(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.Errorf("failed getting Kubernetes node pool: %s", err)
	}

	// Reconciliation failures are remote state, not refresh failures. Keeping
	// them readable lets Terraform plan an update or destroy operation.
	return setResourceDataFromKubernetesNodePool(d, nodePool)
}

func resourceKubernetesNodePoolUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := newKubernetesServiceAPI[kubernetesNodePool](m.(providerContext).legacyClient, "node_pool")
	definition := kubernetesNodePoolUpdateDefinition(d)
	if len(definition) == 0 {
		return resourceKubernetesNodePoolRead(ctx, d, m)
	}

	if _, err := a.Update(ctx, d.Id(), definition); err != nil {
		return diag.Errorf("failed to update Kubernetes node pool: %s", err)
	}

	nodePool, err := awaitKubernetesNodePoolReconciliation(ctx, a, d.Id())
	if err != nil {
		return diag.Errorf("failed awaiting Kubernetes node pool reconciliation: %s", err)
	}

	return setResourceDataFromKubernetesNodePool(d, nodePool)
}

func resourceKubernetesNodePoolDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := newKubernetesServiceAPI[kubernetesNodePool](m.(providerContext).legacyClient, "node_pool")

	if err := a.Delete(ctx, d.Id()); err != nil && !isLegacyNotFoundError(err) {
		return diag.Errorf("failed deleting Kubernetes node pool: %s", err)
	}

	if err := awaitKubernetesNodePoolDeletion(ctx, a, d.Id()); err != nil {
		return diag.Errorf("failed awaiting Kubernetes node pool deletion: %s", err)
	}

	d.SetId("")
	return nil
}

func kubernetesNodePoolCreateDefinition(d *schema.ResourceData) map[string]any {
	return map[string]any{
		"name":             d.Get("name").(string),
		"replicas":         d.Get("initial_replicas").(int),
		"cpus":             d.Get("cpus").(int),
		"memory":           int64(d.Get("memory_gib").(int)) * gibiFactor,
		"disk_size":        int64(d.Get("disk").([]any)[0].(map[string]any)["size_gib"].(int)) * gibiFactor,
		"operating_system": d.Get("operating_system").(string),
		"cluster":          d.Get("cluster").(string),
	}
}

func kubernetesNodePoolCreateUpdateDefinition(d *schema.ResourceData) map[string]any {
	return kubernetesNodePoolCreateUpdateDefinitionWithConfiguredFields(d, func(field string) bool {
		return kubernetesFieldIsConfigured(d, field)
	})
}

func kubernetesNodePoolCreateUpdateDefinitionWithConfiguredFields(d *schema.ResourceData, isConfigured func(string) bool) map[string]any {
	definition := make(map[string]any)

	// The service currently rejects some advertised v2 fields on POST. Apply
	// explicitly configured v2 fields immediately through the PATCH endpoint.
	optionalFields := map[string]string{
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
	for terraformName, apiName := range optionalFields {
		if isConfigured(terraformName) {
			definition[apiName] = d.Get(terraformName)
		}
	}

	primaryDisk := d.Get("disk").([]any)[0].(map[string]any)
	if isConfigured("disk") {
		performanceType := primaryDisk["performance_type"].(string)
		if performanceType != "" {
			definition["disk_performance_type"] = performanceType
		}
	}
	if disks := kubernetesNodePoolAdditionalDisksDefinition(d); isConfigured("additional_disks") && len(disks) != 0 {
		definition["additional_disks"] = disks
	}
	if networks := kubernetesNodePoolNetworksDefinition(d); isConfigured("networks") && len(networks) != 0 {
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
	set("cpu_performance_type", nodePool.CPUPerformanceType.ID)
	set("autoscaler_enabled", nodePool.AutoscalerEnabled)
	set("autoscaler_min_nodes", nodePool.AutoscalerMinNodes)
	set("autoscaler_max_nodes", nodePool.AutoscalerMaxNodes)
	set("dns_override_ipv4", nodePool.DNSOverrideIPv4)
	set("dns_ipv4_1", nodePool.DNSv4Entry1)
	set("dns_ipv4_2", nodePool.DNSv4Entry2)
	set("dns_override_ipv6", nodePool.DNSOverrideIPv6)
	set("dns_ipv6_1", nodePool.DNSv6Entry1)
	set("dns_ipv6_2", nodePool.DNSv6Entry2)
	set("ssh_public_keys", nodePool.SSHPublicKeys)
	set("taints", nodePool.Taints)
	set("labels", nodePool.Labels)
	set("annotations", nodePool.Annotations)
	set("state", nodePool.State.ID)
	set("state_text", nodePool.State.Text)

	set("disk", []map[string]any{{
		"size_gib":         nodePool.DiskSizeBytes / gibiFactor,
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
	a kubernetesServiceAPI[kubernetesNodePool],
	identifier string,
) (kubernetesNodePool, error) {
	for {
		nodePool, err := a.Get(ctx, identifier)
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
	a kubernetesServiceAPI[kubernetesNodePool],
	identifier string,
) error {
	for {
		_, err := a.Get(ctx, identifier)
		if isLegacyNotFoundError(err) {
			return nil
		}
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(kubernetesReconciliationPollInterval):
		}
	}
}
