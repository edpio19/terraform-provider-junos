---
layout: "junos"
page_title: "Junos: junos_interface_physical"
sidebar_current: "docs-junos-resource-interface-physical"
description: |-
  Create/configure a physical interface
---

# junos_interface_physical

Provides a physical interface resource.

## Example Usage

```hcl
# Configure interface of switch
resource junos_interface_physical "interface_switch_demo" {
  name         = "ge-0/0/0"
  description  = "interfaceSwitchDemo"
  trunk        = true
  vlan_members = ["100"]
}
# Configure interface for L3 logical interface on Junos Router or firewall
resource junos_interface_physical "interface_fw_demo" {
  name         = "ge-0/0/1"
  description  = "interfaceFwDemo"
  vlan_tagging = true
}
```

## Argument Reference

~> **NOTE:** This resource computes the maximum number of aggregate interfaces required with the current configuration (searches lines `ether-options 802.3ad` and `ae` interfaces set) then add/remove `chassis aggregated-devices ethernet device-count` line with this maximum.

The following arguments are supported:

* `name` - (Required, Forces new resource)(`String`) Name of physical interface (without dot).
* `no_disable_on_destroy` - (Optional)(`Bool`) When destroy this resource, delete all configurations => do not add `disable` + `descrition NC` or `apply-groups` with `group_interface_delete` provider argument on interface.
* `ae_lacp` - (Optional,**DEPRECATED**)(`String`) Add lacp option in aggregated-ether-options. Need to be 'active' or 'passive' for initiate transmission or respond.  
Deprecated argument, use the `parent_ether_opts` argument instead.
* `ae_link_speed` - (Optional,**DEPRECATED**)(`String`) Link speed of individual interface that joins the AE.  
Deprecated argument, use the `parent_ether_opts` argument instead.
* `ae_minimum_links` - (Optional,**DEPRECATED**)(`Int`) Minimum number of aggregated links.  
Deprecated argument, use the `parent_ether_opts` argument instead.
* `description` - (Optional)(`String`) Description for interface.
* `esi` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified only once to define ESI Config parameters. See the [`esi` arguments](#esi-arguments) block.
* `ether_opts` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified only once to declare 'ether-options' configuration. Conflict with 'gigether_opts'.
  * `ae_8023ad` - (Optional)(`String`) Name of an aggregated Ethernet interface to join.
  * `auto_negotiation` - (Optional)(`Bool`) Enable auto-negotiation.
  * `no_auto_negotiation` - (Optional)(`Bool`) Don't enable auto-negotiation.
  * `flow_control` - (Optional)(`Bool`) Enable flow control.
  * `no_flow_control` - (Optional)(`Bool`) Don't enable flow control.
  * `loopback` - (Optional)(`Bool`) Enable loopback.
  * `no_loopback` - (Optional)(`Bool`) Don't enable loopback.
  * `redundant_parent` - (Optional)(`String`) Name of a redundant ethernet interface to join.
* `ether802_3ad` - (Optional,**DEPRECATED**)(`String`) Name of aggregated device for add this interface to link of 802.3ad interface.  
Deprecated argument, use the `ether_opts` or `gigether_opts` argument instead.
* `gigether_opts` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified only once to declare 'gigether-options' configuration. Conflict with 'ether_opts'.
  * `ae_8023ad` - (Optional)(`String`) Name of an aggregated Ethernet interface to join.
  * `auto_negotiation` - (Optional)(`Bool`) Enable auto-negotiation.
  * `no_auto_negotiation` - (Optional)(`Bool`) Don't enable auto-negotiation.
  * `flow_control` - (Optional)(`Bool`) Enable flow control.
  * `no_flow_control` - (Optional)(`Bool`) Don't enable flow control.
  * `loopback` - (Optional)(`Bool`) Enable loopback.
  * `no_loopback` - (Optional)(`Bool`) Don't enable loopback.
  * `redundant_parent` - (Optional)(`String`) Name of a redundant ethernet interface to join.
* `parent_ether_opts` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified only once to declare 'aggregated-ether-options' or 'redundant-ether-options' configuration (it depends on the interface 'name'). See the [`parent_ether_opts` arguments](#parent_ether_opts-arguments) block.
* `trunk` - (Optional)(`Bool`) Interface mode is trunk.
* `vlan_members` - (Optional)(`ListOfString`) List of vlan for membership for this interface.
* `vlan_native` - (Optional)(`Int`) Vlan for untagged frames.
* `vlan_tagging` - (Optional)(`Bool`) Add 802.1q VLAN tagging support.

---

### esi arguments

* `mode` - (Required)(`String`) ESI Mode
* `identifier` - (Optional)(`String`) The ESI value for the interface
* `auto_derive_lacp` - (Optional)(`Bool`) Auto-derive ESI value for the interface
* `df_election_type` - (Optional)(`String`) DF Election Type
* `source_bmac` - (Optional)(`String`) Unicast Source B-MAC address per ESI for PBB-EVPN

---

### parent_ether_opts arguments

* `bfd_liveness_detection` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified only once to declare 'bfd-liveness-detection' in 'aggregated-ether-options' configuration. See the [`bfd_liveness_detection` arguments in parent_ether_opts](#bfd_liveness_detection-arguments-in-parent_ether_opts) block.
* `flow_control` - (Optional)(`Bool`) Enable flow control.
* `no_flow_control` - (Optional)(`Bool`) Don't enable flow control.
* `lacp` - (Optional)([attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html)) Can be specified only once to declare 'lacp' configuration.
  * `mode` - (Required)(`String`) Active or passive.
  * `admin_key` - (Optional)(`Int`) Node's administrative key.
  * `periodic` - (Optional)(`String`) Timer interval for periodic transmission of LACP packets. Need to be 'fast' or 'slow'.
  * `sync_reset` - (Optional)(`String`) On minimum-link failure notify out of sync to peer. Need to be 'disable' or 'enable'.
  * `system_id` - (Optional)(`String`) Node's System ID, encoded as a MAC address
  * `system_priority` - (Optional)(`Int`) Priority of the system (0 ... 65535).
* `loopback` - (Optional)(`Bool`) Enable loopback.
* `no_loopback` - (Optional)(`Bool`) Don't enable loopback.
* `link_speed` - (Optional)(`String`) Link speed of individual interface that joins the AE.
* `minimum_bandwidth` - (Optional)(`String`) Minimum bandwidth configured for aggregated bundle. Need to be 'N (k|g|m)?bps' format.
* `minimum_links` - (Optional)(`Int`) Minimum number of aggregated/active links (1..64).
* `redundancy_group` - (Optional)(`Int`) Redundancy group of this interface (1..128) for reth interface.
* `source_address_filter` - (Optional)(`ListOfString`) Source address filters.
* `source_filtering` - (Optional)(`Bool`) Enable source address filtering.

---

### bfd_liveness_detection arguments in parent_ether_opts

* `local_address` - (Required)(`String`) BFD local address.
* `authentication_algorithm` - (Optional)(`String`) Authentication algorithm name.
* `authentication_key_chain` - (Optional)(`String`) Authentication Key chain name.
* `authentication_loose_check` - (Optional)(`Bool`) Verify authentication only if authentication is negotiated.
* `detection_time_threshold` - (Optional)(`Int`) High detection-time triggering a trap (milliseconds).
* `holddown_interval` - (Optional)(`Int`) Time to hold the session-UP notification to the client (0..255000 milliseconds).
* `minimum_interval` - (Optional)(`Int`) Minimum transmit and receive interval (1..255000 milliseconds).
* `minimum_receive_interval` - (Optional)(`Int`) Minimum receive interval (1..255000 milliseconds).
* `multiplier` - (Optional)(`Int`) Detection time multiplier (1..255).
* `neighbor` - (Optional)(`String`) BFD neighbor address.
* `no_adaptation` - (Optional)(`Bool`) Disable adaptation.
* `transmit_interval_minimum_interval` - (Optional)(`Int`) Minimum transmit interval (1..255000 milliseconds).
* `transmit_interval_threshold` - (Optional)(`Int`) High transmit interval triggering a trap (milliseconds).
* `version` - (Optional)(`String`) BFD protocol version number.

## Import

Junos interface can be imported using an id made up of `<name>`, e.g.

```shell
$ terraform import junos_interface_physical.interface_switch_demo ge-0/0/0
$ terraform import junos_interface_physical.interface_fw_demo_100 ge-0/0/1
```
