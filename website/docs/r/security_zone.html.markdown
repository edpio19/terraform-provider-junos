---
layout: "junos"
page_title: "Junos: junos_security_zone"
sidebar_current: "docs-junos-resource-security-zone"
description: |-
  Create a security zone (when Junos device supports it)
---

# junos_security_zone

Provides a security zone resource.

## Example Usage

```hcl
# Add a security zone
resource junos_security_zone "demo_zone" {
  name              = "DemoZone"
  inbound_protocols = ["bgp"]
  address_book {
    name    = "DemoAddress"
    network = "192.0.2.0/25"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, Forces new resource)(`String`) The name of security zone.
* `address_book` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified multiple times for each address to declare.
  * `name` - (Required)(`String`) Name of address.
  * `network` - (Required)(`String`) CIDR of address.
  * `description` - (Optional)(`String`) Description of address.
* `address_book_configure_singly` - (Optional)(`Bool`) Disable management of address-book in this resource to be able to manage them with specific resources. Conflict with `address_book_*`.
* `address_book_dns` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified multiple times for each dns-name address to declare.
  * `name` - (Required)(`String`) Name of address.
  * `fqdn` - (Required)(`String`) Fully qualified domain name.
  * `description` - (Optional)(`String`) Description of address.
  * `ipv4_only` - (Optional)(`Bool`) IPv4 dns address.
  * `ipv6_only` - (Optional)(`Bool`) IPv6 dns address.
* `address_book_range` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified multiple times for each range-address to declare.
  * `name` - (Required)(`String`) Name of address.
  * `from` - (Required)(`String`) Lower limit of address range.
  * `to` - (Required)(`String`) Upper limit of address range.
  * `description` - (Optional)(`String`) Description of address.
* `address_book_set` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified multiple times for each address-set to declare.
  * `name` - (Required)(`String`) Name of address-set.
  * `address` - (Required)(`ListOfString`) List of address names.
  * `description` - (Optional)(`String`) Description of address-set.
* `address_book_wildcard` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified multiple times for each wildcard-address to declare.
  * `name` - (Required)(`String`) Name of address.
  * `network` - (Required)(`String`) Numeric IPv4 wildcard address with in the form of a.d.d.r/netmask.
  * `description` - (Optional)(`String`) Description of address.
* `advance_policy_based_routing_profile` - (Optional)(`String`) Enable Advance Policy Based Routing on this zone with a profile.
* `application_tracking` - (Optional)(`Bool`) Enable Application tracking support for this zone.
* `description` - (Optional)(`String`) Text description of zone.
* `inbound_protocols` - (Optional)(`ListOfString`) The inbound protocols allowed. Must be a list of Junos protocols.
* `inbound_services` - (Optional)(`ListOfString`) The inbound services allowed. Must be a list of Junos services.
* `reverse_reroute` - (Optional)(`Bool`) Enable Reverse route lookup when there is change in ingress interface.
* `screen` - (Optional)(`String`) Name of ids option object (screen) applied to the zone.
* `source_identity_log` - (Optional)(`Bool`) Show user and group info in session log for this zone.
* `tcp_rst` - (Optional)(`Bool`) Send RST for NON-SYN packet not matching TCP session.

## Import

Junos security zone can be imported using an id made up of `<name>`, e.g.

```shell
$ terraform import junos_security_zone.demo_zone DemoZone
```
