// NOTE for customers using the LEGACY dbt_cloud provider:
// use dbt_cloud_environment_variable instead of dbtcloud_environment_variable for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_environment_variable" "dbt_my_env_var" {
  name       = "DBT_MY_ENV_VAR"
  project_id = dbtcloud_project.dbt_project.id
  environment_values = {
    "project" : "my_project_level_value",
    "Dev" : "my_env_level_value",
    "CI" : "my_ci_override_value",
    "Prod" "my_prod_override_value"
  }
  depends_on = [
    dbtcloud_project.dbt_project,
    dbtcloud_environment.dev_env,
    dbtcloud_environment.ci_env,
    dbtcloud_environment.prod_env,
  ]
}