# Designing your Tenancy model

When using MACH composer, chances are you want to manage multiple use-cases with it, from a single basis. At least, this is what MACH composer is intended for: building platforms that span many use-cases that share a big part of how they are built. 

MACH composer allows you to orchestrate multi-tenant environments and implement use-case specific requirements where needed, without sacrificing the scalability (and perhaps more importantly, its maintainability) for it.

That being said, this isn't a 'wild card' to just add as many project/environments/etc that you wish. We always recommend to apply the principle: use the lowest possible number of environments, as each of them will come with additional complexity, management overhead and cost. 

The less environments, the better - while of course always making sure that 'it fits'. If there are good reasons to setup additional environments, then set them up instead of trying to fit a square into a circle.

!!! info "Multi tenancy"
    When we talk about the Tenancy model, we explain how we implement multi-tenancy within a MACH composer project. Multi tenancy is used to serve many use-cases, from a single platform. Typical use-cases are multi-brand, multi-country and multi-country environments, that require some 'context' at the application level, to separate one environment from the other, to facilitate the organisation working with the system.

    Conways law comes into play here: *Any organisation that designs a system (defined broadly) will produce a design whose structure is a copy of the organisation's communication structure.* ([source](https://en.wikipedia.org/wiki/Conway%27s_law)).

    Usually the tenancy model of a MACH composer platform, ends up being a reflection of how the organisation that uses it is structured.


## MACH systems tenancy

MACH systems (such as commercetools, contentful or Amplience) are usually multi-tenant by default. They often have a 'single platform' that hosts all of their customers. And within customer accounts, they provide structures and controls to design a sane structure that matches your organisational needs.


### commercetools tenancy

*What to look at when deciding on your commercetools tenancy model:*

1. Consider which commercetools `organisations` you will create (usually one per geographic region is sufficient)
2. Consider when to create a new commercetools `project`.
3. Consider how to look at the different geographic regions that commercetools has
4. Consider how commercetools `project` multi-lingual and multi-country support can help when building multi-lingual and country commerce sites?
5. Consider using commercetools' `shipping zones`
6. Consider when to work with inventory and supply `channels`, including connecting these to `stores`
7. Consider what commercetools `stores` can do for you within a `project` context

<br/>
*Relevant commercetools resources:*

- [Projects documentation](https://docs.commercetools.com/api/projects/project)
- [Channels documentation](https://docs.commercetools.com/api/projects/channels)
- [Stores documentation](https://docs.commercetools.com/api/projects/stores)
- [Organizations documentation (Merchant Center)](https://docs.commercetools.com/merchant-center/organizations)


<br/>
*Things to be aware of:*

1. Known limitations when scaling commercetools `projects` across use-cases:
    - Currently, commercetools is implementing `assortments` in their product catalogue, which will enable you to create store-specific assortments (and thus catalogues). Currently this is not supported yet, but expected in Q1 2021
    - The Merchant Center does not always provide the `store` 'context'. This means that managers of store A might be able to see resources of store B.
        - If this is a problem, this might be a reason to create an additional `project` instead of solving it with a `store`
2. When hosting multiple commercetools regions, commercetools might incur extra costs depending on the region (prices may differ).
3. Extra costs might be incurred for extra `projects` that are created in your commercetools organisation. This is usually contract-bound.
    - This includes setting up non-production (dev, test, qa) `projects` in commercetools


### contentful tenancy

!!! TODO
    describe/link to how contentful is structured and how that relates to MACH composer projects.

### Amplience tenancy

!!! TODO
    describe/link to how Amplience is structured and how that relates to MACH composer projects.


## Cloud provider tenancy

Cloud provider tenancy is aimend at partitioning the platform to a level where certain resources are grouped together logically, and securely. Structuring this suffifient, will allow you to scale the platform across many users and (3rd) parties working on it, while limiting the 'attack surface' in case any environment would be compromised.


### AWS tenancy

How to structure an AWS environment?

In short:

1. Follow [AWS account best practices](https://aws.amazon.com/organizations/getting-started/best-practices/).
2. Create **one AWS account per commercetools project** that you create, which hosts all of the AWS resources tied to that commercetools project.
3. Create one dedicated '*service*' or '*shared*' account that hosts shared resources like component artifacts and centralized routing.

!!! tip ""
      See the instructions on [how to setup AWS](../prerequisites/aws.md) to get a good overview of the way a typical MACH-ready AWS environment is set up

### Azure tenancy

How to structure an Azure environment?

In short:

1. Follow [Azure Resource Group best-practices](https://docs.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-setup-guide/organize-resources?tabs=AzureManagementGroupsAndHierarchy).
2. Create a 'shared' Resource Groups that host shared resources (i.e. CDN, front-ends, encryption keys and the components repository)
3. Have all resources per site be created in their own separate Azure Resource Group. MACH composer will do this for you automatically. [^1]

!!! tip ""
      See the instructions on [how to setup Azure](../prerequisites/azure.md) to get a good overview of the way a typical MACH-ready Azure environment is set up
## Building 'context aware' components

!!! TODO
    Describe how to build components (microservices) that are able to work across tenancy contexts; i.e. they should be parameterised with context-settings, as well as able to 'detect' what store is currently active, for example.

[^1]: You can overwrite this behaviour by creating a Resource Group yourself and define them using the [`resource_group`](../syntax/sites.md#azure) option