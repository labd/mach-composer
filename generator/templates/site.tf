# This file is auto-generated by MACH composer
# Site: {{ site.Identifier }}

terraform {
  {% if global.TerraformConfig.AzureRemoteState %}
  {%- set azure_config = global.TerraformConfig.AzureRemoteState -%}
  backend "azurerm" {
    resource_group_name  = {{ azure_config.resource_group|tf }}
    storage_account_name = {{ azure_config.storage_account|tf }}
    container_name       = {{ azure_config.container_name|tf }}
    key                  = "{{ azure_config.state_folder}}/{{ site.identifier }}"
  }
  {% elif global.TerraformConfig.AwsRemoteState %}
  {%- set aws_config = global.TerraformConfig.AwsRemoteState -%}
  backend "s3" {
    bucket         = {{ aws_config.Bucket|tf }}
    key            = "{{ aws_config.KeyPrefix}}/{{ site.Identifier }}"
    region         = {{ aws_config.Region|tf }}
    {% if aws_config.RoleARN -%}
    role_arn       = {{ aws_config.RoleARN|tf }}
    {% endif -%}
    {%- if aws_config.LockTable -%}
    dynamodb_table = {{ aws_config.LockTable|tf }}
    {% endif -%}
    encrypt        = {% if aws_config.Encrypt %}true{% else %}false{% endif %}
  }
  {%- endif %}
}

{% include "partials/providers.tf" %}

{% if config.variables_encrypted %}
data "local_file" "variables" {
  filename = "variables.yml"
}

data "sops_external" "variables" {
  source     = data.local_file.variables.content
  input_type = "yaml"
}
{% endif %}

{%- if global.SentryConfig.AuthToken %}
provider "sentry" {
  token = {{ global.SentryConfig.AuthToken|tf }}
  base_url = {% if global.SentryConfig.BaseURL %}{{ global.SentryConfig.BaseURL|tf }}{% else %}"https://sentry.io/api/"{% endif %}
}
{%- endif %}


{% if site.Commercetools -%}
  {% include "partials/commercetools.tf" with commercetools=site.Commercetools %}
{%- endif %}
{% if site.Contentful %}{% include "partials/contentful.tf" %}{% endif %}
{% if site.Amplience %}{% include "partials/amplience.tf" with amplience=site.Amplience %}{% endif %}
{% if site.AWS %}{% include "partials/aws.tf" with aws=site.AWS %}{% endif %}
{% if site.Azure %}{% include "partials/azure.tf" %}{% endif %}

{% for component in site.Components %}
  {% include "partials/component.tf" %}
{% endfor %}

