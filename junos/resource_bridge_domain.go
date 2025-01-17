package junos

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type bridgeDomainOptions struct {
	domainTypeBridge bool
	domainID         int
	isolatedVlan     int
	serviceID        int
	vlanID           int
	description      string
	name             string
	routingInstance  string
	routingInterface string
	communityVlans   []string
	vlanIDList       []string
	vxlan            []map[string]interface{}
}

func resourceBridgeDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBridgeDomainCreate,
		ReadContext:   resourceBridgeDomainRead,
		UpdateContext: resourceBridgeDomainUpdate,
		DeleteContext: resourceBridgeDomainDelete,
		Importer: &schema.ResourceImporter{
			State: resourceBridgeDomainImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"routing_instance": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          defaultWord,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
			},
			"community_vlans": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 15),
			},
			"domain_type_bridge": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"isolated_vlan": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"routing_interface": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !checkStringHasPrefixInList(value, []string{"irb.", "vlan."}) {
						errors = append(errors, fmt.Errorf(
							"%q for %q is not start with 'irb.' or 'vlan.'", value, k))
					}

					return
				},
			},
			"service_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"vlan_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
			"vlan_id_list": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"vlan_id"},
			},
			"vxlan": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vni": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 16777214),
						},
						"vni_extend_evpn": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"decapsulate_accept_inner_vlan": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"encapsulate_inner_vlan": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"ingress_node_replication": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"multicast_group": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"ovsdb_managed": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"unreachable_vtep_aging_timer": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(300, 1800),
						},
					},
				},
			},
		},
	}
}

func resourceBridgeDomainCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sess := m.(*Session)
	if sess.junosFakeCreateSetFile != "" {
		if err := setBridgeDomain(d, m, nil); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(d.Get("name").(string) + idSeparator + d.Get("routing_instance").(string))

		return nil
	}
	jnprSess, err := sess.startNewSession()
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(jnprSess)
	if !checkCompatibilityRouter(jnprSess) {
		return diag.FromErr(fmt.Errorf("bridge domain "+
			"not compatible with Junos device %s", jnprSess.SystemInformation.HardwareModel))
	}
	sess.configLock(jnprSess)
	var diagWarns diag.Diagnostics
	if d.Get("routing_instance").(string) != defaultWord {
		instanceExists, err := checkRoutingInstanceExists(d.Get("routing_instance").(string), m, jnprSess)
		if err != nil {
			appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

			return append(diagWarns, diag.FromErr(err)...)
		}
		if !instanceExists {
			appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

			return append(diagWarns,
				diag.FromErr(fmt.Errorf("routing instance %v doesn't exist", d.Get("routing_instance").(string)))...)
		}
	}
	bridgeDomainExists, err := checkBridgeDomainExists(
		d.Get("name").(string), d.Get("routing_instance").(string), m, jnprSess)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	if bridgeDomainExists {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(fmt.Errorf("bridge domain %v already exists in routing_instance %s",
			d.Get("name").(string), d.Get("routing_instance").(string)))...)
	}
	if err := setBridgeDomain(d, m, jnprSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("create resource junos_bridge_domain", jnprSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	bridgeDomainExists, err = checkBridgeDomainExists(
		d.Get("name").(string), d.Get("routing_instance").(string), m, jnprSess)
	if err != nil {
		return append(diagWarns, diag.FromErr(err)...)
	}
	if bridgeDomainExists {
		d.SetId(d.Get("name").(string) + idSeparator + d.Get("routing_instance").(string))
	} else {
		return append(diagWarns,
			diag.FromErr(fmt.Errorf("bridge domain %v not exists in routing_instance %v after commit "+
				"=> check your config", d.Get("name").(string), d.Get("routing_instance").(string)))...)
	}

	return append(diagWarns, resourceBridgeDomainReadWJnprSess(d, m, jnprSess)...)
}

func resourceBridgeDomainRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sess := m.(*Session)
	jnprSess, err := sess.startNewSession()
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(jnprSess)

	return resourceBridgeDomainReadWJnprSess(d, m, jnprSess)
}

func resourceBridgeDomainReadWJnprSess(
	d *schema.ResourceData, m interface{}, jnprSess *NetconfObject) diag.Diagnostics {
	mutex.Lock()
	bridgeDomainOptions, err := readBridgeDomain(d.Get("name").(string), d.Get("routing_instance").(string),
		m, jnprSess)
	mutex.Unlock()
	if err != nil {
		return diag.FromErr(err)
	}
	if bridgeDomainOptions.name == "" {
		d.SetId("")
	} else {
		fillBridgeDomainData(d, bridgeDomainOptions)
	}

	return nil
}

func resourceBridgeDomainUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	sess := m.(*Session)
	jnprSess, err := sess.startNewSession()
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(jnprSess)
	sess.configLock(jnprSess)
	var diagWarns diag.Diagnostics
	if d.HasChange("vxlan") {
		oldVxlan, _ := d.GetChange("vxlan")
		if err := delBridgeDomainOpts(d.Get("name").(string), d.Get("routing_instance").(string), oldVxlan.([]interface{}),
			m, jnprSess); err != nil {
			appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

			return append(diagWarns, diag.FromErr(err)...)
		}
	} else if err := delBridgeDomainOpts(
		d.Get("name").(string), d.Get("routing_instance").(string), d.Get("vxlan").([]interface{}), m, jnprSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	if err := setBridgeDomain(d, m, jnprSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("update resource junos_bridge_domain", jnprSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	d.Partial(false)

	return append(diagWarns, resourceBridgeDomainReadWJnprSess(d, m, jnprSess)...)
}

func resourceBridgeDomainDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sess := m.(*Session)
	jnprSess, err := sess.startNewSession()
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(jnprSess)
	sess.configLock(jnprSess)
	var diagWarns diag.Diagnostics
	if err := delBridgeDomain(d.Get("name").(string), d.Get("routing_instance").(string), d.Get("vxlan").([]interface{}),
		m, jnprSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("delete resource junos_bridge_domain", jnprSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(jnprSess))

		return append(diagWarns, diag.FromErr(err)...)
	}

	return diagWarns
}

func resourceBridgeDomainImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	sess := m.(*Session)
	jnprSess, err := sess.startNewSession()
	if err != nil {
		return nil, err
	}
	defer sess.closeSession(jnprSess)
	result := make([]*schema.ResourceData, 1)
	idSplit := strings.Split(d.Id(), idSeparator)
	if len(idSplit) < 2 {
		return nil, fmt.Errorf("missing element(s) in id with separator %v", idSeparator)
	}
	bridgeDomainExists, err := checkBridgeDomainExists(idSplit[0], idSplit[1], m, jnprSess)
	if err != nil {
		return nil, err
	}
	if !bridgeDomainExists {
		return nil, fmt.Errorf("don't find bridge domain with id '%v' (id must be "+
			"<name>"+idSeparator+"<routing_instance>)", d.Id())
	}
	bridgeDomainOptions, err := readBridgeDomain(idSplit[0], idSplit[1], m, jnprSess)
	if err != nil {
		return nil, err
	}
	fillBridgeDomainData(d, bridgeDomainOptions)

	result[0] = d

	return result, nil
}

func checkBridgeDomainExists(name string, instance string, m interface{},
	jnprSess *NetconfObject) (bool, error) {
	sess := m.(*Session)
	var bridgeDomainConfig string
	var err error
	if instance == defaultWord {
		bridgeDomainConfig, err = sess.command("show configuration"+
			" bridge-domains \""+name+"\" | display set", jnprSess)
		if err != nil {
			return false, err
		}
	} else {
		bridgeDomainConfig, err = sess.command("show configuration routing-instances "+instance+
			" bridge-domains \""+name+"\" | display set", jnprSess)
		if err != nil {
			return false, err
		}
	}

	if bridgeDomainConfig == emptyWord {
		return false, nil
	}

	return true, nil
}

func setBridgeDomain(d *schema.ResourceData, m interface{}, jnprSess *NetconfObject) error {
	sess := m.(*Session)
	configSet := make([]string, 0)

	var setPrefix string
	if d.Get("routing_instance").(string) == defaultWord {
		setPrefix = "set bridge-domains \"" + d.Get("name").(string) + "\" "
	} else {
		setPrefix = "set routing-instances " + d.Get("routing_instance").(string) +
			" bridge-domains \"" + d.Get("name").(string) + "\" "
	}
	for _, v := range d.Get("community_vlans").(*schema.Set).List() {
		configSet = append(configSet, setPrefix+"community-vlans "+v.(string))
	}
	if v := d.Get("description").(string); v != "" {
		configSet = append(configSet, setPrefix+"description \""+v+"\"")
	}
	if v := d.Get("domain_id").(int); v != 0 {
		configSet = append(configSet, setPrefix+"domain-id "+strconv.Itoa(v))
	}
	if d.Get("domain_type_bridge").(bool) {
		configSet = append(configSet, setPrefix+"domain-type bridge")
	}
	if v := d.Get("isolated_vlan").(int); v != 0 {
		configSet = append(configSet, setPrefix+"isolated-vlan "+strconv.Itoa(v))
	}
	if v := d.Get("routing_interface").(string); v != "" {
		configSet = append(configSet, setPrefix+"routing-interface "+v)
	}
	if v := d.Get("service_id").(int); v != 0 {
		configSet = append(configSet, setPrefix+"service-id "+strconv.Itoa(v))
	}
	if v := d.Get("vlan_id").(int); v != 0 {
		configSet = append(configSet, setPrefix+"vlan-id "+strconv.Itoa(v))
	}
	for _, v := range d.Get("vlan_id_list").(*schema.Set).List() {
		configSet = append(configSet, setPrefix+"vlan-id-list "+v.(string))
	}
	for _, v := range d.Get("vxlan").([]interface{}) {
		vxlan := v.(map[string]interface{})
		configSet = append(configSet, setPrefix+"vxlan vni "+strconv.Itoa(vxlan["vni"].(int)))

		if vxlan["vni_extend_evpn"].(bool) {
			if d.Get("routing_instance").(string) == defaultWord {
				configSet = append(configSet, "set protocols evpn extended-vni-list "+strconv.Itoa(vxlan["vni"].(int)))
			} else {
				configSet = append(configSet, "set routing-instances "+d.Get("routing_instance").(string)+
					" protocols evpn extended-vni-list "+strconv.Itoa(vxlan["vni"].(int)))
			}
		}
		if vxlan["decapsulate_accept_inner_vlan"].(bool) {
			configSet = append(configSet, setPrefix+"vxlan decapsulate-accept-inner-vlan")
		}
		if vxlan["encapsulate_inner_vlan"].(bool) {
			configSet = append(configSet, setPrefix+"vxlan encapsulate-inner-vlan")
		}
		if vxlan["ingress_node_replication"].(bool) {
			configSet = append(configSet, setPrefix+"vxlan ingress-node-replication")
		}
		if vxlan["multicast_group"].(string) != "" {
			configSet = append(configSet, setPrefix+"vxlan multicast-group "+vxlan["multicast_group"].(string))
		}
		if vxlan["ovsdb_managed"].(bool) {
			configSet = append(configSet, setPrefix+"vxlan ovsdb-managed")
		}
		if vxlan["unreachable_vtep_aging_timer"].(int) != 0 {
			configSet = append(configSet, setPrefix+
				"vxlan unreachable-vtep-aging-timer "+strconv.Itoa(vxlan["unreachable_vtep_aging_timer"].(int)))
		}
	}

	return sess.configSet(configSet, jnprSess)
}

func readBridgeDomain(name string, instance string, m interface{},
	jnprSess *NetconfObject) (bridgeDomainOptions, error) {
	sess := m.(*Session)
	var confRead bridgeDomainOptions
	var nameConfig string
	var err error

	if instance == defaultWord {
		nameConfig, err = sess.command("show configuration"+
			" bridge-domains \""+name+"\" | display set relative", jnprSess)
	} else {
		nameConfig, err = sess.command("show configuration routing-instances "+instance+
			" bridge-domains \""+name+"\" | display set relative", jnprSess)
	}
	if err != nil {
		return confRead, err
	}

	if nameConfig != emptyWord {
		confRead.name = name
		confRead.routingInstance = instance
		for _, item := range strings.Split(nameConfig, "\n") {
			if strings.Contains(item, "<configuration-output>") {
				continue
			}
			if strings.Contains(item, "</configuration-output>") {
				break
			}
			itemTrim := strings.TrimPrefix(item, setLineStart)
			switch {
			case strings.HasPrefix(itemTrim, "community-vlans "):
				confRead.communityVlans = append(confRead.communityVlans, strings.TrimPrefix(itemTrim, "community-vlans "))
			case strings.HasPrefix(itemTrim, "description "):
				confRead.description = strings.Trim(strings.TrimPrefix(itemTrim, "description "), "\"")
			case strings.HasPrefix(itemTrim, "domain-id "):
				var err error
				confRead.domainID, err = strconv.Atoi(strings.TrimPrefix(itemTrim, "domain-id "))
				if err != nil {
					return confRead, fmt.Errorf("failed to convert value from '%s' to integer : %w", itemTrim, err)
				}
			case itemTrim == "domain-type bridge":
				confRead.domainTypeBridge = true
			case strings.HasPrefix(itemTrim, "isolated-vlan "):
				var err error
				confRead.isolatedVlan, err = strconv.Atoi(strings.TrimPrefix(itemTrim, "isolated-vlan "))
				if err != nil {
					return confRead, fmt.Errorf("failed to convert value from '%s' to integer : %w", itemTrim, err)
				}
			case strings.HasPrefix(itemTrim, "routing-interface "):
				confRead.routingInterface = strings.TrimPrefix(itemTrim, "routing-interface ")
			case strings.HasPrefix(itemTrim, "service-id "):
				var err error
				confRead.serviceID, err = strconv.Atoi(strings.TrimPrefix(itemTrim, "service-id "))
				if err != nil {
					return confRead, fmt.Errorf("failed to convert value from '%s' to integer : %w", itemTrim, err)
				}
			case strings.HasPrefix(itemTrim, "vlan-id "):
				var err error
				confRead.vlanID, err = strconv.Atoi(strings.TrimPrefix(itemTrim, "vlan-id "))
				if err != nil {
					return confRead, fmt.Errorf("failed to convert value from '%s' to integer : %w", itemTrim, err)
				}
			case strings.HasPrefix(itemTrim, "vlan-id-list "):
				confRead.vlanIDList = append(confRead.vlanIDList, strings.TrimPrefix(itemTrim, "vlan-id-list "))
			case strings.HasPrefix(itemTrim, "vxlan "):
				if len(confRead.vxlan) == 0 {
					confRead.vxlan = append(confRead.vxlan, map[string]interface{}{
						"vni":                           -1,
						"vni_extend_evpn":               false,
						"decapsulate_accept_inner_vlan": false,
						"encapsulate_inner_vlan":        false,
						"ingress_node_replication":      false,
						"multicast_group":               "",
						"ovsdb_managed":                 false,
						"unreachable_vtep_aging_timer":  0,
					})
				}
				vxlan := confRead.vxlan[0]
				switch {
				case strings.HasPrefix(itemTrim, "vxlan vni "):
					vxlan["vni"], err = strconv.Atoi(strings.TrimPrefix(itemTrim, "vxlan vni "))
					if err != nil {
						return confRead, fmt.Errorf("failed to convert value from '%s' to integer : %w", itemTrim, err)
					}
					if vxlan["vni"] != -1 {
						var evpnConfig string
						var err error
						if confRead.routingInstance == defaultWord {
							evpnConfig, err = sess.command("show configuration protocols evpn | display set relative", jnprSess)
							if err != nil {
								return confRead, err
							}
						} else {
							evpnConfig, err = sess.command("show configuration routing-instances "+instance+
								" protocols evpn | display set relative", jnprSess)
							if err != nil {
								return confRead, err
							}
						}
						if evpnConfig != emptyWord {
							for _, item := range strings.Split(evpnConfig, "\n") {
								if strings.Contains(item, "<configuration-output>") {
									continue
								}
								if strings.Contains(item, "</configuration-output>") {
									break
								}
								itemTrim := strings.TrimPrefix(item, setLineStart)
								if strings.HasPrefix(itemTrim, "extended-vni-list "+strconv.Itoa(vxlan["vni"].(int))) {
									vxlan["vni_extend_evpn"] = true
								}
							}
						}
					}
				case itemTrim == "vxlan decapsulate-accept-inner-vlan":
					vxlan["decapsulate_accept_inner_vlan"] = true
				case itemTrim == "vxlan encapsulate-inner-vlan":
					vxlan["encapsulate_inner_vlan"] = true
				case itemTrim == "vxlan ingress-node-replication":
					vxlan["ingress_node_replication"] = true
				case strings.HasPrefix(itemTrim, "vxlan multicast-group "):
					vxlan["multicast_group"] = strings.TrimPrefix(itemTrim, "vxlan multicast-group ")
				case itemTrim == "vxlan ovsdb-managed":
					vxlan["ovsdb_managed"] = true
				case strings.HasPrefix(itemTrim, "vxlan unreachable-vtep-aging-timer "):
					vxlan["unreachable_vtep_aging_timer"], err = strconv.Atoi(strings.TrimPrefix(itemTrim,
						"vxlan unreachable-vtep-aging-timer "))
					if err != nil {
						return confRead, fmt.Errorf("failed to convert value from '%s' to integer : %w", itemTrim, err)
					}
				}
			}
		}
	}

	return confRead, nil
}

func delBridgeDomainOpts(
	name string, instance string, vxlan []interface{}, m interface{}, jnprSess *NetconfObject) error {
	sess := m.(*Session)
	configSet := make([]string, 0)
	delPrefix := "delete bridge-domains \"" + name + "\" "
	if instance != defaultWord {
		delPrefix = "delete routing-instances " + instance + " bridge-domains \"" + name + "\" "
	}
	configSet = append(configSet,
		delPrefix+"community-vlans",
		delPrefix+"description",
		delPrefix+"domain-id",
		delPrefix+"domain-type",
		delPrefix+"isolated-vlan",
		delPrefix+"routing-interface",
		delPrefix+"service-id",
		delPrefix+"vlan-id",
		delPrefix+"vlan-id-list",
		delPrefix+"vxlan",
	)
	for _, v := range vxlan {
		vxlanParams := v.(map[string]interface{})
		if vxlanParams["vni_extend_evpn"].(bool) {
			if instance == defaultWord {
				configSet = append(configSet, "delete protocols evpn extended-vni-list "+strconv.Itoa(vxlanParams["vni"].(int)))
			} else {
				configSet = append(configSet, "delete routing-instances "+instance+
					" protocols evpn extended-vni-list "+strconv.Itoa(vxlanParams["vni"].(int)))
			}
		}
	}

	return sess.configSet(configSet, jnprSess)
}

func delBridgeDomain(name string, instance string, vxlan []interface{}, m interface{}, jnprSess *NetconfObject) error {
	sess := m.(*Session)
	configSet := make([]string, 0, 1)
	if instance == defaultWord {
		configSet = append(configSet, "delete bridge-domains \""+name+"\"")
	} else {
		configSet = append(configSet, "delete routing-instances "+instance+" bridge-domains \""+name+"\"")
	}
	for _, v := range vxlan {
		vxlanParams := v.(map[string]interface{})
		if vxlanParams["vni_extend_evpn"].(bool) {
			if instance == defaultWord {
				configSet = append(configSet, "delete protocols evpn extended-vni-list "+strconv.Itoa(vxlanParams["vni"].(int)))
			} else {
				configSet = append(configSet, "delete routing-instances "+instance+
					" protocols evpn extended-vni-list "+strconv.Itoa(vxlanParams["vni"].(int)))
			}
		}
	}

	return sess.configSet(configSet, jnprSess)
}

func fillBridgeDomainData(d *schema.ResourceData, bridgeDomainOptions bridgeDomainOptions) {
	if tfErr := d.Set("name", bridgeDomainOptions.name); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("routing_instance", bridgeDomainOptions.routingInstance); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("community_vlans", bridgeDomainOptions.communityVlans); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", bridgeDomainOptions.description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_id", bridgeDomainOptions.domainID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_type_bridge", bridgeDomainOptions.domainTypeBridge); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("isolated_vlan", bridgeDomainOptions.isolatedVlan); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("routing_interface", bridgeDomainOptions.routingInterface); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("service_id", bridgeDomainOptions.serviceID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("vlan_id", bridgeDomainOptions.vlanID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("vlan_id_list", bridgeDomainOptions.vlanIDList); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("vxlan", bridgeDomainOptions.vxlan); tfErr != nil {
		panic(tfErr)
	}
}
