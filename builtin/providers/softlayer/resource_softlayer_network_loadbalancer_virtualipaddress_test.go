package softlayer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccSoftLayerVirtualIpAddress_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSoftLayerVirtualIpAddressDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckSoftLayerVirtualIpAddressConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"softlayer_network_loadbalancer_virtualipaddress.testacc_vip", "load_balancing_method", "lc"),
					resource.TestCheckResourceAttr(
						"softlayer_network_loadbalancer_virtualipaddress.testacc_vip", "name", "test_load_balancer_vip"),
					resource.TestCheckResourceAttr(
						"softlayer_network_loadbalancer_virtualipaddress.testacc_vip", "source_port", "80"),
					resource.TestCheckResourceAttr(
						"softlayer_network_loadbalancer_virtualipaddress.testacc_vip", "type", "HTTP"),
					resource.TestCheckResourceAttr(
						"softlayer_network_loadbalancer_virtualipaddress.testacc_vip", "virtual_ip_address", "23.246.204.65"),
				),
			},
		},
	})
}

func testAccCheckSoftLayerVirtualIpAddressDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).networkApplicationDeliveryControllerService

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "softlayer_network_loadbalancer_virtualipaddress" {
			continue
		}

		nadcId, _ := strconv.Atoi(rs.Primary.Attributes["nad_controller_id"])
		vipName, _ := rs.Primary.Attributes["name"]

		// Try to find the vip
		result, err := client.GetVirtualIpAddress(nadcId, vipName)

		if err != nil {
			return fmt.Errorf("Error fetching virtual ip")
		}

		if len(result.VirtualIpAddress) != 0 {
			return fmt.Errorf("Virtual ip address still exists")
		}
	}

	return nil
}

var testAccCheckSoftLayerVirtualIpAddressConfig_basic = `
resource "softlayer_network_loadbalancer_virtualipaddress" "testacc_vip" {
    name = "test_load_balancer_vip"
    nad_controller_id = 18171
    load_balancing_method = "lc"
    source_port = 80
    type = "HTTP"
    virtual_ip_address = "23.246.204.65"
}`
