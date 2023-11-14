package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceRepository() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRepositoryCreate,
		ReadContext:   resourceRepositoryRead,
		UpdateContext: resourceRepositoryUpdate,
		DeleteContext: resourceRepositoryDelete,

		Schema: map[string]*schema.Schema{
			"repository_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Repository Identifier",
			},
			"is_active": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the repository is active",
			},
			"project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID to create the repository in",
			},
			"remote_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Git URL for the repository or \\<Group>/\\<Project> for Gitlab",
			},
			"git_clone_strategy": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "deploy_key",
				ForceNew:    true,
				Description: "Git clone strategy for the repository. Can be `deploy_key` (default) for cloning via SSH Deploy Key, `github_app` for GitHub native integration, `deploy_token` for the GitLab native integration and `azure_active_directory_app` for ADO native integration",
			},
			"repository_credentials_id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Credentials ID for the repository (From the repository side not the dbt Cloud ID)",
			},
			"gitlab_project_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Identifier for the Gitlab project -  (for GitLab native integration only)",
			},
			"github_installation_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Identifier for the GitHub App - (for GitHub native integration only)",
			},
			"azure_active_directory_project_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
				Description: "The Azure Dev Ops project ID. It can be retrieved via the Azure API or using the data source `dbtcloud_azure_dev_ops_project` and the project name - (for ADO native integration only)",
			},
			"azure_active_directory_repository_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
				Description: "The Azure Dev Ops repository ID. It can be retrieved via the Azure API or using the data source `dbtcloud_azure_dev_ops_repository` along with the ADO Project ID and the repository name - (for ADO native integration only)",
			},
			"azure_bypass_webhook_registration_failure": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "If set to False (the default), the connection will fail if the service user doesn't have access to set webhooks (required for auto-triggering CI jobs). If set to True, the connection will be successful but no automated CI job will be triggered - (for ADO native integration only)",
			},
			"fetch_deploy_key": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether we should return the public deploy key - (for the `deploy_key` strategy)",
			},
			"deploy_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Public key generated by dbt when using `deploy_key` clone strategy",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRepositoryCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	isActive := d.Get("is_active").(bool)
	projectId := d.Get("project_id").(int)
	remoteUrl := d.Get("remote_url").(string)
	gitCloneStrategy := d.Get("git_clone_strategy").(string)
	gitlabProjectID := d.Get("gitlab_project_id").(int)
	githubInstallationID := d.Get("github_installation_id").(int)
	azureActiveDirectoryProjectID := d.Get("azure_active_directory_project_id").(string)
	azureActiveDirectoryRepositoryID := d.Get("azure_active_directory_repository_id").(string)
	azureBypassWebhookRegistrationFailure := d.Get("azure_bypass_webhook_registration_failure").(bool)

	repository, err := c.CreateRepository(
		projectId,
		remoteUrl,
		isActive,
		gitCloneStrategy,
		gitlabProjectID,
		githubInstallationID,
		azureActiveDirectoryProjectID,
		azureActiveDirectoryRepositoryID,
		azureBypassWebhookRegistrationFailure,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%d%s%d", repository.ProjectID, dbt_cloud.ID_DELIMITER, *repository.ID))

	resourceRepositoryRead(ctx, d, m)

	return diags
}

func resourceRepositoryRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	repositoryIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]
	fetchDeployKey := d.Get("fetch_deploy_key").(bool)

	repository, err := c.GetRepository(repositoryIdString, projectIdString, fetchDeployKey)
	if err != nil {
		if strings.HasPrefix(err.Error(), "resource-not-found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("repository_id", repository.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_active", repository.State == dbt_cloud.STATE_ACTIVE); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project_id", repository.ProjectID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("remote_url", repository.RemoteUrl); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("git_clone_strategy", repository.GitCloneStrategy); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("repository_credentials_id", repository.RepositoryCredentialsID); err != nil {
		return diag.FromErr(err)
	}
	// CC-709: we currently don't get back the GitLab project ID from the API
	if err := d.Set("gitlab_project_id", d.Get("gitlab_project_id").(int)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("github_installation_id", repository.GithubInstallationID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("deploy_key", repository.DeployKey.PublicKey); err != nil {
		return diag.FromErr(err)
	}
	// the following values are not sent back by the API so we set them as they are in the config
	if err := d.Set("azure_active_directory_project_id", d.Get("azure_active_directory_project_id").(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_active_directory_repository_id", d.Get("azure_active_directory_repository_id").(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_bypass_webhook_registration_failure", d.Get("azure_bypass_webhook_registration_failure").(bool)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceRepositoryUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	repositoryIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]
	fetchDeployKey := d.Get("fetch_deploy_key").(bool)

	if d.HasChange("is_active") {
		repository, err := c.GetRepository(repositoryIdString, projectIdString, fetchDeployKey)
		if err != nil {
			return diag.FromErr(err)
		}

		if d.HasChange("is_active") {
			isActive := d.Get("is_active").(bool)
			if isActive {
				repository.State = dbt_cloud.STATE_ACTIVE
			} else {
				repository.State = dbt_cloud.STATE_DELETED
			}
		}

		_, err = c.UpdateRepository(repositoryIdString, projectIdString, *repository)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceRepositoryRead(ctx, d, m)
}

func resourceRepositoryDelete(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	projectIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[0]
	repositoryIdString := strings.Split(d.Id(), dbt_cloud.ID_DELIMITER)[1]

	_, err := c.DeleteRepository(repositoryIdString, projectIdString)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}
