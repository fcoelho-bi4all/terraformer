package vpn_connection

import (
	"waze/terraform/aws_terraforming/aws_generator"
	"waze/terraform/terraform_utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var ignoreKey = map[string]bool{
	"tunnel2_vgw_inside_address": true,
	"id":                         true,
	"tunnel2_cgw_inside_address": true,
	"tunnel2_bgp_holdtime":       true,
	"tunnel2_bgp_asn":            true,
	"tunnel2_address":            true,
	"tunnel1_vgw_inside_address": true,
	"tunnel1_cgw_inside_address": true,
	"tunnel1_bgp_holdtime":       true,
	"tunnel1_bgp_asn":            true,
	"tunnel1_address":            true,
}

var allowEmptyValues = map[string]bool{
	"tags.": true,
}

type VpnConnectionGenerator struct {
	aws_generator.BasicGenerator
}

func (VpnConnectionGenerator) createResources(vpncs *ec2.DescribeVpnConnectionsOutput) []terraform_utils.TerraformResource {
	resoures := []terraform_utils.TerraformResource{}
	for _, vpnc := range vpncs.VpnConnections {
		resourceName := ""
		if len(vpnc.Tags) > 0 {
			for _, tag := range vpnc.Tags {
				if aws.StringValue(tag.Key) == "Name" {
					resourceName = aws.StringValue(tag.Value)
					break
				}
			}
		}
		resoures = append(resoures, terraform_utils.TerraformResource{
			ResourceType: "aws_vpn_connection",
			ResourceName: resourceName,
			Item:         nil,
			ID:           aws.StringValue(vpnc.VpnConnectionId),
			Provider:     "aws",
		})
	}
	return resoures
}

func (g VpnConnectionGenerator) Generate(region string) error {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
	svc := ec2.New(sess)
	vpncs, err := svc.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{})
	if err != nil {
		return err
	}
	resources := g.createResources(vpncs)
	err = terraform_utils.GenerateTfState(resources)
	if err != nil {
		return err
	}
	converter := terraform_utils.TfstateConverter{}
	metadata := terraform_utils.NewResourcesMetaData(resources, ignoreKey, allowEmptyValues, map[string]string{})
	resources, err = converter.Convert("terraform.tfstate", metadata)
	if err != nil {
		return err
	}
	err = terraform_utils.GenerateTf(resources, "vpn_connection", region, "aws")
	if err != nil {
		return err
	}
	return nil

}