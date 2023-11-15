package anxcloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	e5ev1internal "github.com/anexia-it/terraform-provider-anxcloud/anxcloud/internal/apis/e5e/v1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"go.anx.io/go-anxcloud/pkg/api"
	e5ev1 "go.anx.io/go-anxcloud/pkg/apis/e5e/v1"
)

func resourceE5EFunction() *schema.Resource {
	var availableStorageBackends = []string{"storage_backend_git", "storage_backend_archive", "storage_backend_s3"}
	return &schema.Resource{
		Description:   "A function is the collection of all the metadata as well as the code itself that is needed to execute your application on the e5e platform.",
		CreateContext: resourceE5EFunctionCreate,
		ReadContext:   resourceE5EFunctionRead,
		UpdateContext: resourceE5EFunctionUpdate,
		DeleteContext: resourceE5EFunctionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Function identifier.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Function name.",
			},
			"application": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Functions application assignment.",
			},
			"runtime": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Function runtime.",
			},
			"entrypoint": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Function entrypoint.",
			},
			"storage_backend_s3": {
				Type:        schema.TypeList,
				Description: "S3 storage backend configuration.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint":    {Type: schema.TypeString, Required: true},
						"bucket_name": {Type: schema.TypeString, Required: true},
						"object_path": {Type: schema.TypeString, Optional: true},
						"access_key":  {Type: schema.TypeString, Optional: true, Sensitive: true},
						"secret_key":  {Type: schema.TypeString, Optional: true, Sensitive: true},
					},
				},
				ConflictsWith: []string{"storage_backend_git", "storage_backend_archive"},
				ExactlyOneOf:  availableStorageBackends,
			},
			"storage_backend_git": {
				Type:        schema.TypeList,
				Description: "Git storage backend configuration.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url":         {Type: schema.TypeString, Required: true},
						"branch":      {Type: schema.TypeString, Optional: true},
						"private_key": {Type: schema.TypeString, Optional: true, Sensitive: true},
						"username":    {Type: schema.TypeString, Optional: true},
						"password":    {Type: schema.TypeString, Optional: true, Sensitive: true},
					},
				},
				ConflictsWith: []string{"storage_backend_s3", "storage_backend_archive"},
				ExactlyOneOf:  availableStorageBackends,
			},
			"storage_backend_archive": {
				Type:        schema.TypeList,
				Description: "Archive storage backend configuration.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content": {Type: schema.TypeString, Required: true},
						"name":    {Type: schema.TypeString, Required: true},
					},
				},
				ConflictsWith: []string{"storage_backend_s3", "storage_backend_git"},
				ExactlyOneOf:  availableStorageBackends,
			},
			"env": {
				Type: schema.TypeList,
				Description: "Environment variables available to the function." +
					" Note: the provider is not aware of external changes to secret environment variables.",
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secret": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"hostname": {
				Type:        schema.TypeList,
				Description: "Custom host entries that are available when running your function. These hostnames can override existing DNS entries.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"keep_alive": {
				Type:        schema.TypeInt,
				Description: "Keep-alive time.",
				Optional:    true,
				Computed:    true,
			},
			"quota_storage": {
				Type:        schema.TypeInt,
				Description: "Space in MiB e5e will grant your function to write any sort of files.",
				Optional:    true,
				Computed:    true,
			},
			"quota_memory": {
				Type:        schema.TypeInt,
				Description: "Memory in MiB e5e will grant your function.",
				Optional:    true,
				Computed:    true,
			},
			"quota_cpu": {
				Type:        schema.TypeInt,
				Description: "CPU time in percent the e5e platform will grant your function on execution.",
				Optional:    true,
				Computed:    true,
			},
			"quota_timeout": {
				Type:        schema.TypeInt,
				Description: "Time in seconds your function can take to execute before it is killed by the e5e platform.",
				Optional:    true,
				Computed:    true,
			},
			"quota_concurrency": {
				Type:        schema.TypeInt,
				Description: "Number of parallel executions of the function there can be.",
				Optional:    true,
				Computed:    true,
			},
			"worker_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"revision": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Revision is an optional attribute which can be used to trigger a new deployment." +
					" The value can be any arbitrary string (e.g. `COMMIT_SHA` or md5 hash of the code binary passed in via variables).",
			},
		},
	}
}

func resourceE5EFunctionCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	function := e5eFunctionFromResourceData(d)

	if err := a.Create(ctx, &function); err != nil {
		return diag.Errorf("failed to create resource: %s", err)
	}

	d.SetId(function.Identifier)

	if d.HasChange("revision") {
		if err := resourceE5EFunctionDeploy(ctx, a, function.Identifier); err != nil {
			return diag.Errorf("deploy function: %s", err)
		}
	}

	return resourceE5EFunctionRead(ctx, d, m)
}

func resourceE5EFunctionRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	function := e5ev1.Function{Identifier: d.Id()}
	if err := a.Get(ctx, &function); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed getting resource: %s", err)
	} else if err != nil {
		d.SetId("")
		return nil
	}

	return e5eFunctionToResourceData(function, d)
}

func resourceE5EFunctionUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	function := e5eFunctionFromResourceData(d)

	if err := a.Update(ctx, &function); err != nil {
		return diag.Errorf("failed to update resource: %s", err)
	}

	if d.HasChange("revision") {
		if err := resourceE5EFunctionDeploy(ctx, a, function.Identifier); err != nil {
			return diag.Errorf("deploy function: %s", err)
		}
	}

	return resourceE5EFunctionRead(ctx, d, m)
}

func resourceE5EFunctionDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	a := apiFromProviderConfig(m)

	if err := a.Destroy(ctx, &e5ev1.Function{Identifier: d.Id()}); api.IgnoreNotFound(err) != nil {
		return diag.Errorf("failed to delete resource: %s", err)
	}

	return nil
}

func resourceE5EFunctionDeploy(ctx context.Context, a api.API, id string) error {
	if err := a.Create(ctx, &e5ev1internal.E5EFunctionDeployment{FunctionIdentifier: id}); err != nil {
		return err
	}
	return retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		function := e5ev1.Function{Identifier: id}
		if err := a.Get(ctx, &function); err != nil {
			return retry.NonRetryableError(err)
		}

		if function.DeploymentState == "pending" {
			return retry.RetryableError(errors.New("function deployment pending"))
		}

		if function.DeploymentState != "deployed" {
			return retry.NonRetryableError(fmt.Errorf("unexpected deployment state %q", function.DeploymentState))
		}

		return nil
	})
}

func e5eFunctionToResourceData(function e5ev1.Function, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	setVal := func(key string, val any) {
		if err := d.Set(key, val); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	setVal("name", function.Name)
	setVal("application", function.ApplicationIdentifier)
	setVal("runtime", function.Runtime)
	setVal("entrypoint", function.Entrypoint)
	setVal("keep_alive", function.KeepAlive)
	setVal("quota_storage", function.QuotaStorage)
	setVal("quota_memory", function.QuotaMemory)
	setVal("quota_cpu", function.QuotaCPU)
	setVal("quota_timeout", function.QuotaTimeout)
	setVal("quota_concurrency", function.QuotaConcurrency)
	setVal("worker_type", function.WorkerType)

	// set storage backend
	switch function.StorageBackend {
	case "s3":
		setVal("storage_backend_s3", []any{map[string]any{
			"endpoint":    function.StorageBackendMeta.StorageBackendMetaS3.Endpoint,
			"bucket_name": function.StorageBackendMeta.StorageBackendMetaS3.BucketName,
			"object_path": function.StorageBackendMeta.StorageBackendMetaS3.ObjectPath,
			"access_key":  function.StorageBackendMeta.StorageBackendMetaS3.AccessKey,
			"secret_key":  function.StorageBackendMeta.StorageBackendMetaS3.SecretKey,
		}})
	case "git":
		// private_key and password are not returned as cleartext -> take over from previous state
		var (
			previousPrivateKey = ""
			previousPassword   = ""
		)
		if previousConfig, ok := d.GetOk("storage_backend_git"); ok {
			previousConfig := previousConfig.([]any)
			if len(previousConfig) > 0 {
				previousConfig := previousConfig[0].(map[string]any)
				if val, ok := previousConfig["private_key"]; ok {
					previousPrivateKey = val.(string)
				}
				if val, ok := previousConfig["password"]; ok {
					previousPassword = val.(string)
				}
			}
		}

		setVal("storage_backend_git", []any{map[string]any{
			"url":         function.StorageBackendMeta.StorageBackendMetaGit.URL,
			"branch":      function.StorageBackendMeta.StorageBackendMetaGit.Branch,
			"private_key": previousPrivateKey,
			"username":    function.StorageBackendMeta.StorageBackendMetaGit.Username,
			"password":    previousPassword,
		}})
	case "archive":
		// do nothing
	}

	// set environment variables

	// e5e function api does not return the values of secret environment variables
	// therefore we retrieve the old environment variables to be used instead
	previousEnvMap := map[string]string{}
	for _, env := range d.Get("env").([]any) {
		env := env.(map[string]any)
		previousEnvMap[env["name"].(string)] = env["value"].(string)
	}

	var envVars []map[string]any
	for _, envVar := range *function.EnvironmentVariables {
		envVarMap := map[string]any{
			"name":   envVar.Name,
			"value":  envVar.Value,
			"secret": envVar.Secret,
		}
		if envVar.Secret {
			if val, ok := previousEnvMap[envVar.Name]; ok {
				envVarMap["value"] = val
			}
		}
		envVars = append(envVars, envVarMap)
	}
	setVal("env", envVars)

	// set hostnames
	var hostnames []map[string]any
	for _, hostname := range *function.Hostnames {
		hostnames = append(hostnames, map[string]any{
			"hostname": hostname.Hostname,
			"ip":       hostname.IP,
		})
	}
	setVal("hostname", hostnames)

	return diags
}

func e5eFunctionFromResourceData(d *schema.ResourceData) e5ev1.Function {
	function := e5ev1.Function{
		Identifier:            d.Id(),
		Name:                  d.Get("name").(string),
		ApplicationIdentifier: d.Get("application").(string),
		Runtime:               d.Get("runtime").(string),
		Entrypoint:            d.Get("entrypoint").(string),
		KeepAlive:             d.Get("keep_alive").(int),
		QuotaStorage:          d.Get("quota_storage").(int),
		QuotaMemory:           d.Get("quota_memory").(int),
		QuotaCPU:              d.Get("quota_cpu").(int),
		QuotaTimeout:          d.Get("quota_timeout").(int),
		QuotaConcurrency:      d.Get("quota_concurrency").(int),
		WorkerType:            d.Get("worker_type").(string),
		StorageBackendMeta:    &e5ev1.StorageBackendMeta{},
	}

	if envVariables, ok := d.GetOk("env"); ok {
		vars := []e5ev1.EnvironmentVariable{}
		for _, variable := range envVariables.([]any) {
			variable := variable.(map[string]any)
			vars = append(vars, e5ev1.EnvironmentVariable{
				Name:   variable["name"].(string),
				Value:  variable["value"].(string),
				Secret: variable["secret"].(bool),
			})
		}

		function.EnvironmentVariables = &vars
	}

	if hostnames, ok := d.GetOk("hostname"); ok {
		names := []e5ev1.Hostname{}
		for _, name := range hostnames.([]any) {
			name := name.(map[string]any)
			names = append(names, e5ev1.Hostname{
				Hostname: name["hostname"].(string),
				IP:       name["ip"].(string),
			})
		}

		function.Hostnames = &names
	}

	getStorageBackendMeta := func(key string) (map[string]any, bool) {
		if meta, ok := d.GetOk(key); ok {
			if meta := meta.([]any)[0]; meta != nil {
				return meta.(map[string]any), true
			}
		}
		return nil, false
	}

	if meta, ok := getStorageBackendMeta("storage_backend_s3"); ok {
		function.StorageBackend = "s3"
		function.StorageBackendMeta.StorageBackendMetaS3 = &e5ev1.StorageBackendMetaS3{
			Endpoint:   meta["endpoint"].(string),
			BucketName: meta["bucket_name"].(string),
			ObjectPath: meta["object_path"].(string),
			AccessKey:  meta["access_key"].(string),
			SecretKey:  meta["secret_key"].(string),
		}
	} else if meta, ok := getStorageBackendMeta("storage_backend_git"); ok {
		function.StorageBackend = "git"
		function.StorageBackendMeta.StorageBackendMetaGit = &e5ev1.StorageBackendMetaGit{
			URL:        meta["url"].(string),
			Branch:     meta["branch"].(string),
			PrivateKey: meta["private_key"].(string),
			Username:   meta["username"].(string),
			Password:   meta["password"].(string),
		}
	} else if meta, ok := getStorageBackendMeta("storage_backend_archive"); ok {
		function.StorageBackend = "archive"
		function.StorageBackendMeta.StorageBackendMetaArchive = &e5ev1.StorageBackendMetaArchive{
			Content: meta["content"].(string),
			Name:    meta["name"].(string),
		}
	}

	return function
}
