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
  {% if site.aws.route53_zone_name %}
  data "aws_route53_zone" "main" {
    name = "{{ site.aws.route53_zone_name }}"
  }
  {% endif %}

  {% for endpoint in site.used_endpoints %}
    {% include 'partials/endpoints/aws_api_gateway.tf' %}
    
  {% endfor %}
{% endif %}