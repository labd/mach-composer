# Changelog

## 1.0.0 (unreleased)

### New platforms
- Add Amplience support
- Add Apollo Federation support
- Add Sentry DSN management options

**General**
- Add `mach_composer` configuration block to configure required MACH composer
  version
- Add `--ignore-version` to disable the MACH composer version check
- Improved development workflow:
  - Improved git log parsing
  - Add `mach bootstrap` commands:
      - `mach bootstrap config` for creating a new MACH configuration
      - `mach bootstrap component` for creating a new MACH component
  - Add `--site` option to the `generate`, `plan` and `apply` commands
  - Add `--component` option to the `plan` and `apply` commands
  - Add `--reuse` flag to the `plan` and `apply` commands to suppress a
    `terraform init` call
  - Add support for relative paths to components
  - Add extra component definition settings `artifacts` to facilitate local
    deployments
- Improved dependencies between components and MACH-managed commercetools
  configurations
- Add option to override Terraform provider versions
- Add support for multiple API endpoints:
    - `base_url` replaced with `endpoints`
    - `has_public_api` replaced with `endpoints`
    - Supports a `default` endpoint that doesn't require custom domain
      settings

**commercetools**
- Move `currencies`, `languages`, `countries`, `messages_enabled` to `project_settings` configuration block
- Add support for commercetools Store-specific variables and secrets on
  components included in new variable: `ct_stores`
- Add `managed` setting to commercetools store. Set to false it will indicate the store should not be managed by MACH composer
- Add support for commercetools shipping zones
- Make commercetools frontend API client scopes configurable with new
  `frontend` configuration block

**AWS**
- AWS: Set `auto-deploy` on API gateway stage
- AWS: Add new component variable `tags`

**Azure**
- Add configuration options for Azure service plans
- Upgraded Terraform to `0.14.5`
- Upgraded Terraform commercetools provider to `0.25.3`
- Upgraded Terraform AWS provider to `3.28.0`
- Upgraded Terraform Azure provider to `2.47.0`
- Azure: Remove `project_key` from `var.tags` and add `Environment` and `Site`
- Azure: Add `--with-sp-login` option to `mach plan` command
- Azure: Remove function app sync bash command: this is now the
  responsibility of the component


### Breaking changes

**Generic**

- **config**: Rename `general_config` to `global`
- **config**: `base_url` has been replaced by the `endpoints` settings:<br>
  ```yaml
  sites:
  - identifier: mach-site-eu
    base_url: https://api.eu-tst.mach-example.net
  ```
  becomes
  ```yaml
  sites:
  - identifier: mach-site-eu
    endpoints:
      main: https://api.eu-tst.mach-example.net
  ```
  When you name the endpoint that replaces `base_url` "main", it will have
  the least effect on your existing Terraform state.<br><br>
  When endpoints are defined on a component, the component needs to define
  endpoint Terraform variables
  ([AWS](https://docs.machcomposer.io/reference/components/aws.html#with-endpoints)
  and
  [Azure](https://docs.machcomposer.io/reference/components/azure.html#with-endpoints))
- **config**: commercetools `create_frontend_credentials` is replaced with new `frontend` block:
  ```terraform
  commercetools:
    frontend:
      create_credentials: false
  ```
  default is still `true`
- **config** Default scopes for commercetools frontend API client changed:
    - If you want to maintain previous scope set, define the following in the `frontend` block:
    ```terraform
    commercetools:
      frontend:
        permission_scopes: [manage_my_profile, manage_my_orders, view_states, manage_my_shopping_lists, view_products, manage_my_payments, create_anonymous_token, view_project_settings]
    ```
    - Old scope set didn't include store-specific [`manage_my_profile:project:store`](https://docs.commercetools.com/api/scopes#manage_my_profileprojectkeystorekey) scope. If you're using the old set as described above, MACH will need to re-create the store-specific API clients in order to add the extra scope. For migration options, see next point
    - In case the scope needs to be updated but (production) frontend
      implementations are already using the current API client credentials, a
      way to migrate is to;
      1. Remove the old API client resource with `terraform state rm commercetools_api_client.frontend_credentials`
      2. Repeat step for the store-specific API clients in your Terraform
         state
      3. Perform `mach apply` to create the new API clients with updated
         scope
      4. Your commercetools project will now contain API clients with the
         same name. Once the frontend implementation is migrated, the older one
         can safely be removed.
- **component**: Components with a `commercetools` integration require a new variable `ct_stores`:
  ```terraform
  variable "ct_stores" {
    type = map(object({
      key       = string
      variables = map(string)
      secrets   = map(string)
    }))
    default = {}
  }
  ```
- **component**: The folowing deprecated values in the `var.variables` are removed:
  ```terraform
  var.variables["CT_PROJECT_KEY"]
  var.variables["CT_API_URL"]
  var.variables["CT_AUTH_URL"]
  ```
  See [0.5.0 release notes](#050-2020-11-09)
- **component**: The `var.environment_variables` won't be set by MACH anymore. Use `var.variables` for this


**AWS**

- **config**: The AWS `route53_zone_name` setting has been removed in favour of multiple endpoint support
- **config**: The `deploy_role` setting has been renamed to `deploy_role_name`
- **component**: Introduced new variable `tags`:
  ```terraform
  variable "tags" {
    type        = map(string)
    description = "Tags to be used on resources."
  }
  ```
- **component**: Add `aws_endpoint_*` variable when the `endpoints` configuration option is used. [More information](https://docs.machcomposer.io/reference/components/aws.html#with-endpoints) on defining and using endpoints in AWS.

**Azure**

- **config**: The `front_door` configuration block has been renamed to `frontdoor`
- **config**: The Azure frontdoor settings `dns_zone` and `ssl_key_*` settings have been removed;<br>
  Certificates are now managed by Frontdoor and dns_zone is auto-detected.
- **config**: Moved component `short_name` to new `azure` configuration block
- **state**: The Terraform `azurerm_dns_cname_record` resources have been renamed; they now take the name of the associated endpoint key. For the smoothest transition, rename them in your Terraform state:<br>
  ```bash
  terraform state mv azurerm_dns_cname_record.<project-key> azurerm_dns_cname_record.<endpoint-key>
  ```
- **component**: Prefixed **all** Azure-specific variables with `azure_`
- **component**: The `FRONTDOOR_ID` value is removed from the `var.variables`
  of a component. Replaced with `var.azure_endpoint_*`. [More
  information](https://docs.machcomposer.io/reference/components/azure.html#with-endpoints)
  on defining and using endpoints in Azure.
- **component**: `app_service_plan_id` has been replaced with
  `azure_app_service_plan` containing both an `id` and `name` so the
  [azurerm_app_service_plan](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/app_service_plan)
  data source can be used in a component.<br>

  It will *only be set* when `service_plan` is configured in the component
  [definition](https://docs.machcomposer.io/reference/syntax/components.html#azure)
  or [site configuration](https://docs.machcomposer.io/reference/syntax/sites.html#azure_1)
  ```terraform
  variable "azure_app_service_plan" {
    type = object({
      id                  = string
      name                = string
      resource_group_name = string
    })
  }
  ```
- **component**: Replaced `resource_group_name` and `resource_group_location` with `azure_resource_group`:
  ```terraform
  variable "azure_resource_group" {
    type = object({
      name     = string
      location = string
    })
  }
  ```

## 0.5.1 (2020-11-10)
- Removed `aws` block in general_config
- Add `branch` option to component definitions to be able to perform a `mach
  update` and stay within a certain branch (during development)


## 0.5.0 (2020-11-09)
- Add new CLI options:
    - `mach components` to list all components
    - `mach sites` to list all sites
- Improved `update` command:
    - Supports updating (or checking for updates) on all components based on their git history
    - This can now also be used to manually update a single component; `mach update my-component v1.0.4`
    - Add `--commit` argument to automatically create a git commit message
- Add new AWS configuration option `route53_zone_name`
- Remove unused `api_gateway` attribute on AWS config
- Remove restriction from `environment` value; can now be any. Fixes #9

### Breaking changes

- Require `ct_api_url` and `ct_auth_url` for components with `commercetools` integration

### Deprecations

In a component, the use of the following variables have been deprecated;

```
var.variables["CT_PROJECT_KEY"]
var.variables["CT_API_URL"]
var.variables["CT_AUTH_URL"]
```

Instead you should use:

```
var.ct_project_key
var.ct_api_url
var.ct_auth_url
```

## 0.4.3 (2020-11-04)
- Make AWS role definitions optional so MACH can run without an 'assume role' context


## 0.4.2 (2020-11-02)
- Add 'encrypt' option to AWS state backend
- Correctly depend component modules to the commercetools project settings resource
- Extend Azure regions mapping


## 0.4.1 (2020-10-27)
- Fixed TypeError when using `resource_group` on site Azure configuration


## 0.4.0 (2020-10-27)
- Add Contentful support

### Breaking changes

- `is_software_component` has been replaced by the `integrations` settings

```
components:
  - name: my-product-types
    source: git::ssh://git@github.com/example/product-types/my-product-types.git
    version: 81cd828
    is_software_component: false
```

becomes

```
components:
  - name: my-product-types
    source: git::ssh://git@github.com/example/product-types/my-product-types.git
    version: 81cd828
    integrations: ["commercetools"]
```

or `integrations: []` if no integrations are needed at all.


## 0.3.0 (2020-10-21)
- Add option to specify custom resource group per site

### Breaking changes

- All `resource_group_name` attributes is renamed to `resource_group`
- The `storage_account_name` attribute is renamed to `storage_account`


## 0.2.2 (2020-10-15)
- Fixed Azure config merge: not all generic settings where merged with
  site-specific ones
- Only validate short_name length check for Azure implementations
- Setup Frontdoor per 'public api' component regardless of global Frontdoor
  settings


## 0.2.1 (2020-10-06)
- Fixed rendering of STORE environment variables in components
- Updated Terraform version to 0.13.4
- Fix `--auto-approve` option on `mach apply` command


0.2.0 (2020-10-06)
=================
- Add AWS support
- Add new required attribute `cloud` in general config


## 0.1.0 (2020-10-01)
- Initial release
