package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"orphan/awsfunctions"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

type MyCollector struct {
	unusedSecGroups     *prometheus.Desc
	unusedTargetGroups  *prometheus.Desc
	unusedLoadBalancers *prometheus.Desc
}

func (c *MyCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.unusedLoadBalancers
	ch <- c.unusedSecGroups
	ch <- c.unusedTargetGroups

}

func (c *MyCollector) Collect(ch chan<- prometheus.Metric) {

	//Parse the AWS Region from environment variables
	AWS_REGION := kingpin.Flag("AWS_REGION", "AWS Region").Envar("AWS_REGION").Default("us-west-1").String()
	kingpin.Parse()

	//Load the aws config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(*AWS_REGION))

	//Check for errors
	if err != nil {
		log.Fatal(err)
	}

	//Create an ec2 client
	client := ec2.NewFromConfig(cfg)
	//Input for describing the security groups, blank means describe all
	secGroupInput := &ec2.DescribeSecurityGroupsInput{}

	//Get security groups
	resultSecGroup, err := awsfunctions.GetSecGroups(context.TODO(), client, secGroupInput)

	//Check for errors
	if err != nil {
		fmt.Printf("\nGot an error retrieving information about your Amazon EC2 Security Groups: %v\n", err)
		return
	}

	// A map for storing the all security groups
	SecGroups := make(map[string]bool)

	fmt.Println("\tSecurity Groups")

	//Loop over the result and store all security groups into map
	for _, r := range resultSecGroup.SecurityGroups {

		//Set the value to false. Later, the security groups in use will be set to true
		SecGroups[*r.GroupId] = false

		//Print security group info
		fmt.Printf("\nGroup Name: %v\nGroup ID: %v\n", *r.GroupName, *r.GroupId)

	}

	// A parameter needed to list all instances, blank means list all instances.
	instanceInput := &ec2.DescribeInstancesInput{}

	//Get the instances
	result, err := awsfunctions.GetInstances(context.TODO(), client, instanceInput)

	//Check for errors
	if err != nil {
		fmt.Printf("\nGot an error retrieving information about your Amazon EC2 instances: %v\n", err)
		return
	}

	fmt.Println("\tInstance Details")

	//Loop over the result
	for _, r := range result.Reservations {

		//Print reservation details
		fmt.Printf("\n%s\nReservation ID: %v\n", strings.Repeat("-", 50), *r.ReservationId)

		//Loop over the instances
		for _, i := range r.Instances {

			//Print Instance details
			fmt.Println("Instance ID: " + *i.InstanceId)

			//Loop over the attached security groups
			fmt.Println("\nLinkedSecGroups:")

			for _, s := range i.SecurityGroups {

				//Mark the used security groups to true
				SecGroups[*s.GroupId] = true

				//Print the details of current security group
				fmt.Printf("\nGroup Name: %v\n", *s.GroupName)
				fmt.Printf("Group ID: %v\n\n", *s.GroupId)
			}
		}

		//Put a divider on the output
		fmt.Println(strings.Repeat("-", 50))
	}

	// An input parameter to get the load balancers, blank means list all load balancers(alb, nlb, glb)
	loadBalancerInput := &elb.DescribeLoadBalancersInput{}

	//Check for errors
	if err != nil {
		fmt.Printf("\nGot an error while loading the config: %v\n", err)
		return
	}

	//Create a elb client since this is a different service like ec2
	elbclient := elb.NewFromConfig(cfg)

	//Make the API call to get the load balancers
	resultAlb, err := awsfunctions.GetALB(context.TODO(), elbclient, loadBalancerInput)

	//Check for errors
	if err != nil {
		fmt.Printf("\nGot an error retrieving information about your Load Balancers: %v\n", err)
		return
	}

	// A map to store all load balancers
	LoadBalancers := make(map[string]bool)

	// A map to get load balancer name by its ARN
	MapARN := make(map[string]string)

	//Iterate through load balancers
	fmt.Println("\n\tLoad Balancers")
	for _, currDescription := range resultAlb.LoadBalancers {

		//Store the current load balancer by adding it to the map
		LoadBalancers[*currDescription.LoadBalancerArn] = false

		//Link the ARN to its name
		MapARN[*currDescription.LoadBalancerArn] = *currDescription.LoadBalancerName

		//Print info of the Load Balancer
		fmt.Printf("\n\nLoad Balancer Name: %v\t ARN: %v\n", *currDescription.LoadBalancerName, *currDescription.LoadBalancerArn)

		//Iterate over the security groups of load balancer
		for _, s := range currDescription.SecurityGroups {

			//set the used security groups to true
			SecGroups[s] = true

			//Print the security group id
			fmt.Printf("Security Group ID::%v\n", s)
		}
	}

	//Iterate over the security group map and print the elements with the value "false"
	fmt.Println("\n\tSecurity Groups Not Linked to an Instance or a Load Balancer")
	for key, value := range SecGroups {

		//Unused security groups
		if !value {
			ch <- prometheus.MustNewConstMetric(
				c.unusedSecGroups,
				prometheus.GaugeValue,
				1.0, key,
			)
			fmt.Println(key)

			//Used security groups
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.unusedSecGroups,
				prometheus.GaugeValue,
				0.0, key,
			)
		}
	}

	// Parameter to list the target groups, blank means list all
	targetGroupsInput := &elb.DescribeTargetGroupsInput{}

	//Get the target groups
	resultTargetGroups, err := awsfunctions.GetTargetGroups(context.TODO(), elbclient, targetGroupsInput)

	//Check for errors
	if err != nil {
		fmt.Printf("\nGot an error retrieving information about your Load Balancers: %v\n", err)
		return
	}

	TargetGroups := make(map[string]bool)
	//Iterate over target groups
	for _, i := range resultTargetGroups.TargetGroups {

		//add to target group list
		TargetGroups[*i.TargetGroupName] = true

		//Iterate over the load balancers of the current target group
		for _, lbarn := range i.LoadBalancerArns {
			//set the used load balancers to true
			LoadBalancers[lbarn] = true
		}

		//Unused target groups
		if len(i.LoadBalancerArns) == 0 {
			TargetGroups[*i.TargetGroupName] = false
		}

	}

	//Print the orphan load balancers by printing the elements with value "false"
	fmt.Printf("\n\tLoad Balancers without a target group attached\n")

	//Export the metrics for load balancers
	for lb, value := range LoadBalancers {

		//unused load balancers
		if !value {
			fmt.Println(MapARN[lb]) //Get the name ıf the load balancer using its ARN
			ch <- prometheus.MustNewConstMetric(
				c.unusedLoadBalancers,
				prometheus.GaugeValue,
				1.0, MapARN[lb],
			)

			//used load balancers
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.unusedLoadBalancers,
				prometheus.GaugeValue,
				0.0, MapARN[lb],
			)
		}
	}

	//Export the metrics for target groups
	for tg, value := range TargetGroups {

		//unused target groups
		if !value {
			ch <- prometheus.MustNewConstMetric(
				c.unusedTargetGroups,
				prometheus.GaugeValue,
				1.0, tg,
			)

			//used target groups
		} else {
			ch <- prometheus.MustNewConstMetric(
				c.unusedTargetGroups,
				prometheus.GaugeValue,
				0.0, tg,
			)
		}
	}

	//Put a divider on the output
	fmt.Printf("\n%s\n", strings.Repeat("-", 50))

}

func NewMyCollector() *MyCollector {
	return &MyCollector{
		unusedSecGroups:     prometheus.NewDesc(prometheus.BuildFQName("aws", "", "security_group_not_used"), "Unused security groups. Value is 1 if the resource is not used", []string{"resource_id"}, nil),
		unusedTargetGroups:  prometheus.NewDesc(prometheus.BuildFQName("aws", "", "load_balancer_not_used"), "Unused LoadBalancers. Value is 1 if the resource is not used", []string{"resource_name"}, nil),
		unusedLoadBalancers: prometheus.NewDesc(prometheus.BuildFQName("aws", "", "target_group_not_used"), "Unused Target Groups. Value is 1 if the resource is not used", []string{"resource_name"}, nil),
	}
}

func main() {

	//Call the recording function
	prometheus.MustRegister(NewMyCollector())

	//Define an endpoint for metrics
	http.Handle("/metrics", promhttp.Handler())

	//Expose on localhost
	log.Println("Listening on", ":9169")
	log.Fatal(http.ListenAndServe(":9169", nil))

}
