package anxcloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceKubernetesClusterImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return importKubernetesResource(d, "cluster")
}

func resourceKubernetesNodePoolImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	return importKubernetesResource(d, "node pool")
}

func importKubernetesResource(d *schema.ResourceData, resourceName string) ([]*schema.ResourceData, error) {
	importID := d.Id()
	apiEnvironment := kubernetesAPIEnvironmentProd
	identifier := importID

	if environment, resourceID, found := strings.Cut(importID, ":"); found {
		apiEnvironment = environment
		identifier = resourceID
	}

	switch apiEnvironment {
	case kubernetesAPIEnvironmentProd, kubernetesAPIEnvironmentStage, kubernetesAPIEnvironmentDev:
	default:
		return nil, fmt.Errorf(
			"invalid Kubernetes %s import ID %q: API environment must be one of prod, stage, or dev",
			resourceName,
			importID,
		)
	}
	if identifier == "" || strings.Contains(identifier, ":") {
		identifierName := strings.ReplaceAll(resourceName, " ", "-") + "-id"
		return nil, fmt.Errorf(
			"invalid Kubernetes %s import ID %q: expected <%s> or <api-environment>:<%s>",
			resourceName,
			importID,
			identifierName,
			identifierName,
		)
	}
	if err := d.Set("api_environment", apiEnvironment); err != nil {
		return nil, fmt.Errorf("failed setting Kubernetes API environment during import: %w", err)
	}

	d.SetId(identifier)
	return []*schema.ResourceData{d}, nil
}
