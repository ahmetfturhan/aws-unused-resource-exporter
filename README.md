# AWS Unused Resource Exporter

This tool fetches the resource details from AWS and exports any unused Security Group, Load Balancer and Target Group.

### Needed Environmental Variables
AWS_ACCESS_KEY_ID  
AWS_SECRET_ACCESS_KEY  
AWS_REGION  

### API Calls
ec2.DescribeInstances  
ec2.DescribeSecurityGroups  
elasticloadbalancingv2.DescribeLoadBalancers  
elasticloadbalancingv2.DescribeTargetGroups  

### Exported Information

|      name     |     label     |     value     |
| ------------- | ------------- | ------------- |
| aws_security_group_not_used  | resource_id  | ID of the Security Group  |
| aws_load_balancer_not_used  | resource_name  | Name of the Load Balancer  |
| aws_target_group_not_used  | resource_name  | Name of the Target Group  |


### Docker
You can build using,  
```
docker build --tag orphan-finder .
```

You can run using,  
```
docker run -e AWS_ACCESS_KEY_ID={your access key id} -e AWS_SECRET_ACCESS_KEY={your secret access key} -e AWS_REGION={aws region} orphan-finder
```
