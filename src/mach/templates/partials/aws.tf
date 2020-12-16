{% set aws = site.aws %}
provider "aws" {
  region  = "{{ aws.region }}"
  {% if aws.deploy_role_arn %}
  assume_role {
    role_arn = "arn:aws:iam::{{ aws.account_id }}:role/{{ aws.deploy_role_arn }}"
  }
  {% endif %}
}

{% for provider in aws.extra_providers %}
provider "aws" {
  alias   = "{{ provider.name }}"
  region  = "{{ provider.region }}"

  {% if aws.deploy_role_arn %}
  assume_role {
    role_arn = "arn:aws:iam::{{ aws.account_id }}:role/{{ aws.deploy_role_arn }}"
  }
  {% endif %}
}
{% endfor %}

{% if site.used_endpoints %}
  {% for zone in site.aws.route53_zones %}
  data "aws_route53_zone" "{{ zone|slugify }}" {
    name = "{{ zone }}"
  }
  {% endfor %}

  {% for endpoint in site.used_endpoints %}
    {% include 'partials/endpoints/aws_api_gateway.tf' %}
    
  {% endfor %}
{% endif %}