package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLBCookieStickinessPolicy_basic(t *testing.T) {
	lbName := fmt.Sprintf("tf-test-lb-%s", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLBCookieStickinessPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccLBCookieStickinessPolicyConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(
						"aws_elb.lb",
						"aws_lb_cookie_stickiness_policy.foo",
					),
				),
			},
			resource.TestStep{
				Config: testAccLBCookieStickinessPolicyConfigUpdate(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBCookieStickinessPolicy(
						"aws_elb.lb",
						"aws_lb_cookie_stickiness_policy.foo",
					),
				),
			},
		},
	})
}

func testAccCheckLBCookieStickinessPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_cookie_stickiness_policy" {
			continue
		}

		lbName, _, policyName := resourceAwsLBCookieStickinessPolicyParseId(rs.Primary.ID)
		out, err := conn.DescribeLoadBalancerPolicies(
			&elb.DescribeLoadBalancerPoliciesInput{
				LoadBalancerName: aws.String(lbName),
				PolicyNames:      []*string{aws.String(policyName)},
			})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && (ec2err.Code() == "PolicyNotFound" || ec2err.Code() == "LoadBalancerNotFound") {
				continue
			}
			return err
		}

		if len(out.PolicyDescriptions) > 0 {
			return fmt.Errorf("Policy still exists")
		}
	}

	return nil
}

func testAccCheckLBCookieStickinessPolicy(elbResource string, policyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[elbResource]
		if !ok {
			return fmt.Errorf("Not found: %s", elbResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[policyResource]
		if !ok {
			return fmt.Errorf("Not found: %s", policyResource)
		}

		elbconn := testAccProvider.Meta().(*AWSClient).elbconn
		elbName, _, policyName := resourceAwsLBCookieStickinessPolicyParseId(policy.Primary.ID)
		_, err := elbconn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
			LoadBalancerName: aws.String(elbName),
			PolicyNames:      []*string{aws.String(policyName)},
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccLBCookieStickinessPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elb" "lb" {
	name = "%s"
	availability_zones = ["us-west-2a"]
	listener {
		instance_port = 8000
		instance_protocol = "http"
		lb_port = 80
		lb_protocol = "http"
	}
}

resource "aws_lb_cookie_stickiness_policy" "foo" {
	name = "foo-policy"
	load_balancer = "${aws_elb.lb.id}"
	lb_port = 80
}`, rName)
}

// Sets the cookie_expiration_period to 300s.
func testAccLBCookieStickinessPolicyConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_elb" "lb" {
	name = "%s"
	availability_zones = ["us-west-2a"]
	listener {
		instance_port = 8000
		instance_protocol = "http"
		lb_port = 80
		lb_protocol = "http"
	}
}

resource "aws_lb_cookie_stickiness_policy" "foo" {
	name = "foo-policy"
	load_balancer = "${aws_elb.lb.id}"
	lb_port = 80
	cookie_expiration_period = 300
}`, rName)
}
