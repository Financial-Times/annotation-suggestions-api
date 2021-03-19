<!--
    Written in the format prescribed by https://github.com/Financial-Times/runbook.md.
    Any future edits should abide by this format.
-->
# PAC - Draft Content Suggestions API

Retrieves suggestions from UPP for draft CMS content.

## Code

draft-content-suggestions

## Primary URL

https://api.ft.com/drafts/content/%UUID%/suggestions

## Service Tier

Bronze

## Lifecycle Stage

Production

## Host Platform

AWS

## Architecture

The PAC Draft Content Suggestions API reads draft CMS content from the PAC Draft Content Public Read service, retrieves annotation suggestions for it using the UPP Public Suggestions API and returns them to the consumer.

PAC architecture diagram: <https://user-images.githubusercontent.com/3042889/74439601-3aa12180-4e75-11ea-8625-a933bf33ea54.png>

## Contains Personal Data

No

## Contains Sensitive Data

No

<!-- Placeholder - remove HTML comment markers to activate
## Can Download Personal Data
Choose Yes or No

...or delete this placeholder if not applicable to this system
-->

<!-- Placeholder - remove HTML comment markers to activate
## Can Contact Individuals
Choose Yes or No

...or delete this placeholder if not applicable to this system
-->

## Failover Architecture Type

ActiveActive

## Failover Process Type

FullyAutomated

## Failback Process Type

FullyAutomated

## Failover Details

The service is deployed in both PAC clusters. The failover guide is located here:
<https://github.com/Financial-Times/upp-docs/tree/master/failover-guides/pac-cluster>

## Data Recovery Process Type

NotApplicable

## Data Recovery Details

The service does not store data, so it does not require any data recovery steps.

## Release Process Type

PartiallyAutomated

## Rollback Process Type

Manual

## Release Details

The service is a member of the "annotations-curation" health category so a PAC cluster failover is required during release.

<!-- Placeholder - remove HTML comment markers to activate
## Heroku Pipeline Name
Enter descriptive text satisfying the following:
This is the name of the Heroku pipeline for this system. If you don't have a pipeline, this is the name of the app in Heroku. A pipeline is a group of Heroku apps that share the same codebase where each app in a pipeline represents the different stages in a continuous delivery workflow, i.e. staging, production.

...or delete this placeholder if not applicable to this system
-->

## Key Management Process Type

Manual

## Key Management Details

To access the service clients need to provide basic auth credentials.
To rotate credentials you need to login to a particular cluster and update varnish-auth secrets.

## Monitoring

Service in the PAC K8S clusters:

*   PAC-Prod-EU health: <https://pac-prod-eu.upp.ft.com/__health/__pods-health?service-name=draft-content-suggestions>
*   PAC-Prod-US health: <https://pac-prod-us.upp.ft.com/__health/__pods-health?service-name=draft-content-suggestions>

## First Line Troubleshooting

<https://github.com/Financial-Times/upp-docs/tree/master/guides/ops/first-line-troubleshooting>

## Second Line Troubleshooting

Please refer to the GitHub repository README for troubleshooting information.