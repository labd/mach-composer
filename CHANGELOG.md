# Changelog

## 1.0.0 (unreleased)
- Add Sentry DSN management options
- Add Amplience support
- Add Apollo Federation support
- Add support for commercetools Store-specific variables and secrets on components: `store_variables` and `store_secrets`
- Add support for multiple API endpoints:
    - `base_url` replaced with `endpoints`
    - `has_public_api` replaced with `endpoints`
    - Supports a `default` endpoint that doesn't require custom domain settings
- Add configuration options for Azure service plans
- Improved development workflow:
  - Improved git log parsing
  - Add `mach bootstrap` commands:
      - `mach bootstrap config` for creating a new MACH configuration
      - `mach bootstrap component` for creating a new MACH component
  - Add `--site` option to the `generate`, `plan` and `apply` commands
  - Add `--component` option to the `plan` and `apply` commands
  - Add `--reuse` flag to the `plan` and `apply` commands to supress a `terraform init` call
  - Add support for relative paths to components
  - Add extra component definition settings `artifacts` to facilitate local deployments
- Improved dependencies between components and MACH-managed commercetools configurations
- Add option to override Terraform provider versions
- Upgraded Terraform to `0.14.5`
- Upgraded Terraform commercetools provider to `0.25.3`
- Upgraded Terraform AWS provider to `3.28.0`
- Upgraded Terraform Azure provider to `2.47.0`
- AWS: Set `auto-deploy` on API gateway stage
- Azure: Add new required `frontdoor_id` Terraform variable for components with `endpoint` defined
- Azure: Add `--with-sp-login` option to `mach plan` command
- Azure: Add two extra tags to all resources: `environment` and `site`
- Azure: Remove function app sync bash command: this is now the responsibility of the component


### Breaking changes

**Generic**

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
  When you name the endpoint that replaces `base_url` "main", it will have the least effect on your existing Terraform state.
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
- **config**: The `deploy_role` setting has been renamed to `deploy_role_arn`

**Azure**

- **config**: The `front_door` configuration block has been renamed to `frontdoor`
- **config**: The Azure frontdoor settings `dns_zone` and `ssl_key_*` settings have been removed;<br>
  Certificates are now managed by Frontdoor and dns_zone is auto-detected.
- **state**: The Terraform `azurerm_dns_cname_record` resources have been renamed; they now take the name of the associated endpoint key. For the smoothest transition, rename them in your Terraform state:<br>
  ```bash
  terraform state mv azurerm_dns_cname_record.<project-key> azurerm_dns_cname_record.<endpoint-key>
  ```
- **component**: The `FRONTDOOR_ID` value is removed from the `var.variables` of a component. Replaced with `var.frontdoor_id`
- **component**: `app_service_plan_id` has been replaced with `app_service_plan` containing both an `id` and `name` so the [azurerm_app_service_plan](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/app_service_plan) data source can be used in a component.
  ```terraform
  variable "app_service_plan" {
    type = object({
      id   = string
      name = string
    })
  }
  ```

## 0.5.1 (2020-11-10)
- Removed `aws` block in general_config
- Add `branch` option to component definitions to be able to perform a `mach update` and stay within a certain branch (during development)


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
- Fixed Azure config merge: not all generic settings where merged with site-specific ones
- Only validate short_name length check for Azure implementations
- Setup Frontdoor per 'public api' component regardless of global Frontdoor settings


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
