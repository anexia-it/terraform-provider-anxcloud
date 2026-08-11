package anxcloud

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
)

const kubernetesReconciliationPollInterval = 10 * time.Second

func resourceKubernetesCluster() *schema.Resource {
	clusterSchema := schemaKubernetesCluster()
	clusterSchema["wait_until_ready"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Wait for the cluster to reach the ready state during create and update. Intermediate reconciliation states, including Error, do not stop the wait. State transitions are written to provider logs.",
	}

	return &schema.Resource{
		Description: "Resource to create and update Kubernetes clusters.",

		CreateContext: tagsMiddlewareCreate(resourceKubernetesClusterCreate),
		ReadContext:   tagsMiddlewareRead(resourceKubernetesClusterRead),
		UpdateContext: tagsMiddlewareUpdate(resourceKubernetesClusterUpdate),
		DeleteContext: resourceKubernetesClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Read:   schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: withTagsAttribute(clusterSchema),
	}
}

func resourceKubernetesClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	cluster := kubernetesCluster{requestDefinition: kubernetesClusterCreateDefinition(d)}
	if err := a.Create(ctx, &cluster); err != nil {
		return diag.Errorf("failed to create Kubernetes cluster: %s", err)
	}

	d.SetId(cluster.Identifier)
	if definition := kubernetesClusterCreateUpdateDefinition(d); len(definition) != 0 {
		cluster.requestDefinition = definition
		if err := a.Update(ctx, &cluster); err != nil {
			return diag.Errorf("failed to apply Kubernetes cluster v2 fields after creation: %s", err)
		}
	}

	if d.Get("wait_until_ready").(bool) {
		var err error
		cluster, err = awaitKubernetesClusterReconciliation(ctx, a, cluster.Identifier)
		if err != nil {
			return diag.Errorf("failed awaiting Kubernetes cluster reconciliation: %s", err)
		}
	}

	return setResourceDataFromKubernetesClusterV2(d, cluster)
}

func resourceKubernetesClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	cluster, err := getKubernetesCluster(ctx, a, d.Id())
	if err != nil {
		if api.IgnoreNotFound(err) == nil {
			d.SetId("")
			return nil
		}
		return diag.Errorf("failed getting Kubernetes cluster: %s", err)
	}

	// A refresh must succeed even when reconciliation failed. Keeping the
	// remote object in state allows Terraform to plan a repair or deletion.
	return setResourceDataFromKubernetesClusterV2(d, cluster)
}

func resourceKubernetesClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	definition := kubernetesClusterUpdateDefinition(d)
	if len(definition) == 0 {
		cluster, err := getKubernetesCluster(ctx, a, d.Id())
		if err != nil {
			return diag.Errorf("failed getting Kubernetes cluster: %s", err)
		}
		if d.Get("wait_until_ready").(bool) {
			cluster, err = awaitKubernetesClusterReconciliation(ctx, a, d.Id())
			if err != nil {
				return diag.Errorf("failed awaiting Kubernetes cluster reconciliation: %s", err)
			}
		}
		return setResourceDataFromKubernetesClusterV2(d, cluster)
	}

	cluster := kubernetesCluster{Identifier: d.Id(), requestDefinition: definition}
	if err := a.Update(ctx, &cluster); err != nil {
		return diag.Errorf("failed to update Kubernetes cluster: %s", err)
	}

	if d.Get("wait_until_ready").(bool) {
		var err error
		cluster, err = awaitKubernetesClusterReconciliation(ctx, a, d.Id())
		if err != nil {
			return diag.Errorf("failed awaiting Kubernetes cluster reconciliation: %s", err)
		}
	}

	return setResourceDataFromKubernetesClusterV2(d, cluster)
}

func resourceKubernetesClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	a := apiFromProviderConfig(m)
	cluster := kubernetesCluster{Identifier: d.Id()}

	if err := a.Destroy(ctx, &cluster); err != nil && api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed deleting Kubernetes cluster: %s", err)
	}

	if err := awaitKubernetesClusterDeletion(ctx, a, d.Id()); err != nil {
		return diag.Errorf("failed awaiting Kubernetes cluster deletion: %s", err)
	}

	d.SetId("")
	return nil
}

func kubernetesClusterCreateDefinition(d *schema.ResourceData) map[string]any {
	definition := map[string]any{
		"name":                d.Get("name").(string),
		"location":            d.Get("location").(string),
		"needs_service_vms":   d.Get("needs_service_vms").(bool),
		"enable_nat_gateways": d.Get("enable_nat_gateways").(bool),
		"enable_lbaas":        d.Get("enable_lbaas").(bool),
		"autoscaling":         d.Get("enable_autoscaling").(bool),
	}

	if version := d.Get("version").(string); version != "" {
		definition["version"] = version
	}
	if allowlist := kubernetesAPIServerAllowlist(d); allowlist != "" {
		definition["apiserver_allowlist"] = allowlist
	}

	setKubernetesClusterPrefixDefinitionFields(definition, d, false)
	return definition
}

func kubernetesClusterCreateUpdateDefinition(d *schema.ResourceData) map[string]any {
	return kubernetesClusterCreateUpdateDefinitionWithConfiguredFields(d, func(field string) bool {
		return kubernetesFieldIsConfigured(d, field)
	})
}

func kubernetesClusterCreateUpdateDefinitionWithConfiguredFields(d *schema.ResourceData, isConfigured func(string) bool) map[string]any {
	// The service currently rejects some advertised v2 fields on POST. Apply
	// explicitly configured v2 fields immediately through the PATCH endpoint.
	definition := make(map[string]any)
	fields := map[string]string{
		"cni_plugin":                    "cni_plugin",
		"enable_persistent_storage":     "enable_persistent_storage",
		"external_ip_families":          "external_ip_families",
		"enable_oidc_authentication":    "enable_oidc_authentication",
		"oidc_client_id":                "oidc_client_id",
		"oidc_issuer_url":               "oidc_issuer_url",
		"oidc_groups_claim":             "oidc_groups_claim",
		"oidc_username_claim":           "oidc_username_claim",
		"oidc_extra_scopes":             "oidc_extra_scopes",
		"oidc_groups_prefix":            "oidc_groups_prefix",
		"oidc_required_claim":           "oidc_required_claim",
		"oidc_username_prefix":          "oidc_username_prefix",
		"maintenance_window_start_time": "maintenance_window_start_time",
		"maintenance_window_duration":   "maintenance_window_duration",
	}
	for terraformName, apiName := range fields {
		if isConfigured(terraformName) {
			definition[apiName] = d.Get(terraformName)
		}
	}
	return definition
}

func kubernetesFieldIsConfigured(d *schema.ResourceData, field string) bool {
	config := d.GetRawConfig()
	if config.IsNull() || !config.IsKnown() {
		return false
	}

	value := config.GetAttr(field)
	return value.IsKnown() && !value.IsNull()
}

func kubernetesClusterUpdateDefinition(d *schema.ResourceData) map[string]any {
	definition := make(map[string]any)
	fields := map[string]string{
		"name":                          "name",
		"version":                       "version",
		"location":                      "location",
		"needs_service_vms":             "needs_service_vms",
		"enable_nat_gateways":           "enable_nat_gateways",
		"enable_lbaas":                  "enable_lbaas",
		"enable_autoscaling":            "autoscaling",
		"cni_plugin":                    "cni_plugin",
		"enable_persistent_storage":     "enable_persistent_storage",
		"external_ip_families":          "external_ip_families",
		"enable_oidc_authentication":    "enable_oidc_authentication",
		"maintenance_window_start_time": "maintenance_window_start_time",
		"maintenance_window_duration":   "maintenance_window_duration",
	}
	for terraformName, apiName := range fields {
		if d.HasChange(terraformName) {
			definition[apiName] = d.Get(terraformName)
		}
	}
	if d.HasChange("apiserver_allowlist") {
		definition["apiserver_allowlist"] = kubernetesAPIServerAllowlist(d)
	}

	setKubernetesClusterPrefixDefinitionFields(definition, d, true)
	setChangedKubernetesClusterOIDCFields(definition, d)
	return definition
}

func setKubernetesClusterPrefixDefinitionFields(definition map[string]any, d *schema.ResourceData, changedOnly bool) {
	prefixes := map[string]string{
		"internal_ipv4_prefix": "manage_internal_ipv4_prefix",
		"external_ipv4_prefix": "manage_external_ipv4_prefix",
		"external_ipv6_prefix": "manage_external_ipv6_prefix",
	}
	for field, manageField := range prefixes {
		if changedOnly && !d.HasChange(field) {
			continue
		}
		if value, ok := d.GetOk(field); ok {
			definition[field] = value.(string)
			definition[manageField] = false
		}
	}
}

func setChangedKubernetesClusterOIDCFields(definition map[string]any, d *schema.ResourceData) {
	for _, field := range []string{
		"oidc_client_id",
		"oidc_issuer_url",
		"oidc_groups_claim",
		"oidc_username_claim",
		"oidc_extra_scopes",
		"oidc_groups_prefix",
		"oidc_required_claim",
		"oidc_username_prefix",
	} {
		if d.HasChange(field) {
			definition[field] = d.Get(field).(string)
		}
	}
}

func kubernetesAPIServerAllowlist(d *schema.ResourceData) string {
	values := d.Get("apiserver_allowlist").([]interface{})
	allowlist := make([]string, 0, len(values))
	for _, value := range values {
		if cidr := strings.TrimSpace(value.(string)); cidr != "" {
			allowlist = append(allowlist, cidr)
		}
	}
	return strings.Join(allowlist, " ")
}

func awaitKubernetesClusterReconciliation(
	ctx context.Context,
	a api.API,
	identifier string,
) (kubernetesCluster, error) {
	return awaitKubernetesClusterReconciliationWithPollInterval(
		ctx,
		a,
		identifier,
		kubernetesReconciliationPollInterval,
	)
}

func awaitKubernetesClusterReconciliationWithPollInterval(
	ctx context.Context,
	a api.API,
	identifier string,
	pollInterval time.Duration,
) (kubernetesCluster, error) {
	lastState := ""

	for {
		cluster, err := getKubernetesCluster(ctx, a, identifier)
		if err != nil {
			return cluster, err
		}

		currentState := fmt.Sprintf("%s (%s)", cluster.State.Text, cluster.State.ID)
		if currentState != lastState {
			log.Printf("[INFO] Kubernetes cluster %q reconciliation state: %s", identifier, currentState)
			lastState = currentState
		}

		if cluster.State.ID == "0" {
			return cluster, nil
		}

		select {
		case <-ctx.Done():
			return cluster, fmt.Errorf("waiting for Kubernetes cluster %q to become ready; last observed state was %s: %w", identifier, currentState, ctx.Err())
		case <-time.After(pollInterval):
		}
	}
}

func awaitKubernetesClusterDeletion(
	ctx context.Context,
	a api.API,
	identifier string,
) error {
	for {
		_, err := getKubernetesCluster(ctx, a, identifier)
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

func getKubernetesCluster(ctx context.Context, a api.API, identifier string) (kubernetesCluster, error) {
	cluster := kubernetesCluster{Identifier: identifier}
	if err := a.Get(ctx, &cluster); err != nil {
		return cluster, err
	}
	return cluster, nil
}
