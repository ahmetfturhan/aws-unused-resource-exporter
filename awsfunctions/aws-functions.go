package awsfunctions

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
)

// An interface to make the API call
type EC2DescribeInstancesAPI interface {
	DescribeInstances(ctx context.Context,
		params *ec2.DescribeInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// A function to get Instances
func GetInstances(c context.Context, api EC2DescribeInstancesAPI, instanceInput *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return api.DescribeInstances(c, instanceInput)
}

// A function to get Security Groups
func GetSecGroups(c context.Context, api ec2.DescribeSecurityGroupsAPIClient, secGroupInput *ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error) {
	return api.DescribeSecurityGroups(c, secGroupInput)
}

// A function to get Load Balancers
func GetALB(c context.Context, api elb.DescribeLoadBalancersAPIClient, loadBalancerInput *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return api.DescribeLoadBalancers(c, loadBalancerInput)
}

// A function to get Target Groups
func GetTargetGroups(c context.Context, api elb.DescribeTargetGroupsAPIClient, targetGroupsInput *elb.DescribeTargetGroupsInput) (*elb.DescribeTargetGroupsOutput, error) {
	return api.DescribeTargetGroups(c, targetGroupsInput)
}
